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
    load_all?: boolean;
}

export async function getVideoList(data: VideoListParams): Promise<VideoListResponse> {
    // 过滤掉空值，然后转换为 URL 查询字符串
    const filteredData = Object.fromEntries(
        Object.entries(data).filter(([_, v]) => v != null && v !== '')
    );

    // 特殊处理布尔值
    if (filteredData.load_all !== undefined) {
        filteredData.load_all = filteredData.load_all ? '1' : '0';
    }
    
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

export async function deleteVideo(id: number, extend_ids?: number[]): Promise<boolean> {
    const response = await fetch(`/api/videos/${id}`, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({ extend_ids: extend_ids }),
    });
    return response.json();
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