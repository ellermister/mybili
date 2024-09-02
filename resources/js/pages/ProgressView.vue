<template>
    <div value="w-full flex justify-center	">
        <div class="container  mx-auto justify-center p-4 md:p-16" id="main">
            <div class="m-4">
                <h1 class="my-8 text-2xl">
                    <RouterLink to="/">🌸</RouterLink> progress {{ $route.params.id }}
                </h1>
                <h2 class="text-xl">下载进度 {{ progress }}% ({{downloaded}}/{{ count }})</h2>
                <div class="my-8 w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                    <div class="bg-blue-600 h-2.5 rounded-full" :style="{width: progress+'%'}"></div>
                </div>

                <div class="mt-4 grid grid-cols-1 md:grid-cols-4 w-full gap-4">
                    <div class="flex flex-col relative" v-for="item in videoList">
                        <RouterLink :to="{ name: 'video-id', params: { id: item.id } }">
                            <Image class="rounded-lg w-full h-auto md:w-96 md:h-56" :src="image(item.cache_image)"
                            :class="{'grayscale-image': !item.downloaded}"
                                :title="item.title" />
                        </RouterLink>
                        <span class="mt-4 text-center">{{ item.title }}</span>
                        <span class="text-sm">发布时间:{{ formatTimestamp(item.pubtime, "yyyy-mm-dd") }}</span>
                        <span class="text-sm">收藏时间:{{ formatTimestamp(item.fav_time, "yyyy-mm-dd") }}</span>
                        <span
                            class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                                item.media_count }}</span>
                    </div>
                </div>
            </div>

        </div>
    </div>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import Image from '@/components/Image.vue';
import { formatTimestamp, image } from "../lib/helper"
const videoList = ref([])
const count = ref(0)
const downloaded = ref(0)
const progress = ref(0)

fetch(`/api/progress`).then(async (rsp) => {
    if (rsp.ok) {
        const jsonData = await rsp.json()
        videoList.value = jsonData.data
        count.value = jsonData.count
        downloaded.value = jsonData.downloaded
        progress.value = parseInt((downloaded.value / count.value *100).toFixed(2))
    }
})
</script>

<style scoped>
.grayscale-image {
  filter: grayscale(100%) brightness(80%);
}
</style>