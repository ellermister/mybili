<template>
    <div value="w-full flex justify-center	">
        <div class="container  mx-auto justify-center " id="main">
            <div class="m-4">

                <div class="flex justify-between">
                    <h1 class="my-8 text-2xl mt-0">
                        <RouterLink to="/">🌸</RouterLink> {{ t('progress.title') }} {{ $route.params.id }}
                    </h1>
                    <div class="flex gap-4 my-8 text-2xl mt-0 items-center">
                        <!-- 移动端搜索按钮 -->
                        <button @click="showMobileSearch = true"
                            class="md:hidden text-2xl hover:text-blue-600 transition-colors"
                            :title="t('progress.searchButtonTitle')">
                            🔍
                        </button>
                        <!-- PC端搜索按钮 -->
                        <button @click="openDesktopSearch"
                            class="hidden md:block text-2xl hover:text-blue-600 transition-colors"
                            :title="t('progress.searchButtonTitlePC')">
                            🔍
                        </button>
                        <RouterLink to="/videos" class="hover:text-blue-600 transition-colors">
                            🎬<span class="hidden md:inline"> {{ t('navigation.videoManagement') }}</span>
                        </RouterLink>
                        <RouterLink to="/horizon" target="_blank" class="hover:text-blue-600 transition-colors">
                            🔭<span class="hidden md:inline"> {{ t('progress.viewTasks') }}</span>
                        </RouterLink>
                    </div>
                </div>

                <!-- PC端搜索框 -->
                <div v-if="showDesktopSearch" class="hidden md:block mb-4">
                    <div class="relative">
                        <input ref="searchInputRef" v-model="searchQuery" type="text"
                            :placeholder="t('progress.searchPlaceholder')"
                            class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            @keydown.esc="closeSearch" @keydown.enter.prevent="navigateToNextResult" />
                        <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400">🔍</span>
                        <button v-if="searchQuery" @click="clearSearch"
                            class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600">
                            ✕
                        </button>
                    </div>
                    <div v-if="searchQuery && searchResults.length > 0"
                        class="mt-2 text-sm text-gray-600 flex items-center gap-2">
                        <span>{{ t('progress.searchResultsFound', { count: searchResults.length }) }}</span>
                        <span v-if="currentSearchIndex >= 0" class="text-blue-600 font-semibold">
                            ({{ currentSearchIndex + 1 }}/{{ searchResults.length }})
                        </span>
                        <span class="text-xs text-gray-500">{{ t('progress.searchNavigateHint') }}</span>
                    </div>
                    <div v-if="searchQuery && searchResults.length === 0" class="mt-2 text-sm text-red-600">
                        {{ t('progress.searchNoResults') }}
                    </div>
                </div>

                <!-- 移动端搜索框 -->
                <div v-if="showMobileSearch" class="md:hidden mb-4">
                    <div class="relative">
                        <input ref="mobileSearchInputRef" v-model="searchQuery" type="text"
                            :placeholder="t('progress.searchPlaceholderMobile')"
                            class="w-full pl-10 pr-10 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            @keydown.esc="closeSearch" @keydown.enter.prevent="navigateToNextResult" />
                        <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400">🔍</span>
                        <button @click="closeSearch"
                            class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600">
                            ✕
                        </button>
                    </div>
                    <div v-if="searchQuery && searchResults.length > 0"
                        class="mt-2 text-sm text-gray-600 flex items-center gap-2 flex-wrap">
                        <span>{{ t('progress.searchResultsFound', { count: searchResults.length }) }}</span>
                        <span v-if="currentSearchIndex >= 0" class="text-blue-600 font-semibold">
                            ({{ currentSearchIndex + 1 }}/{{ searchResults.length }})
                        </span>
                    </div>
                    <div v-if="searchQuery && searchResults.length === 0" class="mt-2 text-sm text-red-600">
                        {{ t('progress.searchNoResults') }}
                    </div>
                </div>

                <div class="flex flex-col md:flex-row md:justify-between gap-4 md:gap-0">
                    <h2 class="text-xl" :title="t('progress.cacheRateDescription')">{{ t('progress.cacheRate') }} {{
                        progress }}% ({{ stat.downloaded
                        }}/{{ stat.count }})</h2>

                    <div class="flex items-center gap-2">
                        <button @click="showCachedOnly = !showCachedOnly" :class="[
                            'relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
                            showCachedOnly ? 'bg-blue-600' : 'bg-gray-200'
                        ]" role="switch" :aria-checked="showCachedOnly">
                            <span :class="[
                                'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
                                showCachedOnly ? 'translate-x-6' : 'translate-x-1'
                            ]" />
                        </button>
                        <label class="text-sm text-gray-700 cursor-pointer" @click="showCachedOnly = !showCachedOnly">
                            {{ t('progress.showCachedOnly') }}
                        </label>
                    </div>
                </div>



                <div class="my-8 w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                    <div class="bg-blue-600 h-2.5 rounded-full" :style="{ width: progress + '%' }"></div>
                </div>

                <!-- 桌面端筛选器 -->
                <div class="hidden md:grid grid-cols-4 w-full my-4">
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500  py-4 rounded-l-lg"
                        :class="{ 'bg-gradient-to-r': filter.class == null }" @click="setFilter(null)">
                        <span class="text-2xl" :title="t('progress.allVideosDescription')">{{ t('progress.allVideos')
                            }}</span>
                        <span class="text-xl font-semibold">{{ stat.count }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4"
                        :class="{ 'bg-gradient-to-r': filter.class == 'valid' }" @click="setFilter('valid')">
                        <span class="text-2xl" :title="t('progress.validVideosDescription')">{{
                            t('progress.validVideos') }}</span>
                        <span class="text-xl font-semibold">{{ stat.valid }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4"
                        :class="{ 'bg-gradient-to-r': filter.class == 'invalid' }" @click="setFilter('invalid')">
                        <span class="text-2xl" :title="t('progress.invalidVideosDescription')">{{
                            t('progress.invalidVideos') }}</span>
                        <span class="text-xl font-semibold">{{ stat.invalid }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4 rounded-r-lg"
                        :class="{ 'bg-gradient-to-r': filter.class == 'frozen' }" @click="setFilter('frozen')">
                        <span class="text-2xl" :title="t('progress.frozenVideosDescription')">{{
                            t('progress.frozenVideos') }}</span>
                        <span class="text-xl font-semibold">{{ stat.frozen }}</span>
                    </div>
                </div>

                <!-- 移动端筛选器 -->
                <div ref="filterRef" class="md:hidden w-full transition-all duration-300"
                    :class="isScrolled ? 'sticky top-0 z-50 my-0' : 'my-4'">
                    <!-- 完整模式：当前选中的筛选器显示 -->
                    <div v-if="!isScrolled"
                        class="mb-4 p-4 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg text-white shadow-lg">
                        <div class="flex items-center space-x-3">
                            <div class="w-3 h-3 bg-white rounded-full"></div>
                            <div>
                                <div class="text-lg font-semibold">
                                    {{ getCurrentFilterLabel() }}
                                </div>
                                <div class="text-sm opacity-90">
                                    {{ getCurrentFilterCount() }} 个视频
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- 筛选器选项 - 完整模式 -->
                    <div v-if="!isScrolled" class="grid grid-cols-2 gap-3">
                        <div class="bg-white rounded-lg p-4 shadow-sm border-2 transition-all"
                            :class="filter.class == null ? 'border-blue-500 bg-blue-50' : 'border-gray-200 hover:border-gray-300'"
                            @click="setFilter(null)">
                            <div class="text-center">
                                <div class="text-2xl mb-1">📺</div>
                                <div class="text-sm font-medium text-gray-700">{{ t('progress.allVideos') }}</div>
                                <div class="text-lg font-bold text-gray-900">{{ stat.count }}</div>
                            </div>
                        </div>
                        <div class="bg-white rounded-lg p-4 shadow-sm border-2 transition-all"
                            :class="filter.class == 'valid' ? 'border-green-500 bg-green-50' : 'border-gray-200 hover:border-gray-300'"
                            @click="setFilter('valid')">
                            <div class="text-center">
                                <div class="text-2xl mb-1">✅</div>
                                <div class="text-sm font-medium text-gray-700">{{ t('progress.validVideos') }}</div>
                                <div class="text-lg font-bold text-gray-900">{{ stat.valid }}</div>
                            </div>
                        </div>
                        <div class="bg-white rounded-lg p-4 shadow-sm border-2 transition-all"
                            :class="filter.class == 'invalid' ? 'border-red-500 bg-red-50' : 'border-gray-200 hover:border-gray-300'"
                            @click="setFilter('invalid')">
                            <div class="text-center">
                                <div class="text-2xl mb-1">❌</div>
                                <div class="text-sm font-medium text-gray-700">{{ t('progress.invalidVideos') }}</div>
                                <div class="text-lg font-bold text-gray-900">{{ stat.invalid }}</div>
                            </div>
                        </div>
                        <div class="bg-white rounded-lg p-4 shadow-sm border-2 transition-all"
                            :class="filter.class == 'frozen' ? 'border-orange-500 bg-orange-50' : 'border-gray-200 hover:border-gray-300'"
                            @click="setFilter('frozen')">
                            <div class="text-center">
                                <div class="text-2xl mb-1">🧊</div>
                                <div class="text-sm font-medium text-gray-700">{{ t('progress.frozenVideos') }}</div>
                                <div class="text-lg font-bold text-gray-900">{{ stat.frozen }}</div>
                            </div>
                        </div>
                    </div>

                    <!-- 最小模式：Tab样式筛选器 -->
                    <div v-else class="bg-white border-b border-gray-300 shadow-md">
                        <div class="grid grid-cols-4 divide-x divide-gray-300">
                            <button
                                class="flex flex-col items-center justify-center py-3 px-1 transition-all duration-200 relative"
                                :class="filter.class == null ? 'bg-blue-50 text-blue-600' : 'text-gray-600 hover:text-gray-900 active:bg-gray-50'"
                                @click="setFilter(null)">
                                <span class="text-xl leading-none">📺</span>
                                <span v-if="filter.class == null"
                                    class="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-600"></span>
                            </button>
                            <button
                                class="flex flex-col items-center justify-center py-3 px-1 transition-all duration-200 relative"
                                :class="filter.class == 'valid' ? 'bg-green-50 text-green-600' : 'text-gray-600 hover:text-gray-900 active:bg-gray-50'"
                                @click="setFilter('valid')">
                                <span class="text-xl leading-none">✅</span>
                                <span v-if="filter.class == 'valid'"
                                    class="absolute bottom-0 left-0 right-0 h-0.5 bg-green-600"></span>
                            </button>
                            <button
                                class="flex flex-col items-center justify-center py-3 px-1 transition-all duration-200 relative"
                                :class="filter.class == 'invalid' ? 'bg-red-50 text-red-600' : 'text-gray-600 hover:text-gray-900 active:bg-gray-50'"
                                @click="setFilter('invalid')">
                                <span class="text-xl leading-none">❌</span>
                                <span v-if="filter.class == 'invalid'"
                                    class="absolute bottom-0 left-0 right-0 h-0.5 bg-red-600"></span>
                            </button>
                            <button
                                class="flex flex-col items-center justify-center py-3 px-1 transition-all duration-200 relative"
                                :class="filter.class == 'frozen' ? 'bg-orange-50 text-orange-600' : 'text-gray-600 hover:text-gray-900 active:bg-gray-50'"
                                @click="setFilter('frozen')">
                                <span class="text-xl leading-none">🧊</span>
                                <span v-if="filter.class == 'frozen'"
                                    class="absolute bottom-0 left-0 right-0 h-0.5 bg-orange-600"></span>
                            </button>
                        </div>
                    </div>
                </div>

                <DynamicScroller class="mt-4 w-full scroller-container" :items="groupedDataList" :min-item-size="380"
                    key-field="id" :buffer="200" :emit-update="true" #default="{ item, active, index }">
                    <DynamicScrollerItem :item="item" :active="active" :data-index="index">
                        <template #default>
                            <div class="virtual-row grid grid-cols-1 md:grid-cols-4 w-full gap-4 pb-4">
                                <div class="flex flex-col relative" v-for="video in item.videos" :key="video.id"
                                    :data-video-id="video.id">
                                    <RouterLink :to="{ name: 'video-id', params: { id: video.id } }">
                                        <div class="image-container rounded-lg overflow-hidden" :style="{
                                            aspectRatio: '4/3'
                                        }">
                                            <Image
                                                class="w-full h-full object-cover hover:scale-105 transition-all duration-300"
                                                :src="video.cover_info?.image_url ?? '/assets/images/notfound.webp'"
                                                :class="{ 'grayscale-image': video.video_downloaded_num == 0 }"
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
                    </DynamicScrollerItem>
                </DynamicScroller>
            </div>

        </div>
    </div>
</template>
<script lang="ts" setup>
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { DynamicScroller, DynamicScrollerItem } from 'vue-virtual-scroller';
import Image from '@/components/Image.vue';
import { formatTimestamp, image } from "../lib/helper"
import type { Cover } from '../api/cover';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const videoGridRef = ref<HTMLElement>();

const videoList = ref<VideoType[]>([])
const progress = ref(0)
const showCachedOnly = ref(false)
const isScrolled = ref(false) // 是否已滚动
const filterRef = ref<HTMLElement>() // 筛选器元素的引用
const searchQuery = ref('') // 搜索关键词
const showMobileSearch = ref(false) // 移动端搜索框显示状态
const showDesktopSearch = ref(false) // PC端搜索框显示状态
const searchInputRef = ref<HTMLInputElement>() // PC端搜索输入框引用
const mobileSearchInputRef = ref<HTMLInputElement>() // 移动端搜索输入框引用
const currentSearchIndex = ref(-1) // 当前搜索结果索引

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
    video_downloaded_at: string
    invalid: boolean
    frozen: boolean
    pubtime: number
    fav_time: number
    page: number
    video_downloaded_num: number
    cover_info: Cover | null
}

// 从URL参数初始化过滤器状态
const initFilterFromUrl = () => {
    const filterParam = route.query.filter as string;
    if (filterParam && ['valid', 'invalid', 'frozen'].includes(filterParam)) {
        filter.value.class = filterParam;
    } else {
        filter.value.class = null;
    }
}

// 设置过滤器并更新URL
const setFilter = (filterValue: string | null) => {
    filter.value.class = filterValue;

    // 更新URL参数
    const query = { ...route.query };
    if (filterValue) {
        query.filter = filterValue;
    } else {
        delete query.filter;
    }

    router.replace({ query });
}


// 搜索结果
const searchResults = computed(() => {
    if (!searchQuery.value.trim()) {
        return []
    }
    const query = searchQuery.value.trim().toLowerCase()
    return videoList.value.filter(video =>
        video.title.toLowerCase().includes(query)
    )
})

const dataList = computed(() => {
    let list = videoList.value.filter(i => {
        // 如果启用了只显示缓存视频的选项，则过滤掉未缓存的视频
        if (showCachedOnly.value && i.video_downloaded_num === 0) {
            return false
        }

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

    // 如果有关键词，则进一步过滤搜索结果
    if (searchQuery.value.trim()) {
        const query = searchQuery.value.trim().toLowerCase()
        list = list.filter(video =>
            video.title.toLowerCase().includes(query)
        )
    }

    return list
})

// 将数据按行分组，用于虚拟列表
// 注意：不再固定列数，使用响应式 grid 布局（grid-cols-1 md:grid-cols-4）
const groupedDataList = computed(() => {
    const list = dataList.value;
    const cols = 4; // 固定的分组数，实际渲染由 CSS grid 控制
    const grouped = [];

    for (let i = 0; i < list.length; i += cols) {
        grouped.push({
            id: `row-${i}`, // 为每行生成唯一ID
            videos: list.slice(i, i + cols)
        });
    }

    return grouped;
})

// 监听路由变化，更新过滤器状态
watch(() => route.query.filter, () => {
    initFilterFromUrl();
}, { immediate: true });

// 滚动监听处理函数
const handleScroll = () => {
    if (window.innerWidth >= 768) { // 只在移动端生效
        isScrolled.value = false;
        return;
    }

    const scrollY = window.scrollY || window.pageYOffset;

    // 获取筛选器元素
    if (!filterRef.value) {
        return;
    }

    // 获取筛选器完整模式的高度（大约是250-300px）
    // 当滚动超过筛选器完整模式的高度时，切换到最小模式
    // 或者在筛选器距离顶部很近时也切换到最小模式
    const filterRect = filterRef.value.getBoundingClientRect();
    const filterHeight = filterRef.value.offsetHeight;

    // 当滚动距离超过筛选器高度，或者筛选器已经接近顶部时，切换到最小模式
    isScrolled.value = scrollY > 200 || (filterRect.top < 100 && scrollY > 50);
};

onMounted(() => {
    // 初始化过滤器状态
    initFilterFromUrl();

    // 添加滚动监听
    window.addEventListener('scroll', handleScroll, { passive: true });
    // 添加窗口大小变化监听
    window.addEventListener('resize', handleScroll, { passive: true });
    // 添加键盘事件监听（搜索功能）
    document.addEventListener('keydown', handleKeyDown);
    // 初始检查一次
    nextTick(() => {
        handleScroll();
    });
});

onUnmounted(() => {
    // 移除滚动监听
    window.removeEventListener('scroll', handleScroll);
    // 移除窗口大小变化监听
    window.removeEventListener('resize', handleScroll);
    // 移除键盘事件监听
    document.removeEventListener('keydown', handleKeyDown);
});


// 数据加载
fetch(`/api/progress`).then(async (rsp) => {
    if (rsp.ok) {
        const jsonData = await rsp.json()

        videoList.value = jsonData.data
        stat.value = jsonData.stat

        progress.value = parseInt((stat.value.downloaded / stat.value.count * 100).toFixed(2))
        console.log('Loading, video count:', jsonData.data.length);
    }
})

// 获取当前筛选器标签
const getCurrentFilterLabel = () => {
    switch (filter.value.class) {
        case 'valid':
            return t('progress.validVideos')
        case 'invalid':
            return t('progress.invalidVideos')
        case 'frozen':
            return t('progress.frozenVideos')
        default:
            return t('progress.allVideos')
    }
}

// 获取当前筛选器数量
const getCurrentFilterCount = () => {
    switch (filter.value.class) {
        case 'valid':
            return stat.value.valid
        case 'invalid':
            return stat.value.invalid
        case 'frozen':
            return stat.value.frozen
        default:
            return stat.value.count
    }
}

// 搜索相关方法
const openDesktopSearch = () => {
    if (showDesktopSearch.value) {
        showDesktopSearch.value = false
        // 如果已经显示，则隐藏并清空搜索
        // searchQuery.value = ''
        // currentSearchIndex.value = -1
    } else {
        // 如果没显示，则显示并聚焦
        showDesktopSearch.value = true
        nextTick(() => {
            searchInputRef.value?.focus()
        })
    }
}

const closeSearch = () => {
    searchQuery.value = ''
    showMobileSearch.value = false
    showDesktopSearch.value = false
    currentSearchIndex.value = -1
}

const clearSearch = () => {
    searchQuery.value = ''
    currentSearchIndex.value = -1
}

// 处理Ctrl+F快捷键
const handleKeyDown = (e: KeyboardEvent) => {
    // Ctrl+F 或 Cmd+F (Mac)
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'f') {
        e.preventDefault()
        // 阻止浏览器默认搜索
        if (window.innerWidth >= 768) {
            // PC端：显示并聚焦搜索框
            openDesktopSearch()
        } else {
            // 移动端：显示搜索框
            showMobileSearch.value = true
            nextTick(() => {
                mobileSearchInputRef.value?.focus()
            })
        }
        return false
    }

    // F3 导航到下一个结果（全局快捷键）
    if (e.key === 'F3' && !e.shiftKey) {
        if (searchQuery.value.trim() && searchResults.value.length > 0) {
            e.preventDefault()
            navigateToNextResult()
        }
    }

    // Shift+F3 导航到上一个结果
    if (e.shiftKey && e.key === 'F3') {
        if (searchQuery.value.trim() && searchResults.value.length > 0) {
            e.preventDefault()
            navigateToPrevResult()
        }
    }
}

// 导航到下一个搜索结果
const navigateToNextResult = () => {
    if (searchResults.value.length === 0) return

    // 如果还没有开始导航，从第一个结果开始
    if (currentSearchIndex.value < 0) {
        currentSearchIndex.value = 0
    } else {
        currentSearchIndex.value = (currentSearchIndex.value + 1) % searchResults.value.length
    }
    scrollToSearchResult()
}

// 导航到上一个搜索结果
const navigateToPrevResult = () => {
    if (searchResults.value.length === 0) return

    currentSearchIndex.value = currentSearchIndex.value <= 0
        ? searchResults.value.length - 1
        : currentSearchIndex.value - 1
    scrollToSearchResult()
}

// 滚动到搜索结果
const scrollToSearchResult = () => {
    if (currentSearchIndex.value < 0 || currentSearchIndex.value >= searchResults.value.length) {
        return
    }

    const targetVideo = searchResults.value[currentSearchIndex.value]
    if (!targetVideo) return

    // 在 dataList 中找到目标视频的索引
    const videoIndex = dataList.value.findIndex(v => v.id === targetVideo.id)
    if (videoIndex === -1) return

    // 计算目标视频所在的行（每行4个视频）
    const targetRow = Math.floor(videoIndex / 4)

    // 使用 nextTick 确保 DOM 已更新，并等待虚拟列表渲染
    nextTick(() => {
        setTimeout(() => {
            // 在虚拟列表中查找包含目标视频的元素
            const videoElements = document.querySelectorAll(`[data-video-id="${targetVideo.id}"]`)
            if (videoElements.length > 0) {
                const element = videoElements[0] as HTMLElement
                // 滚动到元素
                element.scrollIntoView({
                    behavior: 'smooth',
                    block: 'center'
                })
                // 添加高亮效果
                element.classList.add('search-highlight')
                setTimeout(() => {
                    element.classList.remove('search-highlight')
                }, 2000)
            } else {
                // 如果元素还没渲染，尝试滚动到对应的行
                const scroller = document.querySelector('.scroller-container') as HTMLElement
                if (scroller) {
                    const rowHeight = 380 // 估算的行高
                    scroller.scrollTo({
                        top: targetRow * rowHeight,
                        behavior: 'smooth'
                    })
                    // 延迟后再尝试定位具体元素
                    setTimeout(() => {
                        const elements = document.querySelectorAll(`[data-video-id="${targetVideo.id}"]`)
                        if (elements.length > 0) {
                            (elements[0] as HTMLElement).scrollIntoView({
                                behavior: 'smooth',
                                block: 'center'
                            })
                            elements[0].classList.add('search-highlight')
                            setTimeout(() => {
                                elements[0].classList.remove('search-highlight')
                            }, 2000)
                        }
                    }, 300)
                }
            }
        }, 100)
    })
}

// 监听搜索结果变化，重置索引
watch(searchQuery, () => {
    currentSearchIndex.value = -1
})

// 监听移动端搜索框显示状态，自动聚焦
watch(showMobileSearch, (show) => {
    if (show) {
        nextTick(() => {
            mobileSearchInputRef.value?.focus()
        })
    }
})

// 监听PC端搜索框显示状态，自动聚焦
watch(showDesktopSearch, (show) => {
    if (show) {
        nextTick(() => {
            searchInputRef.value?.focus()
        })
    }
})
</script>

<style scoped>
.grayscale-image {
    filter: grayscale(100%) brightness(80%);
}

.scroller-container {
    height: calc(100vh - 400px);
    /* 视口高度减去顶部内容的高度 */
    min-height: 500px;
    /* 最小高度保证可用性 */
    overflow-y: auto;
}

/* 图片容器 - 使用 aspect-ratio 预先撑起空间 */
.image-container {
    width: 100%;
    position: relative;
    background-color: #f3f4f6;
    /* 加载时的占位背景色 */
}

.image-container img {
    display: block;
}

/* 虚拟列表行容器 - 防止高度测量问题导致的重叠 */
.virtual-row {
    box-sizing: border-box;
    overflow: visible;
    will-change: transform;
}

/* 移动端调整 */
@media (max-width: 768px) {
    .scroller-container {
        height: calc(100vh);
        min-height: 400px;
    }
}

/* 搜索结果高亮 */
.search-highlight {
    animation: searchHighlight 2s ease-in-out;
    outline: 3px solid #3b82f6;
    outline-offset: 2px;
    border-radius: 8px;
}

@keyframes searchHighlight {
    0% {
        background-color: rgba(59, 130, 246, 0.3);
        outline-color: #3b82f6;
    }

    50% {
        background-color: rgba(59, 130, 246, 0.5);
        outline-color: #2563eb;
    }

    100% {
        background-color: transparent;
        outline-color: transparent;
    }
}
</style>