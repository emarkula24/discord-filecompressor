

<template>
    <div>
        <video controls width="480" crossorigin="anonymous" @error="onVideoError">
            <source :src="videoUrl" type="video/mp4" >
        </video>
    </div>
    <a :href="videoUrl">Download</a>
</template>

<script setup lang="ts">
defineProps<{
    videoUrl: string
}>()
function onVideoError(e: Event) {
  console.error('Video failed to load:', e)
}
</script>