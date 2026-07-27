<template>
    <div>
        <h2>État du serveur</h2>
        <p>Statut : {{ healthStatus }}</p>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { getApiUrl } from "@/config/api";

const healthStatus = ref("vérification...");

const checkHealth = async () => {
    try {
        const response = await fetch(getApiUrl("/health"));
        const data = await response.json();
        healthStatus.value = data.status === "ok" ? "opérationnel" : data.status;
    } catch (error) {
        console.error("Error fetching /health:", error);
        healthStatus.value = "indisponible";
    }
};

onMounted(() => {
    checkHealth();
});
</script>
