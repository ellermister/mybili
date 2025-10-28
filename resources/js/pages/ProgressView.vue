<template>
    <div value="w-full flex justify-center	">
        <div class="container  mx-auto justify-center " id="main">
            <div class="m-4">

                <div class="flex justify-between">
                    <h1 class="my-8 text-2xl mt-0">
                        <RouterLink to="/">🌸</RouterLink> {{ t('progress.title') }} {{ $route.params.id }}
                    </h1>
                    <div class="flex gap-4 my-8 text-2xl mt-0">
                        <RouterLink to="/videos" class="hover:text-blue-600 transition-colors">
                            🎬<span class="hidden md:inline"> {{ t('navigation.videoManagement') }}</span>
                        </RouterLink>
                        <RouterLink to="/horizon" target="_blank" class="hover:text-blue-600 transition-colors">
                            🔭<span class="hidden md:inline"> {{ t('progress.viewTasks') }}</span>
                        </RouterLink> 
                    </div>
                </div>

                <div class="flex flex-col md:flex-row md:justify-between gap-4 md:gap-0">
                    <h2 class="text-xl" :title="t('progress.cacheRateDescription')">{{ t('progress.cacheRate') }} {{ progress }}% ({{ stat.downloaded
                    }}/{{ stat.count }})</h2>

                    <div class="flex items-center gap-2">
                        <button
                            @click="showCachedOnly = !showCachedOnly"
                            :class="[
                                'relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2',
                                showCachedOnly ? 'bg-blue-600' : 'bg-gray-200'
                            ]"
                            role="switch"
                            :aria-checked="showCachedOnly"
                        >
                            <span
                                :class="[
                                    'inline-block h-4 w-4 transform rounded-full bg-white transition-transform',
                                    showCachedOnly ? 'translate-x-6' : 'translate-x-1'
                                ]"
                            />
                        </button>
                        <label class="text-sm text-gray-700 cursor-pointer" @click="showCachedOnly = !showCachedOnly">
                            {{ t('progress.showCachedOnly') }}
                        </label>
                    </div>
                </div>

          

                <div class="my-8 w-full bg-gray-200 rounded-full h-2.5 dark:bg-gray-700">
                    <div class="bg-blue-600 h-2.5 rounded-full" :style="{ width: progress + '%' }"></div>
                </div>

                <!-- 初始加载骨架占位（在第一次 fetch 完成前） -->
                <div v-if="!dataLoaded" class="mt-4 grid grid-cols-2 md:grid-cols-4 gap-4">
                    <div v-for="n in 8" :key="n" class="animate-pulse p-4 bg-white rounded-lg shadow-sm">
                        <div class="bg-gray-200 h-40 rounded-md mb-3"></div>
                        <div class="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                        <div class="h-3 bg-gray-200 rounded w-1/2"></div>
                    </div>
                </div>

                <!-- 桌面端筛选器 -->
                <div class="hidden md:grid grid-cols-4 w-full my-4">
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500  py-4 rounded-l-lg"
                        :class="{ 'bg-gradient-to-r': filter.class == null }" @click="setFilter(null)">
                        <span class="text-2xl" :title="t('progress.allVideosDescription')">{{ t('progress.allVideos') }}</span>
                        <span class="text-xl font-semibold">{{ stat.count }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4"
                        :class="{ 'bg-gradient-to-r': filter.class == 'valid' }" @click="setFilter('valid')">
                        <span class="text-2xl" :title="t('progress.validVideosDescription')">{{ t('progress.validVideos') }}</span>
                        <span class="text-xl font-semibold">{{ stat.valid }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4"
                        :class="{ 'bg-gradient-to-r': filter.class == 'invalid' }" @click="setFilter('invalid')">
                        <span class="text-2xl" :title="t('progress.invalidVideosDescription')">{{ t('progress.invalidVideos') }}</span>
                        <span class="text-xl font-semibold">{{ stat.invalid }}</span>
                    </div>
                    <div class="flex flex-col text-center text-white bg-blue-400 hover:bg-gradient-to-r from-purple-500 to-pink-500 py-4 rounded-r-lg"
                        :class="{ 'bg-gradient-to-r': filter.class == 'frozen' }" @click="setFilter('frozen')">
                        <span class="text-2xl" :title="t('progress.frozenVideosDescription')">{{ t('progress.frozenVideos') }}</span>
                        <span class="text-xl font-semibold">{{ stat.frozen }}</span>
                    </div>
                </div>

                <!-- 移动端筛选器 -->
                <div class="md:hidden w-full my-4">
                    <!-- 当前选中的筛选器显示 -->
                    <div class="mb-4 p-4 bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg text-white shadow-lg">
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

                    <!-- 筛选器选项卡片 -->
                    <div class="grid grid-cols-2 gap-3">
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
                </div>

                <div v-if="dataLoaded" class="mt-4 grid grid-cols-1 md:grid-cols-4 w-full gap-4" ref="videoGridRef">
                    <div class="flex flex-col relative" v-for="item in dataList" :key="item.id">
                        <RouterLink :to="{ name: 'video-id', params: { id: item.id } }">
                            <Image class="rounded-lg w-full h-auto md:w-96 md:h-56 hover:scale-105 transition-all duration-300" :src="item.cover_info?.image_url ?? '/assets/images/notfound.webp'"
                                :class="{ 'grayscale-image': item.video_downloaded_num == 0 }" :title="item.title" />
                        </RouterLink>
                        <span class="mt-4 text-center h-12 line-clamp-2" :title="item.title">{{ item.title }}</span>
                        <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
                            <span>{{ t('progress.published') }}: {{ formatTimestamp(item.pubtime, "yyyy.mm.dd") }}</span>
                            <span v-if="item.fav_time > 0">{{ t('progress.favorited') }}: {{ formatTimestamp(item.fav_time, "yyyy.mm.dd") }}</span>
                        </div>
                        <span v-if="item.page > 1"
                            class="text-sm text-white bg-gray-600 rounded-lg w-10 text-center  absolute top-2 right-2">{{
                                item.page }}</span>
                    </div>
                </div>

                <!-- 底部加载更多指示（当后台仍在追加时显示） -->
                <div v-if="loadingMore && dataLoaded" class="mt-6 flex items-center justify-center gap-3">
                    <svg class="w-5 h-5 animate-spin text-blue-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
                    </svg>
                    <div class="text-sm text-gray-600">正在后台加载更多视频…（{{ videoList.length }} / {{ stat.count }}）</div>
                </div>

                <!-- 轻量浮动加载卡片（右下角，吸睛但轻） -->
                <div v-if="dataLoaded" class="fixed right-6 bottom-6 z-50">
                    <div class="bg-white/80 backdrop-blur-sm border border-gray-100 shadow-md rounded-xl px-4 py-3 flex items-center gap-3 w-72 group">
                        <div class="flex-shrink-0">
                            <svg v-if="isPaused" class="w-5 h-5 text-amber-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                <path fill-rule="evenodd" d="M6.75 5.25a.75.75 0 01.75-.75H9a.75.75 0 01.75.75v13.5a.75.75 0 01-.75.75H7.5a.75.75 0 01-.75-.75V5.25zm7.5 0A.75.75 0 0115 4.5h1.5a.75.75 0 01.75.75v13.5a.75.75 0 01-.75.75H15a.75.75 0 01-.75-.75V5.25z" clip-rule="evenodd" />
                            </svg>
                            <svg v-else-if="videoList.length >= (stat.count || 0)" class="w-5 h-5 text-green-500" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
                                <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 00-1.414-1.414L8 11.172 4.707 7.879a1 1 0 00-1.414 1.414l4 4a1 1 0 001.414 0l8-8z" clip-rule="evenodd" />
                            </svg>
                            <svg v-else class="w-5 h-5 animate-spin text-indigo-600" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
                            </svg>
                        </div>
                        <div class="flex-1 min-w-0">
                            <div class="text-sm font-medium text-gray-800">
                                <span v-if="isPaused">已暂停加载</span>
                                <span v-else-if="videoList.length >= (stat.count || 0)">加载完成</span>
                                <span v-else>正在加载视频</span>
                            </div>
                            <div class="text-xs text-gray-500">{{ videoList.length }} / {{ stat.count }}</div>
                            <div class="w-full bg-gray-200 h-2 rounded-full overflow-hidden mt-2">
                                <div class="bg-gradient-to-r from-indigo-500 to-pink-500 h-2 transition-all" :class="{ 'opacity-50': isPaused }" :style="{ width: (stat.count ? (videoList.length / stat.count * 100) + '%' : '0%') }"></div>
                            </div>
                        </div>
                        <!-- 暂停/继续按钮 -->
                        <button v-if="videoList.length < (stat.count || 0)"
                            @click="toggleLoading"
                            class="flex-shrink-0 p-1.5 rounded-lg text-gray-500 hover:text-indigo-600 hover:bg-indigo-50 transition-colors"
                            :title="isPaused ? '继续加载' : '暂停加载'"
                        >
                            <svg v-if="isPaused" class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                <path fill-rule="evenodd" d="M4.5 5.653c0-1.426 1.529-2.33 2.779-1.643l11.54 6.348c1.295.712 1.295 2.573 0 3.285L7.28 19.991c-1.25.687-2.779-.217-2.779-1.643V5.653z" clip-rule="evenodd" />
                            </svg>
                            <svg v-else class="w-5 h-5" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                                <path fill-rule="evenodd" d="M6.75 5.25a.75.75 0 01.75-.75H9a.75.75 0 01.75.75v13.5a.75.75 0 01-.75.75H7.5a.75.75 0 01-.75-.75V5.25zm7.5 0A.75.75 0 0115 4.5h1.5a.75.75 0 01.75.75v13.5a.75.75 0 01-.75.75H15a.75.75 0 01-.75-.75V5.25z" clip-rule="evenodd" />
                            </svg>
                        </button>
                    </div>
                </div>
            </div>

        </div>
    </div>
</template>
<script lang="ts" setup>
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Image from '@/components/Image.vue';
import { formatTimestamp, image } from "../lib/helper"
import type { Cover } from '../api/cover';

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const videoGridRef = ref<HTMLElement>();

const videoList = ref<VideoType[]>([])
const progress = ref(0)
const dataLoaded = ref(false)
const isInitialLoad = ref(true)
const showCachedOnly = ref(false)
const currentPage = ref(0)
const isPaused = ref(false)

// 分页 / 渐进加载设置
const PAGE_SIZE = 24
const LOAD_INTERVAL_MS = 180 // 每次追加一页的间隔（ms），可微调
let loadTimer: number | null = null
const loadingMore = ref(false)

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

// 保存滚动位置到sessionStorage
const saveScrollPosition = () => {
    // 如果是初始加载，不保存滚动位置
    if (isInitialLoad.value) {
        return;
    }
    
    const scrollTop = window.scrollY;
    sessionStorage.setItem('progressViewScrollPosition', scrollTop.toString());
    console.log('Save scroll position:', scrollTop);
}

// 恢复滚动位置
const restoreScrollPosition = () => {
    const savedPosition = sessionStorage.getItem('progressViewScrollPosition');
    
    if (savedPosition && dataLoaded.value) {
        nextTick(() => {
            const scrollTop = parseInt(savedPosition);
            console.log('Restore scroll position to:', scrollTop);
            
            window.scrollTo({
                top: scrollTop,
                behavior: 'instant'
            });
        });
    }
}

// 事件处理函数
const handleScroll = () => {
    saveScrollPosition();
}

const dataList = computed(() => {
    return videoList.value.filter(i => {
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
})

// 监听路由变化，更新过滤器状态
watch(() => route.query.filter, () => {
    initFilterFromUrl();
}, { immediate: true });

onMounted(async () => {
    // 初始化过滤器状态
    initFilterFromUrl();
    // 添加事件监听器
    window.addEventListener('scroll', handleScroll);

    // 加载第一页并在后台逐页加载剩余页面
    await fetchPage(1)
    if (videoList.value.length < (stat.value.count || 0)) {
        startBackgroundLoad()
    }
});

onUnmounted(() => {
    // 移除事件监听器
    window.removeEventListener('scroll', handleScroll);
    // 清理后台加载定时器
    stopBackgroundLoad();
});

// 数据加载：按页从后端请求（后端已支持 page 和 page_size）
const fetchPage = async (page: number) => {
    if (loadingMore.value || isPaused.value) return
    loadingMore.value = true
    try {
        const rsp = await fetch(`/api/progress?page=${page}&page_size=${PAGE_SIZE}`)
        if (!rsp.ok) return
        const jsonData = await rsp.json()
        const items: VideoType[] = jsonData.data || jsonData.list || []

        // 追加到渲染列表
        if (page === 1) {
            videoList.value = items
        } else {
            videoList.value = videoList.value.concat(items)
        }

        // 更新统计信息
        if (jsonData.stat) {
            stat.value = jsonData.stat
            progress.value = stat.value.count > 0 ? parseInt((stat.value.downloaded / stat.value.count * 100).toFixed(2)) : 0
        }

        currentPage.value = page

        // 设置已加载标志（首次加载完毕）
        if (!dataLoaded.value) {
            dataLoaded.value = true
            isInitialLoad.value = false
            restoreScrollPosition()
        }
    } catch (e) {
        console.error('Fetch page error', e)
    } finally {
        loadingMore.value = false
    }
}

// 启动后台分页追加（按页请求下一页直到全部加载完毕）
// 优化：提高优先级 —— 立即开始加载下一页（不再等待 requestIdleCallback）
const startBackgroundLoad = () => {
    if (loadingMore.value) return
    isPaused.value = false

    const scheduleNext = async () => {
        const loaded = videoList.value.length
        const total = stat.value.count || 0
        if (loaded >= total) {
            loadingMore.value = false
            if (loadTimer) { clearTimeout(loadTimer); loadTimer = null }
            return
        }

        const nextPage = currentPage.value + 1
        await fetchPage(nextPage)

        // 安排下一页加载（稍作延迟以避免完全占满主线程）
        loadTimer = window.setTimeout(scheduleNext, LOAD_INTERVAL_MS)
    }

    // 立即开始：比 requestIdleCallback 更高的优先级
    scheduleNext()
}

const stopBackgroundLoad = () => {
    if (loadTimer) {
        clearTimeout(loadTimer)
        loadTimer = null
    }
    isPaused.value = true
    loadingMore.value = false
}

// 切换暂停/继续加载
const toggleLoading = () => {
    if (isPaused.value) {
        // 继续加载
        if (videoList.value.length < (stat.value.count || 0)) {
            startBackgroundLoad()
        }
    } else {
        // 暂停加载
        stopBackgroundLoad()
    }
}

// 监听数据加载状态变化，恢复滚动位置
watch(dataLoaded, (newValue) => {
    if (newValue) {
        restoreScrollPosition();
    }
});

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
</script>

<style scoped>
.grayscale-image {
    filter: grayscale(100%) brightness(80%);
}
</style>