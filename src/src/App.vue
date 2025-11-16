
<template>
  <FileInput  @download-ready="handleDownloadUrl"/>
  <VideoPlayer v-if="downloadUrl" :video-url="downloadUrl" />
</template>

<script setup lang="ts">
import VideoPlayer from './components/FileDownload.vue'
import FileInput from './components/FileInput.vue';
import { ref } from 'vue';
const downloadUrl = ref('')
function handleDownloadUrl(url: string) {
  downloadUrl.value = url
  console.log(downloadUrl)
}
</script>

<style scoped>
.logo {
  height: 6em;
  padding: 1.5em;
  will-change: filter;
  transition: filter 300ms;
}
.logo:hover {
  filter: drop-shadow(0 0 2em #646cffaa);
}
.logo.vue:hover {
  filter: drop-shadow(0 0 2em #42b883aa);
}
</style>
