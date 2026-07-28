<template>
  <div class="pixel-tavern">
    <div class="tavern-background">
      <div class="floor-planks"></div>
      
      <div class="title-banner">
        ⚔️ PLANNING POKER TAVERN ⚔️
      </div>

      <!-- Tavern Scene -->
      <div class="table-area">
        <!-- Tables and Stools -->
        <div class="big-table-container">
          <!-- Mass Shadow -->
          <div class="table-shadow"></div>
          
          <!-- The Wooden Table -->
          <div class="tavern-table">
            <!-- Wood grain rings -->
            <div class="table-grain-outer"></div>
            <div class="table-grain-inner"></div>
            
            <!-- Table Decor (Candle/Glow) -->
            <div class="table-center-decor">
              <div class="candle-glow"></div>
              <div class="candle-stick"></div>
            </div>

            <!-- Predefined Stools -->
            <div 
              v-for="(seat, i) in FIXED_SEATS" 
              :key="i"
              class="stool"
              :style="{ 
                left: `${((seat.x - 50) * 4) + 160}px`, 
                top: `${((seat.y - 55) * 4) + 160}px` 
              }"
            >
              <div class="stool-rim"></div>
            </div>
          </div>
        </div>

        <!-- Players seated at seats -->
        <div 
          v-for="(player, index) in allPlayers" 
          :key="player.id"
          class="player-seat"
          :style="getSeatStyle(index)"
        >
          <!-- Aura for current player -->
          <div v-if="player.id === gameStore.currentPlayer?.id" class="me-aura"></div>

          <div class="player-token" :class="[player.character, { 'is-me': player.id === gameStore.currentPlayer?.id }]">
            <div class="token-rim"></div>
            <div class="token-center">
              <span class="char-emoji animate-float">{{ getCharacterEmoji(player.character) }}</span>
            </div>
            
            <transition name="pop-in">
              <div v-if="player.hasVoted" class="status-badge voted">✅</div>
              <div v-else class="status-badge waiting">⏳</div>
            </transition>
          </div>
          
          <div class="player-label" :class="{ 'is-me': player.id === gameStore.currentPlayer?.id }">
            {{ player.id === gameStore.currentPlayer?.id ? 'VOUS' : player.name }}
          </div>

          <!-- Vote Cards stack -->
          <div class="player-vote-cards">
            <transition-group name="card-deal">
              <div 
                v-for="voteIndex in getVoteCount(player.id)" 
                :key="voteIndex"
                class="vote-card hidden"
                :style="{ 
                  transform: `translate(${voteIndex * 2}px, -${voteIndex * 2}px) rotate(${voteIndex * 2}deg)`,
                  zIndex: voteIndex
                }"
              >
                <div class="card-back"></div>
              </div>
            </transition-group>
          </div>
        </div>
      </div>

      <!-- Reveal Carousel Overlay -->
      <transition name="fade">
        <div v-if="isRevealed" class="reveal-overlay">
          <div class="carousel-container" @click.stop>
            <div class="carousel-header">
              <h2 class="reveal-title">🎉 Tavern Results</h2>
              <button class="close-overlay" @click="$emit('dismiss-reveal')">✕</button>
            </div>

            <div class="carousel-main">
              <button 
                class="carousel-nav prev" 
                @click="prevSlide"
                :disabled="currentSlide === 0"
              >◀</button>

              <div class="carousel-slide">
                <div class="ticket-info">
                  <span class="ticket-key">{{ currentStory?.jiraKey }}</span>
                  <h3 class="ticket-title">{{ currentStory?.title }}</h3>
                </div>

                <div class="results-table-container">
                  <table class="results-table">
                    <thead>
                      <tr>
                        <th>Participant</th>
                        <th>Class</th>
                        <th>Cost</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="player in players" :key="player.id">
                        <td class="player-cell">
                          <span class="char-emoji-small">{{ getCharacterEmoji(player.character) }}</span>
                          {{ player.name }}
                        </td>
                        <td class="class-cell">
                          <span class="class-tag" :class="player.character">{{ player.character }}</span>
                        </td>
                        <td class="vote-cell">
                          <div class="mini-card" :class="{ 'has-vote': getVoteForStory(player.id, currentStory?.id) }">
                            {{ getVoteForStory(player.id, currentStory?.id) || '-' }}
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>

                <div class="story-average-bar" v-if="getAverageForStory(currentStory?.id)">
                  <span class="avg-label">Average Cost:</span>
                  <span class="avg-value">{{ getAverageForStory(currentStory?.id) }}</span>
                </div>
              </div>

              <button 
                class="carousel-nav next" 
                @click="nextSlide"
                :disabled="currentSlide === (currentSession?.stories.length || 0) - 1"
              >▶</button>
            </div>

            <div class="carousel-dots">
              <span 
                v-for="(_, i) in currentSession?.stories" 
                :key="i"
                class="dot"
                :class="{ active: i === currentSlide }"
                @click="currentSlide = Number(i)"
              ></span>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useGameStore, GamePhase } from '@/stores/gameStore';

