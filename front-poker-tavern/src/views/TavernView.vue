<template>
    <div class="tavern-view">
        <div class="tavern-container">
            <!-- Left Sidebar: Session & Characters -->
            <div class="left-sidebar wow-panel">
                <div class="panel-header">
                    <h3>🎭 Role & Session</h3>
                    <button
                        class="logout-btn"
                        @click="handleLogout"
                        title="Logout"
                    >
                        🚪
                    </button>
                </div>

                <div class="room-info" v-if="gameStore.currentSession">
                    <div class="session-header">
                        <span class="session-name"
                            >🏰 {{ gameStore.currentSession.name }}</span
                        >
                        <span class="player-count"
                            >👥
                            {{
                                gameStore.currentSession.requiredPlayers
                                    ?.length || 0
                            }}</span
                        >
                    </div>

                    <div class="session-id-row">
                        <code class="session-id-code"
                            >{{
                                gameStore.currentSession.id.substring(0, 8)
                            }}...</code
                        >
                        <button
                            @click="copySessionId"
                            class="copy-btn-small"
                            :class="{ copied: showCopied }"
                            :title="
                                showCopied
                                    ? 'Copied invite link!'
                                    : 'Copy invite link to share with team'
                            "
                        >
                            {{ showCopied ? "✅" : "🔗" }}
                        </button>
                    </div>
                    <div class="invite-help" v-if="showCopied">
                        <small>✨ Share this link with your team!</small>
                    </div>
                </div>

                <div v-if="!gameStore.currentSession" class="no-session-notice">
                    <h3>🏠 No Session</h3>
                    <p>
                        You need to be connected to a session to start
                        estimating.
                    </p>
                </div>

                <!-- Character Selection (Vertical List) -->
                <div class="character-selection-list">
                    <button
                        v-for="character in gameStore.availableCharacters"
                        :key="character.id"
                        :class="[
                            'character-btn-list',
                            {
                                active:
                                    gameStore.currentPlayer?.character ===
                                    character.id,
                            },
                        ]"
                        @click="selectCharacter(character.id)"
                        :title="character.name"
                    >
                        <span class="char-emoji">{{ character.emoji }}</span>
                        <span class="char-name">{{ character.name }}</span>
                    </button>
                </div>

                <!-- Issues Overview (moved from right sidebar) -->
                <div
                    v-if="
                        gameStore.currentSession?.stories &&
                        gameStore.currentSession.stories.length > 0
                    "
                    class="issues-section"
                >
                    <h4>📋 Issues to Estimate</h4>
                    <div class="issues-list-left">
                        <div
                            v-for="(story, index) in gameStore.currentSession
                                ?.stories"
                            :key="story.id"
                            class="issue-item-left"
                            :class="{
                                active: index === gameStore.localStoryIndex,
                            }"
                            @click="gameStore.navigateToStory(index)"
                        >
                            <div class="issue-header-left">
                                <span class="issue-number">{{
                                    index + 1
                                }}</span>
                                <span class="issue-key-left">{{
                                    story.jiraKey
                                }}</span>
                            </div>
                            <div class="issue-title-left">
                                {{ story.title }}
                            </div>
                            <div class="issue-status-left">
                                <span
                                    v-if="hasVotedOnStory(story.id)"
                                    class="voted-badge-mini"
                                    >✅</span
                                >
                                <span v-else class="pending-badge-mini"
                                    >⭕</span
                                >
                            </div>
                        </div>
                    </div>
                </div>

                <!-- Poker Cards (moved from right sidebar) -->
                <div v-if="gameStore.currentSession" class="cards-section-left">
                    <PokerCards @submit="submitVote" />
                </div>

                <!-- Admin Controls (moved from right sidebar) -->
                <div v-if="gameStore.currentSession" class="game-controls-left">
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
                        v-if="
                            gameStore.gamePhase === GamePhase.VOTING &&
                            !gameStore.allStoriesVotedByEveryone &&
                            gameStore.currentSession
                        "
                        class="wow-button test-btn"
                        @click="makeOthersVote"
                    >
                        🤖 Make others vote (test)
                    </button>

                    <button
                        class="wow-button summary-btn"
                        @click="showSummaryModal = true"
                    >
                        📊 View Summary
                    </button>
                </div>
            </div>

            <!-- Center Column: Game + Chat -->
            <div class="center-column">
                <!-- Pixel Art game area -->
                <div class="game-area">
                    <PixelTavern 
                        :players="gameStore.players"
                        :currentSession="gameStore.currentSession"
                        :gamePhase="gameStore.gamePhase"
                        @dismiss-reveal="gameStore.resetPhase()"
                    />

                    <div class="game-status">
                        <div
                            class="status-indicator"
                            :class="gameStore.gamePhase"
                        >
                            {{ getPhaseText(gameStore.gamePhase) }}
                        </div>
                    </div>
                </div>
            </div>

            <!-- Right Sidebar: JIRA Import Only -->
            <div class="right-sidebar wow-panel">
                <div class="panel-header">
                    <h3>🔍 JIRA Import</h3>
                </div>

                <!-- JIRA Import -->
                <JiraImport
                    @issue-selected="onIssueSelected"
                    :show-issues-overview="false"
                />
            </div>
        </div>

        <!-- Vote Summary Modal to put in a component -->
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
import { useGameStore, GamePhase } from "@/stores/gameStore";
import JiraImport from "@/components/JiraImport.vue";
import PokerCards from "@/components/PokerCards.vue";
import PixelTavern from "@/components/PixelTavern.vue";

