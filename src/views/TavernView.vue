<template>
    <div class="tavern-view">
        <div class="tavern-container">
            <!-- Control panel on the left -->
            <div class="control-panel wow-panel">
                <div class="panel-header">
                    <h3>🎭 Choose your character</h3>
                    <button
                        class="logout-btn"
                        @click="handleLogout"
                        title="Logout"
                    >
                        🚪
                    </button>
                </div>
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
                        <strong>Required players:</strong>
                        {{ gameStore.currentSession.requiredPlayers.length }}
                    </p>

                    <div class="voting-status">
                        <h4>🗳️ Global Voting Status</h4>
                        <div class="status-summary">
                            <span class="votes-count">
                                {{ getTotalVotesCount() }} /
                                {{ getTotalRequiredVotes() }}
                                total votes
                            </span>
                            <div class="progress-bar">
                                <div
                                    class="progress-fill"
                                    :style="{
                                        width: getGlobalVotingProgress() + '%',
                                    }"
                                ></div>
                            </div>
                        </div>

                        <div class="global-status">
                            <div
                                v-if="gameStore.allStoriesVotedByEveryone"
                                class="all-complete-notice"
                            >
                                ✅ All stories voted by everyone!
                            </div>
                            <div v-else class="progress-notice">
                                📋 {{ getCompletedStoriesCount() }} /
                                {{ gameStore.currentSession?.stories.length }}
                                stories complete
                            </div>
                        </div>
                    </div>
                </div>

                <div v-if="!gameStore.currentSession" class="no-session-notice">
                    <h3>🏠 No Session</h3>
                    <p>
                        You need to be connected to a session to start
                        estimating.
                    </p>
                    <p>Please log in or create a session first.</p>
                </div>

                <div class="current-vote" v-if="gameStore.currentStory">
                    <div class="story-header">
                        <h3>🎯 Story to estimate</h3>
                        <span class="story-counter">
                            {{ gameStore.currentStoryProgress.current }} /
                            {{ gameStore.currentStoryProgress.total }}
                        </span>
                    </div>

                    <div class="story-card">
                        <div class="story-status">
                            <span
                                v-if="gameStore.isCurrentStoryRevealed"
                                class="status-badge revealed"
                            >
                                ✅ Revealed
                            </span>
                            <span
                                v-else-if="
                                    gameStore.allPlayersVotedCurrentStory
                                "
                                class="status-badge ready"
                            >
                                🎭 Ready to reveal
                            </span>
                            <span
                                v-else-if="hasCurrentPlayerVoted()"
                                class="status-badge voted"
                            >
                                🗳️ You voted
                            </span>
                            <span v-else class="status-badge pending">
                                ⏳ Not voted
                            </span>
                        </div>

                        <h4>{{ gameStore.currentStory.title }}</h4>
                        <p>{{ gameStore.currentStory.description }}</p>
                        <small v-if="gameStore.currentStory.jiraKey">{{
                            gameStore.currentStory.jiraKey
                        }}</small>
                    </div>
                </div>

                <div
                    class="stories-overview"
                    v-if="
                        gameStore.currentSession &&
                        gameStore.currentSession.stories.length > 0
                    "
                >
                    <h3>📋 Stories Overview</h3>
                    <div class="stories-list">
                        <div
                            v-for="(story, index) in gameStore.currentSession
                                .stories"
                            :key="story.id"
                            :class="[
                                'story-item',
                                { active: index === gameStore.localStoryIndex },
                            ]"
                            @click="gameStore.navigateToStory(index)"
                        >
                            <div class="story-item-header">
                                <span class="story-number">{{
                                    index + 1
                                }}</span>
                                <div class="story-item-status">
                                    <span
                                        v-if="isStoryRevealed(story.id)"
                                        class="mini-badge revealed"
                                        title="All voted, revealed"
                                    >
                                        ✅
                                    </span>
                                    <span
                                        v-else-if="isStoryReady(story.id)"
                                        class="mini-badge ready"
                                        title="All voted, ready to reveal"
                                    >
                                        🎭
                                    </span>
                                    <span
                                        v-else-if="hasVotedOnStory(story.id)"
                                        class="mini-badge voted"
                                        title="You voted"
                                    >
                                        🗳️
                                    </span>
                                    <span
                                        v-else
                                        class="mini-badge pending"
                                        title="Not voted"
                                    >
                                        ⏳
                                    </span>
                                </div>
                            </div>
                            <div class="story-item-title">
                                {{ story.title }}
                            </div>
                            <div class="story-votes-count">
                                {{ getStoryVotesCount(story.id) }} /
                                {{
                                    gameStore.currentSession.requiredPlayers
                                        .length
                                }}
                                votes
                            </div>
                        </div>
                    </div>
                </div>

                <div class="poker-cards">
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
                            !gameStore.currentPlayer ||
                            gameStore.currentPlayer?.hasVoted
                        "
                    >
                        {{
                            !gameStore.currentPlayer
                                ? "👤 Select character first"
                                : gameStore.currentPlayer?.hasVoted
                                  ? "✅ Vote submitted"
                                  : "🗳️ Vote"
                        }}
                    </button>
                </div>

                <div class="game-controls">
                    <button
                        v-if="
                            gameStore.gamePhase === 'voting' &&
                            !gameStore.allStoriesVotedByEveryone &&
                            gameStore.currentSession
                        "
                        class="wow-button test-btn"
                        @click="makeOthersVote"
                    >
                        🤖 Make others vote (test)
                    </button>

                    <button
                        v-if="
                            gameStore.canReveal &&
                            !gameStore.isCurrentStoryRevealed
                        "
                        class="wow-button reveal-btn"
                        @click="revealAllVotes"
                    >
                        🎭 Reveal All Votes
                    </button>

                    <button
                        class="wow-button summary-btn"
                        @click="showSummaryModal = true"
                    >
                        📊 View Summary
                    </button>
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

        <!-- Vote Summary Modal -->
        <div
            v-if="showSummaryModal"
            class="modal-overlay"
            @click="showSummaryModal = false"
        >
            <div class="modal-content" @click.stop>
                <div class="modal-header">
                    <h2>📊 Voting Summary</h2>
                    <button class="close-btn" @click="showSummaryModal = false">
                        ✕
                    </button>
                </div>

                <div class="modal-body">
                    <div class="summary-stats">
                        <div class="stat-item">
                            <span class="stat-label">Total Stories:</span>
                            <span class="stat-value">{{
                                gameStore.currentSession?.stories.length || 0
                            }}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Completed:</span>
                            <span class="stat-value">{{
                                getCompletedStoriesCount()
                            }}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">Your Progress:</span>
                            <span class="stat-value">{{
                                getYourVotesCount()
                            }}</span>
                        </div>
                    </div>

                    <div class="stories-summary">
                        <div
                            v-for="(story, index) in gameStore.currentSession
                                ?.stories"
                            :key="story.id"
                            :class="[
                                'summary-story-item',
                                { revealed: isStoryRevealed(story.id) },
                            ]"
                        >
                            <div class="summary-story-header">
                                <span class="summary-story-number">{{
                                    index + 1
                                }}</span>
                                <div class="summary-story-title">
                                    {{ story.title }}
                                </div>
                                <div class="summary-story-status">
                                    <span
                                        v-if="isStoryRevealed(story.id)"
                                        class="status-revealed"
                                        >✅</span
                                    >
                                    <span v-else class="status-pending"
                                        >{{ getStoryVotesCount(story.id) }}/{{
                                            gameStore.currentSession
                                                ?.requiredPlayers.length
                                        }}</span
                                    >
                                </div>
                            </div>

                            <div
                                v-if="isStoryRevealed(story.id)"
                                class="summary-votes"
                            >
                                <div class="votes-grid">
                                    <div
                                        v-for="result in getStoryResults(
                                            story.id,
                                        )"
                                        :key="result.player"
                                        class="vote-item"
                                    >
                                        <span class="vote-player">{{
                                            result.player
                                        }}</span>
                                        <span class="vote-value">{{
                                            result.vote
                                        }}</span>
                                    </div>
                                </div>
                                <div
                                    v-if="getStoryAverage(story.id)"
                                    class="story-average"
                                >
                                    Average:
                                    <strong>{{
                                        getStoryAverage(story.id)
                                    }}</strong>
                                </div>
                            </div>
                        </div>
                    </div>
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
const showSummaryModal = ref(false);

