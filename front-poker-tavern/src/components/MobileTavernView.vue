<template>
    <div class="mobile-layout">
        <!-- Mobile Top Bar -->
        <div class="mobile-top-bar">
            <div class="mobile-top-left">
                <span
                    class="mobile-session-name"
                    v-if="gameStore.currentSession"
                >
                    🏰 {{ gameStore.currentSession.name }}
                </span>
                <span class="mobile-session-name" v-else>🏰 No Session</span>
            </div>
            <div class="mobile-top-right">
                <span
                    v-if="gameStore.currentSession"
                    class="mobile-player-count"
                >
                    👥
                    {{ gameStore.currentSession.requiredPlayers?.length || 0 }}
                </span>
                <button
                    v-if="gameStore.currentSession"
                    @click="copySessionId"
                    class="mobile-share-btn"
                    :title="showCopied ? 'Copied!' : 'Share link'"
                >
                    {{ showCopied ? "✅" : "🔗" }}
                </button>
                <button
                    class="mobile-logout-btn"
                    @click="handleLogout"
                    title="Logout"
                >
                    🚪
                </button>
            </div>
        </div>

        <!-- Mobile Tab Navigation -->
        <div class="mobile-tabs">
            <button
                class="mobile-tab"
                :class="{ active: activeTab === 'tickets' }"
                @click="activeTab = 'tickets'"
            >
                📋 Tickets
            </button>
            <button
                class="mobile-tab"
                :class="{ active: activeTab === 'vote' }"
                @click="activeTab = 'vote'"
            >
                🃏 Vote
            </button>
            <button
                class="mobile-tab"
                :class="{ active: activeTab === 'summary' }"
                @click="activeTab = 'summary'"
            >
                📊 Summary
            </button>
        </div>

        <!-- Mobile Tab Content -->
        <div class="mobile-content">
            <!-- TICKETS TAB -->
            <div v-if="activeTab === 'tickets'" class="mobile-tab-panel">
                <!-- No session notice -->
                <div v-if="!gameStore.currentSession" class="mobile-no-session">
                    <h3>🏠 No Session</h3>
                    <p>
                        You need to be connected to a session to start
                        estimating.
                    </p>
                </div>

                <!-- JIRA Import -->
                <div
                    v-if="gameStore.currentSession"
                    class="mobile-jira-section"
                >
                    <JiraImport
                        @issue-selected="onIssueSelected"
                        :show-issues-overview="false"
                    />
                </div>

                <!-- Issues List -->
                <div
                    v-if="
                        gameStore.currentSession?.stories &&
                        gameStore.currentSession.stories.length > 0
                    "
                    class="mobile-issues-section"
                >
                    <h4>
                        📋 Issues to Estimate ({{
                            gameStore.currentSession.stories.length
                        }})
                    </h4>
                    <div class="mobile-issues-list">
                        <div
                            v-for="(story, index) in gameStore.currentSession
                                ?.stories"
                            :key="story.id"
                            class="mobile-issue-item"
                            :class="{
                                active: index === gameStore.localStoryIndex,
                            }"
                            @click="goToStoryAndVote(index)"
                        >
                            <div class="mobile-issue-header">
                                <span class="mobile-issue-number">{{
                                    index + 1
                                }}</span>
                                <span class="mobile-issue-key">{{
                                    story.jiraKey
                                }}</span>
                                <div class="mobile-issue-badges">
                                    <button
                                        class="mobile-view-btn"
                                        @click.stop="viewIssueDetails(story)"
                                        title="View details"
                                    >
                                        👁️
                                    </button>
                                    <span
                                        v-if="hasVotedOnStory(story.id)"
                                        class="mobile-voted"
                                        >✅</span
                                    >
                                    <span v-else class="mobile-pending"
                                        >⭕</span
                                    >
                                </div>
                            </div>
                            <div class="mobile-issue-title">
                                {{ story.title }}
                            </div>
                        </div>
                    </div>
                </div>

                <div
                    v-else-if="gameStore.currentSession"
                    class="mobile-no-issues"
                >
                    <p>
                        📭 No tickets imported yet. Use JIRA Import above to add
                        tickets.
                    </p>
                </div>
            </div>

            <!-- VOTE TAB -->
            <div v-if="activeTab === 'vote'" class="mobile-tab-panel">
                <div v-if="!gameStore.currentSession" class="mobile-no-session">
                    <h3>🏠 No Session</h3>
                    <p>Connect to a session first.</p>
                </div>

                <template v-else>
                    <!-- Character Selection -->
                    <div class="mobile-character-section">
                        <h4>🎭 Choose your character</h4>
                        <div class="mobile-character-grid">
                            <button
                                v-for="character in gameStore.availableCharacters"
                                :key="character.id"
                                class="mobile-character-btn"
                                :class="{
                                    active:
                                        gameStore.currentPlayer?.character ===
                                        character.id,
                                }"
                                @click="selectCharacter(character.id)"
                            >
                                <span class="mobile-char-emoji">{{
                                    character.emoji
                                }}</span>
                                <span class="mobile-char-name">{{
                                    character.name
                                }}</span>
                            </button>
                        </div>
                    </div>

                    <!-- Current Story -->
                    <div
                        v-if="gameStore.currentStory"
                        class="mobile-current-story"
                    >
                        <div class="mobile-story-nav">
                            <button
                                class="mobile-nav-btn"
                                @click="gameStore.previousStory()"
                                :disabled="gameStore.localStoryIndex <= 0"
                            >
                                ◀
                            </button>
                            <span class="mobile-story-counter">
                                {{ gameStore.currentStoryProgress.current }} /
                                {{ gameStore.currentStoryProgress.total }}
                            </span>
                            <button
                                class="mobile-nav-btn"
                                @click="gameStore.nextStory()"
                                :disabled="
                                    gameStore.localStoryIndex >=
                                    (gameStore.currentSession?.stories
                                        ?.length || 1) -
                                        1
                                "
                            >
                                ▶
                            </button>
                        </div>
                        <div class="mobile-story-info">
                            <span class="mobile-story-key">{{
                                gameStore.currentStory.jiraKey
                            }}</span>
                            <h3>{{ gameStore.currentStory.title }}</h3>
                            <button
                                class="mobile-detail-btn"
                                @click="
                                    viewIssueDetails(gameStore.currentStory)
                                "
                            >
                                👁️ View Details
                            </button>
                        </div>
                    </div>

                    <!-- Poker Cards -->
                    <div class="mobile-cards-section">
                        <PokerCards @submit="submitVote" />
                    </div>

                    <!-- Controls -->
                    <div class="mobile-controls">
                        <button
                            v-if="
                                gameStore.currentSession &&
                                !gameStore.isCurrentStoryRevealed
                            "
                            class="wow-button mobile-reveal-btn"
                            :disabled="!gameStore.canReveal"
                            @click="revealAllVotes"
                        >
                            {{
                                !gameStore.canReveal
                                    ? "⏳ Waiting for votes..."
                                    : "🎭 Reveal All Votes"
                            }}
                        </button>
                        <button
                            v-if="
                                gameStore.gamePhase === GamePhase.VOTING &&
                                !gameStore.allStoriesVotedByEveryone &&
                                gameStore.currentSession
                            "
                            class="wow-button mobile-test-btn"
                            @click="makeOthersVote"
                        >
                            🤖 Make others vote (test)
                        </button>
                    </div>
                </template>
            </div>

            <!-- SUMMARY TAB -->
            <div v-if="activeTab === 'summary'" class="mobile-tab-panel">
                <div v-if="!gameStore.currentSession" class="mobile-no-session">
                    <h3>📊 No Data</h3>
                    <p>Connect to a session to see the summary.</p>
                </div>

                <template v-else>
                    <div class="mobile-summary-stats">
                        <div class="mobile-stat">
                            <span class="mobile-stat-value">{{
                                gameStore.currentSession?.stories.length || 0
                            }}</span>
                            <span class="mobile-stat-label">Total</span>
                        </div>
                        <div class="mobile-stat">
                            <span class="mobile-stat-value">{{
                                getCompletedStoriesCount()
                            }}</span>
                            <span class="mobile-stat-label">Done</span>
                        </div>
                        <div class="mobile-stat">
                            <span class="mobile-stat-value">{{
                                getYourVotesCount()
                            }}</span>
                            <span class="mobile-stat-label">Your Votes</span>
                        </div>
                    </div>

                    <div class="mobile-summary-list">
                        <div
                            v-for="(story, index) in gameStore.currentSession
                                ?.stories"
                            :key="story.id"
                            class="mobile-summary-item"
                            :class="{ revealed: isStoryRevealed(story.id) }"
                        >
                            <div class="mobile-summary-header">
                                <span class="mobile-summary-number">{{
                                    index + 1
                                }}</span>
                                <div class="mobile-summary-title">
                                    {{ story.title }}
                                </div>
                                <span
                                    v-if="isStoryRevealed(story.id)"
                                    class="mobile-status-done"
                                    >✅</span
                                >
                                <span v-else class="mobile-status-pending">
                                    {{ getStoryVotesCount(story.id) }}/{{
                                        gameStore.currentSession
                                            ?.requiredPlayers.length
                                    }}
                                </span>
                            </div>
                            <div
                                v-if="isStoryRevealed(story.id)"
                                class="mobile-summary-votes"
                            >
                                <div class="mobile-votes-row">
                                    <div
                                        v-for="result in getStoryResults(
                                            story.id,
                                        )"
                                        :key="result.player"
                                        class="mobile-vote-chip"
                                    >
                                        <span class="mobile-vote-player">{{
                                            result.player
                                        }}</span>
                                        <span class="mobile-vote-value">{{
                                            result.vote
                                        }}</span>
                                    </div>
                                </div>
                                <div
                                    v-if="getStoryAverage(story.id)"
                                    class="mobile-story-avg"
                                >
                                    Average:
                                    <strong>{{
                                        getStoryAverage(story.id)
                                    }}</strong>
                                </div>
                            </div>
                        </div>
                    </div>
                </template>
            </div>
        </div>

        <!-- Issue Details Modal -->
        <JiraIssueModal
            ref="issueDetailModal"
            :issue="viewingIssue"
            mode="view"
        />
    </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick } from "vue";
