import type { Subscription } from "./subscription";
import type { Upper } from "./upper";

export interface Favorite {
    id: number;
    title: string;
    description: string;
    image: string;
    url: string;
    ctime: number;
    mtime: number;
    cache_image_url: string;
    media_count: number;
    videos: Video[]| null;
}

export interface VideoPartType {
    id: number
    url: string
    title: string
    part: number
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
    cache_image: string;
    page: number;
    fav_time: number;
    created_at: string;
    updated_at: string;
    video_downloaded_at: string;
    video_downloaded_num: number;
    cache_image_url: string;
    favorite: Favorite[]|null;
    video_parts: VideoPartType[]|null;
    danmaku_count: number
    subscriptions: Subscription[]|null;
    upper: Upper|null;
}


export async function getFavList(): Promise<Favorite[]> {
    const response = await fetch('/api/fav');
    return response.json();
}

export async function getFavDetail(id: number): Promise<Favorite> {
    const response = await fetch(`/api/fav/${id}`);
    return response.json();
}
