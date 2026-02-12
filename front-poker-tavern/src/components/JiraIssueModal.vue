<template>
    <Teleport to="body">
        <div class="modal-overlay" v-if="isVisible" @click.self="closeModal">
            <div class="quest-modal">
                <div class="quest-header">
                    <div class="quest-title-section">
                        <div class="quest-icon">📋</div>
                        <h2 class="quest-title">
                            {{ issue?.title || "JIRA Issue" }}
                        </h2>
                    </div>
                    <button class="close-btn" @click="closeModal">✕</button>
                </div>

                <div class="quest-content">
                    <div class="quest-details">
                        <div class="detail-item">
                            <strong>Issue Key:</strong> {{ issue?.jiraKey }}
                        </div>
                        <div class="detail-item" v-if="issue?.type">
                            <strong>Type:</strong>
                            <span class="issue-type-badge">{{
                                issue.type
                            }}</span>
                        </div>
                        <div class="detail-item" v-if="issue?.priority">
                            <strong>Priority:</strong> {{ issue.priority }}
                        </div>
                        <div class="detail-item" v-if="issue?.status">
                            <strong>Status:</strong> {{ issue.status }}
                        </div>
                    </div>

                    <div
                        class="quest-labels"
                        v-if="issue?.labels && issue.labels.length > 0"
                    >
                        <h3>🏷️ Labels</h3>
                        <div class="labels-container">
                            <span
                                v-for="label in issue.labels"
                                :key="label"
                                class="label-tag"
                            >
                                {{ label }}
                            </span>
                        </div>
                    </div>

                    <div class="quest-description">
                        <h3>📋 Description</h3>
                        <div class="description-content">
                            <div
                                v-if="issue?.description"
                                v-html="issue.description"
                                class="formatted-description"
                            ></div>
                            <p v-else class="no-description">
                                No description available for this issue.
                            </p>
                        </div>
                    </div>
                </div>

                <div class="quest-actions" v-if="mode === 'import'">
                    <button class="decline-btn" @click="closeModal">
                        <span>❌</span> Decline
                    </button>
                    <button class="accept-btn" @click="importIssue">
                        <span>✅</span> Accept
                    </button>
                </div>

                <div class="quest-actions" v-else>
                    <button class="close-only-btn" @click="closeModal">
                        <span>✖️</span> Close
                    </button>
                </div>
            </div>
        </div>
    </Teleport>
</template>

<script setup lang="ts">
import { ref } from "vue";

const props = defineProps<{
    issue?: any;
    mode?: "import" | "view"; // 'import' shows Accept/Decline, 'view' shows only Close
}>();

const emit = defineEmits<{
    import: [issue: any];
}>();

const isVisible = ref(false);

function openModal() {
    isVisible.value = true;
}

function closeModal() {
    isVisible.value = false;
}

function importIssue() {
    if (props.issue) {
        emit("import", props.issue);
        closeModal();
    }
}

defineExpose({
    openModal,
    closeModal,
});
</script>

<style scoped>
.modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
    backdrop-filter: blur(4px);
}

