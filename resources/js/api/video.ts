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