<template>
    <div class="tavern-view">
        <div class="tavern-container">
            <!-- Control panel on the left -->
            <div class="control-panel wow-panel">
                <h3>🎭 Choose your character</h3>
                <div class="character-selection">
                    <button
                        v-for="character in gameStore.availableCharacters"
                        :key="character.id"
                        :class="[
                            'character-btn',
                            {
                                active:
                                    gameStore.currentPlayer?.character ===
                                    character.id,
                            },
                        ]"
                        @click="selectCharacter(character.id)"
                    >
                        {{ character.emoji }} {{ character.name }}
                    </button>
                </div>

                <div class="room-info" v-if="gameStore.currentSession">
                    <h3>📋 Current session</h3>
                    <p>
                        <strong>Room:</strong>
                        {{ gameStore.currentSession.name }}
                    </p>
                    <p>
                        <strong>Players:</strong>
                        {{ gameStore.players.length }}/8
                    </p>
                    <div class="players-list">
                        <div
                            v-for="player in gameStore.players"
                            :key="player.id"
                            class="player-item"
                        >
                            {{ player.emoji }} {{ player.name }}
                            <span v-if="player.hasVoted" class="voted-badge"
                                >✅</span
                            >
                            <span v-else class="waiting-badge">⏳</span>
                        </div>
                    </div>
                </div>

                <div class="current-vote" v-if="gameStore.currentStory">
                    <h3>🎯 Story to estimate</h3>
                    <div class="story-card">
                        <h4>{{ gameStore.currentStory.title }}</h4>
                        <p>{{ gameStore.currentStory.description }}</p>
                        <small v-if="gameStore.currentStory.jiraKey">{{
                            gameStore.currentStory.jiraKey
                        }}</small>
                    </div>
                </div>

                <div class="poker-cards" v-if="gameStore.currentPlayer">
                    <h3>🃏 Your cards</h3>
                    <div class="cards-grid">
                        <button
                            v-for="card in gameStore.pokerCards"
                            :key="card.value"
                            :class="[
                                'poker-card',
                                {
                                    selected:
                                        gameStore.selectedCard === card.value,
                                },
                            ]"
                            @click="selectCard(card.value)"
                            :disabled="gameStore.gamePhase !== 'voting'"
                            :title="card.description"
                        >
                            {{ card.label }}
                        </button>
                    </div>

                    <button
                        class="wow-button vote-btn"
                        @click="submitVote"
                        :disabled="
                            !gameStore.selectedCard ||
                            gameStore.currentPlayer?.hasVoted
                        "
                    >
                        {{
                            gameStore.currentPlayer?.hasVoted
                                ? "✅ Vote submitted"
                                : "🗳️ Vote"
                        }}
                    </button>
                </div>

                <div class="game-controls">
                    <button
                        v-if="
                            gameStore.gamePhase === 'voting' &&
                            !gameStore.allPlayersVoted
                        "
                        class="wow-button test-btn"
                        @click="makeOthersVote"
                    >
                        🤖 Make others vote (test)
                    </button>

                    <button
                        v-if="gameStore.canReveal"
                        class="wow-button reveal-btn"
                        @click="revealVotes"
                    >
                        🎭 Reveal votes
                    </button>

                    <button
                        v-if="gameStore.gamePhase === 'revealed'"
                        class="wow-button next-btn"
                        @click="nextStory"
                    >
                        ➡️ Next story
                    </button>

                    <div
                        v-if="gameStore.gamePhase === 'revealed'"
                        class="voting-results"
                    >
                        <h4>📊 Results</h4>
                        <div class="results-summary">
                            <p v-if="gameStore.averageVote">
                                <strong>Average:</strong>
                                {{ gameStore.averageVote }}
                            </p>
                            <div class="votes-detail">
                                <div
                                    v-for="result in gameStore.votingResults"
                                    :key="result.player"
                                    class="vote-result"
                                >
                                    {{ result.player }}:
                                    <strong>{{ result.vote }}</strong>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Phaser game area in center -->
            <div class="game-area">
                <div ref="phaserContainer" id="phaser-game"></div>

                <div class="game-status">
                    <div class="status-indicator" :class="gameStore.gamePhase">
                        {{ getPhaseText(gameStore.gamePhase) }}
                    </div>
                </div>
            </div>

            <!-- Chat on the right -->
            <div class="chat-panel wow-panel">
                <h3>💬 Discussion</h3>
                <div class="chat-messages" ref="chatMessages">
                    <div
                        v-for="message in gameStore.chatMessages"
                        :key="message.id"
                        :class="['chat-message', message.type]"
                    >
                        <div class="message-header">
                            <span class="message-author">{{
                                message.author
                            }}</span>
                            <span class="message-time">{{
                                formatTime(message.timestamp)
                            }}</span>
                        </div>
                        <div class="message-text">{{ message.text }}</div>
                    </div>
                </div>

                <div class="chat-input">
                    <input
                        v-model="newMessage"
                        @keyup.enter="sendMessage"
                        placeholder="Type your message..."
                        class="chat-input-field"
                        :disabled="!gameStore.currentPlayer"
                    />
                    <button
                        @click="sendMessage"
                        class="send-btn"
                        :disabled="!newMessage.trim()"
                    >
                        📤
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from "vue";
import { useGameStore } from "@/stores/gameStore";
import { GameManager } from "@/game/GameManager";

