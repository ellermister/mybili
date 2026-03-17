<template>
    <div class="w-full flex justify-center">
        <div class="container mx-auto" id="main">
        <div class="m-4">

            <!-- 顶部标题栏 -->
            <div class="flex items-center justify-between mb-4">
                <h1 class="text-xl font-bold text-gray-800 flex items-center gap-2">
                    <RouterLink to="/" class="text-pink-500 hover:text-pink-600">🌸</RouterLink>
                    下载队列
                </h1>
                <div class="flex items-center gap-3 text-sm">
                    <span v-if="autoRefresh"
                        class="flex items-center gap-1 text-blue-500 text-xs">
                        <span class="inline-block w-2 h-2 rounded-full bg-blue-400 animate-pulse"></span>
                        自动刷新
                    </span>
                    <button @click="loadData(true)"
                        :disabled="loading"
                        class="flex items-center gap-1 px-3 py-1.5 rounded-lg border border-gray-300 hover:bg-gray-100 transition-colors text-gray-600 disabled:opacity-50">
                        <span :class="{ 'animate-spin': loading }">↻</span>
                        刷新
                    </button>
                    <RouterLink to="/videos"
                        class="px-3 py-1.5 rounded-lg bg-blue-500 hover:bg-blue-600 text-white transition-colors">
                        🎬 视频管理
                    </RouterLink>
                </div>
            </div>

            <!-- 统计卡片 -->
            <div class="grid grid-cols-5 gap-2 mb-4">
                <button v-for="tab in tabs" :key="tab.key"
                    @click="switchTab(tab.key)"
                    class="rounded-xl p-3 text-center transition-all"
                    :class="activeTab === tab.key
                        ? `${tab.activeBg} text-white shadow-md scale-[1.02]`
                        : 'bg-white border border-gray-200 hover:border-gray-300 text-gray-700'">
                    <div class="text-xl font-bold">{{ stat[tab.key] ?? 0 }}</div>
                    <div class="text-xs mt-0.5 opacity-90">{{ tab.label }}</div>
                </button>
            </div>

            <!-- 加载中 -->
            <div v-if="loading && items.length === 0" class="flex justify-center py-16">
                <div class="w-10 h-10 border-4 border-blue-400 border-t-transparent rounded-full animate-spin"></div>
            </div>

            <!-- 空状态 -->
            <div v-else-if="!loading && items.length === 0" class="text-center py-16 text-gray-400">
                <div class="text-5xl mb-3">📭</div>
                <div class="text-base">暂无任务</div>
                <div class="text-sm mt-1">当收藏夹中有视频需要下载时，任务会自动加入队列</div>
            </div>

            <!-- 队列列表 -->
            <div v-else class="space-y-2">
                <TransitionGroup name="list">
                    <div v-for="item in items" :key="item.id"
                        class="bg-white rounded-xl border border-gray-200 overflow-hidden flex gap-0 hover:shadow-sm transition-shadow">

                        <!-- 封面 -->
                        <RouterLink :to="{ name: 'video-id', params: { id: item.video_id } }" class="shrink-0 w-24 md:w-32 bg-gray-100 self-stretch">
                            <img v-if="item.video_cover"
                                :src="item.video_cover.image_url"
                                :alt="item.video_title ?? ''"
                                class="w-full h-full object-cover"
                                style="aspect-ratio:4/3"
                                loading="lazy" />
                            <div v-else class="w-full h-full flex items-center justify-center text-gray-300 text-2xl"
                                style="min-height:72px">🎬</div>
                        </RouterLink>

                        <!-- 内容 -->
                        <div class="flex-1 p-3 flex flex-col gap-1 min-w-0">
                            <!-- 标题行 -->
                            <div class="flex items-start gap-2">
                                <RouterLink :to="{ name: 'video-id', params: { id: item.video_id } }" class="flex-1 text-sm font-medium text-gray-800 line-clamp-2 leading-snug"
                                    :title="item.video_title ?? ''">
                                    {{ item.video_title ?? `ID: ${item.video_id}` }}
                                </RouterLink>
                                <!-- 状态徽章 -->
                                <span class="shrink-0 text-xs px-2 py-0.5 rounded-full font-medium"
                                    :class="statusClass(item.status)">
                                    {{ statusLabel(item.status) }}
                                </span>
                            </div>

                            <!-- 标签行 -->
                            <div class="flex items-center gap-2 flex-wrap">
                                <span class="text-xs px-2 py-0.5 rounded-full"
                                    :class="item.type === 'audio'
                                        ? 'bg-purple-100 text-purple-600'
                                        : 'bg-blue-100 text-blue-600'">
                                    {{ item.type === 'audio' ? '🎵 音频' : '🎬 视频' }}
                                </span>
                                <span class="text-xs text-gray-400">
                                    优先级 {{ item.priority }}
                                </span>
                                <span v-if="item.scheduled_at" class="text-xs text-gray-400">
                                    开始于 {{ formatDate(item.scheduled_at) }}
                                </span>
                                <span v-else class="text-xs text-gray-400">
                                    入队 {{ formatDate(item.created_at) }}
                                </span>
                            </div>

                            <!-- 错误信息 -->
                            <div v-if="item.error_msg"
                                class="text-xs text-red-500 bg-red-50 rounded px-2 py-1 line-clamp-2"
                                :title="item.error_msg">
                                {{ item.error_msg }}
                            </div>

                            <!-- running 进度条 -->
                            <div v-if="item.status === 'running'"
                                class="w-full h-1 bg-gray-100 rounded-full overflow-hidden mt-1">
                                <div class="h-full bg-blue-400 rounded-full animate-progress"></div>
                            </div>

                            <!-- 操作按钮 -->
                            <div class="flex items-center gap-2 mt-1">
                                <!-- 取消 -->
                                <button v-if="item.status === 'pending'"
                                    @click="handleCancel(item)"
                                    :disabled="actionLoading.has(item.id)"
                                    class="text-xs px-2.5 py-1 rounded-lg border border-red-300 text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50">
                                    取消
                                </button>

                                <!-- 重试 -->
                                <button v-if="item.status === 'failed' || item.status === 'cancelled'"
                                    @click="handleRetry(item)"
                                    :disabled="actionLoading.has(item.id)"
                                    class="text-xs px-2.5 py-1 rounded-lg border border-green-300 text-green-600 hover:bg-green-50 transition-colors disabled:opacity-50">
                                    重新排队
                                </button>

                                <!-- 优先级调整（仅 pending） -->
                                <template v-if="item.status === 'pending'">
                                    <button @click="handlePriority(item, item.priority + 10)"
                                        :disabled="actionLoading.has(item.id)"
                                        class="text-xs px-2 py-1 rounded-lg border border-gray-300 text-gray-500 hover:bg-gray-50 transition-colors disabled:opacity-50"
                                        title="提升优先级">
                                        ↑ 优先
                                    </button>
                                    <button v-if="item.priority > 0"
                                        @click="handlePriority(item, Math.max(0, item.priority - 10))"
                                        :disabled="actionLoading.has(item.id)"
                                        class="text-xs px-2 py-1 rounded-lg border border-gray-300 text-gray-500 hover:bg-gray-50 transition-colors disabled:opacity-50"
                                        title="降低优先级">
                                        ↓ 降级
                                    </button>
                                </template>

                                <!-- B站链接 -->
                                <a v-if="item.video_bvid"
                                    :href="`https://www.bilibili.com/video/${item.video_bvid}`"
                                    target="_blank"
                                    class="text-xs px-2.5 py-1 rounded-lg border border-gray-300 text-gray-500 hover:bg-gray-50 transition-colors ml-auto">
                                    B站 ↗
                                </a>
                            </div>
                        </div>
                    </div>
                </TransitionGroup>

                <!-- 加载更多 -->
                <div v-if="hasMore" class="text-center py-4">
                    <button @click="loadMore"
                        :disabled="loadingMore"
                        class="px-6 py-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-100 transition-colors disabled:opacity-50">
                        {{ loadingMore ? '加载中...' : '加载更多' }}
                    </button>
                </div>
            </div>

        </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
    getQueueList, getQueueStat, cancelQueueItem, retryQueueItem, setQueuePriority,
    type QueueItem, type QueueStat, type QueueStatus,
} from '@/api/downloadQueue'

