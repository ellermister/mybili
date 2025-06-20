<template>
    <Breadcrumbs :items="breadcrumbItems">
        <template #actions>
            <div class="flex items-center  gap-2">
                <label class="text-slate-500">Valid</label>
                <div class="checkbox-wrapper-7">
                    <input class="tgl tgl-ios" id="cb2-7" type="checkbox" v-model="isFilterValid" />
                    <label class="tgl-btn" for="cb2-7"></label>
                </div>
            </div>
        </template>
    </Breadcrumbs>

    <div class="grid grid-cols-1 md:grid-cols-4 w-full gap-4">
        <div class="flex flex-col relative" v-for="item in showVideoList">
            <RouterLink :to="{ name: 'video-id', params: { id: item.id } }">
                <Image class="rounded-lg w-full h-auto md:w-96 md:h-56 hover:scale-105 transition-all duration-300"
                    :src="item.cache_image_url ?? '/assets/images/notfound.webp'" :title="item.title" />
            </RouterLink>
            <div class="absolute top-4 left-4" v-if="item.frozen == 1">💾</div>
            <span class="mt-4 text-center  h-12 line-clamp-2" :title="item.title">{{ item.title }}</span>
            <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
                <span>发布: {{ formatTimestamp(item.pubtime, "yyyy.mm.dd") }}</span>
                <span>收藏: {{ formatTimestamp(item.fav_time, "yyyy.mm.dd") }}</span>
            </div>
            <span v-if="item.page > 1"
                class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                    item.page }}</span>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import Image from '@/components/Image.vue';
import Breadcrumbs from '@/components/Breadcrumbs.vue';

import { formatTimestamp, image } from "../lib/helper"
import { getFavDetail, type Favorite, type Video } from '@/api/fav';

const route = useRoute()
const id = route.params.id
const favorite = ref<Favorite | null>(null)

const isFilterValid = ref(false)

const breadcrumbItems = computed(() => {
    return [
        { text: '首页', to: '/' },
        { text: favorite.value?.title ?? '加载中...' }
    ]
})

const showVideoList = computed(() => {
    return (favorite.value?.videos ?? []).filter((value: Video) => {
        if (isFilterValid.value) {
            return value.invalid == 0;
        }
        return true;
    })
})


getFavDetail(Number(id)).then((result) => {
    favorite.value = result
})
</script>
<style scoped></style>