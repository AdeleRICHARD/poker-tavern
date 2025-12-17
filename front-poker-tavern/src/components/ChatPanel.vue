<template>
    <div class="chat-panel wow-panel">
        <h3>💬 Discussion</h3>
        <div class="chat-messages" ref="chatMessagesContainer">
            <div
                v-for="message in gameStore.chatMessages"
                :key="message.id"
                :class="['chat-message', message.type]"
            >
                <div class="message-header">
                    <span class="message-author">{{ message.author }}</span>
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
                type="text"
                placeholder="Type a message..."
                class="chat-input-field"
            />
            <button @click="sendMessage" class="wow-button send-btn">
                Send
            </button>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from "vue";
import { useGameStore } from "@/stores/gameStore";

const gameStore = useGameStore();
const chatMessagesContainer = ref<HTMLElement>();
const newMessage = ref("");

function formatTime(date: Date): string {
    return date.toLocaleTimeString("en-US", {
        hour: "2-digit",
        minute: "2-digit",
    });
}

function sendMessage() {
    if (newMessage.value.trim()) {
        gameStore.sendChatMessage(newMessage.value);
        newMessage.value = "";

        // Scroll to bottom
        nextTick(() => {
            if (chatMessagesContainer.value) {
                chatMessagesContainer.value.scrollTop =
                    chatMessagesContainer.value.scrollHeight;
            }
        });
    }
}
</script>

<style scoped>
/* Chat starts here */
.chat-panel {
    display: flex;
    flex-direction: column;
    height: 35%;
    min-height: 0; /* Important for flex child scrolling */
}

.chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 0.5rem;
    background: rgba(52, 73, 94, 0.3);
    border-radius: 8px;
    margin-bottom: 0.5rem;
}
/* previous styles for message... */

.chat-input {
    display: flex;
    gap: 0.5rem;
    padding: 0.5rem;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 8px;
}
/* ... */

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
    font-size: 0.9rem;
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
    color: #ecf0f1;
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

/* WoW Button Style (Copied) */
.wow-button {
    color: white;
    font-weight: bold;
    border-radius: 8px;
    border: 2px solid transparent;
    cursor: pointer;
    transition: all 0.2s ease;
    text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
    padding: 0.75rem 1.5rem;
}

.send-btn {
    background: linear-gradient(145deg, #3498db, #2980b9) !important;
    border-color: #3498db !important;
    padding: 0.75rem 1rem;
}

.send-btn:hover {
    background: linear-gradient(145deg, #2980b9, #21618c) !important;
    box-shadow: 0 4px 8px rgba(52, 152, 219, 0.3) !important;
}

.wow-panel {
    background: rgba(44, 62, 80, 0.85);
    border: 2px solid #8b4513;
    border-radius: 12px;
    padding: 1rem;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3);
}

.wow-panel h3 {
    margin-top: 0;
    margin-bottom: 1rem;
    color: #ffd700;
    text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
    text-align: center;
    border-bottom: 1px solid rgba(255, 215, 0, 0.3);
    padding-bottom: 0.5rem;
}
</style>