// Lifecycle
onMounted(() => {
    if (phaserContainer.value) {
        gameManager.init("phaser-game");

        // Initialize store
        gameStore.initializeStore();

        // Sync any existing players after delay
        setTimeout(() => {
            if (gameStore.players.length > 0) {
                console.log("Syncing initial players to Phaser");
                gameManager.syncInitialPlayers(gameStore.players);
            }
        }, 500);
    }
});

// Add emit to communicate with parent
const emit = defineEmits<{
    logout: [];
}>();

function handleLogout() {
    gameStore.logout();
    emit("logout");
}

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
            // Global reveal - show votes for all stories
            setTimeout(() => {
                const votes: { [key: string]: string } = {};
                if (gameStore.currentStory && gameStore.currentSession) {
                    const currentStoryVotes =
                        gameStore.currentSession.persistentVotes[
                            gameStore.currentStory.id
                        ] || {};
                    Object.entries(currentStoryVotes).forEach(
                        ([playerId, vote]) => {
                            votes[playerId] = vote;
                        },
                    );
                    gameManager.revealAllVotes(votes);
                }
            }, 500);
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

        // Update all players' vote cards after current player votes
        gameStore.players.forEach((player) => {
            if (player.hasVoted) {
                gameManager.updatePlayerVote(player.id, true);
            }
        });
    }
}

