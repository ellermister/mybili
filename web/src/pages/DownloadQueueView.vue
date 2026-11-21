<template>
    <div class="w-full flex justify-center">
        <div class="container mx-auto" id="main">
        <div class="m-4">

            <div class="flex items-center justify-between mb-4">
                <h1 class="text-xl font-bold text-gray-800 flex items-center gap-2">
                    <RouterLink to="/" class="text-pink-500 hover:text-pink-600">🌸</RouterLink>
                    {{ t('queue.title') }}
                </h1>
                <div class="flex items-center gap-3 text-sm">
                    <span v-if="autoRefresh"
                        class="flex items-center gap-1 text-blue-500 text-xs">
                        <span class="inline-block w-2 h-2 rounded-full bg-blue-400 animate-pulse"></span>
                        {{ t('queue.autoRefresh') }}
                    </span>
                    <button @click="handleClear"
                        :disabled="clearing"
                        class="px-3 py-1.5 rounded-lg border border-red-300 text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50">
                        {{ t('queue.clear') }}
                    </button>
                    <button @click="loadData(true)"
                        :disabled="loading"
                        class="flex items-center gap-1 px-3 py-1.5 rounded-lg border border-gray-300 hover:bg-gray-100 transition-colors text-gray-600 disabled:opacity-50">
                        <span :class="{ 'animate-spin': loading }">↻</span>
                        {{ t('queue.refresh') }}
                    </button>
                    <RouterLink to="/videos"
                        class="px-3 py-1.5 rounded-lg bg-blue-500 hover:bg-blue-600 text-white transition-colors">
                        🎬 {{ t('queue.videoManagement') }}
                    </RouterLink>
                </div>
            </div>

            <div class="grid grid-cols-2 sm:grid-cols-5 gap-2 mb-4">
                <button v-for="tab in tabs" :key="tab.key"
                    @click="switchTab(tab.key)"
                    class="rounded-xl p-3 text-center transition-all"
                    :class="activeTab === tab.key
                        ? `${tab.activeBg} text-white shadow-md scale-[1.02]`
                        : 'bg-white border border-gray-200 hover:border-gray-300 text-gray-700'">
                    <div class="text-xl font-bold">{{ tabCount(tab.key) }}</div>
                    <div class="text-xs mt-0.5 opacity-90">{{ t(tab.label) }}</div>
                </button>
            </div>

            <div v-if="loading && items.length === 0" class="flex justify-center py-16">
                <div class="w-10 h-10 border-4 border-blue-400 border-t-transparent rounded-full animate-spin"></div>
            </div>

            <div v-else-if="!loading && items.length === 0" class="text-center py-16 text-gray-400">
                <div class="text-5xl mb-3">📭</div>
                <div class="text-base">{{ t('queue.empty') }}</div>
                <div class="text-sm mt-1">{{ t('queue.emptyHint') }}</div>
            </div>

            <div v-else class="space-y-2">
                <TransitionGroup name="list">
                    <div v-for="item in items" :key="item.id"
                        class="bg-white rounded-xl border border-gray-200 overflow-hidden hover:shadow-sm transition-shadow flex">
                        <component v-if="item.image_url"
                            :is="item.href ? 'router-link' : 'div'"
                            :to="item.href || undefined"
                            class="shrink-0 w-28 sm:w-36 bg-gray-100 self-stretch">
                            <img :src="item.image_url"
                                :alt="titleOf(item)"
                                referrerpolicy="no-referrer"
                                class="w-full h-full object-cover"
                                style="aspect-ratio:4/3; min-height:84px"
                                loading="lazy" />
                        </component>
                        <div class="p-3 flex flex-col gap-1 min-w-0 flex-1">
                            <div class="flex items-start gap-2">
                                <div class="flex-1 min-w-0">
                                    <component :is="item.href ? 'router-link' : 'div'"
                                        :to="item.href || undefined"
                                        class="text-sm font-medium text-gray-800 leading-snug line-clamp-2"
                                        :title="titleOf(item)">
                                        {{ titleOf(item) }}
                                    </component>
                                    <div v-if="taskDetail(item)" class="text-xs text-gray-500 mt-0.5 truncate">{{ taskDetail(item) }}</div>
                                </div>
                                <span class="shrink-0 text-xs px-2 py-0.5 rounded-full font-medium"
                                    :class="statusClass(item.status)">
                                    {{ statusLabel(item.status) }}
                                </span>
                            </div>

                            <div class="flex items-center gap-2 flex-wrap">
                                <span class="text-xs text-gray-400">{{ timeHint(item) }}</span>
                                <span v-if="item.retry_count > 0" class="text-xs text-gray-400">
                                    {{ t('queue.retried', { count: item.retry_count, max: item.max_retries }) }}
                                </span>
                            </div>

                            <div v-if="item.error_msg"
                                class="text-xs text-red-500 bg-red-50 rounded px-2 py-1 line-clamp-2"
                                :title="item.error_msg">
                                {{ item.error_msg }}
                            </div>

                            <div v-if="item.status === 'running'"
                                class="w-full h-1 bg-gray-100 rounded-full overflow-hidden mt-1">
                                <div v-if="progressPercent(item.id) > 0"
                                    class="h-full bg-blue-500 rounded-full"
                                    :style="{ width: progressPercent(item.id) + '%' }"></div>
                                <div v-else class="h-full bg-blue-400 rounded-full animate-progress"></div>
                            </div>

                            <div v-if="openID === item.id" class="mt-1">
                                <div v-if="progressText(item.id)" class="text-xs text-gray-500 mb-1">{{ progressText(item.id) }}</div>
                                <pre :id="'task-log-' + item.id"
                                    @scroll="onOutputScroll"
                                    class="text-xs leading-5 bg-gray-900 text-gray-100 rounded-lg p-2 max-h-64 overflow-auto whitespace-pre-wrap break-all">{{ outputText(item.id) }}</pre>
                            </div>

                            <div class="flex items-center gap-2 mt-1">
                                <button @click="toggleOutput(item)"
                                    class="text-xs px-2.5 py-1 rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-100 transition-colors">
                                    {{ openID === item.id ? t('queue.hideOutput') : t('queue.output') }}
                                </button>
                                <button v-if="item.status === 'pending' || item.status === 'queued' || item.status === 'running'"
                                    @click="handleCancel(item)"
                                    :disabled="actionLoading.has(item.id)"
                                    class="text-xs px-2.5 py-1 rounded-lg border border-red-300 text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50">
                                    {{ t('queue.cancel') }}
                                </button>
                                <button v-if="item.status === 'failed' || item.status === 'cancelled'"
                                    @click="handleRetry(item)"
                                    :disabled="actionLoading.has(item.id)"
                                    class="text-xs px-2.5 py-1 rounded-lg border border-green-300 text-green-600 hover:bg-green-50 transition-colors disabled:opacity-50">
                                    {{ t('queue.retry') }}
                                </button>
                            </div>
                        </div>
                    </div>
                </TransitionGroup>

                <div v-if="hasMore" class="text-center py-4">
                    <button @click="loadMore"
                        :disabled="loadingMore"
                        class="px-6 py-2 rounded-lg border border-gray-300 text-gray-600 hover:bg-gray-100 transition-colors disabled:opacity-50">
                        {{ loadingMore ? t('queue.loading') : t('queue.loadMore') }}
                    </button>
                </div>
            </div>

        </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import {
    getQueueList, cancelQueueItem, retryQueueItem, clearQueue, getTaskOutput,
    taskTitle, taskDetail,
    type TaskItem, type TaskOutput, type TaskStat, type TaskStatus,
} from '@/api/downloadQueue'

