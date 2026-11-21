import type { Video } from "./fav";

export interface VideoListResponse {
    stat: {
        count: number;
        downloaded: number;
        invalid: number;
        valid: number;
        frozen: number;
    };
    list: Video[];
}

export interface VideoListParams {
    query?: string;
    page?: number;
    status?: string;
    downloaded?: string;
    multi_part?: string;
    fav_id?: string;
}

export interface DeleteVideoOptions {
    extend_ids?: number[];
    permanent?: boolean;
    requeue?: boolean;
}

export async function getVideoList(data: VideoListParams): Promise<VideoListResponse> {
    // 过滤掉空值，然后转换为 URL 查询字符串
    const filteredData = Object.fromEntries(
        Object.entries(data).filter(([_, v]) => v != null && v !== '')
    );
    const params = new URLSearchParams(filteredData as Record<string, string>);
    const url = `/api/videos${params.toString() ? '?' + params.toString() : ''}`;

    const response = await fetch(url, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json',
        },
    });
    return response.json();
}

export interface DeleteVideoResponse {
    code: number;
    message?: string;
    deleted_ids?: number[];
}

export async function deleteVideo(
    id: number,
    extend_ids?: number[],
    options?: Omit<DeleteVideoOptions, 'extend_ids'>,
): Promise<DeleteVideoResponse> {
    const response = await fetch(`/api/videos/${id}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            extend_ids: extend_ids,
            permanent: options?.permanent ?? true,
            requeue: options?.requeue ?? false,
        }),
    });
    return response.json() as Promise<DeleteVideoResponse>;
}

export async function getVideoDanmaku(id: number): Promise<any[]> {
    const response = await fetch(`/api/danmaku?id=${id}`, {
        method: 'GET',
    });
    const data = await response.json()
    return data.data;
}

export async function getVideoInfo(id: number): Promise<Video> {
    const response = await fetch(`/api/videos/${id}`, {
        method: 'GET',
    });
    return response.json();
}

export interface RefreshDanmakuResponse {
    code: number;
    /** 失败时由后端返回说明（如音频、无分 P） */
    message?: string;
    parts_queued: number;
}

export async function refreshVideoDanmaku(videoId: number): Promise<RefreshDanmakuResponse> {
    const response = await fetch(`/api/videos/${videoId}/danmaku/refresh`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
        },
    });
    return response.json() as Promise<RefreshDanmakuResponse>;
}