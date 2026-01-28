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
          v-for="(player, index) in otherPlayers" 
          :key="player.id"
          class="player-seat"
          :style="getSeatStyle(index, otherPlayers.length)"
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
          <div class="player-name">{{ player.name }}</div>

          <!-- Vote Cards -->
          <div class="player-vote-cards">
            <transition-group name="card-deal">
              <div 
                v-for="voteIndex in getVoteCount(player.id)" 
                :key="voteIndex"
                class="vote-card hidden"
                :style="{ transform: `translate(${voteIndex * 4}px, -${voteIndex * 2}px)` }"
              >
                <div class="card-back"></div>
              </div>
            </transition-group>
          </div>
        </div>
      </div>

      <!-- Reveal Overlay (Simplified for context) -->
      <transition name="fade">
        <div v-if="isRevealed" class="reveal-overlay" @click="$emit('dismiss-reveal')">
          <div class="reveal-content">
            <h2 class="reveal-title">🎉 Votes Revealed!</h2>
            <div class="results-grid">
              <div v-for="player in players" :key="player.id" class="result-row">
                <span class="result-name">{{ player.name }}</span>
                <div class="result-cards">
                  <div v-for="(vote, sId) in getPlayerVotes(player.id)" :key="sId" class="vote-card revealed">
                    <span class="vote-value">{{ vote }}</span>
                  </div>
                </div>
              </div>
            </div>
            <p class="click-notice">Click anywhere to continue...</p>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useGameStore, GamePhase } from '@/stores/gameStore';

const props = defineProps<{
  players: any[];
  currentSession: any;
  gamePhase: GamePhase;
}>();

const emit = defineEmits(['dismiss-reveal']);

const gameStore = useGameStore();

const isRevealed = computed(() => props.gamePhase === GamePhase.REVEALED);

const otherPlayers = computed(() => {
  if (gameStore.currentPlayer) {
    return props.players.filter(p => p.id !== gameStore.currentPlayer?.id);
  }
  return props.players;
});

function getCharacterEmoji(characterId: string) {
  const chars: Record<string, string> = {
    mage: '🧙', paladin: '⚔️', rogue: '🗡️', priest: '✨',
    warrior: '🛡️', hunter: '🏹', warlock: '😈', druid: '🌿'
  };
  return chars[characterId] || '👤';
}

