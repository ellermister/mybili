<template>
    <div class="w-full flex justify-center">
        <div class="container mx-auto justify-center" id="main">
            <div class="m-4">
                <div class="flex justify-between items-center">
                    <h1 class="my-8 text-2xl">
                        <RouterLink to="/">🌸</RouterLink> {{ t('videoManagement.title') }}
                    </h1>
                    <div class="flex items-center gap-3 text-sm">
                        <RouterLink to="/progress" class="text-blue-600 hover:text-blue-800">
                            📊 {{ t('navigation.progress') }}
                        </RouterLink>
                        <RouterLink to="/download-queue" class="text-blue-600 hover:text-blue-800">
                            📥 下载队列
                        </RouterLink>
                    </div>
                </div>

                <div class="bg-white rounded-lg shadow-md p-6 mb-6">
                    <div class="mb-4">
                        <div class="relative">
                            <input v-model="searchQuery" type="text"
                                :placeholder="t('videoManagement.searchPlaceholder')"
                                class="w-full px-4 py-3 pl-12 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                                @input="handleSearch" />
                            <span class="absolute left-4 top-3.5 text-gray-400 text-xl">🔍</span>
                            <button v-if="searchQuery" @click="clearSearch"
                                class="absolute right-4 top-3.5 text-gray-400 hover:text-gray-600">✕</button>
                        </div>
                    </div>

                    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">
                                {{ t('videoManagement.favorite') }}
                            </label>
                            <select v-model="filters.favId"
                                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                                <option value="">{{ t('videoManagement.allFavorites') }}</option>
                                <option v-for="fav in favList" :key="fav.id" :value="fav.id">
                                    {{ fav.title }} ({{ fav.media_count }})
                                </option>
                            </select>
                        </div>

                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">
                                {{ t('videoManagement.status') }}
                            </label>
                            <select v-model="filters.status"
                                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                                <option value="all">{{ t('videoManagement.allStatus') }}</option>
                                <option value="valid">{{ t('videoManagement.validOnly') }}</option>
                                <option value="invalid">{{ t('videoManagement.invalidOnly') }}</option>
                                <option value="frozen">{{ t('videoManagement.frozenOnly') }}</option>
                            </select>
                        </div>

                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">
                                {{ t('videoManagement.cacheStatus') }}
                            </label>
                            <select v-model="filters.downloaded"
                                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                                <option value="all">{{ t('videoManagement.allCache') }}</option>
                                <option value="yes">{{ t('videoManagement.cachedOnly') }}</option>
                                <option value="no">{{ t('videoManagement.notCachedOnly') }}</option>
                            </select>
                        </div>

                        <div>
                            <label class="block text-sm font-medium text-gray-700 mb-2">
                                {{ t('videoManagement.multiPart') }}
                            </label>
                            <select v-model="filters.multiPart"
                                class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent">
                                <option value="all">{{ t('videoManagement.allParts') }}</option>
                                <option value="no">{{ t('videoManagement.singlePartOnly') }}</option>
                                <option value="yes">{{ t('videoManagement.multiPartOnly') }}</option>
                            </select>
                        </div>
                    </div>

                    <div class="flex flex-wrap gap-3 mt-4">
                        <button @click="selectAll"
                            class="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-lg transition-colors">
                            {{ selectedVideos.size === filteredVideos.length && filteredVideos.length > 0
                                ? t('videoManagement.deselectAll') : t('videoManagement.selectAll') }}
                        </button>
                        <button v-if="selectedVideos.size > 0" @click="confirmBatchDelete"
                            class="px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg transition-colors">
                            {{ t('videoManagement.deleteSelected') }} ({{ selectedVideos.size }})
                        </button>
                        <button @click="resetFilters"
                            class="px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition-colors">
                            {{ t('videoManagement.resetFilters') }}
                        </button>
                    </div>
                </div>

                <div class="grid grid-cols-2 md:grid-cols-5 gap-4 mb-6">
                    <div class="bg-gradient-to-r from-blue-500 to-blue-600 text-white p-4 rounded-lg text-center">
                        <div class="text-2xl font-bold">{{ stat.count }}</div>
                        <div class="text-sm">{{ t('videoManagement.totalVideos') }}</div>
                    </div>
                    <div class="bg-gradient-to-r from-green-500 to-green-600 text-white p-4 rounded-lg text-center">
                        <div class="text-2xl font-bold">{{ stat.valid }}</div>
                        <div class="text-sm">{{ t('videoManagement.validVideos') }}</div>
                    </div>
                    <div class="bg-gradient-to-r from-red-500 to-red-600 text-white p-4 rounded-lg text-center">
                        <div class="text-2xl font-bold">{{ stat.invalid }}</div>
                        <div class="text-sm">{{ t('videoManagement.invalidVideos') }}</div>
                    </div>
                    <div class="bg-gradient-to-r from-orange-500 to-orange-600 text-white p-4 rounded-lg text-center">
                        <div class="text-2xl font-bold">{{ stat.frozen }}</div>
                        <div class="text-sm">{{ t('videoManagement.frozenVideos') }}</div>
                    </div>
                    <div class="bg-gradient-to-r from-purple-500 to-purple-600 text-white p-4 rounded-lg text-center">
                        <div class="text-2xl font-bold">{{ filteredVideos.length }}</div>
                        <div class="text-sm">{{ t('videoManagement.filteredCount') }}</div>
                    </div>
                </div>

                <div v-if="loading" class="text-center py-12">
                    <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
                    <p class="mt-4 text-gray-600">{{ t('common.loading') }}</p>
                </div>

                <div v-else-if="filteredVideos.length === 0" class="text-center py-12">
                    <div class="text-6xl mb-4">📭</div>
                    <p class="text-gray-600 text-lg">{{ t('videoManagement.noVideos') }}</p>
                </div>

                <div v-else class="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <div v-for="item in filteredVideos" :key="item.id"
                        class="relative flex flex-col bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow"
                        :class="{ 'ring-2 ring-blue-500': selectedVideos.has(item.id) }">
                        <div class="absolute top-2 left-2 z-10">
                            <input type="checkbox" :checked="selectedVideos.has(item.id)"
                                @change="toggleSelection(item.id)"
                                class="w-5 h-5 rounded border-gray-300 text-blue-600 focus:ring-blue-500" />
                        </div>

                        <div class="absolute top-2 right-2 z-10 flex gap-1">
                            <span v-if="item.page > 1" class="text-xs px-2 py-1 bg-gray-600 text-white rounded">
                                {{ item.page }}P
                            </span>
                            <span v-if="item.invalid" class="text-xs px-2 py-1 bg-red-500 text-white rounded">
                                {{ t('videoManagement.invalid') }}
                            </span>
                            <span v-else-if="item.frozen" class="text-xs px-2 py-1 bg-orange-500 text-white rounded">
                                {{ t('videoManagement.frozen') }}
                            </span>
                        </div>

                        <div @click="toggleSelection(item.id)" class="cursor-pointer">
                            <Image class="rounded-t-lg w-full h-48 object-cover hover:scale-105 transition-all duration-300"
                                :src="item.cover_info?.image_url ?? '/assets/images/notfound.webp'"
                                :class="{ 'grayscale-image': item.video_downloaded_num == 0 && item.audio_downloaded_num == 0 }" :title="item.title" />
                        </div>

                        <div class="p-3 flex-1 flex flex-col">
                            <RouterLink :to="{ name: 'video-id', params: { id: item.id } }" target="_blank">
                                <h3 class="text-sm font-medium line-clamp-2 hover:text-blue-600 transition-colors"
                                    :title="item.title">{{ item.title }}</h3>
                            </RouterLink>

                            <div class="mt-auto pt-3 space-y-1">
                                <div class="flex justify-between text-xs text-gray-500">
                                    <span>{{ t('videoManagement.published') }}: {{ formatTimestamp(item.pubtime, "yyyy.mm.dd") }}</span>
                                    <span v-if="item.video_downloaded_num > 0 || item.audio_downloaded_num > 0" class="text-green-600 font-medium">
                                        ✓ {{ t('videoManagement.cached') }}
                                    </span>
                                </div>
                                <div v-if="item.fav_time > 0" class="text-xs text-gray-500">
                                    {{ t('videoManagement.favorited') }}: {{ formatTimestamp(item.fav_time, "yyyy.mm.dd") }}
                                </div>
                            </div>

                            <div class="mt-3 pt-3 border-t flex gap-2">
                                <RouterLink :to="{ name: 'video-id', params: { id: item.id } }" target="_blank"
                                    class="flex-1 px-3 py-1.5 text-center text-sm bg-blue-500 hover:bg-blue-600 text-white rounded transition-colors">
                                    {{ t('videoManagement.view') }}
                                </RouterLink>
                                <button @click="confirmDelete(item)"
                                    class="px-3 py-1.5 text-sm bg-red-500 hover:bg-red-600 text-white rounded transition-colors">
                                    {{ t('videoManagement.delete') }}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>

                <div v-if="!loading && filteredVideos.length > 0" class="text-center py-8">
                    <div v-if="isLoadingMore" class="flex items-center justify-center gap-2">
                        <div class="inline-block animate-spin rounded-full h-6 w-6 border-b-2 border-blue-500"></div>
                        <span class="text-gray-600">{{ t('common.loading') }}</span>
                    </div>
                    <div v-else-if="!hasMore" class="text-gray-500">
                        {{ t('videoManagement.noMoreVideos') }}
                    </div>
                </div>
            </div>
        </div>

        <Teleport to="body">
            <div v-if="showDeleteDialog"
                class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4"
                @click.self="showDeleteDialog = false">
                <div class="bg-white rounded-lg p-6 max-w-md w-full">
                    <h3 class="text-lg font-semibold mb-4">{{ t('videoManagement.confirmDeleteTitle') }}</h3>
                    <p class="text-gray-600 mb-6">
                        {{ deleteTarget ? t('videoManagement.confirmDeleteMessage') :
                            t('videoManagement.confirmBatchDeleteMessage', { count: selectedVideos.size }) }}
                    </p>
                    <div v-if="deleteTarget" class="mb-4 p-3 bg-gray-50 rounded">
                        <p class="text-sm font-medium line-clamp-2">{{ deleteTarget.title }}</p>
                    </div>
                    <div class="flex gap-3 justify-end">
                        <button @click="showDeleteDialog = false"
                            class="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 transition-colors">
                            {{ t('common.cancel') }}
                        </button>
                        <button @click="handleDelete"
                            class="px-4 py-2 bg-red-500 hover:bg-red-600 text-white rounded-lg transition-colors">
                            {{ t('common.confirm') }}
                        </button>
                    </div>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<script lang="ts" setup>
