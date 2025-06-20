<template>

    <div class="flex justify-between">
        <h1 class="my-8 text-2xl">
            <RouterLink to="/">🌸</RouterLink> 收藏夹列表
        </h1>
        <h1 class="my-8 text-2xl">
            <RouterLink to="/progress">🌸</RouterLink> progress
        </h1>
    </div>
    <div class="grid grid-cols-1 md:grid-cols-4 w-full gap-4">
        <div class="flex flex-col relative" v-for="item in favList">
            <RouterLink :to="{ name: 'favlist-id', params: { id: item.id } }">
                <Image class="rounded-lg w-full h-auto md:w-96 md:h-56 hover:scale-105 transition-all duration-300" :src="item.cache_image_url ?? '/assets/images/notfound.webp'"
                    :title="item.title" />
            </RouterLink>
            <span class="mt-4 text-center font-sans" :title="item.title">{{ item.title }}</span>
            <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
                <span>创建: {{ formatTimestamp(item.ctime, "yyyy.mm.dd") }}</span>
                <span>更新: {{ formatTimestamp(item.mtime, "yyyy.mm.dd") }}</span>
            </div>
            <span class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                item.media_count }}</span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import type { Favorite } from '@/api/fav';

const favList = ref<Favorite[]>([]);

import { formatTimestamp, image } from "../lib/helper"
import Image from '@/components/Image.vue';
import { getFavList } from '@/api/fav';


getFavList().then((result) => {
    favList.value = result
})
</script>

<style scoped>
/* 样式代码 */
</style>