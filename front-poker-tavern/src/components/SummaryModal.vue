<template>
    <div class="modal-overlay" @click="$emit('close')">
        <div class="modal-content wow-panel" @click.stop>
            <div class="modal-header">
                <h2>📊 Voting Summary</h2>
                <button class="close-btn" @click="$emit('close')">✕</button>
            </div>

            <div class="modal-body">
                <div class="summary-stats">
                    <div class="stat-box">
                        <span class="stat-label">Total Stories</span>
                        <span class="stat-value">{{ stories.length }}</span>
                    </div>
                    <div class="stat-box">
                        <span class="stat-label">Estimated</span>
                        <span class="stat-value">{{ completedCount }}</span>
                    </div>
                    <div class="stat-box" v-if="yourVotesCount !== undefined">
                        <span class="stat-label">Your Contributions</span>
                        <span class="stat-value">{{ yourVotesCount }}</span>
                    </div>
                </div>

                <div class="stories-table-wrapper">
                    <table class="summary-table">
                        <thead>
                            <tr>
                                <th>#</th>
                                <th>Issue</th>
                                <th>Title</th>
                                <th>Status</th>
                                <th>Consensus</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr
                                v-for="(story, index) in stories"
                                :key="story.id"
                                :class="{ revealed: isRevealed(story.id) }"
                            >
                                <td class="col-num">{{ index + 1 }}</td>
                                <td class="col-key">
                                    <span class="jira-key">{{
                                        story.jiraKey || "---"
                                    }}</span>
                                </td>
                                <td class="col-title">{{ story.title }}</td>
                                <td class="col-status">
                                    <span
                                        v-if="isRevealed(story.id)"
                                        class="status-badge revealed"
                                        >Revealed ✅</span
                                    >
                                    <span v-else class="status-badge pending"
                                        >{{ getVotesCount(story.id) }}/{{
                                            requiredCount
                                        }}
                                        Votes</span
                                    >
                                </td>
                                <td class="col-avg">
                                    <strong v-if="isRevealed(story.id)">{{
                                        getAverage(story.id) || "---"
                                    }}</strong>
                                    <span v-else>---</span>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { Story } from "@/stores/gameStore";

const props = defineProps<{
    stories: Story[];
    persistentVotes: Record<string, Record<string, string>>;
    requiredCount: number;
    yourVotesCount?: number;
}>();

defineEmits<{
    (e: "close"): void;
}>();

const completedCount = computed(() => {
    return props.stories.filter((s) => isRevealed(s.id)).length;
});

function isRevealed(storyId: string): boolean {
    const storyVotes = props.persistentVotes[storyId] || {};
    // A story is revealed if all required players have voted
    // In our simplified logic, if the count matches requiredCount
    return Object.keys(storyVotes).length >= props.requiredCount && props.requiredCount > 0;
}

function getVotesCount(storyId: string): number {
    return Object.keys(props.persistentVotes[storyId] || {}).length;
}

function getAverage(storyId: string): number | null {
    const storyVotes = props.persistentVotes[storyId] || {};
    const numericVotes = Object.values(storyVotes)
        .filter((vote) => vote && !isNaN(Number(vote)))
        .map(Number);

    if (numericVotes.length === 0) return null;
    const sum = numericVotes.reduce((acc, vote) => acc + vote, 0);
    return Math.round((sum / numericVotes.length) * 10) / 10;
}
</script>

<style scoped>
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.85);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
}

.modal-content {
    width: 90%;
    max-width: 900px;
    background: #1e1e2f;
    border: 2px solid #3d3d5c;
    border-radius: 12px;
    box-shadow: 0 20px 50px rgba(0, 0, 0, 0.5);
    color: #e0e0e0;
    overflow: hidden;
    animation: modalScale 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes modalScale {
    from { transform: scale(0.9); opacity: 0; }
    to { transform: scale(1); opacity: 1; }
}

.modal-header {
    background: #252538;
    padding: 1.25rem 2rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    border-bottom: 2px solid #3d3d5c;
}

.modal-header h2 {
    margin: 0;
    color: #ffd700;
    font-size: 1.5rem;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
}

.close-btn {
    background: transparent;
    border: none;
    color: #888;
    font-size: 1.5rem;
    cursor: pointer;
    transition: color 0.2s;
}

.close-btn:hover {
    color: #fff;
}

.modal-body {
    padding: 2rem;
    max-height: 70vh;
    overflow-y: auto;
}

.summary-stats {
    display: flex;
    gap: 1.5rem;
    margin-bottom: 2rem;
}

.stat-box {
    flex: 1;
    background: #2a2a40;
    padding: 1.25rem;
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    align-items: center;
    border: 1px solid #3d3d5c;
}

.stat-label {
    font-size: 0.85rem;
    color: #999;
    text-transform: uppercase;
    letter-spacing: 1px;
    margin-bottom: 0.5rem;
}

.stat-value {
    font-size: 2rem;
    font-weight: bold;
    color: #fff;
}

.stories-table-wrapper {
    background: #161625;
    border-radius: 8px;
    border: 1px solid #3d3d5c;
    overflow: hidden;
}

.summary-table {
    width: 100%;
    border-collapse: collapse;
    text-align: left;
}

.summary-table th {
    background: #252538;
    padding: 1rem;
    font-size: 0.9rem;
    color: #aaa;
    border-bottom: 1px solid #3d3d5c;
}

.summary-table td {
    padding: 1rem;
    border-bottom: 1px solid #2a2a40;
    vertical-align: middle;
}

.summary-table tr:last-child td {
    border-bottom: none;
}

.summary-table tr.revealed {
    background: rgba(30, 255, 100, 0.03);
}

.col-num { color: #666; width: 40px; }
.col-key { width: 120px; }
.jira-key {
    background: #3d3d5c;
    padding: 2px 8px;
    border-radius: 4px;
    font-family: monospace;
    font-size: 0.85rem;
    color: #ffd700;
}

.col-title { font-weight: 500; }

.status-badge {
    padding: 4px 10px;
    border-radius: 20px;
    font-size: 0.75rem;
    font-weight: bold;
    display: inline-block;
}

.status-badge.revealed {
    background: rgba(46, 204, 113, 0.2);
    color: #2ecc71;
    border: 1px solid rgba(46, 204, 113, 0.3);
}

.status-badge.pending {
    background: rgba(241, 196, 15, 0.1);
    color: #f1c40f;
    border: 1px solid rgba(241, 196, 15, 0.3);
}

.col-avg {
    text-align: right;
    color: #ffd700;
    font-size: 1.1rem;
    padding-right: 2rem !important;
}
</style>