// ── 状态 ──────────────────────────────────────────────────
const items       = ref<QueueItem[]>([])
const stat        = ref<QueueStat>({ pending: 0, running: 0, done: 0, failed: 0, cancelled: 0 })
const activeTab   = ref<string>('active')
const loading     = ref(false)
const loadingMore = ref(false)
const actionLoading = ref<Set<number>>(new Set())
const currentPage = ref(1)
const hasMore     = ref(false)
const PER_PAGE    = 30

// ── 标签页配置 ────────────────────────────────────────────
const tabs = [
    { key: 'pending',   label: '待下载', activeBg: 'bg-amber-500' },
    { key: 'running',   label: '下载中', activeBg: 'bg-blue-500'  },
    { key: 'done',      label: '已完成', activeBg: 'bg-green-500' },
    { key: 'failed',    label: '失败',   activeBg: 'bg-red-500'   },
    { key: 'cancelled', label: '已取消', activeBg: 'bg-gray-400'  },
] as const

// "active" 是前端逻辑 tab，映射到 pending+running
const tabToStatusParam = (tab: string): string => {
    if (tab === 'active') return 'pending,running'
    return tab
}

// 自动刷新：pending 或 running 有任务时开启
const autoRefresh = computed(() =>
    activeTab.value === 'active' || activeTab.value === 'pending' || activeTab.value === 'running'
)

