<template>
    <div>
        <!-- Header -->
        <header class="bg-white/80 backdrop-blur-sm border-b border-gray-200/50 sticky top-0 z-10 mb-4">
            <div class="flex items-center justify-between h-16 px-1">
                <RouterLink to="/" class="flex items-center space-x-2 text-2xl font-bold bg-gradient-to-r from-pink-500 to-purple-600 bg-clip-text text-transparent hover:from-pink-600 hover:to-purple-700 transition-all duration-300">
                    <span>🌸</span>
                    <span>my fav</span>
                </RouterLink>
                <div class="text-sm text-gray-600 bg-gray-100 px-3 py-1 rounded-full">
                    {{ $route.params.id }}
                </div>
            </div>
        </header>

        <!-- Video Content -->
        <div class="space-y-4" v-if="videoInfo != null">
            <!-- Video Player Section -->
            <div class="flex flex-col lg:flex-row gap-4">
                <!-- Main Player -->
                <div class="flex-1">
                    <div class="bg-white shadow-lg overflow-hidden border border-gray-200/50">
                        <Player ref="playerRef" />
                    </div>
                </div>

                <!-- Parts Sidebar -->
                <div class="w-full lg:w-72 lg:shrink-0" v-if="videoInfo && videoInfo.video_parts.length > 1">
                    <div class="bg-white/90 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200/50 p-3">
                        <h3 class="text-xl font-semibold mb-3 text-gray-800 flex items-center">
                            <span class="w-2 h-2 bg-gradient-to-r from-pink-500 to-purple-600 rounded-full mr-2"></span>
                            分P列表
                        </h3>
                        <div class="space-y-1">
                            <button 
                                v-for="part in videoInfo?.video_parts" 
                                :key="part.id" 
                                @click="playPart(part.id)"
                                class="w-full px-3 py-2 text-left rounded-lg transition-all duration-300 group relative overflow-hidden"
                                :class="{ 
                                    'bg-gradient-to-r from-pink-500 to-purple-600 text-white shadow': currentPart?.id === part.id, 
                                    'bg-gray-50 hover:bg-gray-100 text-gray-700 hover:text-gray-900': currentPart?.id !== part.id 
                                }"
                            >
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
            <div v-if="videoInfo != null" class="bg-white/90 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200/50 p-4">
                <div class="space-y-4">
                    <!-- Title -->
                    <div>
                        <h2 class="text-2xl font-bold text-gray-800 mb-2 leading-tight">{{ videoInfo.title }}</h2>
                        <div class="w-16 h-1 bg-gradient-to-r from-pink-500 to-purple-600 rounded-full"></div>
                    </div>

                    <!-- Description -->
                    <div v-if="videoInfo.intro" class="bg-gray-50 rounded-lg p-4">
                        <h3 class="text-lg font-semibold text-gray-700 mb-2 flex items-center">
                            <span class="w-1.5 h-1.5 bg-gray-400 rounded-full mr-2"></span>
                            视频简介
                        </h3>
                        <p class="text-gray-600 leading-relaxed">{{ videoInfo.intro }}</p>
                    </div>

                    <!-- Meta Information -->
                    <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                        <div class="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg p-3 border border-blue-200/50">
                            <div class="flex items-center space-x-2">
                                <span class="text-blue-500">📅</span>
                                <span class="text-sm text-gray-600">发布时间</span>
                            </div>
                            <div class="text-base font-semibold text-gray-800 mt-1">
                                {{ formatTimestamp(videoInfo.pubtime, "yyyy-mm-dd hh:ii") }}
                            </div>
                        </div>

                        <div class="bg-gradient-to-r from-green-50 to-emerald-50 rounded-lg p-3 border border-green-200/50">
                            <div class="flex items-center space-x-2">
                                <span class="text-green-500">⭐</span>
                                <span class="text-sm text-gray-600">收藏时间</span>
                            </div>
                            <div class="text-base font-semibold text-gray-800 mt-1">
                                {{ formatTimestamp(videoInfo.fav_time, "yyyy-mm-dd hh:ii") }}
                            </div>
                        </div>

                        <div class="bg-gradient-to-r from-purple-50 to-pink-50 rounded-lg p-3 border border-purple-200/50">
                            <div class="flex items-center space-x-2">
                                <span class="text-purple-500">💬</span>
                                <span class="text-sm text-gray-600">弹幕数量</span>
                            </div>
                            <div class="text-base font-semibold text-gray-800 mt-1">
                                {{ videoInfo.danmaku_count.toLocaleString() }}
                            </div>
                        </div>
                    </div>

                    <!-- External Link -->
                    <div class="flex justify-center pt-2">
                        <a 
                            class="inline-flex items-center space-x-2 px-6 py-3 bg-gradient-to-r from-green-500 to-emerald-600 text-white rounded-lg hover:from-green-600 hover:to-emerald-700 transition-all duration-300 shadow hover:shadow-lg transform hover:-translate-y-0.5" 
                            :href="bilibiliUrl(videoInfo.bvid)"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            <span class="text-lg">📺</span>
                            <span class="font-semibold">在哔哩哔哩观看</span>
                            <span class="text-white/80">↗</span>
                        </a>
                    </div>
                </div>
            </div>
        </div>

        <!-- Not Found State -->
        <div v-if="notfound" class="text-center py-16">
            <div class="text-6xl mb-4">😢</div>
            <div class="text-3xl font-semibold text-gray-700 mb-2">视频未找到</div>
            <div class="text-gray-500">抱歉，您要查找的视频不存在或已被删除</div>
            <RouterLink to="/" class="inline-block mt-6 px-6 py-3 bg-gradient-to-r from-pink-500 to-purple-600 text-white rounded-lg hover:from-pink-600 hover:to-purple-700 transition-all duration-300 shadow-lg hover:shadow-xl">
                返回首页
            </RouterLink>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { formatTimestamp } from '../lib/helper';
import Player from '../components/Player.vue';

const playerRef = ref()

const route = useRoute()
const id = route.params.id

interface VideoType {
    title: string
    id: number
    link: string
    intro: string
    pubtime: number
    fav_time: number
    bvid: string
    attr: number
    page: number
    video_parts: VideoPartType[]
    danmaku_count: number
    _metas: {
        cover: string
    }
}

interface VideoPartType {
    id: number
    url: string
    title: string
    part: number
}

const bilibiliUrl = (bvid: string) => {
    return `https://www.bilibili.com/video/${bvid}`
}

const videoInfo = ref<VideoType | null>()
const notfound = ref(false)

const currentPart = ref<VideoPartType | null>(null)

const playPart = (partId: number) => {
    const part = videoInfo.value?.video_parts.find(part => part.id === partId)
    if (part) {
        // p1 视频, p2 弹幕
        playerRef.value.switchVideo({
            url: part.url,
            type: 'mp4',
        }, {
            id: part.id,
            api: '/api/danmaku/',
        })

        currentPart.value = part as VideoPartType
    }
}

onMounted(() => {
    fetch(`/api/video/${id}`).then(async (rsp) => {
        if (!rsp.ok) {
            notfound.value = true
        } else {
            const jsonData = await rsp.json()
            videoInfo.value = jsonData
            if (jsonData.video_parts.length > 0) {
                const firstVideo = jsonData.video_parts[0]
                playerRef.value.switchVideo(
                    {
                        url: firstVideo.url,
                        type: 'mp4',
                        danmaku_id: firstVideo.id,
                    }
                )
            }
        }
    })
})
</script>
