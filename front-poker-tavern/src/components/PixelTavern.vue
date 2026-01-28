<template>
  <div class="pixel-tavern">
    <div class="tavern-background">
      <div class="floor-planks"></div>
      <div class="wall-torches">
        <div class="torch left">🔥</div>
        <div class="torch right">🔥</div>
      </div>
      
      <div class="title-banner">
        ⚔️ PLANNING POKER TAVERN ⚔️
      </div>

      <!-- Table and Players Container -->
      <div class="table-container">
        <div class="round-table">
          <div class="table-surface"></div>
        </div>

        <!-- Players seated around the table -->
        <div 
          v-for="(player, index) in allPlayers" 
          :key="player.id"
          class="player-seat"
          :style="getSeatStyle(index, allPlayers.length)"
        >
          <div class="player-token" :class="player.character">
            <div class="token-rim"></div>
            <div class="token-center">
              <span class="char-emoji">{{ getCharacterEmoji(player.character) }}</span>
            </div>
            
            <transition name="pop-in">
              <div v-if="player.hasVoted" class="status-badge voted">✅</div>
              <div v-else class="status-badge waiting">⏳</div>
            </transition>
          </div>
          <div class="player-name" :class="{ 'is-me': player.id === gameStore.currentPlayer?.id }">
            {{ player.name }}
            <span v-if="player.id === gameStore.currentPlayer?.id" class="me-tag">(You)</span>
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

const allPlayers = computed(() => {
  // Sort players to have a stable seating arrangement
  return [...props.players].sort((a, b) => a.id.localeCompare(b.id));
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

function getSeatStyle(index: number, total: number) {
  const angle = (index / total) * 360 + 90; // Start from bottom
    
  const radiusX = 35; 
  const radiusY = 22; 
  
  const rad = (angle * Math.PI) / 180;
  const left = 50 + Math.cos(rad) * radiusX;
  const top = 55 + Math.sin(rad) * radiusY;

  return {
    left: `${left}%`,
    top: `${top}%`,
    zIndex: Math.floor(top * 10)
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
  background: #1a0f0a;
  user-select: none;
  border-radius: 12px;
}

.tavern-background {
  width: 100%;
  height: 100%;
  background: radial-gradient(circle at 50% 50%, #3d2a1a 0%, #1a0f0a 100%);
  position: relative;
}

.floor-planks {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background-image: linear-gradient(#2d1f15 38px, #1a0f0a 38px, #1a0f0a 40px);
  background-size: 100% 40px;
  opacity: 0.3;
}

.title-banner {
  position: absolute;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  color: #d4a756;
  font-size: 1.2rem;
  font-weight: bold;
  text-shadow: 2px 2px 0 #000;
  white-space: nowrap;
  z-index: 10;
  letter-spacing: 2px;
}

.wall-torches {
  position: absolute;
  top: 20%;
  width: 100%;
  display: flex;
  justify-content: space-between;
  padding: 0 60px;
}

.torch {
  font-size: 1.8rem;
  animation: flicker 0.15s infinite alternate;
}

@keyframes flicker {
  from { transform: scale(1) translateY(0); filter: drop-shadow(0 0 5px orange); }
  to { transform: scale(1.1) translateY(-2px); filter: drop-shadow(0 0 20px red); }
}

.table-container {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.round-table {
  width: 380px;
  height: 240px;
  background: #2c1a12;
  border-radius: 50%;
  border: 8px solid #8b6914;
  box-shadow: 
    0 12px 0 #1a0f0a,
    inset 0 0 50px rgba(0,0,0,0.8);
  position: relative;
  transition: all 0.3s ease;
}

.table-surface {
  position: absolute;
  top: 10%; left: 10%; right: 10%; bottom: 10%;
  background: #3e2723;
  border-radius: 50%;
  border: 3px solid #4e342e;
}

.player-seat {
  position: absolute;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  width: 120px;
}

.player-token {
  width: 70px;
  height: 70px;
  position: relative;
  transition: transform 0.2s ease;
}

.player-token:hover {
  transform: scale(1.1);
}

.token-rim {
  width: 100%;
  height: 100%;
  background: silver;
  border-radius: 50%;
  border: 4px solid #fff;
  box-shadow: 0 6px 0 rgba(0,0,0,0.6);
}

.token-center {
  position: absolute;
  top: 8px; left: 8px; right: 8px; bottom: 8px;
  background: #111;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
}

/* Enhanced Class Colors */
.mage .token-rim { background: linear-gradient(135deg, #4a90e2, #1a252f); border-color: #74b9ff; }
.paladin .token-rim { background: linear-gradient(135deg, #f1c40f, #8b6914); border-color: #ffeaa7; }
.rogue .token-rim { background: linear-gradient(135deg, #27ae60, #1a2f1a); border-color: #55efc4; }
.priest .token-rim { background: linear-gradient(135deg, #ecf0f1, #bdc3c7); border-color: #ffffff; }
.warrior .token-rim { background: linear-gradient(135deg, #e67e22, #5a2e17); border-color: #fab1a0; }
.hunter .token-rim { background: linear-gradient(135deg, #16a085, #0a3d31); border-color: #81ecec; }
.warlock .token-rim { background: linear-gradient(135deg, #8e44ad, #2d1a3d); border-color: #a29bfe; }
.druid .token-rim { background: linear-gradient(135deg, #795548, #2d1b15); border-color: #d7ccc8; }

.player-name {
  margin-top: 12px;
  color: #fff;
  font-size: 0.9rem;
  font-weight: bold;
  text-shadow: 1px 1px 2px #000;
  background: rgba(0,0,0,0.7);
  padding: 2px 10px;
  border-radius: 10px;
  border: 1px solid rgba(255,255,255,0.1);
  white-space: nowrap;
}

.player-name.is-me {
  color: #f1c40f;
  border-color: #f1c40f;
}

.me-tag {
  font-size: 0.7rem;
  opacity: 0.8;
  margin-left: 2px;
}

.status-badge {
  position: absolute;
  top: -10px;
  right: -10px;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  box-shadow: 0 3px 6px rgba(0,0,0,0.4);
  z-index: 20;
}

.status-badge.voted { background: #2ecc71; border: 2px solid #fff; }
.status-badge.waiting { background: #f39c12; border: 2px solid #fff; }

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
  background-image: repeating-linear-gradient(45deg, transparent, transparent 5px, rgba(212, 167, 86, 0.1) 5px, rgba(212, 167, 86, 0.1) 10px);
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
</style>
