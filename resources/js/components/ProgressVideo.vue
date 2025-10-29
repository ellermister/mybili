<template>
    <div class="flex flex-col relative">
        <RouterLink :to="{ name: 'video-id', params: { id: item.id } }">
            <Image class="rounded-lg w-full h-auto md:w-96 md:h-56 hover:scale-105 transition-all duration-300"
                :src="item.cover_info?.image_url ?? '/assets/images/notfound.webp'"
                :class="{ 'grayscale-image': item.video_downloaded_num == 0 }" :title="item.title" />
        </RouterLink>
        <span class="mt-4 text-center h-12 line-clamp-2" :title="item.title">{{ item.title }}</span>
        <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
            <span>{{ t('progress.published') }}: {{ formatTimestamp(item.pubtime, "yyyy.mm.dd") }}</span>
            <span v-if="item.fav_time > 0">{{ t('progress.favorited') }}: {{ formatTimestamp(item.fav_time,
                "yyyy.mm.dd") }}</span>
        </div>
        <span v-if="item.page > 1"
            class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                item.page }}</span>
    </div>
</template>
<script lang="ts" setup>
import { defineProps } from 'vue';
import Image from './Image.vue';
import { formatTimestamp } from "../lib/helper"
import { useI18n } from 'vue-i18n';
import type { Video } from '../api/fav';

const { t } = useI18n();

const props = defineProps<{
    item: Video;
}>();
</script>