// ── 格式化 ────────────────────────────────────────────────
function formatDate(dateStr: string | null): string {
    if (!dateStr) return '-'
    const d = new Date(dateStr)
    const now = new Date()
    const diff = now.getTime() - d.getTime()
    if (diff < 60_000)  return '刚刚'
    if (diff < 3600_000) return `${Math.floor(diff / 60_000)} 分钟前`
    if (diff < 86400_000) return `${Math.floor(diff / 3600_000)} 小时前`
    return `${d.getMonth()+1}/${d.getDate()} ${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`
}

function statusLabel(s: QueueStatus): string {
    return { pending: '等待中', running: '下载中', done: '已完成', failed: '失败', cancelled: '已取消' }[s] ?? s
}

function statusClass(s: QueueStatus): string {
    return {
        pending:   'bg-amber-100 text-amber-700',
        running:   'bg-blue-100 text-blue-700',
        done:      'bg-green-100 text-green-700',
        failed:    'bg-red-100 text-red-700',
        cancelled: 'bg-gray-100 text-gray-500',
    }[s] ?? 'bg-gray-100 text-gray-500'
}

// ── 数据加载 ──────────────────────────────────────────────
async function loadData(reset = false) {
    if (reset) {
        currentPage.value = 1
        items.value = []
    }
    loading.value = true
    try {
        const res = await getQueueList({
            status: tabToStatusParam(activeTab.value),
            page: currentPage.value,
            per_page: PER_PAGE,
        })
        if (reset) {
            items.value = res.list
        } else {
            items.value.push(...res.list)
        }
        stat.value  = res.stat
        hasMore.value = items.value.length < res.total
    } finally {
        loading.value = false
    }
}

async function loadMore() {
    loadingMore.value = true
    currentPage.value++
    try {
        const res = await getQueueList({
            status: tabToStatusParam(activeTab.value),
            page: currentPage.value,
            per_page: PER_PAGE,
        })
        items.value.push(...res.list)
        stat.value  = res.stat
        hasMore.value = items.value.length < res.total
    } finally {
        loadingMore.value = false
    }
}

function switchTab(key: string) {
    activeTab.value = key
    loadData(true)
}

