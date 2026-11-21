import { ref } from 'vue';

export type ToastType = 'success' | 'error' | 'info';

export interface ToastItem {
    id: number;
    message: string;
    type: ToastType;
}

const MAX_TOASTS = 8;

/** 全局共享：最新一条在数组前（便于桌面端自上而下展示） */
export const toasts = ref<ToastItem[]>([]);

let nextId = 1;
const timers = new Map<number, ReturnType<typeof setTimeout>>();

function clearTimer(id: number): void {
    const t = timers.get(id);
    if (t !== undefined) {
        clearTimeout(t);
        timers.delete(id);
    }
}

function remove(id: number): void {
    clearTimer(id);
    toasts.value = toasts.value.filter((x) => x.id !== id);
}

export interface ShowToastOptions {
    type?: ToastType;
    /** 毫秒，0 表示不自动关闭 */
    duration?: number;
}

function trimOverflow(): void {
    while (toasts.value.length > MAX_TOASTS) {
        const dropped = toasts.value.pop();
        if (dropped) {
            clearTimer(dropped.id);
        }
    }
}

/**
 * 全局 Toast，任意组件 import { useToast } from '@/composables/toast' 后调用。
 */
export function useToast() {
    function show(message: string, options?: ShowToastOptions): number {
        const id = nextId++;
        const type = options?.type ?? 'info';
        const duration = options?.duration ?? 5000;

        toasts.value = [{ id, message, type }, ...toasts.value];
        trimOverflow();

        if (duration > 0) {
            clearTimer(id);
            timers.set(
                id,
                setTimeout(() => {
                    remove(id);
                }, duration),
            );
        }

        return id;
    }

    function dismiss(id: number): void {
        remove(id);
    }

    function success(message: string, duration?: number): number {
        return show(message, { type: 'success', duration });
    }

    function error(message: string, duration?: number): number {
        return show(message, { type: 'error', duration });
    }

    function info(message: string, duration?: number): number {
        return show(message, { type: 'info', duration });
    }

    return {
        toasts,
        show,
        dismiss,
        success,
        error,
        info,
    };
}