const props = defineProps<{
  players: any[];
  currentSession: any;
  gamePhase: GamePhase;
}>();

const emit = defineEmits(['dismiss-reveal']);

const gameStore = useGameStore();
const currentSlide = ref(0);

const isRevealed = computed(() => props.gamePhase === GamePhase.REVEALED);

// Sync currentSlide with gameStore index when opening
watch(isRevealed, (val) => {
  if (val) currentSlide.value = gameStore.localStoryIndex;
});

// Fixed seat positions from user snippet
const FIXED_SEATS = [
  { id: 'seat-1', x: 50, y: 37 }, // Haut
  { id: 'seat-2', x: 65, y: 47 }, // Haut Droite
  { id: 'seat-3', x: 62, y: 67 }, // Bas Droite
  { id: 'seat-4', x: 38, y: 67 }, // Bas Gauche
  { id: 'seat-5', x: 35, y: 47 }, // Haut Gauche
];

const allPlayers = computed(() => {
  // Stability: Sort by ID
  const players = [...props.players].sort((a, b) => a.id.localeCompare(b.id));
  return players;
});

const currentStory = computed(() => {
  if (!props.currentSession?.stories) return null;
  return props.currentSession.stories[currentSlide.value];
});

function getCharacterEmoji(characterId: string) {
  const chars: Record<string, string> = {
    mage: '🧙', paladin: '⚔️', rogue: '🗡️', priest: '✨',
    warrior: '🛡️', hunter: '🏹', warlock: '😈', druid: '🌿'
  };
  return chars[characterId] || '👤';
}

function getSeatStyle(index: number) {
  // Use fixed seats, loop if more than 5 players
  const seat = FIXED_SEATS[index % FIXED_SEATS.length];
  
  return {
    left: `${seat.x}%`,
    top: `${seat.y}%`,
    transform: 'translate(-50%, -50%)',
    zIndex: Math.floor(seat.y * 10)
  };
}

function getVoteCount(playerId: string) {
  if (!props.currentSession) return 0;
  let count = 0;
  const persistentVotes = props.currentSession.persistentVotes || {};
  for (const storyVotes of Object.values(persistentVotes)) {
    if ((storyVotes as any)[playerId]) count++;
  }
  return count;
}

function getVoteForStory(playerId: string, storyId: string | undefined) {
  if (!storyId || !props.currentSession?.persistentVotes) return null;
  return props.currentSession.persistentVotes[storyId]?.[playerId] || null;
}

function getAverageForStory(storyId: string | undefined) {
  if (!storyId || !props.currentSession?.persistentVotes) return null;
  const votes = props.currentSession.persistentVotes[storyId] || {};
  const numericVotes = Object.values(votes)
    .filter(v => v && !isNaN(Number(v)))
    .map(Number);
  
  if (numericVotes.length === 0) return null;
  const sum = numericVotes.reduce((a, b) => a + b, 0);
  return Math.round((sum / numericVotes.length) * 10) / 10;
}

function nextSlide() {
  if (currentSlide.value < (props.currentSession?.stories.length || 0) - 1) {
    currentSlide.value++;
  }
}

function prevSlide() {
  if (currentSlide.value > 0) {
    currentSlide.value--;
  }
}
</script>

<style scoped>
.pixel-tavern {
  width: 100%;
  height: 100%;
  min-height: 500px;
  position: relative;
  overflow: hidden;
  background: #0a0705;
  user-select: none;
  border-radius: 12px;
}

.tavern-background {
  width: 100%;
  height: 100%;
  background: #110906;
  position: relative;
}

