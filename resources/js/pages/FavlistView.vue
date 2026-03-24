<template>
    <div class="m-4">
        <Breadcrumbs :items="breadcrumbItems">
            <template #actions>
                <div class="flex items-center gap-2">
                    <label class="text-slate-500">{{ t('favorites.downloaded') }}</label>
                    <div class="checkbox-wrapper-7">
                        <input class="tgl tgl-ios" id="cb2-7" type="checkbox" v-model="isFilterDownloaded" />
                        <label class="tgl-btn" for="cb2-7"></label>
                    </div>
                </div>
            </template>
        </Breadcrumbs>

        <SearchBar
            v-if="showSearchPanel"
            ref="searchBarRef"
            class="mt-4 mb-7"
            :model-value="searchQuery"
            :placeholder="'搜索标题 / BV号 / ID（Ctrl+F）'"
            :result-count="searchMatchIndexes.length"
            :current-index="currentSearchIndex"
            :result-found-text="`找到 ${searchMatchIndexes.length} 个结果`"
            :navigate-hint-text="'按 Enter 或 F3 跳转到下一个结果'"
            :no-result-text="'未找到匹配结果'"
            :idle-hint-text="'快捷键：Ctrl+F 打开/关闭，Enter/F3 下一个，Shift+F3 上一个，Esc 关闭'"
            @update:model-value="searchQuery = $event"
            @enter="navigateToNextResult"
            @esc="closeSearchPanel"
            @clear="clearSearchKeyword"
        />
        <div v-if="showSearchPanel" class="search-panel-divider" aria-hidden="true"></div>

        <div v-if="loading" class="fav-scroller flex items-center justify-center">
            <div class="text-center">
                <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
                <p class="mt-4 text-gray-600">{{ t('common.loading') }}</p>
            </div>
        </div>

        <VirtualGroupedList
            v-else
            ref="virtualListRef"
            :items="visibleVideoList"
            :columns="columns"
            :keeps="60"
            :size="rowHeight"
            :container-class="'fav-scroller md:px-4 md:-m-4 py-4 px-2'"
        >
            <template #item="{ record }">
                <div class="grid grid-cols-1 md:grid-cols-4 w-full gap-4 pb-4">
                    <div
                        class="flex flex-col relative"
                        :class="{
                            'search-current': isCurrentMatch(item.id),
                            'search-highlight': isPulseMatch(item.id),
                        }"
                        v-for="item in record.videos"
                        :key="item.id"
                        :data-video-id="item.id"
                    >
                        <RouterLink :to="{ name: 'favlist-video-id', params: { id: id, video_id: item.id } }">
                            <Image
                                :class="[favImageClass, { 'grayscale-image': item.video_downloaded_num == 0 && item.audio_downloaded_num == 0 }]"
                                :src="item.cover_image_url ?? item.cover ?? '/assets/images/notfound.webp'"
                                :title="item.title"
                            />
                        </RouterLink>
                        <div class="absolute top-4 left-4" v-if="item.frozen == 1">💾</div>
                        <span
                            class="mt-4 text-center h-12 line-clamp-2"
                            :title="item.title"
                            v-html="renderTitleWithHighlight(item)"
                        ></span>
                        <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
                            <span>{{ t('favorites.published') }}: {{ formatTimestamp(item.pubtime, "yyyy.mm.dd") }}</span>
                            <span v-if="item.fav_time">
                                {{ t('favorites.favorited') }}: {{ formatTimestamp(item.fav_time, "yyyy.mm.dd") }}
                            </span>
                        </div>
                        <span
                            v-if="item.page > 1"
                            class="text-sm text-white bg-gray-600 rounded-lg min-w-10 px-1.5 text-center absolute top-2 right-2"
                        >
                            {{ item.page }}
                        </span>
                    </div>
                </div>
            </template>
        </VirtualGroupedList>
    </div>
</template>
<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Image from '@/components/Image.vue';
import Breadcrumbs from '@/components/Breadcrumbs.vue';
import SearchBar from '@/components/SearchBar.vue';
import VirtualGroupedList from '@/components/VirtualGroupedList.vue';
import { FAV_IMAGE_CLASS } from '@/constants/videoImageClasses';

import { formatTimestamp as _formatTimestamp } from "../lib/helper"

const formatTimestamp = (value: number | string | null, format: string) => {
    if (!value) return '';
    const ts = typeof value === 'number' ? value : Math.floor(new Date(value).getTime() / 1000);
    return _formatTimestamp(ts, format);
};
import { getFavDetail, getFavVideos, type Favorite, type FavVideo } from '@/api/fav';

