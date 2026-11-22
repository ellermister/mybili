<template>
    <div class="bg-white/90 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200/50 p-4">
        <div class="flex items-center justify-between gap-3 mb-3">
            <h3 class="text-lg font-semibold text-gray-800 flex items-center min-w-0">
                <span class="w-1.5 h-1.5 bg-gray-400 rounded-full mr-2 shrink-0"></span>
                <span class="truncate">{{ t('video.danmakuList') }}</span>
            </h3>
            <span class="text-xs text-gray-500 shrink-0 tabular-nums">
                {{ loading ? t('common.loading') : t('video.danmakuMatched', { shown: rows.length.toLocaleString(), total: total.toLocaleString() }) }}
            </span>
        </div>

        <div class="flex flex-col gap-2 mb-3">
            <input v-model="keywordInput" type="search" enterkeyhint="search"
                class="w-full px-3 py-2 text-sm border border-gray-200 rounded-lg bg-white outline-none focus:border-pink-300"
                :placeholder="t('video.danmakuSearch')" />
            <div class="flex items-center gap-2">
                <input v-model="timeFrom" type="text" inputmode="numeric" autocomplete="off"
                    class="min-w-0 flex-1 px-3 py-2 text-sm border border-gray-200 rounded-lg bg-white outline-none focus:border-pink-300 tabular-nums"
                    :placeholder="t('video.danmakuTimeFrom')" :aria-label="t('video.danmakuTimeFrom')" />
                <span class="text-gray-400 shrink-0">–</span>
                <input v-model="timeTo" type="text" inputmode="numeric" autocomplete="off"
                    class="min-w-0 flex-1 px-3 py-2 text-sm border border-gray-200 rounded-lg bg-white outline-none focus:border-pink-300 tabular-nums"
                    :placeholder="t('video.danmakuTimeTo')" :aria-label="t('video.danmakuTimeTo')" />
                <button type="button"
                    class="shrink-0 px-3 py-2 text-sm rounded-lg border border-gray-200 bg-gray-50 text-gray-700 hover:bg-gray-100"
                    @click="descending = !descending">
                    {{ descending ? t('video.danmakuSortDesc') : t('video.danmakuSortAsc') }}
                </button>
            </div>
        </div>

        <div class="h-80 sm:h-96 rounded-lg border border-gray-100 bg-white overflow-hidden">
            <div v-if="loading" class="h-full flex items-center justify-center text-sm text-gray-500">
                {{ t('common.loading') }}
            </div>
            <div v-else-if="rows.length === 0" class="h-full flex items-center justify-center text-sm text-gray-500 px-4 text-center">
                {{ total === 0 ? t('video.danmakuEmpty') : t('video.danmakuNoMatch') }}
            </div>
            <VList v-else :data="rows" :itemSize="44" class="h-full">
                <template #default="{ item }">
                    <div class="w-full h-11 px-2 flex items-center gap-2 text-left border-b border-gray-100 hover:bg-pink-50/70 cursor-pointer"
                        @click="onRowClick(item.time ?? 0)">
                        <span class="w-16 shrink-0 text-xs tabular-nums text-gray-500 select-none">{{ formatPlayTime(item.time) }}</span>
                        <span class="w-2.5 h-2.5 rounded-full ring-1 ring-gray-300 shrink-0 select-none"
                            :style="{ backgroundColor: item.color || '#ffffff' }"></span>
                        <span class="min-w-0 flex-1 overflow-hidden whitespace-nowrap text-sm text-gray-800 select-text cursor-text">{{ item.text }}</span>
                        <span class="shrink-0 text-[11px] text-gray-400 select-none">{{ modeLabel(item.mode) }}</span>
                    </div>
                </template>
            </VList>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed, onUnmounted, ref, shallowRef, toRaw, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { VList } from 'virtua/vue';

export interface DanmakuItem {
    text?: string
    time?: number
    mode?: number
    color?: string
}

const props = defineProps<{
    items: DanmakuItem[]
    loading: boolean
}>()

const emit = defineEmits<{
    seek: [time: number]
}>()

