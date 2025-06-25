<template>
    <div value="w-full flex justify-center	">
        <div class="container  mx-auto justify-center " id="main">
            <div class="m-4">

                <div class="flex justify-between">
                    <h1 class="my-8 text-2xl">
                        <RouterLink to="/">🌸</RouterLink> {{ t('progress.title') }} {{ $route.params.id }}
                    </h1>
                    <h1 class="my-8 text-2xl">
                        <RouterLink to="/horizon" target="_blank">🔭 {{ t('progress.viewTasks') }}</RouterLink> 
                    </h1>
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

                <div class="grid grid-cols-1 md:grid-cols-4 w-full my-4 ">
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


                <div class="mt-4 grid grid-cols-1 md:grid-cols-4 w-full gap-4" ref="videoGridRef">
                    <div class="flex flex-col relative" v-for="item in dataList">
                        <RouterLink :to="{ name: 'video-id', params: { id: item.id } }">
                            <Image class="rounded-lg w-full h-auto md:w-96 md:h-56 hover:scale-105 transition-all duration-300" :src="item.cache_image_url ?? '/assets/images/notfound.webp'"
                                :class="{ 'grayscale-image': item.video_downloaded_num == 0 }" :title="item.title" />
                        </RouterLink>
                        <span class="mt-4 text-center h-12 line-clamp-2" :title="item.title">{{ item.title }}</span>
                        <div class="mt-2 flex justify-between text-xs text-gray-400 px-1">
                            <span>{{ t('progress.published') }}: {{ formatTimestamp(item.pubtime, "yyyy.mm.dd") }}</span>
                            <span>{{ t('progress.favorited') }}: {{ formatTimestamp(item.fav_time, "yyyy.mm.dd") }}</span>
                        </div>
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
import { computed, ref, onMounted, onUnmounted, watch, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Image from '@/components/Image.vue';
import { formatTimestamp, image } from "../lib/helper"

const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const videoGridRef = ref<HTMLElement>();

const videoList = ref<VideoType[]>([])
const progress = ref(0)
const dataLoaded = ref(false)
const isInitialLoad = ref(true)
const showCachedOnly = ref(false)

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

onMounted(() => {
    // 初始化过滤器状态
    initFilterFromUrl();
    // 添加事件监听器
    window.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
    // 移除事件监听器
    window.removeEventListener('scroll', handleScroll);
});

// 数据加载
fetch(`/api/progress`).then(async (rsp) => {
    if (rsp.ok) {
        const jsonData = await rsp.json()
        videoList.value = jsonData.data
        stat.value = jsonData.stat

        progress.value = parseInt((stat.value.downloaded / stat.value.count * 100).toFixed(2))
        
        console.log('Loading, video count:', jsonData.data.length);
        
        // 数据加载完成后恢复滚动位置
        dataLoaded.value = true
        restoreScrollPosition()
        isInitialLoad.value = false;
    }
})

// 监听数据加载状态变化，恢复滚动位置
watch(dataLoaded, (newValue) => {
    if (newValue) {
        restoreScrollPosition();
    }
});
</script>

<style scoped>
.grayscale-image {
    filter: grayscale(100%) brightness(80%);
}
</style>