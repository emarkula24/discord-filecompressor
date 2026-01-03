

<template>
    <div class="download-btn">
        <a :href="videoUrl">Download</a>
    </div>
    
</template>

<script setup lang="ts">
defineProps<{
    videoUrl: string
}>()
</script>

<style scoped>
    .download-btn {
        margin-top: 10px;
    }
</style>