// Global store
const gameStore = useGameStore();

// Vue references
const phaserContainer = ref<HTMLElement>();
const chatMessages = ref<HTMLElement>();

// Local state
const gameManager = new GameManager();
const newMessage = ref("");

// Lifecycle
onMounted(() => {
    if (phaserContainer.value) {
        gameManager.init("phaser-game");

        // Initialize test data
        gameStore.connectToSession("test-session");

        // Sync initial players after delay
        setTimeout(() => {
            console.log("Syncing initial players to Phaser");
            gameManager.syncInitialPlayers(gameStore.players);
        }, 500);
    }
});

onUnmounted(() => {
    gameManager.destroy();
});

// Watchers to sync with Phaser
watch(
    () => gameStore.players,
    (newPlayers, oldPlayers) => {
        console.log("Players updated:", newPlayers.length, "players");
        // Avoid excessive automatic synchronization
        // Synchronization is done manually during important actions
    },
    { deep: true },
);

watch(
    () => gameStore.gamePhase,
    (newPhase) => {
        if (newPhase === "revealed") {
            const votes: { [key: string]: string } = {};
            gameStore.players.forEach((player) => {
                if (player.vote) {
                    votes[player.id] = player.vote;
                }
            });
            gameManager.revealAllVotes(votes);
        } else if (newPhase === "voting") {
            gameManager.resetTable();
        }
    },
);

// Methods
function selectCharacter(characterId: string) {
    gameStore.selectCharacter(characterId);

    if (gameStore.currentPlayer) {
        console.log(
            "Adding current player to Phaser:",
            gameStore.currentPlayer.id,
        );
        gameManager.addPlayer(
            gameStore.currentPlayer.id,
            gameStore.currentPlayer,
        );
    }
}

function selectCard(cardValue: string) {
    gameStore.selectCard(cardValue);
}

function submitVote() {
    console.log("Submitting vote for current player");
    gameStore.submitVote();

    if (gameStore.currentPlayer) {
        console.log(
            "Updating player vote in Phaser:",
            gameStore.currentPlayer.id,
        );
        gameManager.updatePlayerVote(gameStore.currentPlayer.id, true);
    }
}

function revealVotes() {
    console.log("Revealing votes");
    gameStore.revealVotes();
}

function nextStory() {
    gameStore.startNextStory();
}

function makeOthersVote() {
    // Automatically make all players who haven't voted yet vote
    const votes = ["1", "2", "3", "5", "8", "13", "?", "☕"];
    gameStore.players.forEach((player) => {
        if (!player.hasVoted) {
            console.log("Making player vote:", player.id);
            player.hasVoted = true;
            player.vote = votes[Math.floor(Math.random() * votes.length)];
            gameManager.updatePlayerVote(player.id, true);
        }
    });

    gameStore.addChatMessage({
        author: "System",
        text: "Other players voted automatically (test mode)",
        type: "system",
    });
}

function sendMessage() {
    if (newMessage.value.trim()) {
        gameStore.sendChatMessage(newMessage.value);
        newMessage.value = "";

        // Scroll to bottom
        nextTick(() => {
            if (chatMessages.value) {
                chatMessages.value.scrollTop = chatMessages.value.scrollHeight;
            }
        });
    }
}

function formatTime(date: Date): string {
    return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
    });
}

function getPhaseText(phase: string): string {
    const phases = {
        waiting: "⏳ Waiting",
        voting: "🗳️ Voting in progress",
        discussion: "💬 Discussion",
        revealed: "🎭 Votes revealed",
    };
    return phases[phase as keyof typeof phases] || phase;
}
</script>

<style scoped>
.tavern-view {
    min-height: 100vh;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
    padding: 1rem;
}

.tavern-container {
    display: grid;
    grid-template-columns: 300px 1fr 300px;
    gap: 1rem;
    max-width: 1400px;
    margin: 0 auto;
    height: calc(100vh - 2rem);
}

/* Control panel */
.control-panel {
    overflow-y: auto;
    max-height: 100%;
}

.control-panel h3 {
    color: #ffd700;
    margin-bottom: 1rem;
    text-align: center;
    font-size: 1.1rem;
}

.character-selection {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
}