function getCompletedStoriesCount(): number {
    if (!gameStore.currentSession) return 0;
    return gameStore.currentSession.stories.filter((story) =>
        isStoryRevealed(story.id),
    ).length;
}

function getYourVotesCount(): number {
    if (!gameStore.currentSession || !gameStore.currentPlayer) return 0;
    return gameStore.currentSession.stories.filter((story) =>
        hasVotedOnStory(story.id),
    ).length;
}

function getStoryResults(storyId: string) {
    if (!gameStore.currentSession) return [];
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    const playerMap = new Map(gameStore.players.map((p) => [p.id, p]));

    return Object.entries(storyVotes).map(([playerId, vote]) => {
        const player = playerMap.get(playerId);
        return {
            player: player?.name || playerId,
            vote: vote,
        };
    });
}

function getStoryAverage(storyId: string): number | null {
    if (!gameStore.currentSession) return null;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    const numericVotes = Object.values(storyVotes)
        .filter((vote) => vote && !isNaN(Number(vote)))
        .map(Number);

    if (numericVotes.length === 0) return null;
    const sum = numericVotes.reduce((acc, vote) => acc + vote, 0);
    return Math.round((sum / numericVotes.length) * 10) / 10;
}

function hasCurrentPlayerVoted(): boolean {
    return gameStore.hasCurrentPlayerVoted();
}

function makeOthersVote() {
    // Automatically make all players who haven't voted yet vote
    const votes = ["1", "2", "3", "5", "8", "13", "?", "☕"];

    if (!gameStore.currentSession || !gameStore.currentStory) return;

    const currentStoryVotes =
        gameStore.currentSession.persistentVotes[gameStore.currentStory.id] ||
        {};

    gameStore.currentSession.requiredPlayers.forEach((playerId) => {
        if (!currentStoryVotes[playerId]) {
            console.log("Making player vote:", playerId);
            const randomVote = votes[Math.floor(Math.random() * votes.length)];

            // Add to persistent votes
            if (
                !gameStore.currentSession!.persistentVotes[
                    gameStore.currentStory!.id
                ]
            ) {
                gameStore.currentSession!.persistentVotes[
                    gameStore.currentStory!.id
                ] = {};
            }
            gameStore.currentSession!.persistentVotes[
                gameStore.currentStory!.id
            ][playerId] = randomVote;

            // Update local player state
            const player = gameStore.players.find((p) => p.id === playerId);
            if (player) {
                player.hasVoted = true;
                player.vote = randomVote;
                gameManager.updatePlayerVote(player.id, true);
            }
        }
    });

    // Save state and update UI
    gameStore.savePersistedState();
    gameStore.updatePlayerVotedStatus();

    gameStore.addChatMessage({
        author: "System",
        text: "Other players voted automatically (test mode)",
        type: "system",
    });
}

