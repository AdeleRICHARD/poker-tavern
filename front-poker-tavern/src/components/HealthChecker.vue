<template>
    <div class="health-checker" :class="`health-checker--${healthStatus}`">
        <span class="status-dot" aria-hidden="true"></span>
        <span>Serveur {{ statusLabel }}</span>
    </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { getApiUrl } from "@/config/api";

type HealthStatus = "checking" | "online" | "offline";

const healthStatus = ref<HealthStatus>("checking");
const statusLabel = computed(() => {
    if (healthStatus.value === "online") return "opérationnel";
    if (healthStatus.value === "offline") return "indisponible";
    return "en vérification";
});

const checkHealth = async () => {
    try {
        const response = await fetch(getApiUrl("/health"));
        const data = await response.json();
        healthStatus.value = data.status === "ok" ? "online" : "offline";
    } catch (error) {
        console.error("Error fetching /health:", error);
        healthStatus.value = "offline";
    }
};

onMounted(() => {
    checkHealth();
});
</script>

<style scoped>
.health-checker {
    position: fixed;
    top: 1.25rem;
    right: 1.5rem;
    z-index: 20;
    display: inline-flex;
    gap: 0.55rem;
    align-items: center;
    padding: 0.55rem 0.8rem;
    border: 1px solid rgba(255, 246, 229, 0.14);
    border-radius: 999rem;
    background: rgba(24, 14, 9, 0.64);
    color: rgba(255, 246, 229, 0.68);
    font-family: "Lato", sans-serif;
    font-size: 0.68rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    box-shadow: 0 0.5rem 1.5rem rgba(0, 0, 0, 0.18);
    backdrop-filter: blur(1rem);
}

.status-dot {
    width: 0.45rem;
    height: 0.45rem;
    border-radius: 50%;
    background: #dba761;
    box-shadow: 0 0 0.65rem rgba(219, 167, 97, 0.7);
}

.health-checker--online .status-dot {
    background: #76c38f;
    box-shadow: 0 0 0.65rem rgba(118, 195, 143, 0.7);
}

.health-checker--offline .status-dot {
    background: #dc7564;
    box-shadow: 0 0 0.65rem rgba(220, 117, 100, 0.7);
}

@media (max-width: 38rem) {
    .health-checker {
        top: 0.9rem;
        right: 1rem;
    }
}
</style>