.character-btn {
    background: linear-gradient(145deg, #34495e, #2c3e50);
    border: 2px solid #7f8c8d;
    color: #ecf0f1;
    padding: 0.5rem;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 0.9rem;
}

.character-btn:hover {
    border-color: #f39c12;
    transform: translateY(-2px);
}

.character-btn.active {
    background: linear-gradient(145deg, #e67e22, #d35400);
    border-color: #ffd700;
    color: #fff;
}

.room-info {
    margin-bottom: 1.5rem;
}

.room-info p {
    margin: 0.25rem 0;
    font-size: 0.9rem;
}

.players-list {
    margin-top: 0.5rem;
}

.player-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.25rem 0;
    font-size: 0.8rem;
}

.voted-badge {
    color: #27ae60;
}

.waiting-badge {
    color: #f39c12;
}

.current-vote {
    margin-bottom: 1.5rem;
}

.story-card {
    background: rgba(52, 73, 94, 0.8);
    border: 2px solid #7f8c8d;
    border-radius: 8px;
    padding: 1rem;
    margin-top: 0.5rem;
}

.story-card h4 {
    color: #3498db;
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
}

.story-card p {
    margin: 0 0 0.5rem 0;
    line-height: 1.4;
    font-size: 0.9rem;
}

.story-card small {
    color: #95a5a6;
    font-style: italic;
}

.poker-cards {
    margin-bottom: 1.5rem;
}

.cards-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.poker-card {
    aspect-ratio: 3/4;
    background: linear-gradient(145deg, #ecf0f1, #bdc3c7);
    border: 2px solid #95a5a6;
    border-radius: 8px;
    cursor: pointer;
    font-weight: bold;
    font-size: 1rem;
    color: #2c3e50;
    transition: all 0.3s ease;
}

.poker-card:hover:not(:disabled) {
    border-color: #3498db;
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(52, 152, 219, 0.3);
}

.poker-card.selected {
    background: linear-gradient(145deg, #3498db, #2980b9);
    border-color: #ffd700;
    color: white;
}

.poker-card:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.vote-btn,
.reveal-btn,
.next-btn,
.test-btn {
    width: 100%;
    margin: 0.5rem 0;
}

.test-btn {
    background: linear-gradient(145deg, #e74c3c, #c0392b) !important;
    border-color: #e74c3c !important;
}

.test-btn:hover {
    background: linear-gradient(145deg, #c0392b, #a93226) !important;
    box-shadow: 0 4px 8px rgba(231, 76, 60, 0.3) !important;
}

.game-controls {
    margin-top: 1rem;
}

.voting-results {
    margin-top: 1rem;
    padding: 1rem;
    background: rgba(39, 174, 96, 0.2);
    border-radius: 8px;
    border: 2px solid #27ae60;
}

.voting-results h4 {
    color: #27ae60;
    margin: 0 0 0.5rem 0;
}

.votes-detail {
    margin-top: 0.5rem;
}

.vote-result {
    padding: 0.25rem 0;
    font-size: 0.9rem;
}

/* Game area */
.game-area {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: rgba(44, 62, 80, 0.5);
    border-radius: 12px;
    border: 3px solid #8b4513;
}

#phaser-game {
    border-radius: 8px;
    overflow: hidden;
}

.game-status {
    position: absolute;
    top: 1rem;
    left: 50%;
    transform: translateX(-50%);
    z-index: 10;
}

.status-indicator {
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 20px;
    font-weight: bold;
    border: 2px solid #ffd700;
}

.status-indicator.voting {
    border-color: #3498db;
}

.status-indicator.revealed {
    border-color: #27ae60;
}

/* Chat */
.chat-panel {
    display: flex;
    flex-direction: column;
    max-height: 100%;
}

.chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem;
    background: rgba(52, 73, 94, 0.3);
    border-radius: 8px;
    margin-bottom: 1rem;
    max-height: 400px;
}

.chat-message {
    margin-bottom: 0.75rem;
    padding: 0.5rem;
    border-radius: 8px;
}

.chat-message.message {
    background: rgba(52, 152, 219, 0.2);
    border-left: 3px solid #3498db;
}

.chat-message.system {
    background: rgba(241, 196, 15, 0.2);
    border-left: 3px solid #f1c40f;
    font-style: italic;
}

.message-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 0.25rem;
}

.message-author {
    font-weight: bold;
    color: #3498db;
    font-size: 0.9rem;
}

.message-time {
    color: #95a5a6;
    font-size: 0.8rem;
}

.message-text {
    font-size: 0.9rem;
    line-height: 1.4;
}

.chat-input {
    display: flex;
    gap: 0.5rem;
}

.chat-input-field {
    flex: 1;
    padding: 0.75rem;
    border: 2px solid #7f8c8d;
    border-radius: 8px;
    background: rgba(52, 73, 94, 0.8);
    color: white;
    font-size: 0.9rem;
}

.chat-input-field:focus {
    outline: none;
    border-color: #3498db;
}

.chat-input-field::placeholder {
    color: #95a5a6;
}

.send-btn {
    padding: 0.75rem;
    background: linear-gradient(145deg, #27ae60, #229954);
    border: 2px solid #2ecc71;
    border-radius: 8px;
    color: white;
    cursor: pointer;
    transition: all 0.3s ease;
}

.send-btn:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(46, 204, 113, 0.3);
}

.send-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

/* Responsive */
@media (max-width: 1200px) {
    .tavern-container {
        grid-template-columns: 250px 1fr 250px;
    }
}

@media (max-width: 992px) {
    .tavern-container {
        grid-template-columns: 1fr;
        grid-template-rows: auto auto auto;
        gap: 1rem;
    }

    .control-panel,
    .chat-panel {
        max-height: 300px;
    }
}
</style>
