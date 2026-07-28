<template>
    <div class="jira-import" v-if="gameStore.currentSession">
        <h3>📋 Intégration Jira</h3>

        <div class="jira-connect" v-if="!isConnected">
            <button @click="connectToJira" class="wow-button connect-btn">
                🔗 Se connecter à Jira
            </button>
            <p class="connect-info">
                Connectez votre compte Jira pour importer les tickets à estimer
            </p>
        </div>

        <div class="jira-connected" v-else>
            <div class="connected-status">
                <span class="status-badge connected">✅ Connecté à Jira</span>
                <button @click="disconnect" class="disconnect-btn">
                    Se déconnecter
                </button>
            </div>

            <div class="search-form">
                <input
                    v-model="searchQuery"
                    type="text"
                    placeholder="Rechercher des tickets Jira…"
                    class="search-input"
                    @keyup.enter="searchIssues"
                />
                <button
                    @click="searchIssues"
                    class="wow-button search-btn"
                    :disabled="isLoading"
                >
                    {{ isLoading ? "Recherche…" : "🔍 Rechercher" }}
                </button>
            </div>

            <div v-if="searchError" class="search-error">
                <p>❌ {{ searchError }}</p>
            </div>

            <!-- Search Results Section -->
            <div v-if="searchResults.length > 0" class="search-results">
                <h4>🔍 Résultats de recherche</h4>
                <div class="results-list">
                    <div
                        v-for="issue in searchResults"
                        :key="issue.id"
                        class="issue-card"
                    >
                        <div class="issue-title">{{ issue.title }}</div>
                        <div class="issue-actions">
                            <button @click="viewIssue(issue)" class="view-btn">
                                Voir
                            </button>
                            <button
                                @click="importIssueDirectly(issue)"
                                class="import-btn"
                            >
                                Importer
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Issues Overview Section (Always visible if there are imported stories, independent of JIRA connection) -->
        <div
            v-if="props.showIssuesOverview && importedIssues.length > 0"
            class="issues-overview"
        >
            <h4>📋 Vue d’ensemble des tickets</h4>
            <div class="issues-list">
                <div
                    v-for="issue in importedIssues"
                    :key="issue.uniqueKey"
                    class="issue-card"
                >
                    <div class="issue-title">{{ issue.title }}</div>
                    <div class="issue-actions">
                        <div class="voting-indicator">
                            <span
                                v-if="issue.hasVoted"
                                class="vote-status voted"
                                title="Vous avez voté"
                                >✅</span
                            >
                            <span
                                v-else
                                class="vote-status not-voted"
                                title="Pas encore voté"
                                >⭕</span
                            >
                        </div>
                        <button
                            @click="selectIssueForVoting(issue)"
                            class="select-btn"
                            :class="{ active: isSelectedForVoting(issue) }"
                        >
                            {{
                                isSelectedForVoting(issue)
                                    ? "Sélectionné"
                                    : "Sélectionner"
                            }}
                        </button>
                        <button @click="viewIssue(issue)" class="view-btn">
                            Voir
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Issue Details Modal -->
        <JiraIssueModal
            ref="issueModal"
            :issue="selectedIssue"
            :mode="modalMode"
            @import="handleIssueImport"
        />
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, nextTick } from "vue";
import { useGameStore } from "@/stores/gameStore";
import SimpleJiraAuth from "@/auth/SimpleJiraAuth";
import JiraIssueModal from "./JiraIssueModal.vue";
import { getApiUrl } from "@/config/api";

const props = withDefaults(
    defineProps<{
        showIssuesOverview?: boolean;
    }>(),
    {
        showIssuesOverview: true,
    },
);

const gameStore = useGameStore();
const isConnected = ref(false);
const isLoading = ref(false);
const searchQuery = ref(
    'project=IMMO AND Sprint="POKER" AND status="To Do" AND "Story point estimate"=0 ORDER BY rank ASC',
);
const searchResults = ref<any[]>([]);
const searchError = ref<string>("");
const currentUser = ref<any>(null);
const jiraAuth = new SimpleJiraAuth();
const selectedIssueForVoting = ref<any>(null);
const selectedIssue = ref<any>(null);
const modalMode = ref<"import" | "view">("view");
const issueModal = ref<InstanceType<typeof JiraIssueModal>>();

