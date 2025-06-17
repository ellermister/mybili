<template>
    <h1 class="my-8 text-2xl flex-b flex justify-between ">
        <div>
            <RouterLink to="/">🌸</RouterLink> my fav {{ $route.params.id }}
        </div>
        <div class="flex items-center  gap-2">
            <label class="text-slate-500">Valid</label>
            <div class="checkbox-wrapper-7">
                <input class="tgl tgl-ios" id="cb2-7" type="checkbox" v-model="isFilterValid" />
                <label class="tgl-btn" for="cb2-7"></label>
            </div>

        </div>
    </h1>

    <div class="grid grid-cols-1 md:grid-cols-4 w-full gap-4">
        <div class="flex flex-col relative" v-for="item in showVideoList">
            <RouterLink :to="{ name: 'video-id', params: { id: item.id } }">
                <Image class="rounded-lg w-full h-auto md:w-96 md:h-56 hover:scale-105 transition-all duration-300" :src="item.cache_image_url ?? '/assets/images/notfound.webp'"
                    :title="item.title" />
            </RouterLink>
            <div class="absolute top-4 left-4" v-if="item.backup">💾</div>
            <span class="mt-4 text-center">{{ item.title }}</span>
            <span class="text-sm">发布时间:{{ formatTimestamp(item.pubtime, "yyyy-mm-dd") }}</span>
            <span class="text-sm">收藏时间:{{ formatTimestamp(item.fav_time, "yyyy-mm-dd") }}</span>
            <span class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                item.media_count }}</span>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import Image from '@/components/Image.vue';

import { formatTimestamp, image } from "../lib/helper"

const route = useRoute()
const id = route.params.id
const videoList = ref([])

const isFilterValid = ref(false)

const showVideoList = computed(()=>{
    return videoList.value.filter((value)=>{
        if(isFilterValid.value){
            return value.invalid == false;
        }
        return true;
    })
})


fetch('/api/fav/' + id).then(async (response) => {
    const result = await response.json()
    videoList.value = result
})</script>
<style scoped>

</style>