.floor-planks {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background-image: repeating-linear-gradient(0deg, #3d2a1a 0px, #3d2a1a 40px, #000 41px, #000 42px);
  background-size: 100% 84px;
  opacity: 0.1;
}

.title-banner {
  position: absolute;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  color: #f1c40f;
  font-size: 1.2rem;
  font-weight: 900;
  text-shadow: 2px 2px 0 #000;
  white-space: nowrap;
  z-index: 10;
  letter-spacing: 4px;
  text-transform: uppercase;
}

.table-area {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
}

.big-table-container {
  position: absolute;
  top: 55%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.table-shadow {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -30%);
  width: 400px;
  height: 200px;
  background: rgba(0,0,0,0.6);
  filter: blur(60px);
  border-radius: 50%;
}

.tavern-table {
  position: relative;
  width: 320px;
  height: 320px;
  background: #3d2b1f;
  border: 12px solid #251810;
  border-radius: 50%;
  box-shadow: 0 20px 50px rgba(0,0,0,0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 20;
}

.table-grain-outer {
  position: absolute;
  top: 16px; left: 16px; right: 16px; bottom: 16px;
  border: 2px solid rgba(212,167,86,0.1);
  border-radius: 50%;
}

.table-grain-inner {
  position: absolute;
  top: 48px; left: 48px; right: 48px; bottom: 48px;
  border: 1px solid rgba(212,167,86,0.1);
  border-radius: 50%;
}

.table-center-decor {
  position: relative;
  z-index: 10;
}

.candle-glow {
  width: 32px;
  height: 32px;
  background: rgba(255,165,0,0.2);
  border-radius: 50%;
  filter: blur(12px);
  animation: flicker 2s infinite alternate;
}

@keyframes flicker {
  from { opacity: 0.5; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1.1); }
}

.candle-stick {
  width: 8px;
  height: 24px;
  background: rgba(255,140,0,0.4);
  border-radius: 4px;
  margin-top: -12px;
}

.stool {
  position: absolute;
  width: 56px;
  height: 56px;
  background: #251810;
  border: 4px solid #1a110a;
  border-radius: 50%;
  box-shadow: 0 10px 20px rgba(0,0,0,0.5);
  transform: translate(-50%, -50%);
}

.stool-rim {
  position: absolute;
  top: 8px; left: 8px; right: 8px; bottom: 8px;
  border: 1px solid rgba(212,167,86,0.2);
  border-radius: 50%;
}

.player-seat {
  position: absolute;
  display: flex;
  flex-direction: column;
  align-items: center;
  z-index: 40;
  transition: all 0.8s cubic-bezier(0.4, 0, 0.2, 1);
}

.me-aura {
  position: absolute;
  top: 0; left: 0;
  transform: translate(-10%, -10%);
  width: 120%;
  height: 120%;
  background: rgba(243, 156, 18, 0.15);
  filter: blur(25px);
  border-radius: 50%;
  animation: pulse-aura 3s infinite;
}

@keyframes pulse-aura {
  0%, 100% { opacity: 0.5; transform: translate(-10%, -10%) scale(1); }
  50% { opacity: 1; transform: translate(-10%, -10%) scale(1.1); }
}

.player-token {
  width: 80px;
  height: 80px;
  position: relative;
  background: #1a110a;
  border-radius: 50%;
  border: 4px solid #3d2b1f;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.5s ease;
  box-shadow: 0 10px 25px rgba(0,0,0,0.7);
}

.player-token.is-me {
  border-color: #f1c40f;
  transform: scale(1.1);
  box-shadow: 0 0 30px rgba(241, 196, 15, 0.4);
}

.token-center {
  font-size: 2.5rem;
  z-index: 10;
}

.char-emoji.animate-float {
  display: block;
  animation: float 5s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-6px); }
}

.player-label {
  margin-top: 15px;
  padding: 4px 12px;
  border-radius: 20px;
  background: rgba(0,0,0,0.8);
  border: 1px solid rgba(217,119,6,0.5);
  color: rgba(255,255,255,0.6);
  font-size: 10px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 2px;
  white-space: nowrap;
  backdrop-filter: blur(4px);
}

.player-label.is-me {
  background: #f1c40f;
  color: #000;
  border-color: #fff;
  opacity: 1;
}

.status-badge {
  position: absolute;
  top: -5px;
  right: -5px;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
  box-shadow: 0 4px 8px rgba(0,0,0,0.5);
  z-index: 50;
  border: 2px solid #fff;
}

