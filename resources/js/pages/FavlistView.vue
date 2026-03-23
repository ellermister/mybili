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
            :result-count="stat.count"
            :current-index="-1"
            :result-found-text="`找到 ${stat.count} 个结果`"
            :no-result-text="'未找到匹配结果'"
            :idle-hint-text="'快捷键：Ctrl+F 打开/关闭，Enter 搜索，Esc 关闭'"
            @update:model-value="onSearchInput"
            @enter="triggerSearch"
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
            :items="videoList"
            :columns="columns"
            :keeps="60"
            :size="rowHeight"
            :container-class="'fav-scroller md:px-4 md:-m-4 py-4 px-2'"
        >
            <template #item="{ record }">
                <div class="grid grid-cols-1 md:grid-cols-4 w-full gap-4 pb-4">
                    <div
                        class="flex flex-col relative"
                        v-for="item in record.videos"
                        :key="item.id"
                        :data-video-id="item.id"
                    >
                        <RouterLink :to="{ name: 'favlist-video-id', params: { id: id, video_id: item.id } }">
                            <Image
                                :class="[favImageClass, { 'grayscale-image': item.video_downloaded_num == 0 && item.audio_downloaded_num == 0 }]"
                                :src="item.cover_info?.image_url ?? item.cover ?? '/assets/images/notfound.webp'"
                                :title="item.title"
                            />
                        </RouterLink>
                        <div class="absolute top-4 left-4" v-if="item.frozen == 1">💾</div>
                        <span
                            class="mt-4 text-center h-12 line-clamp-2"
                            :title="item.title"
                            v-html="renderTitle(item)"
                        ></span>
                        <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
                            <span>{{ t('favorites.published') }}: {{ formatTimestamp(item.pubtime, "yyyy.mm.dd") }}</span>
                            <span v-if="item.fav_time">
                                {{ t('favorites.favorited') }}: {{ formatTimestamp(item.fav_time, "yyyy.mm.dd") }}
                            </span>
                        </div>
                        <span
                            v-if="item.page > 1"
                            class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center absolute top-2 right-2"
                        >
                            {{ item.page }}
                        </span>
                    </div>
                </div>
            </template>
        </VirtualGroupedList>

        <div v-if="!loading && videoList.length > 0" class="text-center py-4">
            <div v-if="isLoadingMore" class="flex items-center justify-center gap-2">
                <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
                <span class="text-gray-600">{{ t('common.loading') }}</span>
            </div>
            <div v-else-if="!hasMore" class="text-gray-500 text-sm">
                {{ `共 ${stat.count} 个视频` }}
            </div>
        </div>
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

import { formatTimestamp } from "../lib/helper"
import { getFavDetail, type Favorite, type Video } from '@/api/fav';
import { getVideoList } from '@/api/video';

const { t } = useI18n();
const route = useRoute();
const id = Number(route.params.id);
const favMeta = ref<Favorite | null>(null);
const isFilterDownloaded = ref(false);

const showSearchPanel = ref(false);
const searchQuery = ref('');
const searchBarRef = ref<any>(null);
const columns = ref(4);
const rowHeight = ref(260);
const favImageClass = FAV_IMAGE_CLASS;

const videoList = ref<Video[]>([]);
const currentPage = ref(1);
const hasMore = ref(true);
const loading = ref(true);
const isLoadingMore = ref(false);
const searchTimeout = ref<ReturnType<typeof setTimeout> | null>(null);

const stat = ref({
    count: 0,
    downloaded: 0,
    invalid: 0,
    valid: 0,
    frozen: 0,
});

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
        { text: favMeta.value?.title ?? t('common.loading') }
    ];
});

const loadVideos = async (isReset = false) => {
    if (!isReset && (isLoadingMore.value || !hasMore.value)) return;

    isReset ? (loading.value = true) : (isLoadingMore.value = true);

    try {
        const data = await getVideoList({
            fav_id: String(id),
            sort: 'fav_time',
            page: currentPage.value,
            query: searchQuery.value.trim(),
            downloaded: isFilterDownloaded.value ? 'yes' : '',
        });

        const newVideos = data.list ?? [];

        if (newVideos.length === 0) {
            hasMore.value = false;
        } else {
            videoList.value = isReset ? newVideos : [...videoList.value, ...newVideos];
        }

        if (isReset || currentPage.value === 1) {
            stat.value = data.stat ?? stat.value;
        }
    } catch (error) {
        console.error('Failed to load videos:', error);
    } finally {
        loading.value = false;
        isLoadingMore.value = false;
    }
};

