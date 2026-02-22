import type { Cover } from "./cover"

export type QueueStatus = 'pending' | 'running' | 'done' | 'failed' | 'cancelled'
export type QueueType = 'video' | 'audio'

export interface QueueItem {
    id: number
    type: QueueType
    video_id: number
    video_part_id: number | null
    status: QueueStatus
    priority: number
    error_msg: string | null
    scheduled_at: string | null
    completed_at: string | null
    created_at: string
    video_title: string | null
    video_cover: Cover | null
    video_bvid: string | null
    video_type: number | null
}

export interface QueueStat {
    pending: number
    running: number
    done: number
    failed: number
    cancelled: number
}

export interface QueueListResponse {
    list: QueueItem[]
    total: number
    stat: QueueStat
}

export async function getQueueList(params: {
    status?: string
    page?: number
    per_page?: number
}): Promise<QueueListResponse> {
    const p = new URLSearchParams()
    if (params.status)   p.set('status', params.status)
    if (params.page)     p.set('page', String(params.page))
    if (params.per_page) p.set('per_page', String(params.per_page))
    const res = await fetch(`/api/download-queue?${p}`)
    return res.json()
}

export async function getQueueStat(): Promise<QueueStat> {
    const res = await fetch('/api/download-queue/stat')
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

export async function setQueuePriority(id: number, priority: number): Promise<{ success: boolean }> {
    const res = await fetch(`/api/download-queue/${id}/priority`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ priority }),
    })
    return res.json()
}
