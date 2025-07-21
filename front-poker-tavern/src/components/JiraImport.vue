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

        <div class="jira-search" v-else>
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
            
            <div v-if="searchResults.length > 0" class="search-results">
                <div v-for="issue in searchResults" :key="issue.jiraKey" class="issue-item">
                    <div class="issue-header">
                        <span class="issue-key">{{ issue.jiraKey }}</span>
                        <span class="issue-type">Story</span>
                    </div>
                    <div class="issue-title">{{ issue.title }}</div>
                    <button @click="importIssue(issue)" class="import-single-btn">Import</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useGameStore } from '@/stores/gameStore';
import SimpleJiraAuth from '@/auth/SimpleJiraAuth';

const gameStore = useGameStore();
const isConnected = ref(false);
const isLoading = ref(false);
const searchQuery = ref('');
const searchResults = ref<any[]>([]);
const searchError = ref<string>('');
const currentUser = ref<any>(null);
const jiraAuth = new SimpleJiraAuth();

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
    padding: 1rem;
    margin-bottom: 1rem;
}

.jira-import h3 {
    color: #ffd700;
    margin-bottom: 1rem;
}

.jira-connect {
    text-align: center;
    padding: 1rem;
}

.connect-btn {
    background: linear-gradient(145deg, #2c5f2d, #4a7c59);
    border-color: #4a7c59;
    margin-bottom: 1rem;
}

.connect-btn:hover {
    background: linear-gradient(145deg, #4a7c59, #5a9e68);
}

.connect-info {
    color: #bbb;
    font-size: 0.9rem;
    margin: 0;
}

.connected-status {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
}

.status-badge {
    padding: 0.25rem 0.5rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: bold;
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
    font-size: 0.8rem;
    text-decoration: underline;
}

.disconnect-btn:hover {
    color: #c82333;
}

.search-form {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.search-input {
    flex: 1;
    padding: 0.5rem;
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.9);
    color: #333;
    font-size: 0.9rem;
}

.search-input:focus {
    outline: none;
    border-color: #daa520;
    box-shadow: 0 0 0 2px rgba(218, 165, 32, 0.2);
}

.search-btn {
    background: linear-gradient(145deg, #3498db, #2980b9);
    border-color: #3498db;
}

.search-btn:hover {
    background: linear-gradient(145deg, #2980b9, #21618c);
}

.search-error {
    background: rgba(231, 76, 60, 0.2);
    border: 1px solid #e74c3c;
    border-radius: 4px;
    padding: 0.75rem;
    margin-bottom: 1rem;
}

.search-error p {
    margin: 0;
    color: #e74c3c;
    font-size: 0.9rem;
    text-align: center;
}

.search-results {
    max-height: 300px;
    overflow-y: auto;
    border: 1px solid #8b4513;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.05);
}

.issue-item {
    display: flex;
    flex-direction: column;
    padding: 0.75rem;
    border-bottom: 1px solid rgba(139, 69, 19, 0.3);
    position: relative;
}

.issue-item:last-child {
    border-bottom: none;
}

.issue-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
}

.issue-key {
    font-weight: bold;
    color: #3498db;
    font-size: 0.9rem;
}

.issue-type {
    background: rgba(52, 152, 219, 0.2);
    color: #3498db;
    padding: 0.25rem 0.5rem;
    border-radius: 8px;
    font-size: 0.7rem;
    text-transform: uppercase;
}

.issue-title {
    color: #f4f4f4;
    font-size: 0.9rem;
    margin-bottom: 0.5rem;
    line-height: 1.3;
}

.import-single-btn {
    position: absolute;
    top: 0.75rem;
    right: 0.75rem;
    background: linear-gradient(145deg, #2c5f2d, #4a7c59);
    border: 1px solid #4a7c59;
    color: white;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    font-size: 0.8rem;
    cursor: pointer;
    transition: all 0.3s ease;
}

.import-single-btn:hover {
    background: linear-gradient(145deg, #4a7c59, #5a9e68);
    transform: translateY(-1px);
}
</style>