const onRowClick = (time: number) => {
    const selection = window.getSelection()
    if (selection && !selection.isCollapsed) {
        return
    }
    emit('seek', time)
}

const { t } = useI18n()
const keywordInput = ref('')
const timeFrom = ref('')
const timeTo = ref('')
const descending = ref(false)
const appliedKeyword = ref('')
const appliedFrom = ref('')
const appliedTo = ref('')
const revision = ref(0)
const ordered = shallowRef<DanmakuItem[]>([])

let keywordTimer = 0

watch(() => props.items, (value) => {
    const raw = (toRaw(value) || []) as DanmakuItem[]
    ordered.value = raw.slice().sort((prev, next) => (prev.time ?? 0) - (next.time ?? 0))
    revision.value++
}, { immediate: true })

watch([keywordInput, timeFrom, timeTo], () => {
    window.clearTimeout(keywordTimer)
    keywordTimer = window.setTimeout(() => {
        appliedKeyword.value = keywordInput.value
        appliedFrom.value = timeFrom.value
        appliedTo.value = timeTo.value
    }, 200)
})

onUnmounted(() => {
    window.clearTimeout(keywordTimer)
})

const parseTimeInput = (value: string) => {
    const text = value.trim()
    if (!text) {
        return null
    }
    if (/^\d+(\.\d+)?$/.test(text)) {
        return Number(text)
    }
    const parts = text.split(':')
    if (parts.length < 2 || parts.length > 3) {
        return null
    }
    if (parts.some((part) => part === '' || Number.isNaN(Number(part)))) {
        return null
    }
    const nums = parts.map(Number)
    if (parts.length === 2) {
        return nums[0] * 60 + nums[1]
    }
    return nums[0] * 3600 + nums[1] * 60 + nums[2]
}

const rows = computed(() => {
    revision.value
    const query = appliedKeyword.value.trim().toLowerCase()
    let from = parseTimeInput(appliedFrom.value)
    let to = parseTimeInput(appliedTo.value)
    if (from != null && to != null && from > to) {
        const swap = from
        from = to
        to = swap
    }
    const source = ordered.value
    const start = from == null ? 0 : lowerBound(source, from)
    const end = to == null ? source.length : upperBound(source, to)
    if (!query && !descending.value && start === 0 && end === source.length) {
        return source
    }
    const matched: DanmakuItem[] = []
    for (let i = start; i < end; i++) {
        const item = source[i]
        if (query && !String(item.text ?? '').toLowerCase().includes(query)) {
            continue
        }
        matched.push(item)
    }
    if (descending.value) {
        matched.reverse()
    }
    return matched
})

const lowerBound = (list: DanmakuItem[], time: number) => {
    let lo = 0
    let hi = list.length
    while (lo < hi) {
        const mid = (lo + hi) >> 1
        if ((list[mid].time ?? 0) < time) {
            lo = mid + 1
        } else {
            hi = mid
        }
    }
    return lo
}

const upperBound = (list: DanmakuItem[], time: number) => {
    let lo = 0
    let hi = list.length
    while (lo < hi) {
        const mid = (lo + hi) >> 1
        if ((list[mid].time ?? 0) <= time) {
            lo = mid + 1
        } else {
            hi = mid
        }
    }
    return lo
}

const total = computed(() => {
    revision.value
    return ordered.value.length
})

const formatPlayTime = (time?: number) => {
    const totalSeconds = Math.max(0, Math.floor(time ?? 0))
    const hours = Math.floor(totalSeconds / 3600)
    const minutes = Math.floor((totalSeconds % 3600) / 60)
    const seconds = totalSeconds % 60
    const mm = String(minutes).padStart(2, '0')
    const ss = String(seconds).padStart(2, '0')
    if (hours > 0) {
        return `${hours}:${mm}:${ss}`
    }
    return `${minutes}:${ss}`
}

const modeLabel = (mode?: number) => {
    if (mode === 1) {
        return t('video.danmakuModeTop')
    }
    if (mode === 2) {
        return t('video.danmakuModeBottom')
    }
    return t('video.danmakuModeScroll')
}
</script>