const resetAndLoad = () => {
    currentPage.value = 1;
    hasMore.value = true;
    videoList.value = [];
    loadVideos(true);
};

const loadNextPage = () => {
    if (hasMore.value && !isLoadingMore.value && !loading.value) {
        currentPage.value++;
        loadVideos(false);
    }
};

const handleScroll = (e: Event) => {
    const el = e.target as HTMLElement;
    if (el.scrollTop + el.clientHeight >= el.scrollHeight - 300) {
        loadNextPage();
    }
};

const attachScrollListener = () => {
    nextTick(() => {
        const scroller = document.querySelector('.fav-scroller');
        if (scroller) {
            scroller.addEventListener('scroll', handleScroll, { passive: true });
        }
    });
};

const detachScrollListener = () => {
    const scroller = document.querySelector('.fav-scroller');
    if (scroller) {
        scroller.removeEventListener('scroll', handleScroll);
    }
};

// 搜索相关
const onSearchInput = (value: string) => {
    searchQuery.value = value;
    if (searchTimeout.value) clearTimeout(searchTimeout.value);
    searchTimeout.value = setTimeout(() => {
        resetAndLoad();
    }, 500);
};

const triggerSearch = () => {
    if (searchTimeout.value) clearTimeout(searchTimeout.value);
    resetAndLoad();
};

const openSearchPanel = () => {
    showSearchPanel.value = true;
    nextTick(() => {
        searchBarRef.value?.focusInput?.();
        searchBarRef.value?.selectInput?.();
    });
};

const closeSearchPanel = () => {
    showSearchPanel.value = false;
    if (searchQuery.value) {
        searchQuery.value = '';
        resetAndLoad();
    }
};

const clearSearchKeyword = () => {
    searchQuery.value = '';
    resetAndLoad();
};

const toggleSearchPanelByShortcut = () => {
    if (showSearchPanel.value) {
        closeSearchPanel();
        return;
    }
    openSearchPanel();
};

const escapeHtml = (value: string) => {
    return value
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
};

const renderTitle = (video: Video) => {
    const title = video.title ?? '';
    const safeTitle = escapeHtml(title);
    const keyword = searchQuery.value.trim();

    if (!keyword) return safeTitle;

    const lowerTitle = title.toLowerCase();
    const lowerKeyword = keyword.toLowerCase();
    const start = lowerTitle.indexOf(lowerKeyword);
    if (start === -1) return safeTitle;

    const end = start + keyword.length;
    const before = escapeHtml(title.slice(0, start));
    const match = escapeHtml(title.slice(start, end));
    const after = escapeHtml(title.slice(end));
    return `${before}<mark class="fav-search-mark">${match}</mark>${after}`;
};

const handleKeyDown = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') {
        e.preventDefault();
        toggleSearchPanelByShortcut();
    }
};

watch(isFilterDownloaded, () => {
    resetAndLoad();
});

// 监听 loading 变化，在数据加载完成后挂载滚动监听
watch(loading, (newVal, oldVal) => {
    if (oldVal && !newVal) {
        attachScrollListener();
    }
});

onMounted(() => {
    updateLayout();
    window.addEventListener('resize', updateLayout, { passive: true });
    document.addEventListener('keydown', handleKeyDown);
});

onUnmounted(() => {
    window.removeEventListener('resize', updateLayout);
    document.removeEventListener('keydown', handleKeyDown);
    detachScrollListener();
    if (searchTimeout.value) clearTimeout(searchTimeout.value);
});

// 加载元信息
getFavDetail(id).then((result) => {
    favMeta.value = result;
});

// 加载视频列表（分页）
loadVideos(true);
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
.fav-search-mark {
    background: #ff9632;
    color: #111827;
    padding: 0 2px;
    border-radius: 3px;
}
</style>
