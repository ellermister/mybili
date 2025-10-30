<template>
    <div class="virtual-row grid grid-cols-1 md:grid-cols-4 w-full gap-4 pb-4">
        <div class="flex flex-col relative" v-for="video in source.videos" :key="video.id" :data-video-id="video.id">
            <RouterLink :to="{ name: 'video-id', params: { id: video.id } }">
                <div class="image-container rounded-lg overflow-hidden" :style="{
                    aspectRatio: '4/3'
                }">
                    <Image class="w-full h-full object-cover hover:scale-105 transition-all duration-300"
                        :src="video.cover_info?.image_url ?? '/assets/images/notfound.webp'"
                        :class="{ 'grayscale-image': video.video_downloaded_num == 0 }" :title="video.title" />
                </div>
            </RouterLink>
            <span class="mt-4 text-center h-12 line-clamp-2" :title="video.title">{{ video.title
            }}</span>
            <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
                <span>{{ t('progress.published') }}: {{ formatTimestamp(video.pubtime,
                    "yyyy.mm.dd") }}</span>
                <span v-if="video.fav_time > 0">{{ t('progress.favorited') }}: {{
                    formatTimestamp(video.fav_time, "yyyy.mm.dd") }}</span>
            </div>
            <span v-if="video.page > 1"
                class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center absolute top-2 right-2">{{
                    video.page }}</span>
        </div>
    </div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { formatTimestamp } from "../lib/helper"
import Image from './Image.vue';
const props = defineProps<{
    source: any
}>()
// console.log(props.source);
const { t } = useI18n();
</script>
<style scoped>

</style>