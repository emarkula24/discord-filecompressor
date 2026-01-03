
<template>
    <div class="file-input">
        <div v-if="isProcessing" class="loader-box">
            <span class="loader"></span>
        </div>
        <label v-if="label" class="file-input__label">{{ label }}</label>
        <input
            ref="inputRef"
            class="file-input__control"
            type="file"
            :accept="'.mp4, .webm'"
            @change="onFileChange"
        />

        <div class="file-input__list">
            <button type="button" class="file-input__submit" @click="fetchPresignedURL">Upload file</button>
        </div>
        <div v-if="errorMsg">
            <ErrorComponent  :err-msg="errorMsg"/>
        </div>
        
    </div>

</template>

<script setup lang="ts">

import { ref} from 'vue'
import ErrorComponent from './ErrorComponent.vue'

const emit = defineEmits<{
    (e: 'download-ready', url: string): void
}>()

interface PreURL {
    method: string;
    url: string;
    headers: Record<string, string>;
}
interface S3Url {
    job_id: number;
    object_key: string;
    presigned_url: PreURL;
}

const S3URL = ref<S3Url | null>(null)
const file = ref<File | null>(null)
const label = ref<string>('')
const url = import.meta.env.VITE_BACKEND_URL
const errorMsg = ref<string>('') 
const isFailed = ref(false)
const isProcessing = ref(false)

function onFileChange(event: Event) {
    const target = event.target as HTMLInputElement
    file.value = target.files?.[0] || null
}
const fetchPresignedURL = async () => {
    const minSize = 1048576 // 10 MB in bytes
    const fileSize = file.value?.size || 0
    if (fileSize <= minSize) {
        errorMsg.value = "File is already small enough."
        return
    }
    errorMsg.value = ""

    const filename = file.value?.name || "unnamed"
    const res = await fetch(`${url}/upload`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({filename})
    }
    )
    if (res.ok) {
        console.log("Backend URL:", import.meta.env.VITE_BACKEND_URL)
        const data: S3Url = await res.json()
        S3URL.value = data
        console.log(S3URL)
        if (file.value) {
            await uploadToS3(file.value, S3URL.value)
        }
    } else {
        errorMsg.value = "File upload failed."
    }

}

const uploadToS3 = async (file: File, S3URL: S3Url) => {
    isProcessing.value = true
    try {
        const { method, url, headers } = S3URL.presigned_url;
        console.log(method)
        console.log(headers.Host)
        
        const res = await fetch(url, {
            method: method,
            headers: { "Content-Type": "video/mp4"},
            body: file
        })
        console.log(res)
        if (!res.ok) {
            throw new Error(`Upload failed with status ${res.status}`)
        }
        console.log("successfully uploaded to r2 bucket")
        onS3Upload(S3URL.job_id, S3URL.object_key)
        checkStatus()
    } catch (error) {
        console.error(error)
        errorMsg.value = "File upload failed."
    }
}

const onS3Upload = async (job_id: number, object_key: string) => {
    const data = {job_id, object_key}
    try {
        const res = await fetch(`${url}/jobs/upload`, {
            method: "POST",
            headers: {"Content-Type": "application/json"},
            body: JSON.stringify(data)

        })

        const result = await res.json()
        console.log(result)
    } catch(error) {
        console.error(error)
        errorMsg.value = "File upload failed."
    }
}

const getJobStatus = async (job_id: number) => {
    const timeout = 1200000
    const interval = 4000
    const startTime = Date.now()

    isFailed.value = false
    errorMsg.value = ""

    while (Date.now() - startTime < timeout) {
        try {
                const res = await fetch(`${url}/jobs/status?job_id=${job_id}`, {
                    method: "GET",
                    headers: {"Content-Type":"application/json"},
                })
                const result = await res.json()
                console.log(result)

                if (result.event_type === "success" && result.presigned_download_url?.url) {
                    emit('download-ready', result.presigned_download_url.url)
                    isProcessing.value = false
                    return
                }

                if (result.event_type === "failed") {
                    isProcessing.value = false
                    isFailed.value = true
                    throw Error("File compression failed.")
                }
                await new Promise(resolve => setTimeout(resolve, interval))
                
        } catch(error) {
            console.log(error)
            isProcessing.value = false
            errorMsg.value = "File compression failed."
            }
    }
    isProcessing.value = false
    errorMsg.value = "Job status check timed out."
}


const checkStatus = () => {
    
    if (S3URL.value) {
        console.log("Calling getJobStatus with job_id:", S3URL.value.job_id)
        getJobStatus(S3URL.value.job_id)
    } else {
        console.warn("no job id")
        errorMsg.value = "Something went wrong."
    }
}

</script>

<style scoped>
.loader-box {
    position: relative;
}
.file-input {
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-family: system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial;
    justify-content: center;
    align-items: center;
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

.file-input__submit {
    border: 1px;
}
.loader {
  width: 48px;
  height: 48px;
  border: 5px solid #FFF;
  border-radius: 50%;
  display: inline-block;
  box-sizing: border-box;
  position: relative;
  animation: pulse 1s linear infinite;
}
.loader:after {
  content: '';
  position: absolute;
  width: 48px;
  height: 48px;
  border: 5px solid #FFF;
  border-radius: 50%;
  display: inline-block;
  box-sizing: border-box;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  animation: scaleUp 1s linear infinite;
}

@keyframes scaleUp {
  0% { transform: translate(-50%, -50%) scale(0) }
  60% , 100% { transform: translate(-50%, -50%)  scale(1)}
}
@keyframes pulse {
  0% , 60% , 100%{ transform:  scale(1) }
  80% { transform:  scale(1.2)}
}


</style>