import { useGameStore, GamePhase } from "@/stores/gameStore";
import JiraImport from "@/components/JiraImport.vue";
import PokerCards from "@/components/PokerCards.vue";
import JiraIssueModal from "@/components/JiraIssueModal.vue";

const gameStore = useGameStore();
const activeTab = ref<"tickets" | "vote" | "summary">("tickets");
const showCopied = ref(false);
const viewingIssue = ref<any>(null);
const issueDetailModal = ref<InstanceType<typeof JiraIssueModal>>();

const emit = defineEmits<{
    logout: [];
}>();

function handleLogout() {
    emit("logout");
}

function selectCharacter(characterId: string) {
    gameStore.selectCharacter(characterId);
}

function goToStoryAndVote(index: number) {
    gameStore.navigateToStory(index);
    activeTab.value = "vote";
}

function submitVote() {
    console.log("Submitting vote");
    gameStore.submitVote();
}

function copySessionId() {
    if (gameStore.currentSession?.id) {
        const baseUrl = window.location.origin;
        const inviteLink = `${baseUrl}/?sessionId=${gameStore.currentSession.id}`;

        const onSuccess = () => {
            showCopied.value = true;
            setTimeout(() => (showCopied.value = false), 2000);
        };

        const fallbackCopy = () => {
            const textarea = document.createElement("textarea");
            textarea.value = inviteLink;
            textarea.style.position = "fixed";
            textarea.style.opacity = "0";
            document.body.appendChild(textarea);
            textarea.focus();
            textarea.select();
            try {
                document.execCommand("copy");
                onSuccess();
            } catch (err) {
                console.error("Impossible de copier le lien :", err);
            } finally {
                document.body.removeChild(textarea);
            }
        };

        if (navigator.clipboard && window.isSecureContext) {
            navigator.clipboard
                .writeText(inviteLink)
                .then(onSuccess)
                .catch(fallbackCopy);
        } else {
            fallbackCopy();
        }
    }
}

