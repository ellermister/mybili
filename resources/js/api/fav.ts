import type { Subscription } from "./subscription";
import type { Upper } from "./upper";
import type { Cover } from "./cover";

export interface Favorite {
    id: number;
    title: string;
    description: string;
    image: string;
    url: string;
    ctime: number;
    mtime: number;
    media_count: number;
    videos?: Video[]| null;
    cover_info: Cover|null;
}

export interface VideoPartType {
    id: number
    url: string
    title: string
    part: number
    downloaded: boolean
}

export interface Video {
    id: number;
    link: string;
    title: string;
    intro: string;
    cover: string;
    bvid: string;
    pubtime: number;
    attr: number;
    invalid: number;
    frozen: number;
    page: number;
    fav_time: number;
    type: number;
    created_at: string;
    updated_at: string;
    video_downloaded_at: string;
    video_downloaded_num: number;
    audio_downloaded_num: number;
    favorite: Favorite[]|null;
    video_parts: VideoPartType[]|null;
    danmaku_count: number
    subscriptions: Subscription[]|null;
    upper: Upper|null;
    cover_info: Cover|null;
}


export async function getFavList(): Promise<Favorite[]> {
    const response = await fetch('/api/fav');
    return response.json();
}

export async function getFavDetail(id: number): Promise<Favorite> {
    const response = await fetch(`/api/fav/${id}`);
    return response.json();
}

export interface VideoLite {
    id: number;
    title: string;
    bvid: string;
    pubtime: number;
    fav_time: number;
    page: number;
    video_downloaded_num: number;
    audio_downloaded_num: number;
    frozen: number;
    invalid: number;
    cover: string;
    cover_image_url: string | null;
    // 兼容后端历史拼写与新拼写
    cover_image_thumb_url?: string | null;
    created_at: string;
}

export type FavVideo = VideoLite;
export type ProgressVideo = VideoLite;

export async function getFavVideos(id: number): Promise<FavVideo[]> {
    const response = await fetch(`/api/fav/${id}/videos`);
    return response.json();
}
