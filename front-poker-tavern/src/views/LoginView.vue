<template>
    <div class="login-view">
        <div class="login-container">
            <div class="login-card wow-panel">
                <div class="login-header">
                    <h1>🏰 Planning Poker Tavern</h1>
                    <p>Enter the tavern to start estimating</p>
                </div>

                <form @submit.prevent="handleLogin" class="login-form">
                    <div class="form-group">
                        <label for="playerName">🧙‍♂️ Your Name</label>
                        <input
                            id="playerName"
                            v-model="playerName"
                            type="text"
                            placeholder="Enter your name"
                            required
                            class="form-input"
                        />
                    </div>

                    <div class="form-group">
                        <label for="sessionId">🏠 Session ID</label>
                        <input
                            id="sessionId"
                            v-model="sessionId"
                            type="text"
                            placeholder="Enter session ID or leave empty to create"
                            class="form-input"
                        />
                    </div>

                    <div class="form-actions">
                        <button
                            type="submit"
                            class="wow-button login-btn"
                            :disabled="!playerName.trim() || isLoading"
                        >
                            {{
                                isLoading ? "Connecting..." : "🚪 Enter Tavern"
                            }}
                        </button>
                    </div>
                </form>

                <div class="login-divider">
                    <span>or</span>
                </div>

                <div class="create-session">
                    <h3>🆕 Create New Session</h3>
                    <form
                        @submit.prevent="handleCreateSession"
                        class="create-form"
                    >
                        <div class="form-group">
                            <label for="newSessionName">Session Name</label>
                            <input
                                id="newSessionName"
                                v-model="newSessionName"
                                type="text"
                                placeholder="e.g., Sprint 24 Planning"
                                class="form-input"
                            />
                        </div>

                        <button
                            type="submit"
                            class="wow-button create-btn"
                            :disabled="
                                !playerName.trim() ||
                                !newSessionName.trim() ||
                                isLoading
                            "
                        >
                            {{
                                isLoading ? "Creating..." : "🏗️ Create Session"
                            }}
                        </button>
                    </form>
                </div>

                <div v-if="errorMessage" class="error-message">
                    ⚠️ {{ errorMessage }}
                </div>
            </div>

            <div class="login-info">
                <div class="info-card">
                    <h3>🎯 How it works</h3>
                    <ul>
                        <li>Enter your name and join a session</li>
                        <li>Choose your WoW character class</li>
                        <li>Estimate stories with Fibonacci cards</li>
                        <li>Reveal votes when everyone is done</li>
                    </ul>
                </div>

                <div class="info-card">
                    <h3>⚔️ Features</h3>
                    <ul>
                        <li>Real-time collaborative estimation</li>
                        <li>Medieval tavern atmosphere</li>
                        <li>Persistent voting across sessions</li>
                        <li>Team chat and discussion</li>
                    </ul>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { useGameStore } from "@/stores/gameStore";

// Emit event to parent component when login is successful
const emit = defineEmits<{
    loginSuccess: [];
}>();

// Store
const gameStore = useGameStore();

// Form data
const playerName = ref("");
const sessionId = ref("");
const newSessionName = ref("");
const isLoading = ref(false);
const errorMessage = ref("");

// Handlers
async function handleLogin() {
    if (!playerName.value.trim()) return;

    isLoading.value = true;
    errorMessage.value = "";

    try {
        if (sessionId.value.trim()) {
            // Join existing session
            await gameStore.joinSession(
                sessionId.value.trim(),
                playerName.value.trim(),
            );
            await gameStore.connectToSession(sessionId.value.trim());
        } else {
            // Create default session
            const newSessionId = await gameStore.createSession(
                `${playerName.value}'s Session`,
                [],
            );
            await gameStore.connectToSession(newSessionId);
        }

        emit("loginSuccess");
    } catch (error) {
        errorMessage.value = "Failed to connect. Please try again.";
        console.error("Login error:", error);
    } finally {
        isLoading.value = false;
    }
}