// Helper functions for the new UI
function getTotalVotesCount(): number {
    if (!gameStore.currentSession) return 0;
    let totalVotes = 0;
    gameStore.currentSession.stories.forEach((story) => {
        const storyVotes =
            gameStore.currentSession!.persistentVotes[story.id] || {};
        totalVotes += Object.keys(storyVotes).length;
    });
    return totalVotes;
}

function getTotalRequiredVotes(): number {
    if (!gameStore.currentSession) return 0;
    return (
        gameStore.currentSession.stories.length *
        gameStore.currentSession.requiredPlayers.length
    );
}

function getGlobalVotingProgress(): number {
    const total = getTotalRequiredVotes();
    const voted = getTotalVotesCount();
    return total > 0 ? (voted / total) * 100 : 0;
}

function hasPlayerVoted(playerId: string): boolean {
    if (!gameStore.currentSession || !gameStore.currentStory) return false;
    const currentStoryVotes =
        gameStore.currentSession.persistentVotes[gameStore.currentStory.id] ||
        {};
    return !!currentStoryVotes[playerId];
}

function getPlayerInfo(playerId: string) {
    const player = gameStore.players.find((p) => p.id === playerId);
    return player || { name: playerId, emoji: "👤" };
}

function getVoteTime(playerId: string): string {
    // Placeholder for vote timestamp - would come from backend in real implementation
    return "Voted recently";
}

function revealAllVotes() {
    console.log("Revealing all votes globally");
    gameStore.revealVotes();
}

// Helper functions for stories overview
function isStoryRevealed(storyId: string): boolean {
    if (!gameStore.currentSession) return false;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    const requiredPlayers = gameStore.currentSession.requiredPlayers;
    return requiredPlayers.every((playerId) => storyVotes[playerId]);
}

function isStoryReady(storyId: string): boolean {
    return isStoryRevealed(storyId); // Same logic for now
}

function hasVotedOnStory(storyId: string): boolean {
    if (!gameStore.currentSession || !gameStore.currentPlayer) return false;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    return !!storyVotes[gameStore.currentPlayer.id];
}

function getStoryVotesCount(storyId: string): number {
    if (!gameStore.currentSession) return 0;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    return Object.keys(storyVotes).length;
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

.panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
}

.panel-header h3 {
    color: #ffd700;
    text-align: center;
    font-size: 1.1rem;
    margin: 0;
}

