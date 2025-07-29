<template>
    <div class="jira-import" v-if="gameStore.currentSession">
        <h3>📋 JIRA Integration</h3>
        
        <div class="jira-connect" v-if="!isConnected">
            <button @click="connectToJira" class="wow-button connect-btn">
                🔗 Connect to JIRA
            </button>
            <p class="connect-info">
                Connect your JIRA account to import tickets for estimation
            </p>
        </div>

        <div class="jira-connected" v-else>
            <div class="connected-status">
                <span class="status-badge connected">✅ Connected to JIRA</span>
                <button @click="disconnect" class="disconnect-btn">Disconnect</button>
            </div>
            
            <div class="search-form">
                <input
                    v-model="searchQuery"
                    type="text"
                    placeholder="Search JIRA issues..."
                    class="search-input"
                    @keyup.enter="searchIssues"
                />
                <button @click="searchIssues" class="wow-button search-btn" :disabled="isLoading">
                    {{ isLoading ? 'Searching...' : '🔍 Search' }}
                </button>
            </div>
            
            <div v-if="searchError" class="search-error">
                <p>❌ {{ searchError }}</p>
            </div>
            
            <!-- Search Results Section -->
            <div v-if="searchResults.length > 0" class="search-results">
                <h4>🔍 Search Results</h4>
                <div class="results-list">
                    <div v-for="issue in searchResults" :key="issue.id" class="issue-card">
                        <div class="issue-content">
                            <div class="issue-title">{{ issue.title }}</div>
                            <div class="issue-actions">
                                <button @click="showIssueModal(issue)" class="view-btn">
                                    Import
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Issues Overview Section (Always visible if there are imported stories, independent of JIRA connection) -->
        <div v-if="importedIssues.length > 0" class="issues-overview">
            <h4>📋 Issues Overview</h4>
            <div class="issues-list">
                <div v-for="issue in importedIssues" :key="issue.uniqueKey" class="issue-card">
                    <div class="issue-content">
                        <div class="issue-title">{{ issue.title }}</div>
                        <div class="issue-actions">
                            <div class="voting-indicator">
                                <span v-if="issue.hasVoted" class="vote-status voted" title="You have voted">✅</span>
                                <span v-else class="vote-status not-voted" title="Not voted yet">⭕</span>
                            </div>
                            <button @click="selectIssueForVoting(issue)" class="select-btn" :class="{ active: isSelectedForVoting(issue) }">
                                {{ isSelectedForVoting(issue) ? 'Selected' : 'Select' }}
                            </button>
                            <button @click="viewIssue(issue)" class="view-btn">
                                View
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- Quest Modal -->
        <JiraIssueModal 
            ref="issueModal" 
            :issue="selectedIssue" 
            :mode="modalMode"
            @import="handleIssueImport"
        />
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useGameStore } from '@/stores/gameStore';
import SimpleJiraAuth from '@/auth/SimpleJiraAuth';
import JiraIssueModal from './JiraIssueModal.vue';

const gameStore = useGameStore();
const isConnected = ref(false);
const isLoading = ref(false);
const searchQuery = ref('');
const searchResults = ref<any[]>([]);
const searchError = ref<string>('');
const currentUser = ref<any>(null);
const selectedIssue = ref<any>(null);
const modalMode = ref<string>('view');
const issueModal = ref<InstanceType<typeof JiraIssueModal>>();
const jiraAuth = new SimpleJiraAuth();
const selectedIssueForVoting = ref<any>(null);

// Computed property to show imported stories (always visible regardless of JIRA connection)
const importedIssues = computed(() => {
    const issues = [];
    
    // Show imported stories from the session (from localStorage)
    gameStore.currentSession?.stories?.forEach(story => {
        issues.push({
            ...story,
            uniqueKey: `story_${story.id}`,
            hasVoted: false // TODO: Connect to actual voting system
        });
    });
    
    return issues;
});

// Keep the allIssues for backward compatibility (deprecated)
const allIssues = computed(() => {
    return importedIssues.value;
});

function handleIssueAction(issue: any) {
    if (issue.isImported) {
        // View imported issue
        viewImportedStory(issue);
    } else {
        // Import new issue
        showIssueModal(issue);
    }
}