import { computed, ref, onMounted, onUnmounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Image from '@/components/Image.vue'
import { formatTimestamp } from '../lib/helper'
import { deleteVideo, getVideoList } from '@/api/video'
import { getFavList, type Favorite, type Video as VideoType } from '@/api/fav'

const { t } = useI18n()

const videoList = ref<VideoType[]>([])
const loading = ref(true)
const searchQuery = ref('')
const selectedVideos = ref(new Set<number>())
const showDeleteDialog = ref(false)
const deleteTarget = ref<VideoType | null>(null)
const favList = ref<Favorite[]>([])

const currentPage = ref(1)
const hasMore = ref(true)
const isLoadingMore = ref(false)

const stat = ref({
    count: 0,
    downloaded: 0,
    invalid: 0,
    valid: 0,
    frozen: 0,
})

const filters = ref({
    favId: '',
    status: 'all',
    downloaded: 'all',
    multiPart: 'all',
})

const filteredVideos = computed(() => videoList.value)

const handleSearch = () => resetAndLoad()

const clearSearch = () => {
    searchQuery.value = ''
    resetAndLoad()
}

const resetFilters = () => {
    searchQuery.value = ''
    filters.value = {
        favId: '',
        status: 'all',
        downloaded: 'all',
        multiPart: 'all'
    }
    selectedVideos.value.clear()
    resetAndLoad()
}

const resetAndLoad = () => {
    currentPage.value = 1
    hasMore.value = true
    videoList.value = []
    selectedVideos.value.clear()
    loadVideos(500, true)
}

const selectAll = () => {
    if (selectedVideos.value.size === filteredVideos.value.length && filteredVideos.value.length > 0) {
        selectedVideos.value.clear()
    } else {
        selectedVideos.value = new Set(filteredVideos.value.map(v => v.id))
    }
}

const toggleSelection = (id: number) => {
    if (selectedVideos.value.has(id)) {
        selectedVideos.value.delete(id)
    } else {
        selectedVideos.value.add(id)
    }
}

const confirmDelete = (video: VideoType) => {
    deleteTarget.value = video
    showDeleteDialog.value = true
}

const confirmBatchDelete = () => {
    if (selectedVideos.value.size === 0) return
    deleteTarget.value = null
    showDeleteDialog.value = true
}

const handleDelete = async () => {
    const idsToDelete = deleteTarget.value ? [deleteTarget.value.id] : Array.from(selectedVideos.value)

    try {
        await deleteVideo(idsToDelete[0], idsToDelete.slice(1))
        videoList.value = videoList.value.filter(v => !idsToDelete.includes(v.id))
        updateStats()
        selectedVideos.value.clear()
        deleteTarget.value = null
        showDeleteDialog.value = false
        alert(t('videoManagement.deleteSuccess'))
    } catch (error) {
        console.error('Delete failed:', error)
    }
}

const updateStats = () => {
    stat.value.count = videoList.value.length
    stat.value.valid = videoList.value.filter(v => !v.invalid).length
    stat.value.invalid = videoList.value.filter(v => v.invalid).length
    stat.value.frozen = videoList.value.filter(v => v.frozen).length
    stat.value.downloaded = videoList.value.filter(v => v.video_downloaded_num > 0).length
}

const loadingTimeout = ref<ReturnType<typeof setTimeout> | null>(null)

const loadVideos = async (sleep: number = 500, isReset: boolean = false) => {
    if (!isReset && (isLoadingMore.value || !hasMore.value)) return

    if (loadingTimeout.value) clearTimeout(loadingTimeout.value)
    
    loadingTimeout.value = setTimeout(async () => {
        try {
            isReset ? (loading.value = true) : (isLoadingMore.value = true)

            const jsonData = await getVideoList({
                query: searchQuery.value,
                page: currentPage.value,
                status: filters.value.status,
                downloaded: filters.value.downloaded,
                multi_part: filters.value.multiPart,
                fav_id: filters.value.favId,
            })

            const newVideos = jsonData.list ?? []

            if (newVideos.length === 0) {
                hasMore.value = false
            } else {
                videoList.value = isReset ? newVideos : [...videoList.value, ...newVideos]
            }

            if (isReset || currentPage.value === 1) {
                stat.value = jsonData.stat ?? stat.value
            }
        } catch (error) {
            console.error('Failed to load videos:', error)
        } finally {
            loading.value = false
            isLoadingMore.value = false
        }
    }, sleep)
}

const loadNextPage = () => {
    if (hasMore.value && !isLoadingMore.value && !loading.value) {
        currentPage.value++
        loadVideos(0, false)
    }
}

const handleScroll = () => {
    const scrollTop = window.scrollY || document.documentElement.scrollTop
    const windowHeight = window.innerHeight
    const documentHeight = document.documentElement.scrollHeight

    if (scrollTop + windowHeight >= documentHeight - 300) {
        loadNextPage()
    }
}

watch(() => [filters.value.status, filters.value.downloaded, filters.value.multiPart, filters.value.favId], resetAndLoad, { deep: true })

onMounted(() => {
    getFavList().then(res => favList.value = res)
    window.addEventListener('scroll', handleScroll)
    loadVideos(0, true)
})

onUnmounted(() => {
    window.removeEventListener('scroll', handleScroll)
    if (loadingTimeout.value) clearTimeout(loadingTimeout.value)
})
</script>

<style scoped>


.line-clamp-2 {
    display: -webkit-box;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}
</style>