// Computed property to show imported stories (always visible regardless of JIRA connection)
const importedIssues = computed(() => {
    const issues: any[] = [];

    // Show imported stories from the session (from localStorage)
    gameStore.currentSession?.stories?.forEach((story) => {
        // Check if the current player has voted on this story
        const playerId = gameStore.currentPlayer?.id;
        const hasVoted =
            playerId !== undefined &&
            gameStore.currentSession?.persistentVotes?.[story.id]?.[playerId] !== undefined;

        issues.push({
            ...story,
            uniqueKey: `story_${story.id}`,
            hasVoted: hasVoted,
        });
    });

    return issues;
});

// Keep the allIssues for backward compatibility (deprecated)
const allIssues = computed(() => {
    return importedIssues.value;
});

async function connectToJira() {
    if (!gameStore.currentSession?.id) {
        console.error("No session ID available");
        return;
    }

    try {
        const userData = await jiraAuth.authenticate(
            gameStore.currentSession.id,
        );
        currentUser.value = userData;
        isConnected.value = true;
        console.log("JIRA authentication successful:", userData);
    } catch (error) {
        console.error("JIRA authentication failed:", error);
        // You might want to show a user-friendly error message here
    }
}

function disconnect() {
    isConnected.value = false;
    currentUser.value = null;
    searchResults.value = [];
    searchQuery.value = "";
    searchError.value = "";
}

async function searchIssues() {
    if (!searchQuery.value.trim() || !gameStore.currentSession?.id) return;

    isLoading.value = true;
    searchError.value = "";
    searchResults.value = [];

    try {
        const issues = await jiraAuth.searchIssues(
            gameStore.currentSession.id,
            searchQuery.value,
        );
        searchResults.value = issues;
        console.log("Search successful, found", issues.length, "issues");
    } catch (error) {
        console.error("Search failed:", error);
        searchError.value =
            error instanceof Error
                ? error.message
                : "La recherche a échoué. Veuillez réessayer.";
    } finally {
        isLoading.value = false;
    }
}

async function checkConnectionStatus() {
    try {
        // First try API token mode (no session needed)
        let connected = await jiraAuth.isAuthenticated();

        // If not connected via API token, try OAuth mode with session
        if (!connected && gameStore.currentSession?.id) {
            connected = await jiraAuth.isAuthenticated(
                gameStore.currentSession.id,
            );
        }

        isConnected.value = connected;
        console.log("JIRA connection status:", connected);

        // Auto-load POKER tickets if connected
        if (connected && searchResults.value.length === 0) {
            await searchIssues();
        }
    } catch (error) {
        console.error("Failed to check connection status:", error);
    }
}

async function importIssue(issue: any) {
    if (!gameStore.currentSession?.id) {
        console.error("No session ID available");
        return;
    }

    try {
        const story = {
            id: issue.id,
            title: issue.title,
            description: issue.description || "",
            jiraKey: issue.jiraKey,
        };

        // Call backend API to import and broadcast to all users
        const response = await fetch(getApiUrl("/jira/import-issues"), {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({
                sessionId: gameStore.currentSession.id,
                stories: [story],
            }),
        });

        if (!response.ok) {
            throw new Error(`Failed to import issue: ${response.status}`);
        }

        const result = await response.json();
        console.log("Issue imported successfully:", result);

        // Optimistically add to local store to ensure UI updates immediately
        // The backend will also broadcast via WS, but gameStore now handles duplicates safely
        gameStore.addStoriesToSession([story]);

        // Remove from search results
        searchResults.value = searchResults.value.filter(
            (i) => i.jiraKey !== issue.jiraKey,
        );
    } catch (error) {
        console.error("Import failed:", error);
        // Fallback to local addition if API call fails
        const story = {
            id: issue.id,
            title: issue.title,
            description: issue.description || "",
            jiraKey: issue.jiraKey,
        };
        gameStore.addStoriesToSession([story]);
    }
}

// Simple import function that directly imports the issue without modal
function importIssueDirectly(issue: any) {
    importIssue(issue);
}

