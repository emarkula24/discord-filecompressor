
<template>
    <div class="file-input">
        <label v-if="label" class="file-input__label">{{ label }}</label>
        <input
            ref="inputRef"
            class="file-input__control"
            type="file"
            :accept="'.mp4, .webm'"
            @change="onFileChange"
        />
        <div v-if="file" class="file-input__list">
            <div class="file-input__item">
                {{ file.name }}
            </div>
            <button type="button" class="file-input__submit" @click="fetchPresignedURL">Upload file</button>
        </div>
    </div>
    
</template>

<script setup lang="ts">

import { ref } from 'vue'

const uploadUrl = ref<string>('')
const file = ref<File | null>(null)
const inputRef = ref<HTMLInputElement | null >(null)
const label = ref<string>('')

function onFileChange(event: Event) {
    const target = event.target as HTMLInputElement
    file.value = target.files?.[0] || null
}
async function fetchPresignedURL() {
    const filename = file.value?.name || "unnamed"
    const res = await fetch(`${import.meta.env.VITE_BACKEND_URL}/upload`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({filename})
    }
    )
    uploadUrl.value = await res.json()

}


console.log(uploadUrl)
</script>

<script lang="ts">

async function uploadToS3() {

}

</script>

<style scoped>

.file-input {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-family: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial;
}
.file-input__label {
    font-size: 14px;
    
}
.file-input__control {
    padding: 6px;
}
.file-input__list {
    display: flex;
    gap: 10px;
    align-items: center;
    flex-wrap: wrap;
}
.file-input__item {
    padding: 6px 8px;
    border-radius: 4px;
    font-size: 13px;
}
.file-input__submit {
    border: 1px;
}
</style>