const { t } = useI18n()

const items = ref<TaskItem[]>([])
const stat = ref<TaskStat>({ pending: 0, queued: 0, running: 0, done: 0, failed: 0, cancelled: 0 })
const activeTab = ref('waiting')
const loading = ref(false)
const loadingMore = ref(false)
const clearing = ref(false)
const actionLoading = ref(new Set<number>())
const openID = ref<number | null>(null)
const outputs = ref<Record<number, TaskOutput>>({})
const stickBottom = ref(true)

function setBusy(id: number, on: boolean) {
    const next = new Set(actionLoading.value)
    if (on) next.add(id)
    else next.delete(id)
    actionLoading.value = next
}
const currentPage = ref(1)
const hasMore = ref(false)
const PER_PAGE = 30

const tabs = [
    { key: 'waiting', label: 'queue.tabs.waiting', activeBg: 'bg-amber-500' },
    { key: 'running', label: 'queue.tabs.running', activeBg: 'bg-blue-500' },
    { key: 'done', label: 'queue.tabs.done', activeBg: 'bg-green-500' },
    { key: 'failed', label: 'queue.tabs.failed', activeBg: 'bg-red-500' },
    { key: 'cancelled', label: 'queue.tabs.cancelled', activeBg: 'bg-gray-400' },
] as const

