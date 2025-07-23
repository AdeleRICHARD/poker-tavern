<template>
    <div>
        <h2>Backend Health Check</h2>
        <p>Status: {{ healthStatus }}</p>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { getApiUrl } from "@/config/api";

const healthStatus = ref("loading...");

const checkHealth = async () => {
    try {
        const response = await fetch(getApiUrl("/health"));
        const data = await response.json();
        healthStatus.value = data.status;
    } catch (error) {
        console.error("Error fetching /health:", error);
        healthStatus.value = "error";
    }
};

onMounted(() => {
    checkHealth();
});
</script>