// View function that shows issue details in a modal
async function viewIssue(issue: any) {
    selectedIssue.value = issue;
    modalMode.value = "view";

    // Wait for DOM update to ensure ref is bound
    await nextTick();

    if (issueModal.value?.openModal) {
        issueModal.value.openModal();
    }
}

// Show modal function for import
async function showIssueModal(issue: any) {
    selectedIssue.value = issue;
    modalMode.value = "import";
    await nextTick();
    issueModal.value?.openModal();
}

// Issue selection functions
function selectIssueForVoting(issue: any) {
    selectedIssueForVoting.value = issue;
    // Emit event to parent component (TavernView) so it can update the voting UI
    emit("issue-selected", issue);
    console.log("Selected issue for voting:", issue.title);
}

function isSelectedForVoting(issue: any): boolean {
    return selectedIssueForVoting.value?.id === issue.id;
}

// Handle import from modal
function handleIssueImport(issue: any) {
    importIssue(issue);
}

// Define emits
const emit = defineEmits<{
    "issue-selected": [issue: any];
}>();

// Check connection status on mount
onMounted(() => {
    checkConnectionStatus();
});
</script>

<style scoped>
.jira-import {
    background: rgba(44, 62, 80, 0.8);
    border: 2px solid #8b4513;
    border-radius: 8px;
    padding: 0.75rem;
    margin-bottom: 1rem;
}

.jira-import h3 {
    color: #ffd700;
    margin-bottom: 0.75rem;
    font-size: 1rem;
}

.jira-connect {
    text-align: center;
    padding: 0.75rem;
}

.connect-btn {
    background: linear-gradient(145deg, #2c5f2d, #4a7c59);
    border-color: #4a7c59;
    margin-bottom: 0.75rem;
    font-size: 0.9rem;
    padding: 0.5rem 0.75rem;
    width: 100%;
}

.connect-btn:hover {
    background: linear-gradient(145deg, #4a7c59, #5a9e68);
}

.connect-info {
    color: #bbb;
    font-size: 0.8rem;
    margin: 0;
    line-height: 1.3;
}

.connected-status {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    align-items: stretch;
}

.status-badge {
    padding: 0.25rem 0.5rem;
    border-radius: 12px;
    font-size: 0.75rem;
    font-weight: bold;
    text-align: center;
}

.status-badge.connected {
    background: rgba(46, 204, 113, 0.2);
    color: #2ecc71;
    border: 1px solid #2ecc71;
}

.disconnect-btn {
    background: none;
    border: none;
    color: #dc3545;
    cursor: pointer;
    font-size: 0.75rem;
    text-decoration: underline;
    padding: 0.25rem;
    align-self: center;
}

.disconnect-btn:hover {
    color: #c82333;
}

.search-form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
}

.search-input {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.9);
    color: #333;
    font-size: 0.85rem;
    box-sizing: border-box;
}

.search-input:focus {
    outline: none;
    border-color: #daa520;
    box-shadow: 0 0 0 2px rgba(218, 165, 32, 0.2);
}

.search-btn {
    background: linear-gradient(145deg, #3498db, #2980b9);
    border-color: #3498db;
    width: 100%;
    padding: 0.5rem;
    font-size: 0.85rem;
    border-radius: 4px;
    color: white;
    cursor: pointer;
    border: 1px solid #3498db;
    transition: all 0.3s ease;
}

.search-btn:hover:not(:disabled) {
    background: linear-gradient(145deg, #2980b9, #21618c);
}

.search-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}

.search-error {
    background: rgba(231, 76, 60, 0.2);
    border: 1px solid #e74c3c;
    border-radius: 4px;
    padding: 0.5rem;
    margin-bottom: 0.75rem;
}

.search-error p {
    margin: 0;
    color: #e74c3c;
    font-size: 0.8rem;
    margin-bottom: 0.75rem;
}

.search-results {
    /* Removed fixed max-height to let it grow if needed, or set a larger one */
    max-height: 400px;
    overflow-y: auto;
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.05);
    margin-bottom: 0.75rem;
}

.results-list {
    /* Let it fill the container */
    overflow-y: auto;
}

.search-results h4,
.issues-overview h4 {
    color: #ffd700;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
}

.issues-overview {
    margin-top: 0.75rem;
    background: rgba(52, 73, 94, 0.6);
    border: 1px solid #8b4513;
    border-radius: 6px;
    padding: 0.75rem;
}