const { t } = useI18n();
const route = useRoute();
const id = Number(route.params.id);
const favorite = ref<Favorite | null>(null);
const videoList = ref<FavVideo[]>([]);
const loading = ref(true);
const isFilterDownloaded = ref(false);

const showSearchPanel = ref(false);
const searchQuery = ref('');
const searchBarRef = ref<any>(null);
const currentSearchIndex = ref(-1);
const highlightedVideoId = ref<number | null>(null);
const pulseVideoId = ref<number | null>(null);
const pulseTimer = ref<number | null>(null);
const columns = ref(4);
const rowHeight = ref(260);
const favImageClass = FAV_IMAGE_CLASS;

const updateLayout = () => {
    if (window.innerWidth >= 768) {
        columns.value = 4;
        rowHeight.value = 260;
    } else {
        columns.value = 1;
        rowHeight.value = 320;
    }
};
const virtualListRef = ref<any>(null);
const breadcrumbItems = computed(() => {
    return [
        { text: t('navigation.home'), to: '/' },
        { text: favorite.value?.title ?? t('common.loading') }
    ];
});

const visibleVideoList = computed(() => {
    return videoList.value.filter((value: FavVideo) => {
        if (isFilterDownloaded.value) {
            return value.video_downloaded_num > 0 || value.audio_downloaded_num > 0;
        }
        return true;
    });
});

const searchMatchIndexes = computed(() => {
    const query = searchQuery.value.trim().toLowerCase();
    if (!query) return [];

    const matches: number[] = [];
    visibleVideoList.value.forEach((video, index) => {
        const videoId = String(video.id ?? '');
        const bvid = (video.bvid ?? '').toLowerCase();
        const title = (video.title ?? '').toLowerCase();

        if (title.includes(query) || bvid.includes(query) || videoId.includes(query)) {
            matches.push(index);
        }
    });

    return matches;
});

const openSearchPanel = () => {
    showSearchPanel.value = true;
    nextTick(() => {
        searchBarRef.value?.focusInput?.();
        searchBarRef.value?.selectInput?.();
    });
};

const closeSearchPanel = () => {
    showSearchPanel.value = false;
    searchQuery.value = '';
    currentSearchIndex.value = -1;
    highlightedVideoId.value = null;
    pulseVideoId.value = null;
};

const clearSearchKeyword = () => {
    searchQuery.value = '';
    currentSearchIndex.value = -1;
    highlightedVideoId.value = null;
    pulseVideoId.value = null;
};

const toggleSearchPanelByShortcut = () => {
    if (showSearchPanel.value) {
        closeSearchPanel();
        return;
    }
    openSearchPanel();
};

const isCurrentMatch = (videoId: number | string) => {
    return highlightedVideoId.value === Number(videoId);
};

const isPulseMatch = (videoId: number | string) => {
    return pulseVideoId.value === Number(videoId);
};

const triggerPulseHighlight = (videoId: number) => {
    // 先清空再设置，确保同一个视频也能重播动画
    pulseVideoId.value = null;
    nextTick(() => {
        pulseVideoId.value = videoId;
    });

    if (pulseTimer.value !== null) {
        window.clearTimeout(pulseTimer.value);
    }
    pulseTimer.value = window.setTimeout(() => {
        if (pulseVideoId.value === videoId) {
            pulseVideoId.value = null;
        }
    }, 2200);
};

const escapeHtml = (value: string) => {
    return value
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
};

const renderTitleWithHighlight = (video: FavVideo) => {
    const title = video.title ?? '';
    const safeTitle = escapeHtml(title);
    const keyword = searchQuery.value.trim();

    // 仅对当前命中项做关键词高亮，避免页面噪音
    if (!keyword || !isCurrentMatch(video.id)) {
        return safeTitle;
    }

    const lowerTitle = title.toLowerCase();
    const lowerKeyword = keyword.toLowerCase();
    const start = lowerTitle.indexOf(lowerKeyword);
    if (start === -1) {
        return safeTitle;
    }

    const end = start + keyword.length;
    const before = escapeHtml(title.slice(0, start));
    const match = escapeHtml(title.slice(start, end));
    const after = escapeHtml(title.slice(end));
    return `${before}<mark class="fav-search-mark">${match}</mark>${after}`;
};

