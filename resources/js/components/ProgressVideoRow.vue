<template>
    <div class="virtual-row grid grid-cols-1 md:grid-cols-4 w-full gap-4 pb-4">
        <div class="flex flex-col relative" v-for="video in source.videos" :key="video.id" :data-video-id="video.id">
            <RouterLink :to="{ name: 'video-id', params: { id: video.id } }">
                <div class="image-container rounded-lg overflow-hidden" :style="{
                    aspectRatio: '4/3'
                }">
                    <ImagePreload  :class="[imageClass, { 'grayscale-image': video.video_downloaded_num == 0 && video.audio_downloaded_num == 0 }]"
                        :src="resolveCoverUrl(video)"
                        :thumbSrc="resolveCoverUrl(video, true)"
                        :title="video.title" />
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
import ImagePreload from './ImagePreload.vue';
import { PROGRESS_IMAGE_CLASS } from '../constants/videoImageClasses';
import type { ProgressVideo } from '../api/fav';

interface ProgressVideoRowData {
    id: string;
    videos: ProgressVideo[];
}

const props = defineProps<{
    source: ProgressVideoRowData
    imageClass?: string
}>()
// console.log(props.source);
const { t } = useI18n();
const imageClass = props.imageClass ?? PROGRESS_IMAGE_CLASS;

const resolveCoverUrl = (video: ProgressVideo, isThumb = false) => {
    const origin = video.cover_image_url ?? video.cover ?? '/assets/images/notfound.webp';
    if (!origin || origin.includes('notfound')) {
        return '/assets/images/notfound.webp';
    }
    const url = isThumb ? video.cover_image_thumb_url : origin;
    return url;
}
</script>
<style scoped>

</style>