.issues-overview h4 {
    color: #ffd700;
    margin: 0 0 0.75rem 0;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    text-align: left;
}

.issue-item {
    padding: 0.5rem;
    border-bottom: 1px solid rgba(139, 69, 19, 0.3);
    position: relative;
}

.issue-item:last-child {
    border-bottom: none;
}

.issue-header {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    flex-wrap: wrap;
}

.issue-key {
    font-weight: bold;
    color: #3498db;
    font-size: 0.8rem;
}

.issue-type {
    background: rgba(52, 152, 219, 0.2);
    color: #3498db;
    padding: 0.15rem 0.4rem;
    border-radius: 6px;
    font-size: 0.65rem;
    text-transform: uppercase;
}

.issue-title {
    color: #f4f4f4;
    font-size: 0.8rem;
    margin-bottom: 0.5rem;
    line-height: 1.2;
    padding-right: 50px; /* Space for import button */
    word-wrap: break-word;
    overflow-wrap: break-word;
}

.view-quest-btn {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: linear-gradient(145deg, #8b6914, #daa520);
    border: 1px solid #daa520;
    color: #2c1810;
    padding: 0.2rem 0.4rem;
    border-radius: 3px;
    font-size: 0.7rem;
    cursor: pointer;
    transition: all 0.3s ease;
    min-width: 60px;
    font-weight: bold;
    text-shadow: 1px 1px 2px rgba(255, 255, 255, 0.3);
}

.view-quest-btn:hover {
    background: linear-gradient(145deg, #daa520, #ffd700);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.issues-list {
    max-height: 250px;
    overflow-y: auto;
    overflow-x: hidden; /* Prevent horizontal scrollbar */
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.05);
}

.issue-card {
    padding: 0.75rem;
    border-bottom: 1px solid rgba(139, 69, 19, 0.3);
    transition: background-color 0.2s ease;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
}

.issue-card:last-child {
    border-bottom: none;
}

.issue-card:hover {
    background: rgba(255, 255, 255, 0.08);
}

.issue-title {
    color: #f4f4f4;
    font-size: 0.85rem;
    line-height: 1.3;
    word-wrap: break-word;
    overflow-wrap: break-word;
    margin: 0;
    padding: 0;
    white-space: normal; /* Allow text to wrap */
}

.issue-actions {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    width: 100%;
    justify-content: flex-end;
}

.voting-indicator {
    display: flex;
    align-items: center;
    margin-right: 0.25rem;
}

.vote-status {
    font-size: 1rem;
    transition: transform 0.2s ease;
}

.vote-status.voted {
    opacity: 1;
}

.vote-status.not-voted {
    opacity: 0.6;
}

.view-btn {
    background: linear-gradient(145deg, #8b6914, #daa520);
    border: 1px solid #daa520;
    color: #2c1810;
    padding: 0.3rem 0.6rem; /* Slightly smaller padding */
    border-radius: 4px;
    font-size: 0.7rem; /* Slightly smaller font */
    cursor: pointer;
    transition: all 0.3s ease;
    font-weight: bold;
    text-shadow: 1px 1px 2px rgba(255, 255, 255, 0.3);
    white-space: nowrap;
}

.view-btn:hover {
    background: linear-gradient(145deg, #daa520, #ffd700);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.select-btn {
    background: linear-gradient(145deg, #3498db, #2980b9);
    border: 1px solid #3498db;
    color: white;
    padding: 0.3rem 0.6rem; /* Slightly smaller padding */
    border-radius: 4px;
    font-size: 0.7rem; /* Slightly smaller font */
    cursor: pointer;
    transition: all 0.3s ease;
    font-weight: bold;
    white-space: nowrap;
}

.select-btn:hover {
    background: linear-gradient(145deg, #2980b9, #21618c);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(52, 152, 219, 0.3);
}

.select-btn.active {
    background: linear-gradient(145deg, #27ae60, #229954);
    border-color: #2ecc71;
}

.select-btn.active:hover {
    background: linear-gradient(145deg, #229954, #1e8449);
    box-shadow: 0 2px 4px rgba(46, 204, 113, 0.3);
}

.issue-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
    flex-wrap: wrap;
}

.voting-status {
    margin-left: auto;
}

.status-icon {
    font-size: 1rem;
    display: inline-block;
    transition: transform 0.2s ease;
}

.status-icon.imported {
    opacity: 0.8;
}

.status-icon.available {
    animation: sparkle 2s ease-in-out infinite;
}

@keyframes sparkle {
    0%,
    100% {
        transform: scale(1);
        opacity: 0.8;
    }
    50% {
        transform: scale(1.1);
        opacity: 1;
    }
}

.action-btn {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: linear-gradient(145deg, #8b6914, #daa520);
    border: 1px solid #daa520;
    color: #2c1810;
    padding: 0.2rem 0.4rem;
    border-radius: 3px;
    font-size: 0.7rem;
    cursor: pointer;
    transition: all 0.3s ease;
    min-width: 70px;
    font-weight: bold;
    text-shadow: 1px 1px 2px rgba(255, 255, 255, 0.3);
}

.action-btn:hover {
    background: linear-gradient(145deg, #daa520, #ffd700);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

.action-btn.imported {
    background: linear-gradient(145deg, #4a7c59, #5a9e68);
    border-color: #5a9e68;
    color: white;
}

.action-btn.imported:hover {
    background: linear-gradient(145deg, #5a9e68, #6bb76c);
}

/* Scrollbar styling for webkit browsers */
.search-results::-webkit-scrollbar,
.results-list::-webkit-scrollbar,
.issues-list::-webkit-scrollbar {
    width: 6px;
}

.search-results::-webkit-scrollbar-track,
.results-list::-webkit-scrollbar-track,
.issues-list::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
}

.search-results::-webkit-scrollbar-thumb,
.results-list::-webkit-scrollbar-thumb,
.issues-list::-webkit-scrollbar-thumb {
    background: rgba(139, 69, 19, 0.6);
    border-radius: 3px;
}

.search-results::-webkit-scrollbar-thumb:hover,
.results-list::-webkit-scrollbar-thumb:hover,
.issues-list::-webkit-scrollbar-thumb:hover {
    background: rgba(139, 69, 19, 0.8);
}

/* Imported Stories Section */
.imported-stories {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid rgba(139, 69, 19, 0.3);
}

.imported-stories h4 {
    color: #ffd700;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
}

.imported-stories-list {
    max-height: 150px;
    overflow-y: auto;
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.05);
}

.imported-story-item {
    padding: 0.5rem;
    border-bottom: 1px solid rgba(139, 69, 19, 0.3);
    position: relative;
}

.imported-story-item:last-child {
    border-bottom: none;
}

.imported-story-item .story-header {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    flex-wrap: wrap;
}

.imported-story-item .story-key {
    font-weight: bold;
    color: #3498db;
    font-size: 0.8rem;
}

.imported-story-item .story-title {
    color: #f4f4f4;
    font-size: 0.8rem;
    margin-bottom: 0.5rem;
    line-height: 1.2;
    padding-right: 50px; /* Space for view button */
    word-wrap: break-word;
    overflow-wrap: break-word;
}

.view-story-btn {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: linear-gradient(145deg, #8b6914, #daa520);
    border: 1px solid #daa520;
    color: #2c1810;
    padding: 0.2rem 0.4rem;
    border-radius: 3px;
    font-size: 0.7rem;
    cursor: pointer;
    transition: all 0.3s ease;
    min-width: 60px;
    font-weight: bold;
    text-shadow: 1px 1px 2px rgba(255, 255, 255, 0.3);
}

.view-story-btn:hover {
    background: linear-gradient(145deg, #daa520, #ffd700);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
}

/* Scrollbar styling for imported stories */
.imported-stories-list::-webkit-scrollbar {
    width: 6px;
}

.imported-stories-list::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
}

.imported-stories-list::-webkit-scrollbar-thumb {
    background: rgba(139, 69, 19, 0.6);
    border-radius: 3px;
}

.imported-stories-list::-webkit-scrollbar-thumb:hover {
    background: rgba(139, 69, 19, 0.8);
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
    backdrop-filter: blur(2px);
}

.modal-content {
    background: linear-gradient(145deg, #2c2c54, #2c3e50);
    border: 3px solid #8b4513;
    border-radius: 12px;
    padding: 2rem;
    max-width: 600px;
    max-height: 80vh;
    overflow-y: auto;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6);
    animation: modalAppear 0.3s ease-out;
    min-width: 400px;
}

@keyframes modalAppear {
    from {
        opacity: 0;
        transform: scale(0.9) translateY(-20px);
    }
    to {
        opacity: 1;
        transform: scale(1) translateY(0);
    }
}

.modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    border-bottom: 2px solid #daa520;
    padding-bottom: 1rem;
}

.modal-header h2 {
    color: #ffd700;
    margin: 0;
    font-size: 1.4rem;
    text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
}

.close-btn {
    background: none;
    border: none;
    color: #ecf0f1;
    font-size: 1.5rem;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 50%;
    transition: all 0.3s ease;
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
}

.close-btn:hover {
    background: rgba(231, 76, 60, 0.2);
    color: #e74c3c;
    transform: scale(1.1);
}

.modal-body {
    color: #ecf0f1;
}

.issue-detail-section {
    margin-bottom: 1.5rem;
}

.issue-detail-section h3 {
    color: #ecf0f1;
    margin: 0 0 1rem 0;
    font-size: 1.2rem;
    line-height: 1.4;
    word-wrap: break-word;
}

.issue-meta {
    margin-bottom: 1rem;
}

.jira-key {
    background: rgba(52, 152, 219, 0.2);
    color: #3498db;
    padding: 0.3rem 0.6rem;
    border-radius: 6px;
    font-size: 0.9rem;
    font-weight: bold;
    border: 1px solid #3498db;
    display: inline-block;
}

.issue-description {
    background: rgba(52, 73, 94, 0.3);
    border: 1px solid #7f8c8d;
    border-radius: 8px;
    padding: 1rem;
}

.issue-description h4 {
    color: #3498db;
    margin: 0 0 0.75rem 0;
    font-size: 1rem;
}

.description-content {
    line-height: 1.6;
    color: #bdc3c7;
    font-size: 0.9rem;
}

.description-content p {
    margin: 0 0 1rem 0;
}

.description-content ul,
.description-content ol {
    margin: 0 0 1rem 1.5rem;
}

.description-content li {
    margin-bottom: 0.5rem;
}

.description-content a {
    color: #3498db;
    text-decoration: none;
}

.description-content a:hover {
    color: #5dade2;
    text-decoration: underline;
}

.no-description {
    color: #95a5a6;
    font-style: italic;
    text-align: center;
    margin: 1rem 0;
}

/* Import button styling for modal */
.import-btn {
    background: linear-gradient(145deg, #27ae60, #229954);
    border: 1px solid #2ecc71;
    color: white;
    padding: 0.3rem 0.6rem;
    border-radius: 4px;
    font-size: 0.7rem;
    cursor: pointer;
    transition: all 0.3s ease;
    font-weight: bold;
    white-space: nowrap;
}

.import-btn:hover {
    background: linear-gradient(145deg, #229954, #1e8449);
    transform: translateY(-1px);
    box-shadow: 0 2px 4px rgba(46, 204, 113, 0.3);
}

/* Make component more compact on smaller screens */
@media (max-width: 1200px) {
    .jira-import {
        padding: 0.5rem;
    }

    .jira-import h3 {
        font-size: 0.9rem;
        margin-bottom: 0.5rem;
    }

    .search-results,
    .imported-stories-list {
        max-height: 150px;
    }

    .issue-title,
    .imported-story-item .story-title {
        font-size: 0.75rem;
    }

    .issue-key,
    .imported-story-item .story-key {
        font-size: 0.75rem;
    }

    .imported-stories h4 {
        font-size: 0.85rem;
    }
}

/* Modern tavern skin — Jira requests and actions stay unchanged */
.jira-import {
    margin: 0;
    padding: 0;
    border: 0;
    background: transparent;
}

.jira-import h3 {
    margin-bottom: 0.8rem;
    color: #f1c17d;
    font-family: "Cinzel", serif;
    font-size: 0.86rem;
    letter-spacing: 0.06em;
    text-transform: uppercase;
}

.connected-status {
    gap: 0.4rem;
    margin-bottom: 0.8rem;
}

.status-badge {
    padding: 0.45rem 0.7rem;
    border-radius: 999rem;
    font-family: "Lato", sans-serif;
    font-size: 0.68rem;
    letter-spacing: 0.04em;
}

.status-badge.connected {
    border-color: rgba(105, 188, 126, 0.36);
    background: rgba(51, 117, 68, 0.16);
    color: #9bd1a8;
}

.disconnect-btn {
    color: rgba(239, 153, 128, 0.7);
    font-family: "Lato", sans-serif;
    text-decoration: none;
}

.search-form {
    gap: 0.65rem;
    margin-bottom: 1rem;
}

.search-input {
    min-height: 2.8rem;
    padding: 0.65rem 0.75rem;
    border: 1px solid rgba(255, 236, 207, 0.14);
    border-radius: 0.7rem;
    background: rgba(9, 6, 4, 0.5);
    color: #fff6e7;
    font-family: "Lato", sans-serif;
    font-size: 0.78rem;
}

.search-input:focus {
    border-color: rgba(232, 177, 103, 0.58);
    box-shadow: 0 0 0 0.18rem rgba(232, 177, 103, 0.08);
}

.search-btn {
    min-height: 2.7rem;
    border: 1px solid #d99955;
    border-radius: 0.7rem;
    background: linear-gradient(135deg, #d58a42, #a95629);
    color: #1b0e07;
    font-family: "Lato", sans-serif;
    font-size: 0.75rem;
    font-weight: 700;
}

.search-btn:hover:not(:disabled) {
    background: linear-gradient(135deg, #e29a50, #bb6330);
    box-shadow: 0 0.65rem 1.4rem rgba(171, 83, 34, 0.24);
    transform: translateY(-0.08rem);
}

.search-results,
.issues-list,
.imported-stories-list {
    border: 1px solid rgba(232, 177, 103, 0.14);
    border-radius: 0.8rem;
    background: rgba(8, 5, 3, 0.32);
}

.search-results {
    max-height: 31rem;
    padding: 0.65rem;
}

.search-results h4,
.issues-overview h4,
.imported-stories h4 {
    color: #e8b36e;
    font-family: "Cinzel", serif;
    font-size: 0.76rem;
    letter-spacing: 0.04em;
}

.issue-card,
.issue-item,
.imported-story-item {
    gap: 0.6rem;
    padding: 0.75rem 0.65rem;
    border-bottom: 1px solid rgba(255, 236, 207, 0.08);
}

.issue-card:hover {
    background: rgba(137, 77, 39, 0.18);
}

.issue-key,
.imported-story-item .story-key {
    color: #e5a75f;
    font-family: "Lato", sans-serif;
    font-size: 0.72rem;
}

.issue-title,
.imported-story-item .story-title {
    color: rgba(255, 248, 234, 0.76);
    font-family: "Lato", sans-serif;
    font-size: 0.76rem;
    line-height: 1.4;
}

.view-btn,
.select-btn,
.import-btn {
    min-height: 1.9rem;
    padding: 0.35rem 0.65rem;
    border-radius: 0.5rem;
    font-family: "Lato", sans-serif;
    font-size: 0.65rem;
    text-shadow: none;
}

.view-btn {
    border-color: rgba(232, 177, 103, 0.4);
    background: rgba(232, 177, 103, 0.1);
    color: #efbc77;
}

.view-btn:hover {
    background: rgba(232, 177, 103, 0.2);
}

.select-btn,
.import-btn {
    border-color: rgba(103, 181, 123, 0.42);
    background: rgba(56, 124, 73, 0.2);
    color: #a6d7b1;
}

.select-btn:hover,
.import-btn:hover {
    background: rgba(65, 148, 86, 0.34);
    box-shadow: none;
}

.search-results::-webkit-scrollbar,
.results-list::-webkit-scrollbar,
.issues-list::-webkit-scrollbar {
    width: 0.4rem;
}

.search-results::-webkit-scrollbar-thumb,
.results-list::-webkit-scrollbar-thumb,
.issues-list::-webkit-scrollbar-thumb {
    border-radius: 999rem;
    background: rgba(217, 153, 80, 0.36);
}
</style>