async function connectToJira() {
    if (!gameStore.currentSession?.id) {
        console.error('No session ID available');
        return;
    }
    
    try {
        const userData = await jiraAuth.authenticate(gameStore.currentSession.id);
        currentUser.value = userData;
        isConnected.value = true;
        console.log('JIRA authentication successful:', userData);
    } catch (error) {
        console.error('JIRA authentication failed:', error);
        // You might want to show a user-friendly error message here
    }
}

function disconnect() {
    isConnected.value = false;
    currentUser.value = null;
    searchResults.value = [];
    searchQuery.value = '';
    searchError.value = '';
}

async function searchIssues() {
    if (!searchQuery.value.trim() || !gameStore.currentSession?.id) return;
    
    isLoading.value = true;
    searchError.value = '';
    searchResults.value = [];
    
    try {
        const issues = await jiraAuth.searchIssues(gameStore.currentSession.id, searchQuery.value);
        searchResults.value = issues;
        console.log('Search successful, found', issues.length, 'issues');
    } catch (error) {
        console.error('Search failed:', error);
        searchError.value = error instanceof Error ? error.message : 'Search failed. Please try again.';
    } finally {
        isLoading.value = false;
    }
}

async function checkConnectionStatus() {
    if (!gameStore.currentSession?.id) return;
    
    try {
        const connected = await jiraAuth.isAuthenticated(gameStore.currentSession.id);
        isConnected.value = connected;
    } catch (error) {
        console.error('Failed to check connection status:', error);
    }
}

async function importIssue(issue: any) {
    try {
        const story = {
            id: issue.id,
            title: issue.title,
            description: issue.description || '',
            jiraKey: issue.jiraKey
        };
        
        gameStore.addStoriesToSession([story]);
        
        // Remove from search results
        searchResults.value = searchResults.value.filter(i => i.jiraKey !== issue.jiraKey);
    } catch (error) {
        console.error('Import failed:', error);
    }
}

function showIssueModal(issue: any) {
    selectedIssue.value = issue;
    modalMode.value = 'import';
    issueModal.value?.openModal();
}

function viewImportedStory(story: any) {
    selectedIssue.value = story;
    modalMode.value = 'view';
    issueModal.value?.openModal();
}

function handleIssueImport(issue: any) {
    importIssue(issue);
}

function viewIssue(issue: any) {
    selectedIssue.value = issue;
    modalMode.value = 'view';
    issueModal.value?.openModal();
}

// Issue selection functions
function selectIssueForVoting(issue: any) {
    selectedIssueForVoting.value = issue;
    // Emit event to parent component (TavernView) so it can update the voting UI
    emit('issue-selected', issue);
    console.log('Selected issue for voting:', issue.title);
}

function isSelectedForVoting(issue: any): boolean {
    return selectedIssueForVoting.value?.id === issue.id;
}

// Define emits
const emit = defineEmits<{
    'issue-selected': [issue: any];
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
    min-height: 0; /* Allow shrinking */
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
    text-align: center;
    line-height: 1.3;
}

.search-results {
    max-height: 200px;
    overflow-y: auto;
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.05);
    margin-bottom: 0.75rem;
}

.results-list {
    max-height: 200px;
    overflow-y: auto;
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.05);
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
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.05);
}

.issue-card {
    padding: 0.75rem;
    border-bottom: 1px solid rgba(139, 69, 19, 0.3);
    transition: background-color 0.2s ease;
}

.issue-card:last-child {
    border-bottom: none;
}

.issue-card:hover {
    background: rgba(255, 255, 255, 0.08);
}

.issue-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
}

.issue-title {
    color: #f4f4f4;
    font-size: 0.85rem;
    line-height: 1.3;
    flex: 1;
    word-wrap: break-word;
    overflow-wrap: break-word;
    margin: 0;
    padding: 0;
}

.issue-actions {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex-shrink: 0;
}

.voting-indicator {
    display: flex;
    align-items: center;
}

.vote-status {
    font-size: 1.2rem;
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
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    font-size: 0.75rem;
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
    padding: 0.4rem 0.8rem;
    border-radius: 4px;
    font-size: 0.75rem;
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
    0%, 100% { transform: scale(1); opacity: 0.8; }
    50% { transform: scale(1.1); opacity: 1; }
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
</style>