function getSeatStyle(index: number, total: number) {
  const startAngle = 180 + 20;
  const endAngle = 360 - 20;
  const angle = total > 1 
    ? startAngle + (index / (total - 1)) * (endAngle - startAngle)
    : 270;

  const radiusX = 40;
  const radiusY = 25;
  
  const rad = (angle * Math.PI) / 180;
  const left = 50 + Math.cos(rad) * radiusX;
  const top = 60 + Math.sin(rad) * radiusY;

  return {
    left: `${left}%`,
    top: `${top}%`,
    zIndex: Math.floor(top)
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

function getPlayerVotes(playerId: string) {
  if (!props.currentSession) return {};
  const votes: Record<string, string> = {};
  const persistentVotes = props.currentSession.persistentVotes || {};
  for (const [storyId, storyVotes] of Object.entries(persistentVotes)) {
    if ((storyVotes as any)[playerId]) {
      votes[storyId] = (storyVotes as any)[playerId];
    }
  }
  return votes;
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
  image-rendering: auto; /* Changed from pixelated to auto to avoid blur on text */
  user-select: none;
  border-radius: 8px;
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
  opacity: 0.5;
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
}

.wall-torches {
  position: absolute;
  top: 30%;
  width: 100%;
  display: flex;
  justify-content: space-between;
  padding: 0 40px;
}

.torch {
  font-size: 1.5rem;
  animation: flicker 0.15s infinite alternate;
}

@keyframes flicker {
  from { transform: scale(1); filter: drop-shadow(0 0 5px orange); }
  to { transform: scale(1.1); filter: drop-shadow(0 0 15px red); }
}

.table-container {
  position: absolute;
  top: 55%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 90%;
  height: 80%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.round-table {
  width: 80%;
  height: 50%;
  background: #2c1a12;
  border-radius: 50%;
  border: 6px solid #8b6914;
  box-shadow: 
    0 8px 0 #1a0f0a,
    inset 0 0 30px rgba(0,0,0,0.8);
  position: relative;
}

.table-surface {
  position: absolute;
  top: 10%; left: 10%; right: 10%; bottom: 10%;
  background: #3e2723;
  border-radius: 50%;
  border: 2px solid #4e342e;
}

.player-seat {
  position: absolute;
  transform: translate(-50%, -50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  width: 100px;
}

.player-token {
  width: 60px;
  height: 60px;
  position: relative;
}

.token-rim {
  width: 100%;
  height: 100%;
  background: silver;
  border-radius: 50%;
  border: 3px solid #fff;
  box-shadow: 0 4px 0 rgba(0,0,0,0.5);
}

.token-center {
  position: absolute;
  top: 6px; left: 6px; right: 6px; bottom: 6px;
  background: #222;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.8rem;
}

/* Character colors */
.mage .token-rim { background: #4a90e2; }
.paladin .token-rim { background: #f5a623; }
.rogue .token-rim { background: #7ed321; }
.priest .token-rim { background: #d0021b; }
.warrior .token-rim { background: #b68d00; }
.hunter .token-rim { background: #50e3c2; }
.warlock .token-rim { background: #9013fe; }
.druid .token-rim { background: #4a4a4a; }

.player-name {
  margin-top: 8px;
  color: #ffd700;
  font-size: 0.8rem;
  text-shadow: 1px 1px 0 #000;
  background: rgba(0,0,0,0.6);
  padding: 1px 6px;
  border-radius: 4px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.status-badge {
  position: absolute;
  top: -8px;
  right: -8px;
  width: 24px;
  height: 24px;
  background: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
  box-shadow: 0 2px 4px rgba(0,0,0,0.3);
  z-index: 5;
}

.status-badge.voted { background: #7ed321; border: 2px solid #fff; }
.status-badge.waiting { background: #f5a623; border: 2px solid #fff; }

.player-vote-cards {
  position: absolute;
  top: 70px;
  width: 40px;
  height: 55px;
  display: flex;
  justify-content: center;
}

.vote-card {
  width: 32px;
  height: 44px;
  border-radius: 4px;
  border: 1.5px solid #daa520;
  background: #8b4513;
  position: absolute;
  box-shadow: 2px 2px 4px rgba(0,0,0,0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-back {
  width: 80%;
  height: 80%;
  border: 1px solid rgba(218, 165, 32, 0.4);
}

.vote-card.revealed {
  background: #fff;
  border-color: #333;
  color: #333;
  font-weight: bold;
  font-size: 1rem;
  font-family: Arial, sans-serif;
}

/* Animations */
.pop-in-enter-active { animation: pop-in 0.3s; }
@keyframes pop-in {
  0% { transform: scale(0); }
  70% { transform: scale(1.2); }
  100% { transform: scale(1); }
}

.card-deal-enter-active { animation: card-deal 0.4s ease-out; }
@keyframes card-deal {
  0% { transform: translate(0, -40px) scale(0); opacity: 0; }
  100% { transform: translate(var(--tw-translate-x), var(--tw-translate-y)) scale(1); opacity: 1; }
}

/* Reveal Overlay */
.reveal-overlay {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.9);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.reveal-content {
  background: #2c1a12;
  border: 4px solid #d4a756;
  border-radius: 12px;
  padding: 30px;
  max-width: 90%;
  width: 500px;
  box-shadow: 0 0 40px rgba(0,0,0,1);
  text-align: center;
}

.reveal-title {
  color: #d4a756;
  font-size: 2rem;
  margin-bottom: 25px;
}

.results-grid {
  display: flex;
  flex-direction: column;
  gap: 15px;
  margin-bottom: 30px;
  max-height: 300px;
  overflow-y: auto;
  padding-right: 10px;
}

.result-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(212, 167, 86, 0.2);
  padding-bottom: 10px;
}

.result-name {
  color: #fff;
  font-size: 1rem;
  font-weight: bold;
}

.result-cards {
  display: flex;
  gap: 6px;
}

.click-notice {
  color: #a69b8d;
  font-style: italic;
  font-size: 0.9rem;
}

.fade-enter-active, .fade-leave-active { transition: opacity 0.3s; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