// Global store
const gameStore = useGameStore();

// Vue references
const sessionIdInput = ref<HTMLInputElement>();

// Local state
const showSummaryModal = ref(false);
const showCopied = ref(false);
const selectedIssueForVoting = ref<any>(null);

// Lifecycle
onMounted(() => {
    // Always initialize store to load persisted state
    gameStore.initializeStore();
});

// Add emit to communicate with parent
const emit = defineEmits<{
    logout: [];
}>();

function handleLogout() {
    gameStore.logout();
    emit("logout");
}

// Watchers (Phaser sync removed, PixelTavern uses props)

// Methods
function selectCharacter(characterId: string) {
    gameStore.selectCharacter(characterId);
}

function selectCard(cardValue: string) {
    gameStore.selectCard(cardValue);
}

// Helper to count total votes for a player across all stories
function getPlayerVoteCount(playerId: string): number {
    if (!gameStore.currentSession) return 0;
    let count = 0;
    for (const storyVotes of Object.values(
        gameStore.currentSession.persistentVotes,
    )) {
        if (storyVotes && storyVotes[playerId]) {
            count++;
        }
    }
    return count;
}

function submitVote() {
    console.log("Submitting vote for current player");
    gameStore.submitVote();
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

// Helper functions used in template
function getPhaseText(phase: string): string {
    const phases = {
        waiting: "⏳ Waiting",
        voting: "🗳️ Voting in progress",
        discussion: "💬 Discussion",
        revealed: "🎭 Votes revealed",
    };
    return phases[phase as keyof typeof phases] || phase;
}

function copySessionId() {
    if (gameStore.currentSession?.id) {
        const baseUrl = window.location.origin;
        const inviteLink = `${baseUrl}/?sessionId=${gameStore.currentSession.id}`;
        navigator.clipboard.writeText(inviteLink).then(() => {
            showCopied.value = true;
            setTimeout(() => (showCopied.value = false), 2000);
        });
    }
}

function hasVotedOnStory(storyId: string): boolean {
    if (!gameStore.currentSession || !gameStore.currentPlayer) return false;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    return !!storyVotes[gameStore.currentPlayer.id];
}

function isStoryRevealed(storyId: string): boolean {
    if (!gameStore.currentSession) return false;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    const requiredPlayers = gameStore.currentSession.requiredPlayers;
    return requiredPlayers.length > 0 && requiredPlayers.every((playerId) => storyVotes[playerId]);
}

function getStoryVotesCount(storyId: string): number {
    if (!gameStore.currentSession) return 0;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    return Object.keys(storyVotes).length;
}

function revealAllVotes() {
    gameStore.revealVotes();
}


// Issue selection handler
function onIssueSelected(issue: any) {
    selectedIssueForVoting.value = issue;
    console.log("Issue selected for voting:", issue.title);

    // Find the index of this story in the session and navigate to it
    if (gameStore.currentSession?.stories) {
        const storyIndex = gameStore.currentSession.stories.findIndex(
            (s) => s.id === issue.id,
        );
        if (storyIndex !== -1) {
            gameStore.navigateToStory(storyIndex);
            console.log("Navigated to story index:", storyIndex);
        }
    }

    // Add a chat message about the selection
    gameStore.addChatMessage({
        author: "System",
        text: `Issue selected for voting: ${issue.title}`,
        type: "system",
    });
}
</script>

<style scoped>
.tavern-view {
    min-height: 100vh;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
    padding: 1rem;
    width: 100%;
}

.tavern-container {
    display: grid;
    grid-template-columns: 240px 1fr 340px; /* Refined widths */
    gap: 1rem;
    width: 100%; /* Full width */
    max-width: none; /* No max width */
    margin: 0;
    height: calc(100vh - 85px); /* Full height minus approx header */
    overflow: hidden;
    padding: 0.5rem; /* Tiny padding for edge spacing */
}

/* Left Sidebar */
.left-sidebar {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    height: 100%;
    overflow-y: auto;
    padding-right: 0.5rem;
}

/* Center Column: Game + Chat */
.center-column {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    height: 100%;
    min-width: 0;
}

.game-area {
    flex: 2;
    min-height: 400px;
    position: relative;
    background: #000;
    border: 3px solid #8b4513;
    border-radius: 12px;
    box-shadow: inset 0 0 20px rgba(0, 0, 0, 0.8);
    display: flex;
    justify-content: center;
    align-items: center;
    overflow: hidden;
}

/* Right Sidebar */
.right-sidebar {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    height: 100%;
    overflow-y: auto;
    padding-right: 0.5rem;
}

/* Vertical Character Selection List */
.character-selection-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    flex-shrink: 0;
}

