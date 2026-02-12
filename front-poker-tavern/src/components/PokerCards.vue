<template>
    <div class="poker-cards">
        <h3>🃏 Your cards</h3>
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
                    ? "👤 Select character first"
                    : gameStore.currentPlayer?.hasVoted
                      ? "✅ Vote submitted"
                      : "🗳️ Vote"
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
</style>
