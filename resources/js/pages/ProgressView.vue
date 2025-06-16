<template>
    <div value="w-full flex justify-center	">
        <div class="container  mx-auto justify-center " id="main">
            <div class="m-4">
                <div class="flex justify-between">
                    <h1 class="my-8 text-2xl">
                        <RouterLink to="/">🌸</RouterLink> progress {{ $route.params.id }}
                    </h1>
                    <h1 class="my-8 text-2xl">
                        <RouterLink to="/horizon" target="_blank">🔭</RouterLink> 查看任务
                    </h1>
                </div>
                <h2 class="text-xl" title="如果你的收藏夹中出现了无效视频那么就会低于100%">缓存的视频率 {{ progress }}% ({{ stat.downloaded
                    }}/{{ stat.count }})</h2>


                <div class="my-8 w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                    <div class="bg-blue-600 h-2.5 rounded-full" :style="{ width: progress + '%' }"></div>
                </div>

                <div class="grid grid-cols-1 md:grid-cols-4 w-full my-4 ">
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500  py-4 rounded-l-lg"
                        :class="{ 'bg-gradient-to-r': filter.class == null }" @click="filter.class = null">
                        <span class="text-2xl" title="你所有收藏的视频数">所有视频</span>
                        <span class="text-xl font-semibold">{{ stat.count }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4"
                        :class="{ 'bg-gradient-to-r': filter.class == 'valid' }" @click="filter.class = 'valid'">
                        <span class="text-2xl" title="目前仍可以在线观看的视频">有效视频</span>
                        <span class="text-xl font-semibold">{{ stat.valid }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4"
                        :class="{ 'bg-gradient-to-r': filter.class == 'invalid' }" @click="filter.class = 'invalid'">
                        <span class="text-2xl" title="收藏的视频无效被下架">无效视频</span>
                        <span class="text-xl font-semibold">{{ stat.invalid }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4 rounded-r-lg"
                        :class="{ 'bg-gradient-to-r': filter.class == 'frozen' }" @click="filter.class = 'frozen'">
                        <span class="text-2xl" title="当你收藏的视频缓存了之后, 如果视频被删除下架那么就会将该视频归纳为冻结">冻结视频</span>
                        <span class="text-xl font-semibold">{{ stat.frozen }}</span>
                    </div>
                </div>


                <div class="mt-4 grid grid-cols-1 md:grid-cols-4 w-full gap-4">
                    <div class="flex flex-col relative" v-for="item in dataList">
                        <RouterLink :to="{ name: 'video-id', params: { id: item.id } }">
                            <Image class="rounded-lg w-full h-auto md:w-96 md:h-56" :src="item.cache_image_url"
                                :class="{ 'grayscale-image': item.video_downloaded_num == 0 }" :title="item.title" />
                        </RouterLink>
                        <span class="mt-4 text-center">{{ item.title }}</span>
                        <span class="text-sm">发布时间:{{ formatTimestamp(item.pubtime, "yyyy-mm-dd") }}</span>
                        <span class="text-sm">收藏时间:{{ formatTimestamp(item.fav_time, "yyyy-mm-dd") }}</span>
                        <span v-if="item.page > 1"
                            class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                                item.page }}</span>
                    </div>
                </div>
            </div>

        </div>
    </div>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import Image from '@/components/Image.vue';
import { formatTimestamp, image } from "../lib/helper"
const videoList = ref<VideoType[]>([])
const progress = ref(0)

const stat = ref({
    count: 0,
    downloaded: 0,
    invalid: 0,
    valid: 0,
    frozen: 0,
})

const filter = ref<{
    class: null | string
}>({
    class: null
})

interface VideoType {
    id: string
    title: string
    cache_image: string
    cache_image_url: string
    video_downloaded_at: string
    invalid: boolean
    frozen: boolean
    pubtime: number
    fav_time: number
    page: number
    video_downloaded_num: number
}


const dataList = computed(() => {
    return videoList.value.filter(i => {
        if (filter.value.class == null) {
            return true
        }

        if (filter.value.class == 'invalid' && i.invalid) {
            return true
        } else if (filter.value.class == 'valid' && !i.invalid) {
            return true
        } else if (filter.value.class == 'frozen' && i.frozen) {
            return true
        }

        return false;
    })
})

fetch(`/api/progress`).then(async (rsp) => {
    if (rsp.ok) {
        const jsonData = await rsp.json()
        videoList.value = jsonData.data
        stat.value = jsonData.stat

        progress.value = parseInt((stat.value.downloaded / stat.value.count * 100).toFixed(2))
    }
})
</script>

<style scoped>
.grayscale-image {
    filter: grayscale(100%) brightness(80%);
}
</style>