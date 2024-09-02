<template>
    <div class="flex flex-col align-items justify-center mt-8">
        <h2 class="my-4">文件状态：<span class="text-white p-2"
                :class="{ 'bg-green-600': state.exist, 'bg-red-600': !state.exist }">{{ state.exist ? '存在' : '不存在'
                }}</span></h2>
        <button class="w-40 text-base bg-sky-600 text-white px-6 py-2 rounded-lg" @click="uploadFile">Upload</button>
        <h2 class="my-4">COOKIE 状态：<span class="text-white p-2"
                :class="{ 'bg-green-600': state.valid, 'bg-red-600': !state.valid }">{{ state.valid ? '有效' : '无效'
                }}</span></h2>
        <button class="w-40 text-base bg-sky-600 text-white px-6 py-2 rounded-lg" @click="checkValid">Check</button>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const state = ref({
    exist: null,
    valid: null,
})

const checkFile = () => {
    fetch(`/api/cookie/exist`).then(async (rsp) => {
        if (rsp.ok) {
            const data = await rsp.json()
            state.value.exist = data.exist
        }
    })
}
const checkValid = () => {
    fetch(`/api/cookie/status`).then(async (rsp) => {
        if (rsp.ok) {
            const data = await rsp.json()
            state.value.valid = data.logged

        }
    })
}

const uploadFile = () => {
    const input: HTMLInputElement = document.createElement('input')
    input.type = "file"
    input.onchange = (res) => {
        if (input.files?.length <= 0) {
            return
        }
        var data = new FormData()
        data.append('file', input.files[0])
        fetch(`/api/cookie/upload`, {
            method: 'POST',
            body: data
        }).then(async (rsp) => {
            checkFile()
            checkValid()
        })
    }
    input.click()
}

checkFile()
checkValid()
</script>