function tabToStatus(tab: string): string {
    if (tab === 'waiting') return 'pending,queued'
    return tab
}

function tabCount(key: string): number {
    if (key === 'waiting') return (stat.value.pending ?? 0) + (stat.value.queued ?? 0)
    return stat.value[key as TaskStatus] ?? 0
}

const autoRefresh = computed(() => activeTab.value === 'waiting' || activeTab.value === 'running')

function formatDate(dateStr: string | null): string {
    if (!dateStr) return '-'
    const d = new Date(dateStr)
    if (Number.isNaN(d.getTime())) return dateStr
    const diff = Date.now() - d.getTime()
    if (diff < 60_000) return t('queue.justNow')
    if (diff < 3600_000) return t('queue.minutesAgo', { n: Math.floor(diff / 60_000) })
    if (diff < 86400_000) return t('queue.hoursAgo', { n: Math.floor(diff / 3600_000) })
    return `${d.getMonth() + 1}/${d.getDate()} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function timeHint(item: TaskItem): string {
    if (item.status === 'running' && item.started_at) return t('queue.startedAt', { time: formatDate(item.started_at) })
    if ((item.status === 'done' || item.status === 'failed' || item.status === 'cancelled') && item.finished_at) {
        return t('queue.finishedAt', { time: formatDate(item.finished_at) })
    }
    return t('queue.updatedAt', { time: formatDate(item.updated_at ?? item.created_at) })
}

function titleOf(item: TaskItem): string {
    return taskTitle((key, params) => t(key, params ?? {}), item)
}

function statusLabel(s: TaskStatus): string {
    return t(`queue.status.${s}`)
}

function statusClass(s: TaskStatus): string {
    return {
        pending: 'bg-amber-100 text-amber-700',
        queued: 'bg-orange-100 text-orange-700',
        running: 'bg-blue-100 text-blue-700',
        done: 'bg-green-100 text-green-700',
        failed: 'bg-red-100 text-red-700',
        cancelled: 'bg-gray-100 text-gray-500',
    }[s] ?? 'bg-gray-100 text-gray-500'
}

async function loadData(reset = false) {
    if (reset) {
        currentPage.value = 1
        items.value = []
    }
    loading.value = true
    try {
        const res = await getQueueList({
            status: tabToStatus(activeTab.value),
            page: currentPage.value,
            per_page: PER_PAGE,
        })
        items.value = reset ? res.list : items.value.concat(res.list)
        stat.value = res.stat
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
            status: tabToStatus(activeTab.value),
            page: currentPage.value,
            per_page: PER_PAGE,
        })
        items.value = items.value.concat(res.list)
        stat.value = res.stat
        hasMore.value = items.value.length < res.total
    } finally {
        loadingMore.value = false
    }
}

function switchTab(key: string) {
    activeTab.value = key
    loadData(true)
}

function bumpStat(from: TaskStatus, to: TaskStatus) {
    stat.value[from] = Math.max(0, (stat.value[from] ?? 0) - 1)
    stat.value[to] = (stat.value[to] ?? 0) + 1
}

function outputOf(id: number): TaskOutput | undefined {
    return outputs.value[id]
}

function progressPercent(id: number): number {
    const progress = outputOf(id)?.progress
    if (!progress || progress.total <= 0) return 0
    return Math.max(0, Math.min(100, Math.round(progress.done * 100 / progress.total)))
}

function progressText(id: number): string {
    return outputOf(id)?.progress.text ?? ''
}

function outputText(id: number): string {
    const lines = outputOf(id)?.lines
    if (!lines || lines.length === 0) return t('queue.noOutput')
    return lines.join('\n')
}

function onOutputScroll(event: Event) {
    const el = event.target as HTMLElement
    stickBottom.value = el.scrollHeight - el.scrollTop - el.clientHeight <= 4
}

async function loadOutput(id: number) {
    const res = await getTaskOutput(id).catch(() => null)
    if (!res || openID.value !== id) return
    const follow = stickBottom.value
    outputs.value = { ...outputs.value, [id]: res }
    if (!follow) return
    await nextTick()
    const el = document.getElementById('task-log-' + id)
    if (el) el.scrollTop = el.scrollHeight
}

function toggleOutput(item: TaskItem) {
    if (openID.value === item.id) {
        openID.value = null
        return
    }
    stickBottom.value = true
    openID.value = item.id
    loadOutput(item.id)
}

async function handleClear() {
    if (!window.confirm(t('queue.clearConfirm'))) return
    clearing.value = true
    try {
        const res = await clearQueue()
        if (res.success) await loadData(true)
    } finally {
        clearing.value = false
    }
}

async function handleCancel(item: TaskItem) {
    setBusy(item.id, true)
    try {
        const res = await cancelQueueItem(item.id)
        if (!res.success) return
        const prev = item.status
        item.status = 'cancelled'
        bumpStat(prev, 'cancelled')
        if (activeTab.value === 'waiting' || activeTab.value === 'running') {
            items.value = items.value.filter(row => row.id !== item.id)
        }
    } finally {
        setBusy(item.id, false)
    }
}

async function handleRetry(item: TaskItem) {
    setBusy(item.id, true)
    try {
        const res = await retryQueueItem(item.id)
        if (!res.success) return
        const prev = item.status
        item.status = 'pending'
        item.error_msg = null
        item.retry_count = 0
        bumpStat(prev, 'pending')
        if (activeTab.value !== 'waiting') {
            items.value = items.value.filter(row => row.id !== item.id)
        }
    } finally {
        setBusy(item.id, false)
    }
}

let timer: ReturnType<typeof setInterval> | null = null

function startAutoRefresh() {
    if (timer) return
    timer = setInterval(async () => {
        if (!autoRefresh.value) return
        const res = await getQueueList({
            status: tabToStatus(activeTab.value),
            page: 1,
            per_page: Math.max(PER_PAGE, items.value.length),
        }).catch(() => null)
        if (!res) return
        stat.value = res.stat
        const fresh = new Map(res.list.map(row => [row.id, row]))
        const valid = tabToStatus(activeTab.value).split(',')
        const kept = items.value.filter(row => fresh.has(row.id) && valid.includes(row.status)).map(row => fresh.get(row.id)!)
        const exist = new Set(kept.map(row => row.id))
        const added = res.list.filter(row => !exist.has(row.id))
        items.value = added.concat(kept)
        hasMore.value = items.value.length < res.total
        if (openID.value) loadOutput(openID.value)
    }, 5000)
}

let outputTimer: ReturnType<typeof setInterval> | null = null

function startOutputRefresh() {
    if (outputTimer) return
    outputTimer = setInterval(() => {
        if (openID.value) loadOutput(openID.value)
    }, 1000)
}

onMounted(() => {
    loadData(true)
    startAutoRefresh()
    startOutputRefresh()
})

onUnmounted(() => {
    if (timer) clearInterval(timer)
    if (outputTimer) clearInterval(outputTimer)
})
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
