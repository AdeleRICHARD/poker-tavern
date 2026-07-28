<template>
    <div id="app">
        <HealthChecker v-if="!isLoggedIn" />
        <header v-if="isLoggedIn" class="app-header">
            <h1>🏰 La Taverne du Planning Poker</h1>
            <p>Estimation collaborative pour votre équipe</p>
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
    position: relative;
    display: flex;
    justify-content: center;
    gap: 1rem;
    align-items: baseline;
    min-height: 5.25rem;
    padding: 1.15rem 2rem;
    border-bottom: 1px solid rgba(232, 177, 103, 0.22);
    background:
        repeating-linear-gradient(176deg, rgba(255, 218, 160, 0.025) 0 0.08rem, transparent 0.08rem 0.48rem),
        linear-gradient(90deg, #1d100a 0%, #4d2a17 50%, #1d100a 100%);
    text-align: center;
    box-shadow: 0 0.75rem 2.5rem rgba(0, 0, 0, 0.36);
}

.app-header h1 {
    margin: 0;
    color: #fff1d7;
    font-family: "Cinzel", serif;
    font-size: clamp(1.25rem, 2vw, 1.65rem);
    font-weight: 600;
    text-shadow: 0 0.15rem 0.5rem rgba(0, 0, 0, 0.72);
}

.app-header p {
    color: rgba(255, 239, 213, 0.48);
    font-family: "Lato", sans-serif;
    font-size: 0.75rem;
    font-style: normal;
    letter-spacing: 0.05em;
    text-transform: uppercase;
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