async function handleCreateSession() {
    if (!playerName.value.trim() || !newSessionName.value.trim()) return;

    isLoading.value = true;
    errorMessage.value = "";

    try {
        // Create new session
        const newSessionId = await gameStore.createSession(
            newSessionName.value.trim(),
            [],
        );

        // Connect to the new session
        await gameStore.connectToSession(newSessionId);

        emit("loginSuccess");
    } catch (error) {
        errorMessage.value = "Failed to create session. Please try again.";
        console.error("Create session error:", error);
    } finally {
        isLoading.value = false;
    }
}
</script>

<style scoped>
.login-view {
    min-height: 100vh;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
}

.login-container {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;
    max-width: 1000px;
    width: 100%;
}

.login-card {
    max-width: 100%;
}

.login-header {
    text-align: center;
    margin-bottom: 2rem;
}

.login-header h1 {
    color: #ffd700;
    font-size: 2rem;
    margin-bottom: 0.5rem;
    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
}

.login-header p {
    color: #ecf0f1;
    font-style: italic;
}

.login-form,
.create-form {
    margin-bottom: 1.5rem;
}

.form-group {
    margin-bottom: 1.5rem;
}

.form-group label {
    display: block;
    color: #ffd700;
    font-weight: bold;
    margin-bottom: 0.5rem;
    font-size: 0.9rem;
}

.form-input {
    width: 100%;
    padding: 0.75rem;
    border: 2px solid #7f8c8d;
    border-radius: 8px;
    background: rgba(52, 73, 94, 0.8);
    color: #ecf0f1;
    font-size: 1rem;
    transition: border-color 0.3s ease;
}

.form-input:focus {
    outline: none;
    border-color: #3498db;
    box-shadow: 0 0 8px rgba(52, 152, 219, 0.3);
}

.form-input::placeholder {
    color: #95a5a6;
}

.form-actions {
    text-align: center;
}

.login-btn,
.create-btn {
    width: 100%;
    padding: 1rem;
    font-size: 1.1rem;
    margin: 0.5rem 0;
}

.login-btn {
    background: linear-gradient(145deg, #27ae60, #229954) !important;
    border-color: #27ae60 !important;
}

.login-btn:hover:not(:disabled) {
    background: linear-gradient(145deg, #229954, #1e8449) !important;
    box-shadow: 0 4px 12px rgba(39, 174, 96, 0.4) !important;
}

.create-btn {
    background: linear-gradient(145deg, #3498db, #2980b9) !important;
    border-color: #3498db !important;
}

.create-btn:hover:not(:disabled) {
    background: linear-gradient(145deg, #2980b9, #21618c) !important;
    box-shadow: 0 4px 12px rgba(52, 152, 219, 0.4) !important;
}

.login-divider {
    text-align: center;
    margin: 2rem 0;
    position: relative;
}

.login-divider::before {
    content: "";
    position: absolute;
    top: 50%;
    left: 0;
    right: 0;
    height: 1px;
    background: #7f8c8d;
}

.login-divider span {
    background: linear-gradient(145deg, #2c2c54, #2c3e50);
    color: #95a5a6;
    padding: 0 1rem;
    position: relative;
}

.create-session h3 {
    color: #3498db;
    margin-bottom: 1rem;
    text-align: center;
}

.error-message {
    background: rgba(231, 76, 60, 0.2);
    border: 2px solid #e74c3c;
    color: #e74c3c;
    padding: 1rem;
    border-radius: 8px;
    text-align: center;
    margin-top: 1rem;
}

.login-info {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
}

.info-card {
    background: rgba(52, 73, 94, 0.8);
    border: 2px solid #7f8c8d;
    border-radius: 12px;
    padding: 1.5rem;
}

.info-card h3 {
    color: #f39c12;
    margin-bottom: 1rem;
    text-align: center;
}

.info-card ul {
    list-style: none;
    padding: 0;
}

.info-card li {
    color: #ecf0f1;
    margin-bottom: 0.75rem;
    padding-left: 1.5rem;
    position: relative;
}

.info-card li::before {
    content: "⚡";
    position: absolute;
    left: 0;
    color: #f39c12;
}

/* Responsive */
@media (max-width: 768px) {
    .login-container {
        grid-template-columns: 1fr;
        gap: 1rem;
    }

    .login-header h1 {
        font-size: 1.5rem;
    }

    .login-view {
        padding: 1rem;
    }
}

button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
}
</style>
