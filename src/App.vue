<template>
    <div id="app">
        <header v-if="isLoggedIn" class="app-header">
            <h1>🏰 Planning Poker Tavern</h1>
            <p>Collaborative estimation for your development team</p>
        </header>

        <main class="app-main">
            <LoginView v-if="!isLoggedIn" @login-success="handleLoginSuccess" />
            <TavernView v-else @logout="handleLogout" />
        </main>
    </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import TavernView from "./views/TavernView.vue";
import LoginView from "./views/LoginView.vue";

// Login state
const isLoggedIn = ref(false);

function handleLoginSuccess() {
    isLoggedIn.value = true;
}

function handleLogout() {
    isLoggedIn.value = false;
}
</script>

<!-- Global styles (not scoped) -->
<style>
/* WoW-inspired styles */
@import url("https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&display=swap");

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: "Cinzel", serif, system-ui;
    background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
    color: #f4f4f4;
    min-height: 100vh;
}

/* Global utility classes */
.wow-button {
    background: linear-gradient(145deg, #8b4513, #a0522d);
    border: 2px solid #daa520;
    color: #ffd700;
    padding: 0.75rem 1.5rem;
    font-family: "Cinzel", serif;
    font-weight: 600;
    cursor: pointer;
    border-radius: 8px;
    transition: all 0.3s ease;
    text-transform: uppercase;
    letter-spacing: 1px;
}

.wow-button:hover {
    background: linear-gradient(145deg, #a0522d, #cd853f);
    box-shadow: 0 4px 12px rgba(218, 165, 32, 0.4);
    transform: translateY(-2px);
}

.wow-panel {
    background: linear-gradient(145deg, #2c2c54, #2c3e50);
    border: 3px solid #8b4513;
    border-radius: 12px;
    padding: 1.5rem;
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.5);
    position: relative;
}

.wow-panel::before {
    content: "";
    position: absolute;
    top: -3px;
    left: -3px;
    right: -3px;
    bottom: -3px;
    background: linear-gradient(45deg, #daa520, #ffd700, #daa520);
    border-radius: 15px;
    z-index: -1;
}
</style>

<!-- Component-specific styles (scoped) -->
<style scoped>
#app {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
}

.app-header {
    background: linear-gradient(90deg, #8b4513 0%, #a0522d 50%, #8b4513 100%);
    padding: 1rem 2rem;
    text-align: center;
    border-bottom: 3px solid #daa520;
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
}

.app-header h1 {
    font-size: 2.5rem;
    color: #ffd700;
    text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
    margin-bottom: 0.5rem;
}

.app-header p {
    font-size: 1.1rem;
    color: #f4f4f4;
    font-style: italic;
}

.app-main {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 2rem;
}
</style>