.character-btn-list {
    background: linear-gradient(145deg, #34495e, #2c3e50);
    border: 2px solid #7f8c8d;
    color: #ecf0f1;
    padding: 0.5rem;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    text-align: left;
}

.character-btn-list:hover {
    border-color: #f39c12;
    transform: translateX(2px);
    background: linear-gradient(145deg, #3e5871, #34495e);
}

.character-btn-list.active {
    background: linear-gradient(145deg, #e67e22, #d35400);
    border-color: #ffd700;
    color: #fff;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.2);
}

.char-emoji {
    font-size: 1.5rem;
    filter: drop-shadow(0 2px 2px rgba(0, 0, 0, 0.3));
}

.char-name {
    font-weight: bold;
    font-size: 0.95rem;
}

/* Voting Progress tweak for right sidebar */
.voting-progress {
    background: rgba(0, 0, 0, 0.3);
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid #8b6914;
}

/* Cards section tweak */

/* General Layout Tweaks */
.panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
    border-bottom: 1px solid rgba(139, 69, 19, 0.5);
    padding-bottom: 0.5rem;
}

/* Previous styles... */
.logout-btn {
    /* ... same as before */
    background: rgba(231, 76, 60, 0.2);
    border: 2px solid #e74c3c;
    color: #e74c3c;
    padding: 0.5rem;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 1rem;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.room-info {
    margin-bottom: 1rem;
    background: linear-gradient(
        145deg,
        rgba(26, 15, 10, 0.8),
        rgba(42, 25, 15, 0.8)
    );
    border: 2px solid #8b6914;
    border-radius: 8px;
    padding: 0.75rem;
}

/* Compact session header */
.session-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
}

.session-name {
    color: #d4a756;
    font-weight: bold;
    font-size: 0.95rem;
}

.player-count {
    color: #e8d5b3;
    font-size: 0.85rem;
}

/* Compact session ID row */
.session-id-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
}

.session-id-code {
    background: rgba(0, 0, 0, 0.3);
    color: #95a5a6;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    flex: 1;
}

.copy-btn-small {
    background: transparent;
    border: 1px solid #8b6914;
    color: #d4a756;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    cursor: pointer;
    transition: all 0.2s ease;
    font-size: 0.8rem;
}

.copy-btn-small:hover {
    background: rgba(139, 105, 20, 0.3);
}

.copy-btn-small.copied {
    border-color: #27ae60;
}

.invite-help {
    margin-top: 0.5rem;
    padding: 0.5rem;
    background: rgba(39, 174, 102, 0.1);
    border: 1px solid rgba(39, 174, 102, 0.3);
    border-radius: 4px;
    animation: fadeIn 0.3s ease-in;
}

.invite-help small {
    color: #2ecc71;
    font-size: 0.75rem;
}

@keyframes fadeIn {
    from {
        opacity: 0;
        transform: translateY(-5px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

/* Minimal voting progress */
.voting-progress {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.progress-bar-mini {
    flex: 1;
    height: 6px;
    background: rgba(0, 0, 0, 0.4);
    border-radius: 3px;
    overflow: hidden;
}

.progress-bar-mini .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, #27ae60, #2ecc71);
    border-radius: 3px;
    transition: width 0.3s ease;
}

.progress-text {
    color: #e8d5b3;
    font-size: 0.75rem;
    white-space: nowrap;
}

/* Session sharing styles - keeping for backwards compatibility but simplified */
.session-sharing {
    margin-top: 1rem;
    padding: 0.75rem;
    background: rgba(52, 73, 94, 0.2);
    border-radius: 8px;
    border: 1px solid #34495e;
}

.session-sharing h4 {
    color: #3498db;
    margin: 0 0 1rem 0;
    font-size: 1rem;
    text-align: center;
}

.session-id-display {
    margin-bottom: 1rem;
}

.session-id-display label {
    display: block;
    color: #95a5a6;
    font-size: 0.9rem;
    margin-bottom: 0.5rem;
    font-weight: bold;
}

.session-id-container {
    display: flex;
    gap: 0.5rem;
    align-items: center;
    min-width: 0; /* Allow flex items to shrink */
}

.session-id-input {
    flex: 1;
    min-width: 0; /* Allow input to shrink */
    padding: 0.75rem;
    border: 2px solid #7f8c8d;
    border-radius: 6px;
    background: rgba(52, 73, 94, 0.8);
    color: #ecf0f1;
    font-size: 0.9rem;
    font-family: monospace;
    cursor: pointer;
    transition: border-color 0.3s ease;
    overflow: hidden; /* Prevent text overflow */
    text-overflow: ellipsis;
    white-space: nowrap;
}

.session-id-input:hover {
    border-color: #3498db;
}

.session-id-input:focus {
    outline: none;
    border-color: #3498db;
    box-shadow: 0 0 5px rgba(52, 152, 219, 0.3);
}

.copy-btn {
    padding: 0.75rem;
    background: linear-gradient(145deg, #27ae60, #229954);
    border: 2px solid #2ecc71;
    border-radius: 6px;
    color: white;
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 1rem;
    min-width: 50px;
    height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.copy-btn:hover {
    background: linear-gradient(145deg, #229954, #1e8449);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(46, 204, 113, 0.3);
}

.copy-btn.copied {
    background: linear-gradient(145deg, #3498db, #2980b9);
    border-color: #3498db;
}

.copy-btn.copied:hover {
    background: linear-gradient(145deg, #2980b9, #21618c);
    box-shadow: 0 4px 8px rgba(52, 152, 219, 0.3);
}

.share-instructions {
    background: rgba(241, 196, 15, 0.1);
    border: 1px solid #f39c12;
    border-radius: 6px;
    padding: 0.75rem;
    text-align: center;
}

.share-instructions p {
    margin: 0.25rem 0;
    font-size: 0.85rem;
    color: #ecf0f1;
}

.share-instructions p:first-child {
    color: #f39c12;
    font-weight: bold;
}

.share-instructions strong {
    color: #ffd700;
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

.current-story-display {
    background: rgba(52, 73, 94, 0.8);
    border: 2px solid #7f8c8d;
    border-radius: 8px;
    padding: 1rem;
    margin-top: 0.5rem;
}

.current-story-display h4 {
    color: #3498db;
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
}

.current-story-display p {
    margin: 0 0 0.5rem 0;
    line-height: 1.4;
    font-size: 0.9rem;
}

.current-story-display small {
    color: #95a5a6;
    font-style: italic;
}

/* Current Issue Display */
.current-issue-display {
    background: rgba(52, 73, 94, 0.8);
    border: 2px solid #f39c12;
    border-radius: 8px;
    padding: 1rem;
    margin-bottom: 1.5rem;
}

.current-issue-display h3 {
    color: #f39c12;
    margin: 0 0 1rem 0;
    font-size: 1rem;
    text-align: center;
}

.issue-info h4 {
    color: #ecf0f1;
    margin: 0 0 0.5rem 0;
    font-size: 0.9rem;
    line-height: 1.3;
}

.issue-info p {
    color: #bdc3c7;
    margin: 0 0 0.75rem 0;
    font-size: 0.8rem;
    line-height: 1.4;
}

.issue-meta {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.jira-key {
    background: rgba(52, 152, 219, 0.2);
    color: #3498db;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: bold;
    border: 1px solid #3498db;
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

/* Chat input and send button (remaining) */
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
/* Issues section in left sidebar */
.issues-section {
    margin-top: 1rem;
    background: rgba(44, 62, 80, 0.8);
    border: 2px solid #8b4513;
    border-radius: 8px;
    padding: 0.75rem;
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
}

.issues-section h4 {
    color: #ffd700;
    margin: 0 0 0.5rem 0;
    font-size: 0.9rem;
    text-align: center;
    border-bottom: 1px solid rgba(255, 215, 0, 0.3);
    padding-bottom: 0.5rem;
}

.issues-list-left {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.issue-item-left {
    background: rgba(52, 73, 94, 0.6);
    border: 2px solid #7f8c8d;
    border-radius: 6px;
    padding: 0.5rem;
    cursor: pointer;
    transition: all 0.2s ease;
    position: relative;
}

.issue-item-left:hover {
    background: rgba(52, 73, 94, 0.8);
    border-color: #95a5a6;
}

.issue-item-left.active {
    background: rgba(52, 152, 219, 0.3);
    border-color: #3498db;
}

.issue-header-left {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
}

.issue-number {
    background: #8b6914;
    color: #ffd700;
    font-weight: bold;
    font-size: 0.7rem;
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
    min-width: 1.5rem;
    text-align: center;
}

.issue-key-left {
    color: #3498db;
    font-weight: bold;
    font-size: 0.75rem;
}

.issue-title-left {
    color: #ecf0f1;
    font-size: 0.75rem;
    line-height: 1.2;
    margin-bottom: 0.25rem;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.issue-status-left {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
}

.voted-badge-mini {
    font-size: 0.9rem;
}

.pending-badge-mini {
    font-size: 0.9rem;
    opacity: 0.5;
}

/* Cards section in left sidebar */
.cards-section-left {
    margin-top: 1rem;
}

/* Scrollbar for issues list */
.issues-list-left::-webkit-scrollbar {
    width: 6px;
}

.issues-list-left::-webkit-scrollbar-track {
    background: rgba(0, 0, 0, 0.2);
    border-radius: 3px;
}

.issues-list-left::-webkit-scrollbar-thumb {
    background: #8b6914;
    border-radius: 3px;
}

.issues-list-left::-webkit-scrollbar-thumb:hover {
    background: #daa520;
}

/* Admin controls in left sidebar */
.game-controls-left {
    margin-top: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

/* Scrollbar styling for left sidebar */
.left-sidebar::-webkit-scrollbar {
    width: 8px;
}

.left-sidebar::-webkit-scrollbar-track {
    background: rgba(0, 0, 0, 0.2);
    border-radius: 4px;
}

.left-sidebar::-webkit-scrollbar-thumb {
    background: #8b6914;
    border-radius: 4px;
}

.left-sidebar::-webkit-scrollbar-thumb:hover {
    background: #daa520;
}

@media (max-width: 1200px) {
    .tavern-container {
        grid-template-columns: 200px 1fr 200px;
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