function onIssueSelected(issue: any) {
    console.log("Issue selected:", issue.title);
    if (gameStore.currentSession?.stories) {
        const storyIndex = gameStore.currentSession.stories.findIndex(
            (s) => s.id === issue.id,
        );
        if (storyIndex >= 0) {
            goToStoryAndVote(storyIndex);
        }
    }
}

async function viewIssueDetails(story: any) {
    viewingIssue.value = story;
    await nextTick();
    issueDetailModal.value?.openModal();
}

function revealAllVotes() {
    gameStore.revealVotes();
}

// Helpers
function hasVotedOnStory(storyId: string): boolean {
    if (!gameStore.currentSession || !gameStore.currentPlayer) return false;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    return !!storyVotes[gameStore.currentPlayer.id];
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

function isStoryRevealed(storyId: string): boolean {
    if (!gameStore.currentSession) return false;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    const requiredPlayers = gameStore.currentSession.requiredPlayers;
    return (
        requiredPlayers.length > 0 &&
        requiredPlayers.every((playerId) => storyVotes[playerId])
    );
}

function getStoryVotesCount(storyId: string): number {
    if (!gameStore.currentSession) return 0;
    const storyVotes = gameStore.currentSession.persistentVotes[storyId] || {};
    return Object.keys(storyVotes).length;
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

function makeOthersVote() {
    const votes = ["1", "2", "3", "5", "8", "13", "?", "☕"];
    if (!gameStore.currentSession || !gameStore.currentStory) return;

    const currentStoryVotes =
        gameStore.currentSession.persistentVotes[gameStore.currentStory.id] ||
        {};

    gameStore.currentSession.requiredPlayers.forEach((playerId) => {
        if (!currentStoryVotes[playerId]) {
            const randomVote = votes[Math.floor(Math.random() * votes.length)];
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

            const player = gameStore.players.find((p) => p.id === playerId);
            if (player) {
                player.hasVoted = true;
                player.vote = randomVote;
            }
        }
    });

    gameStore.savePersistedState();
    gameStore.updatePlayerVotedStatus();
}
</script>

<style scoped>
/* =============================================
   MOBILE STYLES
   ============================================= */

.mobile-layout {
    display: flex;
    flex-direction: column;
    height: 100dvh;
    overflow: hidden;
}

/* Mobile Top Bar */
.mobile-top-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0.75rem;
    background: linear-gradient(90deg, #3e2723, #5d4037);
    border-bottom: 2px solid #d4af37;
    flex-shrink: 0;
    min-height: 48px;
}

.mobile-top-left {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex: 1;
    min-width: 0;
}

.mobile-session-name {
    color: #ffd700;
    font-weight: bold;
    font-size: 0.9rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.mobile-top-right {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    flex-shrink: 0;
}

.mobile-player-count {
    color: #e0d8c3;
    font-size: 0.85rem;
}

.mobile-share-btn,
.mobile-logout-btn {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(212, 175, 55, 0.4);
    border-radius: 8px;
    padding: 0.4rem 0.6rem;
    font-size: 1rem;
    cursor: pointer;
    min-width: 40px;
    min-height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.2s;
}

.mobile-share-btn:active,
.mobile-logout-btn:active {
    background: rgba(255, 255, 255, 0.2);
}

.mobile-logout-btn {
    border-color: rgba(231, 76, 60, 0.4);
}

/* Mobile Tab Navigation */
.mobile-tabs {
    display: flex;
    background: rgba(44, 26, 18, 0.95);
    border-bottom: 2px solid #d4af37;
    flex-shrink: 0;
}

.mobile-tab {
    flex: 1;
    padding: 0.75rem 0.5rem;
    background: transparent;
    border: none;
    color: #a69b8d;
    font-family: "Cinzel", serif;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    text-align: center;
    position: relative;
    min-height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.mobile-tab.active {
    color: #ffd700;
    background: rgba(212, 175, 55, 0.1);
}

.mobile-tab.active::after {
    content: "";
    position: absolute;
    bottom: -2px;
    left: 10%;
    right: 10%;
    height: 3px;
    background: #ffd700;
    border-radius: 3px 3px 0 0;
}

.mobile-tab:active {
    background: rgba(212, 175, 55, 0.15);
}

/* Mobile Content Area */
.mobile-content {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
    -webkit-overflow-scrolling: touch;
}

.mobile-tab-panel {
    padding: 1rem;
    min-height: 100%;
}

/* Mobile No Session */
.mobile-no-session,
.mobile-no-issues {
    text-align: center;
    padding: 2rem 1rem;
    background: rgba(149, 165, 166, 0.1);
    border: 2px solid rgba(149, 165, 166, 0.3);
    border-radius: 12px;
    margin-bottom: 1rem;
}

.mobile-no-session h3 {
    color: #a69b8d;
    margin-bottom: 0.5rem;
}

.mobile-no-session p,
.mobile-no-issues p {
    color: #e0d8c3;
    font-size: 0.9rem;
    line-height: 1.4;
}

/* Mobile JIRA Section */
.mobile-jira-section {
    margin-bottom: 1rem;
}

/* Mobile Issues Section */
.mobile-issues-section {
    margin-bottom: 1rem;
}

.mobile-issues-section h4 {
    color: #ffd700;
    font-size: 1rem;
    margin-bottom: 0.75rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid rgba(212, 175, 55, 0.3);
}

.mobile-issues-list {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.mobile-issue-item {
    background: rgba(44, 26, 18, 0.8);
    border: 2px solid rgba(93, 64, 55, 0.6);
    border-radius: 10px;
    padding: 0.75rem;
    cursor: pointer;
    transition: all 0.2s;
    -webkit-tap-highlight-color: transparent;
}

.mobile-issue-item:active {
    transform: scale(0.98);
}

.mobile-issue-item.active {
    border-color: #d4af37;
    background: rgba(212, 175, 55, 0.1);
    box-shadow: 0 0 12px rgba(212, 175, 55, 0.15);
}

.mobile-issue-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.35rem;
}

.mobile-issue-number {
    background: #5d4037;
    color: #ffd700;
    font-weight: bold;
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 6px;
    min-width: 1.5rem;
    text-align: center;
}

.mobile-issue-key {
    color: #64b5f6;
    font-weight: bold;
    font-size: 0.8rem;
    flex: 1;
}

.mobile-issue-badges {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-shrink: 0;
}

.mobile-view-btn {
    background: rgba(212, 175, 55, 0.15);
    border: 1px solid rgba(212, 175, 55, 0.3);
    border-radius: 6px;
    padding: 0.3rem 0.5rem;
    font-size: 0.9rem;
    cursor: pointer;
    min-width: 36px;
    min-height: 36px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.2s;
}

.mobile-view-btn:active {
    background: rgba(212, 175, 55, 0.3);
}

.mobile-voted {
    font-size: 1rem;
}

.mobile-pending {
    font-size: 1rem;
    opacity: 0.5;
}

.mobile-issue-title {
    color: #e0d8c3;
    font-size: 0.85rem;
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

/* Mobile Character Section */
.mobile-character-section {
    margin-bottom: 1rem;
}

.mobile-character-section h4 {
    color: #ffd700;
    font-size: 0.95rem;
    margin-bottom: 0.75rem;
}

.mobile-character-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.5rem;
}

.mobile-character-btn {
    background: rgba(93, 64, 55, 0.5);
    border: 2px solid rgba(93, 64, 55, 0.8);
    border-radius: 10px;
    padding: 0.5rem 0.25rem;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.2rem;
    transition: all 0.2s;
    min-height: 60px;
    -webkit-tap-highlight-color: transparent;
}

.mobile-character-btn.active {
    background: rgba(212, 175, 55, 0.2);
    border-color: #d4af37;
    box-shadow: 0 0 8px rgba(212, 175, 55, 0.3);
}

.mobile-character-btn:active {
    transform: scale(0.95);
}

.mobile-char-emoji {
    font-size: 1.4rem;
}

.mobile-char-name {
    font-size: 0.6rem;
    color: #e0d8c3;
    font-weight: 600;
    text-align: center;
}

/* Mobile Current Story */
.mobile-current-story {
    margin-bottom: 1rem;
    background: rgba(44, 26, 18, 0.8);
    border: 2px solid #d4af37;
    border-radius: 12px;
    overflow: hidden;
}

.mobile-story-nav {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    padding: 0.5rem;
    background: rgba(212, 175, 55, 0.1);
    border-bottom: 1px solid rgba(212, 175, 55, 0.3);
}

.mobile-nav-btn {
    background: rgba(93, 64, 55, 0.8);
    border: 2px solid #d4af37;
    color: #ffd700;
    border-radius: 8px;
    width: 44px;
    height: 44px;
    font-size: 1.1rem;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
}

.mobile-nav-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
}

.mobile-nav-btn:active:not(:disabled) {
    transform: scale(0.9);
}

.mobile-story-counter {
    color: #ffd700;
    font-weight: bold;
    font-size: 1rem;
    min-width: 60px;
    text-align: center;
}

.mobile-story-info {
    padding: 0.75rem;
}

.mobile-story-key {
    display: inline-block;
    background: rgba(100, 181, 246, 0.15);
    color: #64b5f6;
    padding: 0.2rem 0.6rem;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: bold;
    border: 1px solid rgba(100, 181, 246, 0.3);
    margin-bottom: 0.5rem;
}

.mobile-story-info h3 {
    color: #e0d8c3;
    font-size: 1rem;
    line-height: 1.3;
    margin: 0 0 0.5rem 0;
}

.mobile-detail-btn {
    background: rgba(212, 175, 55, 0.15);
    border: 1px solid rgba(212, 175, 55, 0.4);
    color: #ffd700;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.85rem;
    font-family: "Cinzel", serif;
    transition: background 0.2s;
    min-height: 44px;
}

.mobile-detail-btn:active {
    background: rgba(212, 175, 55, 0.3);
}

/* Mobile Cards Section */
.mobile-cards-section {
    margin-bottom: 1rem;
}

/* Mobile Controls */
.mobile-controls {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.mobile-reveal-btn {
    width: 100%;
    padding: 0.85rem 1rem !important;
    font-size: 0.9rem !important;
}

.mobile-test-btn {
    width: 100%;
    background: linear-gradient(145deg, #e74c3c, #c0392b) !important;
    border-color: #e74c3c !important;
    padding: 0.85rem 1rem !important;
    font-size: 0.85rem !important;
}

/* Mobile Summary Stats */
.mobile-summary-stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 0.75rem;
    margin-bottom: 1.25rem;
}

.mobile-stat {
    background: rgba(44, 26, 18, 0.8);
    border: 2px solid rgba(93, 64, 55, 0.6);
    border-radius: 12px;
    padding: 0.75rem 0.5rem;
    text-align: center;
}

.mobile-stat-value {
    display: block;
    color: #ffd700;
    font-size: 1.5rem;
    font-weight: bold;
    margin-bottom: 0.25rem;
}

.mobile-stat-label {
    display: block;
    color: #a69b8d;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

/* Mobile Summary List */
.mobile-summary-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
}

.mobile-summary-item {
    background: rgba(44, 26, 18, 0.8);
    border: 2px solid rgba(93, 64, 55, 0.6);
    border-radius: 10px;
    padding: 0.75rem;
    transition: border-color 0.2s;
}

.mobile-summary-item.revealed {
    border-color: rgba(46, 204, 113, 0.5);
    background: rgba(46, 204, 113, 0.05);
}

.mobile-summary-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.mobile-summary-number {
    background: #5d4037;
    color: #ffd700;
    font-weight: bold;
    font-size: 0.75rem;
    padding: 0.2rem 0.5rem;
    border-radius: 50%;
    min-width: 1.5rem;
    height: 1.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}

.mobile-summary-title {
    flex: 1;
    color: #e0d8c3;
    font-size: 0.85rem;
    font-weight: bold;
    line-height: 1.3;
}

.mobile-status-done {
    font-size: 1rem;
    flex-shrink: 0;
}

.mobile-status-pending {
    color: #f39c12;
    font-size: 0.8rem;
    flex-shrink: 0;
    white-space: nowrap;
}

.mobile-summary-votes {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid rgba(93, 64, 55, 0.5);
}

.mobile-votes-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
}

.mobile-vote-chip {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    background: rgba(100, 181, 246, 0.1);
    border: 1px solid rgba(100, 181, 246, 0.3);
    border-radius: 8px;
    padding: 0.3rem 0.6rem;
}

.mobile-vote-player {
    color: #e0d8c3;
    font-size: 0.8rem;
}

.mobile-vote-value {
    color: #64b5f6;
    font-weight: bold;
    font-size: 0.85rem;
    background: rgba(100, 181, 246, 0.1);
    padding: 0.15rem 0.4rem;
    border-radius: 4px;
}

.mobile-story-avg {
    text-align: center;
    color: #f39c12;
    font-weight: bold;
    padding: 0.5rem;
    background: rgba(243, 156, 18, 0.1);
    border: 1px solid rgba(243, 156, 18, 0.3);
    border-radius: 8px;
    font-size: 0.9rem;
}
</style>
