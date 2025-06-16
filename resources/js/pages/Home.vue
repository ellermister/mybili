<template>

    <div class="flex justify-between">
        <h1 class="my-8 text-2xl">
            <RouterLink to="/">🌸</RouterLink> my fav
        </h1>
        <h1 class="my-8 text-2xl">
            <RouterLink to="/progress">🌸</RouterLink> progress
        </h1>
    </div>
    <div class="grid grid-cols-1 md:grid-cols-4 w-full gap-4">
        <div class="flex flex-col relative" v-for="item in favList">
            <RouterLink :to="{ name: 'favlist-id', params: { id: item.id } }">
                <Image class="rounded-lg w-full h-auto md:w-96 md:h-56" :src="item.cache_image_url"
                    :title="item.title" />
            </RouterLink>
            <span class="mt-4 text-center">{{ item.title }}</span>
            <span class="text-sm">创建时间:{{ formatTimestamp(item.ctime, "yyyy-mm-dd") }}</span>
            <span class="text-sm">更新时间:{{ formatTimestamp(item.mtime, "yyyy-mm-dd") }}</span>
            <span class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                item.media_count }}</span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
const favList = ref([])

import { formatTimestamp, image } from "../lib/helper"
import Image from '@/components/Image.vue';

fetch('/api/fav').then(async (response) => {
    const result = await response.json()
    favList.value = result
})
</script>

<style scoped>
/* 样式代码 */
</style>