.quest-modal {
    background: linear-gradient(145deg, #f4f2e8, #e8e4d6);
    border: 4px solid #8b6914;
    border-radius: 16px;
    box-shadow:
        0 0 20px rgba(139, 105, 20, 0.6),
        inset 0 2px 4px rgba(255, 255, 255, 0.3),
        inset 0 -2px 4px rgba(0, 0, 0, 0.1);
    width: 800px;
    max-width: 95vw;
    max-height: 90vh;
    overflow-y: auto;
    animation: questAppear 0.3s ease-out;
    position: relative;
    font-size: 1.1rem;
}

@keyframes questAppear {
    from {
        opacity: 0;
        transform: scale(0.9) translateY(-20px);
    }
    to {
        opacity: 1;
        transform: scale(1) translateY(0);
    }
}

.quest-header {
    background: linear-gradient(145deg, #d4a574, #c49660);
    border-bottom: 2px solid #8b6914;
    padding: 1rem 1.5rem;
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    border-radius: 12px 12px 0 0;
    position: relative;
}

.quest-title-section {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    flex: 1;
}

.quest-icon {
    font-size: 2rem;
    background: rgba(255, 215, 0, 0.2);
    padding: 0.5rem;
    border-radius: 50%;
    border: 2px solid #ffd700;
}

.quest-title {
    color: #1a1a1a;
    font-size: 1.4rem;
    font-weight: bold;
    margin: 0;
    text-shadow: 1px 1px 2px rgba(255, 255, 255, 0.5);
    line-height: 1.3;
    flex: 1;
    word-break: break-word;
}

.close-btn {
    background: rgba(220, 53, 69, 0.1);
    border: 2px solid #dc3545;
    color: #dc3545;
    border-radius: 50%;
    width: 32px;
    height: 32px;
    display: flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    font-weight: bold;
    transition: all 0.2s ease;
    font-size: 1rem;
    flex-shrink: 0;
    margin-left: 1rem;
}

.close-btn:hover {
    background: rgba(220, 53, 69, 0.2);
    transform: scale(1.1);
}

.quest-content {
    padding: 1.5rem;
}

.quest-description {
    background: rgba(255, 255, 255, 0.6);
    border: 2px solid #d4a574;
    border-radius: 8px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    color: #1a1a1a;
    line-height: 1.6;
    font-size: 1.05rem;
}

.quest-description h3 {
    color: #8b6914;
    margin-bottom: 0.75rem;
    font-size: 1.1rem;
    border-bottom: 2px solid #d4a574;
    padding-bottom: 0.5rem;
    margin-top: 0;
}

.description-content {
    max-height: 350px;
    overflow-y: auto;
    font-style: normal;
    font-weight: 500;
}

.description-content p {
    margin: 0;
    white-space: pre-wrap;
    word-wrap: break-word;
}

.no-description {
    font-style: italic;
    color: #6c757d;
    text-align: center;
}

.quest-details {
    background: rgba(255, 255, 255, 0.2);
    border: 1px solid #d4a574;
    border-radius: 8px;
    padding: 1rem;
    margin-bottom: 1.5rem;
}

.detail-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.5rem 0;
    border-bottom: 1px solid rgba(139, 105, 20, 0.2);
    color: #2c1810;
}

.detail-item:last-child {
    border-bottom: none;
}

.issue-type-badge {
    background: linear-gradient(145deg, #3498db, #2980b9);
    color: white;
    padding: 0.2rem 0.6rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: bold;
    text-transform: uppercase;
}

.quest-labels {
    background: rgba(255, 255, 255, 0.3);
    border: 2px solid #d4a574;
    border-radius: 8px;
    padding: 1.5rem;
    margin-bottom: 1.5rem;
    color: #1a1a1a;
}

.quest-labels h3 {
    color: #8b6914;
    margin-bottom: 0.75rem;
    font-size: 1.1rem;
    border-bottom: 2px solid #d4a574;
    padding-bottom: 0.5rem;
    margin-top: 0;
}

.labels-container {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 0.75rem;
}

.label-tag {
    background: linear-gradient(145deg, #17a2b8, #138496);
    color: white;
    padding: 0.3rem 0.8rem;
    border-radius: 16px;
    font-size: 0.85rem;
    font-weight: bold;
    border: 2px solid #17a2b8;
    transition: all 0.2s ease;
    display: inline-flex;
    align-items: center;
    text-transform: capitalize;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.label-tag:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
    background: linear-gradient(145deg, #138496, #117a8b);
}

.quest-actions {
    background: linear-gradient(145deg, #e8e4d6, #f4f2e8);
    border-top: 2px solid #d4a574;
    padding: 1rem 1.5rem;
    display: flex;
    gap: 1rem;
    justify-content: center;
    border-radius: 0 0 12px 12px;
}

.decline-btn,
.accept-btn {
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: bold;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 120px;
    justify-content: center;
}

.decline-btn {
    background: linear-gradient(145deg, #6c757d, #5a6268);
    border: 2px solid #6c757d;
    color: white;
}

.decline-btn:hover {
    background: linear-gradient(145deg, #5a6268, #495057);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

.accept-btn {
    background: linear-gradient(145deg, #28a745, #1e7e34);
    border: 2px solid #28a745;
    color: white;
}

.accept-btn:hover {
    background: linear-gradient(145deg, #1e7e34, #155724);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

.close-only-btn {
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: bold;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-width: 120px;
    justify-content: center;
    background: linear-gradient(145deg, #3498db, #2980b9);
    border: 2px solid #3498db;
    color: white;
}

.close-only-btn:hover {
    background: linear-gradient(145deg, #2980b9, #21618c);
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
}

/* Scrollbar styling */
.quest-modal::-webkit-scrollbar {
    width: 8px;
}

.quest-modal::-webkit-scrollbar-track {
    background: rgba(212, 165, 116, 0.2);
    border-radius: 4px;
}

.quest-modal::-webkit-scrollbar-thumb {
    background: #d4a574;
    border-radius: 4px;
}

.quest-modal::-webkit-scrollbar-thumb:hover {
    background: #c49660;
}

/* Mobile responsive */
@media (max-width: 768px) {
    .quest-modal {
        width: 100vw;
        max-width: 100vw;
        max-height: 100dvh;
        height: 100dvh;
        border-radius: 0;
        border: none;
        font-size: 1rem;
    }

    .quest-header {
        border-radius: 0;
        padding: 0.75rem 1rem;
        position: sticky;
        top: 0;
        z-index: 10;
    }

    .quest-icon {
        font-size: 1.5rem;
        padding: 0.35rem;
    }

    .quest-title {
        font-size: 1.1rem;
    }

    .close-btn {
        width: 40px;
        height: 40px;
        font-size: 1.2rem;
    }

    .quest-content {
        padding: 1rem;
    }

    .quest-description {
        padding: 1rem;
    }

    .description-content {
        max-height: none;
    }

    .quest-actions {
        position: sticky;
        bottom: 0;
        border-radius: 0;
        padding: 1rem;
    }

    .decline-btn,
    .accept-btn,
    .close-only-btn {
        padding: 0.85rem 1.5rem;
        min-height: 48px;
        font-size: 1rem;
    }
}
</style>
