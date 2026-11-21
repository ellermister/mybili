import type { Cover } from "./cover";

export interface Upper {
    mid: number;
    name: string;
    face: string;
    cover_info: Cover|null;
}
