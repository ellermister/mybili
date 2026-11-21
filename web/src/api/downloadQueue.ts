export type TaskStatus = 'pending' | 'queued' | 'running' | 'done' | 'failed' | 'cancelled'

export interface TaskSubject {
    title?: string
    bvid?: string
    page?: number
    part?: string
    kind?: string
    pull_all?: boolean
}

export interface TaskItem {
    id: number
    queue: string
    type: string
    unique_key: string
    status: TaskStatus
    subject?: TaskSubject
    image_url: string
    href: string
    payload: string
    error_msg: string | null
    retry_count: number
    max_retries: number
    created_at: string | null
    started_at: string | null
    finished_at: string | null
    updated_at: string | null
}

type Translate = (key: string, params?: Record<string, unknown>) => string

const typeKey: Record<string, string> = {
    'download-video': 'queue.types.downloadVideo',
    'download-danmaku': 'queue.types.downloadDanmaku',
    'download-audio': 'queue.types.downloadAudio',
    'fav-list': 'queue.types.favList',
    'fix-freeze': 'queue.types.fixFreeze',
    'sync-video': 'queue.types.syncVideo',
    'sync-audio': 'queue.types.syncAudio',
    'readable-names': 'queue.types.readableNames',
    'readable-link': 'queue.types.readableLink',
}

export function taskTitle(translate: Translate, item: TaskItem): string {
    const subject = item.subject ?? {}
    const name = (subject.title || subject.bvid || '').trim()
    const page = subject.page ?? 0
    let key = typeKey[item.type] ?? 'queue.types.unknown'
    if (item.type === 'fav-videos') {
        key = page > 0 ? 'queue.types.favVideosPage' : 'queue.types.favVideosAll'
    } else if (item.type === 'cover') {
        if (subject.kind === 'avatar') key = 'queue.types.coverAvatar'
        else if (subject.kind === 'favorite') key = 'queue.types.coverFavorite'
        else if (subject.kind === 'video') key = 'queue.types.coverVideo'
        else key = 'queue.types.cover'
    } else if (item.type === 'subscription') {
        key = subject.pull_all ? 'queue.types.subscriptionAll' : 'queue.types.subscription'
    } else if (item.type === 'fix-invalid-fav') {
        key = name ? 'queue.types.fixInvalid' : 'queue.types.fixInvalidPlain'
    }
    return translate(key, { name, page, type: item.type }).replace(/\s+/g, ' ').trim()
}

export function taskDetail(item: TaskItem): string {
    const subject = item.subject
    if (!subject) return ''
    const bits: string[] = []
    if (subject.bvid && subject.title) bits.push(subject.bvid)
    const showPage = (subject.page ?? 0) > 1 || (item.type === 'readable-link' && (subject.page ?? 0) > 0)
    if (showPage) bits.push(`P${subject.page}`)
    if (subject.part && item.type === 'readable-link') bits.push(subject.part)
    return bits.join(' ')
}

export interface TaskStat {
    pending: number
    queued: number
    running: number
    done: number
    failed: number
    cancelled: number
}

export interface TaskListResponse {
    list: TaskItem[]
    total: number
    stat: TaskStat
}

export async function getQueueList(params: {
    status?: string
    page?: number
    per_page?: number
}): Promise<TaskListResponse> {
    const p = new URLSearchParams()
    if (params.status) p.set('status', params.status)
    if (params.page) p.set('page', String(params.page))
    if (params.per_page) p.set('per_page', String(params.per_page))
    const res = await fetch(`/api/download-queue?${p}`)
    if (!res.ok) {
        throw new Error('加载任务失败')
    }
    return res.json()
}

export async function cancelQueueItem(id: number): Promise<{ success: boolean }> {
    const res = await fetch(`/api/download-queue/${id}/cancel`, { method: 'POST' })
    return res.json()
}

export async function retryQueueItem(id: number): Promise<{ success: boolean }> {
    const res = await fetch(`/api/download-queue/${id}/retry`, { method: 'POST' })
    return res.json()
}

export async function clearQueue(): Promise<{ success: boolean }> {
    const res = await fetch('/api/download-queue/clear', { method: 'POST' })
    return res.json()
}

export interface TaskProgress {
    done: number
    total: number
    text: string
}

export interface TaskOutput {
    lines: string[]
    progress: TaskProgress
}

export async function getTaskOutput(id: number): Promise<TaskOutput> {
    const res = await fetch(`/api/download-queue/${id}/output`)
    if (!res.ok) {
        throw new Error('加载输出失败')
    }
    return res.json()
}