.status-badge.voted { background: #2ecc71; }
.status-badge.waiting { background: #f39c12; }

.player-vote-cards {
  position: absolute;
  top: 80px;
  height: 60px;
}

.vote-card {
  width: 35px;
  height: 50px;
  border-radius: 5px;
  border: 1.5px solid #d4a756;
  background: #5d4037;
  position: absolute;
  box-shadow: 2px 2px 5px rgba(0,0,0,0.6);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-back {
  width: 80%;
  height: 80%;
  border: 1px solid rgba(212, 167, 86, 0.4);
  background: repeating-linear-gradient(45deg, transparent, transparent 5px, rgba(212, 167, 86, 0.1) 5px, rgba(212, 167, 86, 0.1) 10px);
}

/* Reveal Carousel Overlay Styles */
.reveal-overlay {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.85);
  backdrop-filter: blur(8px);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
}

.carousel-container {
  background: #2c1a12;
  border: 4px solid #8b6914;
  border-radius: 20px;
  width: 95%;
  max-width: 1200px;
  height: 90%;
  display: flex;
  flex-direction: column;
  box-shadow: 0 0 80px rgba(0,0,0,1);
  overflow: hidden;
}

.carousel-header {
  padding: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #3e2723;
  border-bottom: 2px solid #8b6914;
}

.reveal-title {
  color: #f1c40f;
  margin: 0;
  font-size: 1.8rem;
  text-shadow: 2px 2px 0 #000;
}

.close-overlay {
  background: rgba(255,255,255,0.1);
  border: none;
  color: #fff;
  font-size: 1.5rem;
  cursor: pointer;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  transition: all 0.2s;
}

.close-overlay:hover { background: rgba(231, 76, 60, 0.5); }

.carousel-main {
  flex: 1;
  display: flex;
  align-items: center;
  padding: 20px;
  gap: 20px;
}

.carousel-nav {
  background: #8b6914;
  border: none;
  color: white;
  width: 50px;
  height: 50px;
  border-radius: 50%;
  cursor: pointer;
  font-size: 1.5rem;
  transition: all 0.2s;
  flex-shrink: 0;
}

.carousel-nav:disabled { opacity: 0.2; cursor: not-allowed; }
.carousel-nav:hover:not(:disabled) { transform: scale(1.1); background: #f1c40f; }

.carousel-slide {
  flex: 1;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 20px;
  overflow: hidden;
}

.ticket-info {
  background: rgba(0,0,0,0.3);
  padding: 15px;
  border-radius: 10px;
  border-left: 5px solid #f1c40f;
}

.ticket-key { color: #f1c40f; font-family: monospace; font-size: 1.1rem; }
.ticket-title { color: #fff; margin: 5px 0 0 0; font-size: 1.2rem; }

.results-table-container {
  flex: 1;
  background: rgba(0,0,0,0.2);
  border-radius: 10px;
  overflow-y: auto;
  border: 1px solid rgba(255,255,255,0.05);
}

.results-table {
  width: 100%;
  border-collapse: collapse;
}

.results-table th {
  background: rgba(139, 105, 20, 0.3);
  padding: 12px;
  text-align: left;
  color: #f1c40f;
  text-transform: uppercase;
  font-size: 0.8rem;
  position: sticky;
  top: 0;
}

.results-table td {
  padding: 12px;
  border-bottom: 1px solid rgba(255,255,255,0.05);
  color: #fff;
}

.player-cell { display: flex; align-items: center; gap: 10px; font-weight: bold; }
.char-emoji-small { font-size: 1.4rem; }

.class-tag {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 0.7rem;
  text-transform: uppercase;
  font-weight: bold;
}

.vote-cell { width: 80px; }

.mini-card {
  width: 40px;
  height: 55px;
  background: #34495e;
  border-radius: 5px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  color: #fff;
  border: 2px solid rgba(255,255,255,0.1);
}

.mini-card.has-vote {
  background: #fff;
  color: #333;
  border-color: #f1c40f;
  box-shadow: 0 0 10px rgba(241, 196, 15, 0.3);
}

.story-average-bar {
  padding: 15px;
  background: linear-gradient(to right, #8b6914, #3e2723);
  border-radius: 10px;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 15px;
}

.avg-label { color: #fff; font-weight: bold; }
.avg-value { font-size: 1.8rem; color: #f1c40f; font-weight: bold; text-shadow: 1px 1px 0 #000; }

.carousel-dots {
  padding: 15px;
  display: flex;
  justify-content: center;
  gap: 10px;
}

.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: rgba(255,255,255,0.2);
  cursor: pointer;
  transition: all 0.2s;
}

.dot.active { background: #f1c40f; transform: scale(1.3); }

/* Animations */
.fade-enter-active, .fade-leave-active { transition: opacity 0.4s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

.pop-in-enter-active { animation: pop-in 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275); }
@keyframes pop-in {
  0% { transform: scale(0); }
  100% { transform: scale(1); }
}

.card-deal-enter-active { animation: card-deal 0.5s ease-out; }
@keyframes card-deal {
  0% { transform: translate(0, -100px) scale(0) rotate(-45deg); opacity: 0; }
  100% { opacity: 1; }
}

/* Modern tavern skin — scene structure and player placement stay unchanged */
.pixel-tavern {
  min-height: 32rem;
  border-radius: 1.3rem;
  background: #0d0805;
}

.tavern-background {
  background:
    radial-gradient(circle at 50% 48%, rgba(194, 103, 39, 0.16), transparent 35%),
    linear-gradient(145deg, rgba(50, 28, 17, 0.72), rgba(10, 7, 5, 0.94));
}

.floor-planks {
  background-image:
    repeating-linear-gradient(0deg, transparent 0 3.4rem, rgba(226, 171, 99, 0.035) 3.4rem 3.48rem),
    linear-gradient(90deg, transparent, rgba(230, 153, 72, 0.05), transparent);
  background-size: 100% auto;
  opacity: 1;
}

.title-banner {
  top: 1.45rem;
  padding: 0.55rem 1.1rem;
  border: 1px solid rgba(232, 177, 103, 0.2);
  border-radius: 999rem;
  background: rgba(17, 10, 6, 0.52);
  color: rgba(255, 238, 211, 0.68);
  font-family: "Cinzel", serif;
  font-size: 0.72rem;
  font-weight: 600;
  letter-spacing: 0.14em;
  text-shadow: none;
  backdrop-filter: blur(0.65rem);
}

.table-shadow {
  background: rgba(0, 0, 0, 0.72);
  filter: blur(4rem);
}

.tavern-table {
  border: 0.75rem solid #24150d;
  background:
    radial-gradient(circle at 50% 42%, rgba(224, 153, 79, 0.18), transparent 28%),
    repeating-radial-gradient(ellipse at center, transparent 0 1.5rem, rgba(238, 182, 110, 0.035) 1.55rem 1.62rem),
    linear-gradient(135deg, #654027, #3b2417 58%, #2b190f);
  box-shadow:
    0 2rem 4rem rgba(0, 0, 0, 0.68),
    inset 0 0 0 1px rgba(242, 194, 130, 0.14),
    inset 0 -1.5rem 2.5rem rgba(0, 0, 0, 0.32);
}

.table-grain-outer,
.table-grain-inner {
  border-color: rgba(238, 183, 111, 0.12);
}

.candle-glow {
  background: rgba(255, 159, 67, 0.34);
  filter: blur(1rem);
}

.stool {
  border-color: #1b100a;
  background: linear-gradient(145deg, #3b2417, #21130c);
  box-shadow: 0 0.75rem 1.5rem rgba(0, 0, 0, 0.48);
}

.player-token {
  border: 0.22rem solid rgba(193, 125, 66, 0.48);
  background: linear-gradient(145deg, rgba(47, 29, 18, 0.96), rgba(15, 10, 7, 0.98));
  box-shadow:
    0 0.75rem 1.75rem rgba(0, 0, 0, 0.58),
    inset 0 1px rgba(255, 255, 255, 0.04);
}

.player-token.is-me {
  border-color: #e3aa63;
  box-shadow: 0 0 2rem rgba(220, 142, 65, 0.34);
}

.player-label {
  margin-top: 0.8rem;
  padding: 0.3rem 0.75rem;
  border-color: rgba(217, 148, 77, 0.3);
  background: rgba(13, 8, 5, 0.8);
  color: rgba(255, 246, 231, 0.7);
  font-family: "Lato", sans-serif;
  font-size: 0.62rem;
  letter-spacing: 0.11em;
}

.player-label.is-me {
  border-color: #e3aa63;
  background: linear-gradient(135deg, #d78c43, #a95528);
  color: #1b0e07;
}

.status-badge {
  border: 1px solid rgba(255, 248, 234, 0.7);
  background: rgba(28, 17, 10, 0.9);
  font-size: 1rem;
}

.status-badge.voted {
  background: #3d7850;
}

.status-badge.waiting {
  background: #9b642e;
}
</style>