// ── 操作 ──────────────────────────────────────────────────
async function handleCancel(item: QueueItem) {
    actionLoading.value.add(item.id)
    try {
        const res = await cancelQueueItem(item.id)
        if (res.success) {
            item.status = 'cancelled'
            stat.value.pending = Math.max(0, stat.value.pending - 1)
            stat.value.cancelled++
            // 若当前在 pending 标签则移出列表
            if (activeTab.value === 'pending') {
                items.value = items.value.filter(i => i.id !== item.id)
            }
        }
    } finally {
        actionLoading.value.delete(item.id)
    }
}

async function handleRetry(item: QueueItem) {
    actionLoading.value.add(item.id)
    try {
        const res = await retryQueueItem(item.id)
        if (res.success) {
            const prev = item.status as QueueStatus
            item.status = 'pending'
            if (prev === 'failed')    stat.value.failed    = Math.max(0, stat.value.failed - 1)
            if (prev === 'cancelled') stat.value.cancelled = Math.max(0, stat.value.cancelled - 1)
            stat.value.pending++
            // 若当前不是 pending 标签则移出列表
            if (activeTab.value !== 'active' && activeTab.value !== 'pending') {
                items.value = items.value.filter(i => i.id !== item.id)
            }
        }
    } finally {
        actionLoading.value.delete(item.id)
    }
}

async function handlePriority(item: QueueItem, newPriority: number) {
    actionLoading.value.add(item.id)
    try {
        const res = await setQueuePriority(item.id, newPriority)
        if (res.success) {
            item.priority = newPriority
            // 重新按优先级排序
            items.value.sort((a, b) => b.priority - a.priority || a.id - b.id)
        }
    } finally {
        actionLoading.value.delete(item.id)
    }
}

// ── 自动刷新 ──────────────────────────────────────────────
let timer: ReturnType<typeof setInterval> | null = null

function startAutoRefresh() {
    if (timer) return
    timer = setInterval(async () => {
        if (!autoRefresh.value) return
        // 只刷新 stat + 当前页第一页数据（不重置滚动位置，只更新已有项）
        const res = await getQueueList({
            status: tabToStatusParam(activeTab.value),
            page: 1,
            per_page: Math.max(PER_PAGE, items.value.length),
        }).catch(() => null)
        if (res) {
            stat.value = res.stat
            // 增量合并：用新数据更新现有条目状态，同时插入新增条目
            const newMap = new Map(res.list.map(i => [i.id, i]))
            // 更新已有条目（不在新列表里的直接移除，避免列表不缩减导致“不同步”）
            items.value = items.value
                .filter(i => newMap.has(i.id))
                .map(i => newMap.get(i.id)!)
            // 过滤掉不再属于当前 tab 的条目（如 running→done）
            const validStatuses = tabToStatusParam(activeTab.value).split(',')
            items.value = items.value.filter(i => validStatuses.includes(i.status))
            // 插入新增条目（排在前面）
            const existIds = new Set(items.value.map(i => i.id))
            const newItems = res.list.filter(i => !existIds.has(i.id))
            if (newItems.length > 0) {
                items.value = [...newItems, ...items.value]
            }
        }
    }, 5000)
}

function stopAutoRefresh() {
    if (timer) { clearInterval(timer); timer = null }
}

onMounted(() => {
    loadData(true)
    startAutoRefresh()
})

onUnmounted(() => stopAutoRefresh())
</script>

<style scoped>
.list-enter-active,
.list-leave-active {
    transition: all 0.3s ease;
}
.list-enter-from,
.list-leave-to {
    opacity: 0;
    transform: translateX(-12px);
}

@keyframes progress-indeterminate {
    0%   { transform: translateX(-100%) scaleX(0.5); }
    50%  { transform: translateX(0%)    scaleX(1);   }
    100% { transform: translateX(100%)  scaleX(0.5); }
}
.animate-progress {
    animation: progress-indeterminate 1.6s ease-in-out infinite;
    width: 40%;
}
</style>
