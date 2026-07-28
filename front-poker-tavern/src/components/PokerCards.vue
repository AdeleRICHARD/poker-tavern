<template>
    <div class="poker-cards">
        <h3>🃏 Vos cartes</h3>
        <div class="cards-grid">
            <button
                v-for="card in gameStore.pokerCards"
                :key="card.value"
                :class="[
                    'poker-card',
                    {
                        selected: gameStore.selectedCard === card.value,
                    },
                ]"
                @click="gameStore.selectCard(card.value)"
                :disabled="gameStore.gamePhase !== GamePhase.VOTING"
                :title="card.description"
            >
                {{ card.label }}
            </button>
        </div>

        <button
            class="wow-button vote-btn"
            @click="$emit('submit')"
            :disabled="
                !gameStore.selectedCard ||
                !gameStore.currentPlayer ||
                gameStore.currentPlayer?.hasVoted
            "
        >
            {{
                !gameStore.currentPlayer
                    ? "👤 Choisissez d’abord une classe"
                    : gameStore.currentPlayer?.hasVoted
                      ? "✅ Vote envoyé"
                      : "🗳️ Voter"
            }}
        </button>
    </div>
</template>

<script setup lang="ts">
import { useGameStore, GamePhase } from "@/stores/gameStore";

const gameStore = useGameStore();

defineEmits<{
    (e: "submit"): void;
}>();
</script>

<style scoped>
.poker-cards {
    margin-bottom: 1.5rem;
}

.cards-grid {
    display: grid;
    grid-template-columns: repeat(5, 1fr);
    gap: 0.5rem;
    margin-bottom: 1rem;
}

.poker-card {
    aspect-ratio: 3/4;
    background: linear-gradient(145deg, #ecf0f1, #bdc3c7);
    border: 2px solid #95a5a6;
    border-radius: 8px;
    cursor: pointer;
    font-weight: bold;
    font-size: 1rem;
    color: #2c3e50;
    transition: all 0.3s ease;
}

.poker-card:hover:not(:disabled) {
    border-color: #3498db;
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(52, 152, 219, 0.3);
}

.poker-card.selected {
    background: linear-gradient(145deg, #3498db, #2980b9);
    border-color: #ffd700;
    color: white;
}

.poker-card:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.vote-btn {
    width: 100%;
    margin-top: 0.5rem;
    font-size: 1.2rem;
    background: linear-gradient(145deg, #2ecc71, #27ae60) !important;
    border-color: #2ecc71 !important;
}

.vote-btn:disabled {
    background: linear-gradient(145deg, #7f8c8d, #95a5a6) !important;
    border-color: #7f8c8d !important;
}

/* WoW Button Style (Copied from Global/Tavern) */
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

/* Mobile responsive */
@media (max-width: 768px) {
    .cards-grid {
        grid-template-columns: repeat(5, 1fr);
        gap: 0.4rem;
    }

    .poker-card {
        aspect-ratio: 1;
        font-size: 1.1rem;
        border-radius: 10px;
        min-height: 44px;
    }

    .vote-btn {
        font-size: 1rem;
        padding: 0.85rem 1rem;
        min-height: 48px;
    }

    .poker-cards h3 {
        font-size: 0.95rem;
        margin-bottom: 0.5rem;
    }
}

/* Modern tavern skin — card selection and submission stay unchanged */
.poker-cards {
    margin-bottom: 0;
}

.poker-cards h3 {
    margin-bottom: 0.7rem;
    color: #efbd78;
    font-family: "Cinzel", serif;
    font-size: 0.78rem;
    letter-spacing: 0.05em;
    text-transform: uppercase;
}

.cards-grid {
    gap: 0.4rem;
    margin-bottom: 0.75rem;
}

.poker-card {
    min-height: 2.65rem;
    border: 1px solid rgba(255, 236, 207, 0.16);
    border-radius: 0.55rem;
    background: linear-gradient(145deg, rgba(255, 247, 232, 0.92), rgba(211, 190, 161, 0.88));
    color: #2b190f;
    font-family: "Cinzel", serif;
    font-size: 0.8rem;
}

.poker-card:hover:not(:disabled) {
    border-color: #e1a65f;
    box-shadow: 0 0.45rem 0.85rem rgba(177, 89, 35, 0.2);
    transform: translateY(-0.08rem);
}

.poker-card.selected {
    border-color: #f1c181;
    background: linear-gradient(145deg, #d68a42, #a65327);
    color: #fff8ea;
}

.poker-card:disabled {
    opacity: 0.34;
}

.vote-btn,
.vote-btn:disabled {
    min-height: 2.8rem;
    margin-top: 0;
    border-radius: 0.7rem;
    font-family: "Lato", sans-serif;
    font-size: 0.76rem;
}

.vote-btn {
    border-color: #d99955 !important;
    background: linear-gradient(135deg, #d58a42, #a95629) !important;
    color: #1b0e07;
}

.vote-btn:disabled {
    border-color: rgba(255, 236, 207, 0.12) !important;
    background: rgba(8, 5, 3, 0.38) !important;
    color: rgba(255, 248, 234, 0.35);
}
</style>