const scrollToCurrentSearchResult = () => {
    if (currentSearchIndex.value < 0 || currentSearchIndex.value >= searchMatchIndexes.value.length) {
        return;
    }

    const targetVideoIndex = searchMatchIndexes.value[currentSearchIndex.value];
    const targetVideo = visibleVideoList.value[targetVideoIndex];
    if (!targetVideo) return;
    highlightedVideoId.value = Number(targetVideo.id);

    const targetRowIndex = Math.floor(targetVideoIndex / columns.value);
    if (virtualListRef.value) {
        virtualListRef.value?.scrollToIndex(targetRowIndex);
    }

    nextTick(() => {
        setTimeout(() => {
            const element = document.querySelector(`[data-video-id="${targetVideo.id}"]`) as HTMLElement | null;
            if (!element) return;
            element.scrollIntoView({ behavior: 'smooth', block: 'center' });
            triggerPulseHighlight(Number(targetVideo.id));
        }, 140);
    });
};

const navigateToNextResult = () => {
    if (searchMatchIndexes.value.length === 0) return;
    if (currentSearchIndex.value < 0) {
        currentSearchIndex.value = 0;
    } else {
        currentSearchIndex.value = (currentSearchIndex.value + 1) % searchMatchIndexes.value.length;
    }
    scrollToCurrentSearchResult();
};

const navigateToPrevResult = () => {
    if (searchMatchIndexes.value.length === 0) return;
    if (currentSearchIndex.value <= 0) {
        currentSearchIndex.value = searchMatchIndexes.value.length - 1;
    } else {
        currentSearchIndex.value -= 1;
    }
    scrollToCurrentSearchResult();
};

const handleKeyDown = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') {
        e.preventDefault();
        toggleSearchPanelByShortcut();
        return;
    }

    if (e.key === 'F3' && !e.shiftKey) {
        if (searchQuery.value.trim()) {
            e.preventDefault();
            navigateToNextResult();
        }
    }

    if (e.key === 'F3' && e.shiftKey) {
        if (searchQuery.value.trim()) {
            e.preventDefault();
            navigateToPrevResult();
        }
    }
};

watch([searchQuery, isFilterDownloaded], () => {
    currentSearchIndex.value = -1;
});

onMounted(() => {
    updateLayout();
    window.addEventListener('resize', updateLayout, { passive: true });
    document.addEventListener('keydown', handleKeyDown);
});

onUnmounted(() => {
    window.removeEventListener('resize', updateLayout);
    document.removeEventListener('keydown', handleKeyDown);
    if (pulseTimer.value !== null) {
        window.clearTimeout(pulseTimer.value);
    }
});

getFavDetail(id).then((result) => {
    favorite.value = result;
});

getFavVideos(id).then((result) => {
    videoList.value = result;
}).finally(() => {
    loading.value = false;
});
</script>
<style scoped>
.fav-scroller {
    height: calc(100vh - 180px);
    min-height: 520px;
    overflow-y: auto;
}

.search-panel-divider {
    height: 14px;
    margin-top: -20px;
    margin-bottom: 8px;
    pointer-events: none;
    background: linear-gradient(to bottom, rgba(15, 23, 42, 0.14), rgba(15, 23, 42, 0));
}

@media (max-width: 768px) {
    .fav-scroller {
        height: calc(100vh - 140px);
        min-height: 420px;
    }
}
</style>
<style>
.search-highlight {
    border-radius: 12px;
    animation: search-focus-ring 2200ms cubic-bezier(0.22, 1, 0.36, 1) 1;
}

.search-current {
    border-radius: 12px;
    background: linear-gradient(180deg, rgba(59, 130, 246, 0.14), rgba(59, 130, 246, 0.06));
    box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.52);
}

.fav-search-mark {
    background: #ff9632;
    color: #111827;
    padding: 0 2px;
    border-radius: 3px;
}

@keyframes search-focus-ring {
    0% {
        box-shadow: 0 0 0 0 rgba(59, 130, 246, 0.7), 0 0 0 0 rgba(59, 130, 246, 0.25);
        background: linear-gradient(180deg, rgba(59, 130, 246, 0.24), rgba(59, 130, 246, 0.12));
        transform: translateY(0) scale(1);
    }

    35% {
        box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.82), 0 0 0 10px rgba(59, 130, 246, 0.2);
        background: linear-gradient(180deg, rgba(59, 130, 246, 0.2), rgba(59, 130, 246, 0.08));
        transform: translateY(-2px) scale(1.005);
    }

    100% {
        box-shadow: 0 0 0 0 rgba(59, 130, 246, 0), 0 0 0 0 rgba(59, 130, 246, 0);
        background: linear-gradient(180deg, rgba(59, 130, 246, 0.14), rgba(59, 130, 246, 0.06));
        transform: translateY(0) scale(1);
    }
}
</style>