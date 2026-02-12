<template>
    <div id="app">
        <HealthChecker />
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
import HealthChecker from "./components/HealthChecker.vue";
import { hasActiveSession, clearAllSessionData } from "./utils/sessionStorage";

// Login state
const isLoggedIn = ref(false);

// Check for existing session in localStorage on app load
if (hasActiveSession()) {
    isLoggedIn.value = true;
}

function handleLoginSuccess() {
    isLoggedIn.value = true;
}

function handleLogout() {
    isLoggedIn.value = false;
    clearAllSessionData(); // Clear all session data on logout
}
</script>

<style>
/* WoW-inspired Fantasy Tavern Styles */
@import url("https://fonts.googleapis.com/css2?family=Cinzel:wght@400;600;700&display=swap");
@import url("https://fonts.googleapis.com/css2?family=Lato:wght@400;700&display=swap");

:root {
    /* "Fantasy Tavern" Palette */
    --color-bg-dark: #120c08;
    --color-wood-dark: #2c1a12;
    --color-wood-medium: #3e2723;
    --color-wood-light: #5d4037;
    --color-gold: #d4af37;
    --color-gold-glow: #ffd700;
    --color-text-parchment: #e0d8c3;
    --color-text-muted: #a69b8d;
    --color-fire: #ff5722;
    
    --shadow-deep: 0 4px 12px rgba(0, 0, 0, 0.8);
    --border-gold: 2px solid var(--color-gold);
}

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: "Cinzel", serif, system-ui;
    /* Deep, smoky tavern corner background */
    background: radial-gradient(circle at 50% 30%, #2c1a12 0%, #120c08 90%);
    color: var(--color-text-parchment);
    min-height: 100vh;
}

/* Global utility classes */
.wow-button {
    background: linear-gradient(180deg, #5d4037 0%, #3e2723 100%);
    border: 2px solid var(--color-gold);
    color: var(--color-gold-glow);
    padding: 0.75rem 1.5rem;
    font-family: "Cinzel", serif;
    font-weight: 700;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.2s ease;
    text-transform: uppercase;
    letter-spacing: 1px;
    box-shadow: 0 2px 4px rgba(0,0,0,0.5);
    text-shadow: 1px 1px 2px rgba(0,0,0,0.8);
}

.wow-button:hover:not(:disabled) {
    background: linear-gradient(180deg, #6d4c41 0%, #4e342e 100%);
    border-color: var(--color-gold-glow);
    box-shadow: 0 0 10px rgba(212, 175, 55, 0.4);
    transform: translateY(-1px);
}

.wow-button:disabled {
    background: #2c1a12;
    border-color: #5d4037;
    color: #5d4037;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
}

.wow-panel {
    background: rgba(44, 26, 18, 0.95); /* Dark wood semi-transparent */
    border: 2px solid var(--color-gold);
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: var(--shadow-deep), inset 0 0 30px rgba(0,0,0,0.8);
    position: relative;
    backdrop-filter: blur(5px);
}

/* Decorative corner accents for panels could go here */
.wow-panel::after {
    content: "";
    position: absolute;
    top: 4px; left: 4px; right: 4px; bottom: 4px;
    border: 1px solid rgba(212, 175, 55, 0.3);
    border-radius: 4px;
    pointer-events: none;
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
    justify-content: center; /* keep center but items should use full width */
    align-items: stretch; /* Stretch children */
    padding: 0; /* Remove padding for full immersion */
    overflow: hidden;
}

/* Mobile: hide the header to maximize screen space (session info is in the mobile top bar) */
@media (max-width: 768px) {
    .app-header {
        display: none;
    }

    .app-main {
        min-height: 100dvh;
    }
}
</style>