.logout-btn {
    background: rgba(231, 76, 60, 0.2);
    border: 2px solid #e74c3c;
    color: #e74c3c;
    padding: 0.5rem;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 1rem;
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.logout-btn:hover {
    background: rgba(231, 76, 60, 0.4);
    transform: scale(1.1);
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

.story-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
}

.story-header h3 {
    margin: 0;
}

.story-navigation {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.nav-btn {
    background: linear-gradient(145deg, #34495e, #2c3e50);
    border: 2px solid #7f8c8d;
    color: #ecf0f1;
    padding: 0.5rem;
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 1rem;
    width: 36px;
    height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.nav-btn:hover:not(:disabled) {
    border-color: #3498db;
    transform: translateY(-1px);
}

.nav-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
}

.story-counter {
    font-weight: bold;
    color: #3498db;
    font-size: 0.9rem;
    min-width: 50px;
    text-align: center;
}

.story-status {
    margin-bottom: 0.5rem;
}

.status-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: bold;
    display: inline-block;
}

.status-badge.revealed {
    background: rgba(46, 204, 113, 0.2);
    color: #2ecc71;
    border: 1px solid #2ecc71;
}

.status-badge.ready {
    background: rgba(241, 196, 15, 0.2);
    color: #f1c40f;
    border: 1px solid #f1c40f;
}

.status-badge.voted {
    background: rgba(52, 152, 219, 0.2);
    color: #3498db;
    border: 1px solid #3498db;
}

.status-badge.pending {
    background: rgba(149, 165, 166, 0.2);
    color: #95a5a6;
    border: 1px solid #95a5a6;
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
.test-btn,
.summary-btn {
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

.summary-btn {
    background: linear-gradient(145deg, #9b59b6, #8e44ad) !important;
    border-color: #9b59b6 !important;
}

.summary-btn:hover {
    background: linear-gradient(145deg, #8e44ad, #7d3c98) !important;
    box-shadow: 0 4px 8px rgba(155, 89, 182, 0.3) !important;
}

.reveal-btn {
    background: linear-gradient(145deg, #e67e22, #d35400) !important;
    border-color: #e67e22 !important;
}

.reveal-btn:hover {
    background: linear-gradient(145deg, #d35400, #ca6f1e) !important;
    box-shadow: 0 4px 8px rgba(230, 126, 34, 0.3) !important;
}

.global-status {
    margin-top: 1rem;
    text-align: center;
}

.all-complete-notice {
    background: rgba(46, 204, 113, 0.2);
    border: 2px solid #2ecc71;
    color: #2ecc71;
    padding: 1rem;
    border-radius: 8px;
    font-weight: bold;
    font-size: 1.1rem;
}

.progress-notice {
    background: rgba(52, 152, 219, 0.2);
    border: 2px solid #3498db;
    color: #3498db;
    padding: 0.75rem;
    border-radius: 8px;
    font-weight: bold;
}

.no-session-notice {
    text-align: center;
    padding: 2rem;
    background: rgba(149, 165, 166, 0.2);
    border: 2px solid #95a5a6;
    border-radius: 8px;
    margin-bottom: 1.5rem;
}

.no-session-notice h3 {
    color: #95a5a6;
    margin-bottom: 1rem;
}

.no-session-notice p {
    color: #ecf0f1;
    margin-bottom: 0.5rem;
    font-style: italic;
}

/* Modal Styles */
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
}

.modal-content {
    background: linear-gradient(145deg, #2c2c54, #2c3e50);
    border: 3px solid #8b4513;
    border-radius: 12px;
    padding: 2rem;
    max-width: 800px;
    max-height: 80vh;
    overflow-y: auto;
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.5);
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    border-bottom: 2px solid #daa520;
    padding-bottom: 1rem;
}

.modal-header h2 {
    color: #ffd700;
    margin: 0;
}

.close-btn {
    background: none;
    border: none;
    color: #ecf0f1;
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 50%;
    transition: background 0.3s ease;
}

.close-btn:hover {
    background: rgba(231, 76, 60, 0.2);
    color: #e74c3c;
}

.summary-stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 1rem;
    margin-bottom: 2rem;
}

.stat-item {
    background: rgba(52, 73, 94, 0.8);
    padding: 1rem;
    border-radius: 8px;
    text-align: center;
    border: 2px solid #7f8c8d;
}

.stat-label {
    display: block;
    color: #95a5a6;
    font-size: 0.9rem;
    margin-bottom: 0.5rem;
}

.stat-value {
    display: block;
    color: #ecf0f1;
    font-size: 1.5rem;
    font-weight: bold;
}

.stories-summary {
    max-height: 400px;
    overflow-y: auto;
}

.summary-story-item {
    background: rgba(52, 73, 94, 0.5);
    border: 2px solid #7f8c8d;
    border-radius: 8px;
    padding: 1rem;
    margin-bottom: 1rem;
}

.summary-story-item.revealed {
    border-color: #2ecc71;
    background: rgba(46, 204, 113, 0.1);
}

.summary-story-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 0.5rem;
}

.summary-story-number {
    background: #3498db;
    color: white;
    border-radius: 50%;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: bold;
}

.summary-story-title {
    flex: 1;
    font-weight: bold;
    color: #ecf0f1;
}

.summary-story-status {
    font-size: 1.2rem;
}

.status-revealed {
    color: #2ecc71;
}

.status-pending {
    color: #f39c12;
    font-size: 0.9rem;
}

.summary-votes {
    margin-top: 1rem;
    padding-top: 1rem;
    border-top: 1px solid #7f8c8d;
}

.votes-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.vote-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem;
    background: rgba(52, 152, 219, 0.2);
    border-radius: 6px;
    border: 1px solid #3498db;
}

.vote-player {
    font-size: 0.9rem;
    color: #ecf0f1;
}

.vote-value {
    font-weight: bold;
    color: #3498db;
    background: rgba(255, 255, 255, 0.1);
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
}

.story-average {
    text-align: center;
    color: #f39c12;
    font-weight: bold;
    padding: 0.5rem;
    background: rgba(241, 196, 15, 0.1);
    border-radius: 6px;
    border: 1px solid #f39c12;
}

.stories-overview {
    margin-bottom: 1.5rem;
}

.stories-overview h3 {
    color: #ffd700;
    margin-bottom: 1rem;
    text-align: center;
    font-size: 1.1rem;
}

.stories-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-height: 300px;
    overflow-y: auto;
}

.story-item {
    background: rgba(52, 73, 94, 0.5);
    border: 2px solid #7f8c8d;
    border-radius: 8px;
    padding: 0.75rem;
    cursor: pointer;
    transition: all 0.3s ease;
}

.story-item:hover {
    border-color: #3498db;
    background: rgba(52, 152, 219, 0.2);
}

.story-item.active {
    border-color: #f39c12;
    background: rgba(243, 156, 18, 0.2);
    box-shadow: 0 0 8px rgba(243, 156, 18, 0.3);
}

.story-item-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
}

.story-number {
    background: #3498db;
    color: white;
    border-radius: 50%;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 0.8rem;
    font-weight: bold;
}

.story-item.active .story-number {
    background: #f39c12;
}

.story-item-status {
    display: flex;
    align-items: center;
}

.mini-badge {
    font-size: 1rem;
    opacity: 0.8;
}

.story-item-title {
    font-weight: bold;
    color: #ecf0f1;
    font-size: 0.9rem;
    margin-bottom: 0.25rem;
    line-height: 1.2;
}

.story-votes-count {
    font-size: 0.8rem;
    color: #95a5a6;
}

.game-controls {
    margin-top: 1rem;
}

.voting-status h4 {
    color: #3498db;
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
}

.status-summary {
    margin-bottom: 1rem;
}

.votes-count {
    font-weight: bold;
    color: #ecf0f1;
    font-size: 0.9rem;
}

.progress-bar {
    width: 100%;
    height: 8px;
    background: rgba(127, 140, 141, 0.3);
    border-radius: 4px;
    margin-top: 0.5rem;
    overflow: hidden;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #3498db, #2ecc71);
    transition: width 0.3s ease;
}

.voting-info {
    margin-bottom: 1rem;
    padding: 1rem;
    border-radius: 8px;
}

.all-voted-notice {
    background: rgba(46, 204, 113, 0.2);
    border: 2px solid #2ecc71;
    text-align: center;
}

.all-voted-notice p {
    margin: 0.25rem 0;
    color: #2ecc71;
    font-weight: bold;
}

.auto-reveal-notice {
    font-size: 0.8rem !important;
    font-style: italic !important;
    opacity: 0.8 !important;
}

.waiting-notice {
    background: rgba(241, 196, 15, 0.2);
    border: 2px solid #f1c40f;
    text-align: center;
}

.waiting-notice p {
    margin: 0.25rem 0;
    color: #f1c40f;
    font-weight: bold;
}

.pending-players {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    justify-content: center;
    margin-top: 0.5rem;
}

.pending-player {
    background: rgba(231, 76, 60, 0.3);
    color: #e74c3c;
    padding: 0.25rem 0.5rem;
    border-radius: 12px;
    font-size: 0.8rem;
    border: 1px solid #e74c3c;
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
