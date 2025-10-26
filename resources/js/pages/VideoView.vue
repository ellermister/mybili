<template>
    <div>
        <Breadcrumbs :items="breadcrumbItems">
            <template #actions>
                <div class="text-sm text-gray-600 bg-gray-100 px-3 py-1 rounded-full hidden md:block">
                    {{ videoId }}
                </div>
            </template>
        </Breadcrumbs>


        <!-- Video Content -->
        <div class="space-y-4" v-if="videoInfo != null">
            <!-- Video Player Section -->
            <div class="flex flex-col lg:flex-row gap-4">
                <!-- Main Player -->
                <div class="flex-1">
                    <div ref="playerContainer" id="playerContainer" class="-mx-2 md:mx-0 bg-white shadow-lg overflow-hidden md:border border-gray-200/50">
                        <Player ref="playerRef" @ready="onPlayerReady" :danmaku="danmaku" :url="currentPart?.url ?? ''" />
                    </div>
                </div>

                <!-- Parts Sidebar -->
                <div class="w-full lg:w-72 lg:shrink-0"
                    v-if="videoInfo && videoInfo.video_parts && videoInfo.video_parts.length > 1">
                    <div class="bg-white/90 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200/50 p-3 flex flex-col"
                        :style="{ height: sidebarHeight }">
                        <h3 class="text-xl font-semibold mb-3 text-gray-800 flex items-center flex-shrink-0">
                            <span class="w-2 h-2 bg-gradient-to-r from-pink-500 to-purple-600 rounded-full mr-2"></span>
                            {{ t('video.videoParts') }}&nbsp;<span class="text-gray-500 text-sm font-normal">({{
                                videoInfo.video_parts.findIndex(part => part.id === currentPart?.id) + 1}}/{{
                                    videoInfo.video_parts.length }})</span>
                        </h3>
                        <div class="space-y-1 overflow-y-auto flex-1 min-h-0 pr-1 custom-scrollbar">
                            <button v-for="part in videoInfo?.video_parts" :key="part.id" @click="playPart(part.id)"
                                class="w-full px-3 py-2 text-left rounded-lg transition-all duration-300 group relative overflow-hidden"
                                :class="{
                                    'bg-gradient-to-r from-pink-500 to-purple-600 text-white shadow': currentPart?.id === part.id && part.downloaded,
                                    'bg-gray-50 hover:bg-gray-100 text-gray-700 hover:text-gray-900': currentPart?.id !== part.id && part.downloaded,
                                    'bg-gray-100 text-gray-500 cursor-not-allowed': !part.downloaded
                                }">
                                <div class="flex items-center justify-between">
                                    <span class="font-medium truncate">{{ part.title }}</span>
                                    <span v-if="currentPart?.id === part.id" class="text-white/80">▶</span>
                                </div>
                            </button>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Video Info Section -->
            <div v-if="videoInfo != null"
                class="bg-white/90 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200/50 p-4">
                <div class="space-y-4">
                    <!-- Title -->
                    <div>
                        <h2 class="text-2xl font-bold text-gray-800 mb-2 leading-tight">{{ videoInfo.title }}</h2>
                        <div class="w-16 h-1 bg-gradient-to-r from-pink-500 to-purple-600 rounded-full"></div>
                    </div>

                    <!-- UP主信息 -->
                    <div v-if="videoInfo.upper"
                        class="flex items-center justify-between p-3 bg-gradient-to-r from-gray-50 to-gray-100 rounded-lg border border-gray-200/50">
                        <div class="flex items-center space-x-3">
                            <!-- 预留头像位置 -->
                            <div
                                class="w-10 h-10 bg-gradient-to-br from-pink-400 to-purple-500 rounded-full flex items-center justify-center flex-shrink-0">
                                <span class="text-white font-semibold text-sm">UP</span>
                            </div>
                            <div class="min-w-0 flex-1" @click="openUpperSpace(videoInfo.upper.mid)">
                                <div class="flex items-center space-x-2">
                                    <h3 class="font-semibold text-gray-800 truncate" >{{ videoInfo.upper.name }}</h3>
                                    <span class="text-xs text-gray-500 bg-gray-200 px-2 py-0.5 rounded-full">UID: {{
                                        videoInfo.upper.mid }}</span>
                                </div>
                            </div>
                        </div>
                        <a :href="upperSpaceUrl(videoInfo.upper.mid)" target="_blank" rel="noopener noreferrer"
                            class="hidden md:inline-flex items-center space-x-1 px-3 py-1.5 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 hover:border-gray-400 transition-all duration-200 text-sm font-medium text-gray-700 hover:text-gray-900 flex-shrink-0">
                            <span>{{ t('video.visitSpace') }}</span>
                            <span class="text-gray-400">↗</span>
                        </a>
                    </div>

                    <!-- Description -->
                    <div v-if="videoInfo.intro" class="bg-gray-50 rounded-lg p-4">
                        <h3 class="text-lg font-semibold text-gray-700 mb-2 flex items-center">
                            <span class="w-1.5 h-1.5 bg-gray-400 rounded-full mr-2"></span>
                            {{ t('video.videoDescription') }}
                        </h3>
                        <p class="text-gray-600 leading-relaxed break-words">{{ videoInfo.intro }}</p>
                    </div>

                    <!-- Meta Information -->
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                        <div
                            class="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg p-3 border border-blue-200/50">
                            <div class="flex items-center space-x-2">
                                <span class="text-blue-500">📅</span>
                                <span class="text-sm text-gray-600">{{ t('video.publishTime') }}</span>
                            </div>
                            <div class="text-base font-semibold text-gray-800 mt-1">
                                {{ formatTimestamp(videoInfo.pubtime, "yyyy-mm-dd hh:ii") }}
                            </div>
                        </div>

                        <div
                            class="bg-gradient-to-r from-green-50 to-emerald-50 rounded-lg p-3 border border-green-200/50">
                            <div class="flex items-center space-x-2">
                                <span class="text-green-500">⭐</span>
                                <span class="text-sm text-gray-600">{{ t('video.favoriteTime') }}</span>
                            </div>
                            <div class="text-base font-semibold text-gray-800 mt-1">
                                {{ videoInfo.fav_time ? formatTimestamp(videoInfo.fav_time, "yyyy-mm-dd hh:ii") : '-' }}
                            </div>
                        </div>

                        <div
                            class="bg-gradient-to-r from-purple-50 to-pink-50 rounded-lg p-3 border border-purple-200/50">
                            <div class="flex items-center space-x-2">
                                <span class="text-purple-500">💬</span>
                                <span class="text-sm text-gray-600">{{ t('video.danmakuCount') }}</span>
                            </div>
                            <div class="text-base font-semibold text-gray-800 mt-1">
                                {{ videoInfo.danmaku_count.toLocaleString() }}
                            </div>
                        </div>
                    </div>

                    <!-- External Link -->
                    <div class="flex justify-center pt-2">
                        <a class="inline-flex items-center space-x-2 px-6 py-3 bg-gradient-to-r from-green-500 to-emerald-600 text-white rounded-lg hover:from-green-600 hover:to-emerald-700 transition-all duration-300 shadow hover:shadow-lg transform hover:-translate-y-0.5"
                            :href="bilibiliUrl(videoInfo.bvid)" target="_blank" rel="noopener noreferrer">
                            <span class="text-lg">📺</span>
                            <span class="font-semibold">{{ t('video.watchOnBilibili') }}</span>
                            <span class="text-white/80">↗</span>
                        </a>
                    </div>

                    <!-- Download Actions -->
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-3 pt-4 border-t border-gray-200/50">
                        <!-- Download Video Button -->
                        <button @click="downloadVideo"
                            class="flex flex-col items-center p-4 bg-gradient-to-br from-blue-50 to-blue-100 border border-blue-200/50 rounded-xl hover:from-blue-100 hover:to-blue-150 hover:border-blue-300/50 transition-all duration-300 group hover:shadow-md">
                            <div
                                class="w-12 h-12 bg-gradient-to-br from-blue-500 to-blue-600 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform duration-300">
                                <span class="text-xl text-white">🎬</span>
                            </div>
                            <span class="text-sm font-medium text-blue-700 group-hover:text-blue-800">{{
                                t('video.downloadVideo') }}</span>
                        </button>

                        <!-- Download Danmaku Button -->
                        <button @click="downloadDanmaku"
                            class="flex flex-col items-center p-4 bg-gradient-to-br from-purple-50 to-purple-100 border border-purple-200/50 rounded-xl hover:from-purple-100 hover:to-purple-150 hover:border-purple-300/50 transition-all duration-300 group hover:shadow-md">
                            <div
                                class="w-12 h-12 bg-gradient-to-br from-purple-500 to-purple-600 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform duration-300">
                                <span class="text-xl text-white">💬</span>
                            </div>
                            <span class="text-sm font-medium text-purple-700 group-hover:text-purple-800">{{
                                t('video.downloadDanmaku') }}</span>
                        </button>

                        <!-- Download Cover Button -->
                        <button @click="downloadCover"
                            class="flex flex-col items-center p-4 bg-gradient-to-br from-pink-50 to-pink-100 border border-pink-200/50 rounded-xl hover:from-pink-100 hover:to-pink-150 hover:border-pink-300/50 transition-all duration-300 group hover:shadow-md">
                            <div
                                class="w-12 h-12 bg-gradient-to-br from-pink-500 to-pink-600 rounded-full flex items-center justify-center mb-2 group-hover:scale-110 transition-transform duration-300">
                                <span class="text-xl text-white">🖼️</span>
                            </div>
                            <span class="text-sm font-medium text-pink-700 group-hover:text-pink-800">{{
                                t('video.downloadCover') }}</span>
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Not Found State -->
        <div v-if="notfound" class="text-center py-16">
            <div class="text-6xl mb-4">😢</div>
            <div class="text-3xl font-semibold text-gray-700 mb-2">{{ t('video.videoNotFound') }}</div>
            <div class="text-gray-500">{{ t('video.videoNotFoundDescription') }}</div>
            <RouterLink to="/"
                class="inline-block mt-6 px-6 py-3 bg-gradient-to-r from-pink-500 to-purple-600 text-white rounded-lg hover:from-pink-600 hover:to-purple-700 transition-all duration-300 shadow-lg hover:shadow-xl">
                {{ t('video.backToHome') }}
            </RouterLink>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { computed, onMounted, ref, nextTick, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { formatTimestamp } from '../lib/helper';
import Player from '../components/Player.vue';
import Breadcrumbs from '../components/Breadcrumbs.vue';
import type { Video, VideoPartType } from '@/api/fav';
import { getVideoDanmaku, getVideoInfo } from '@/api/video';

const { t } = useI18n();
const playerRef = ref()
const playerContainer = ref<HTMLDivElement | null>(null)
const sidebarHeight = ref('auto')

const route = useRoute()

const videoId = ref(route.params.id)

if (route.name == "subscription-video-id" || route.name == "favlist-video-id") {
    videoId.value = route.params.video_id
}

const breadcrumbItems = computed(() => {
    if (route.name == "subscription-video-id") {
        let subscriptionId = route.params.id
        return [
            { text: t('navigation.home'), to: '/' },
            { text: videoInfo.value?.subscriptions?.[0]?.name ?? t('video.loading'), to: '/subscription/' + subscriptionId },
            { text: videoInfo.value?.title ?? t('video.loading') }
        ]
    } else if (route.name == "favlist-video-id") {
        let favId = route.params.id as string
        let title = t('video.favorite')
        let fav = videoInfo.value?.favorite?.find(fav => fav.id == parseInt(favId))
        if (fav) {
            title = fav.title
        } else {
            let subscription = videoInfo.value?.subscriptions?.find(sub => -sub.id == parseInt(favId))
            if (subscription) {
                title = subscription.name
            }
        }
        return [
            { text: t('navigation.home'), to: '/' },
            { text: title, to: '/fav/' + favId },
            { text: videoInfo.value?.title ?? t('video.loading') }
        ]
    } else {
        if (videoInfo.value?.favorite && videoInfo.value?.favorite?.length > 0) {
            return [
                { text: t('navigation.home'), to: '/' },
                { text: (videoInfo.value?.favorite?.[0]?.title ?? t('video.favorite')), to: '/fav/' + (videoInfo.value?.favorite?.[0]?.id ?? '') },
                { text: videoInfo.value?.title ?? t('video.loading') }
            ]
        } else if (videoInfo.value?.subscriptions && videoInfo.value?.subscriptions?.length > 0) {
            return [
                { text: t('navigation.home'), to: '/' },
                { text: (videoInfo.value?.subscriptions?.[0]?.name ?? t('video.favorite')), to: '/subscription/' + (videoInfo.value?.subscriptions?.[0]?.id ?? '') },
                { text: videoInfo.value?.title ?? t('video.loading') }
            ]
        } else {
            return [
                { text: t('navigation.home'), to: '/' },
                { text: videoInfo.value?.title ?? t('video.loading') }
            ]
        }

    }
})

const bilibiliUrl = (bvid: string) => {
    return `https://www.bilibili.com/video/${bvid}`
}

const upperSpaceUrl = (mid: number) => {
    return `https://space.bilibili.com/${mid}`
}

const openUpperSpace = (mid: number) => {
    window.open(upperSpaceUrl(mid), '_blank')
}

const downloadFile = (url: string, name: string) => {
    const a = document.createElement('a')
    a.href = url
    a.download = name
    a.click()
}

const downloadVideo = () => {
    const parts = videoInfo.value?.video_parts
    if (parts) {
        for (let i in parts) {
            const part = parts[i]
            const url = part.url
            if (url) {
                // 下载视频
                downloadFile(url, part.title + ".mp4")
            }
        }
    }
    console.log('Download video clicked for:', videoInfo.value?.bvid)
}

const downloadDanmaku = () => {
    const parts = videoInfo.value?.video_parts
    if (parts) {
        for (let i in parts) {
            const part = parts[i]
            const partId = part.id
            fetch(`/api/danmaku/v3/?id=${partId}`).then(async (rsp) => {
                if (rsp.ok) {
                    const jsonData = await rsp.json()
                    const danmaku = jsonData.data
                    if (danmaku) {
                        // 按照第一个元素排序,时间
                        danmaku.sort((a: any, b: any) => a[0] - b[0])
                        const json = JSON.stringify(danmaku)
                        const file = new File([json], part.title + ".json", { type: "application/json" })
                        const url = URL.createObjectURL(file)
                        downloadFile(url, part.title + ".json")
                    }
                }
            })
        }
    }
    console.log('Download danmaku clicked for:', videoInfo.value?.bvid)
}

const downloadCover = () => {
    const coverURL = videoInfo.value?.cover_info?.image_url
    if (coverURL) {
        // 下载图片
        downloadFile(coverURL, videoInfo.value?.title + ".jpg")
    }
    console.log('Download cover clicked for:', videoInfo.value?.bvid)
}

const videoInfo = ref<Video | null>()
const notfound = ref(false)

const currentPart = ref<VideoPartType | null>(null)

const danmaku = ref<any[]>([])

// Player 准备就绪时的回调
const onPlayerReady = () => {
    playPart(videoInfo.value?.video_parts?.[0]?.id ?? 0)
}


const playPart = (partId: number) => {
    console.log('playPart', partId)
    const part = videoInfo.value?.video_parts?.find(part => part.id === partId)
    if (part && playerRef.value) {
        console.log('playPart switchVideo', part.id)
        getVideoDanmaku(part.id).then(danmaku => {
            console.log('playPart switchVideo', {
                url: part.url,
                danmaku: danmaku,
            })
            playerRef.value?.switchVideo({
                url: part.url,
                danmaku: danmaku,
            })
        })
        currentPart.value = part as VideoPartType
    }
}

// 更新侧边栏高度
const updateSidebarHeight = () => {
    if (playerContainer.value) {
        const height = playerContainer.value.offsetHeight
        sidebarHeight.value = height > 0 ? `${height}px` : 'auto'
    }
}

// 监听窗口大小变化
let resizeObserver: ResizeObserver | null = null

onMounted(() => {
    getVideoInfo(parseInt(videoId.value as string)).then(async (data) => {
        if (data) {
            videoInfo.value = data
            updateSidebarHeight()
            // // 等待DOM更新后设置高度监听
            nextTick(() => {
                updateSidebarHeight()

                // 使用ResizeObserver监听播放器容器大小变化
                if (playerContainer.value && window.ResizeObserver) {
                    resizeObserver = new ResizeObserver(() => {
                        updateSidebarHeight()
                    })
                    resizeObserver.observe(playerContainer.value)
                }
            })
        }
    })
})

// 组件卸载时清理
onUnmounted(() => {
    if (resizeObserver) {
        resizeObserver.disconnect()
        resizeObserver = null
    }
})
</script>

<style scoped>
.custom-scrollbar {
    scrollbar-width: thin;
    scrollbar-color: #cbd5e1 #f1f5f9;
}

.custom-scrollbar::-webkit-scrollbar {
    width: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
    background: #f1f5f9;
    border-radius: 3px;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
    background: #cbd5e1;
    border-radius: 3px;
    transition: background-color 0.2s ease;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: #94a3b8;
}
</style>
