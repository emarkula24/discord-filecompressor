<script setup lang="ts">

defineProps<{errMsg: string}>()


</script>

<template>
    <div>
        <p style="color: red;">{{ errMsg }}</p>
    </div>
</template>

