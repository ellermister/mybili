<template>
    <img
        ref="imgEle"
        class="bg-gray-200"
        :src="src"
        :data-src="resolvedSrc"
        :alt="props.title"
        :loading="resolvedMode === 'direct' ? 'eager' : 'lazy'"
        decoding="async"
        @load="onLoad"
    >
</template>
<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

const defaultSrc = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';

const props = defineProps<{
    src?: string;
    originalSrc?: string;
    title: string;
    directLoad?: boolean;
    mode?: 'observer' | 'observer-preload' | 'direct';
}>();

const imgEle = ref<HTMLImageElement | null>(null);
const src = ref(defaultSrc);
const isLoaded = ref(false);
const io = ref<IntersectionObserver | null>(null);
const fallbackTimer = ref<number | null>(null);
const observerRoot = ref<HTMLElement | null>(null);

const resolvedMode = computed(() => {
    if (props.mode) return props.mode;
    return props.directLoad ? 'direct' : 'observer';
});

const resolvedSrc = computed(() => props.originalSrc ?? props.src ?? '');
const resolvedRootMargin = computed(() => resolvedMode.value === 'observer-preload' ? '800px 0px' : '0px');
const resolvedThreshold = computed(() => resolvedMode.value === 'observer-preload' ? 0 : 0.1);

const cleanupObserver = () => {
    if (io.value) {
        io.value.disconnect();
        io.value = null;
    }
};

const clearFallbackTimer = () => {
    if (fallbackTimer.value !== null) {
        window.clearTimeout(fallbackTimer.value);
        fallbackTimer.value = null;
    }
};

const loadImage = () => {
    if (!resolvedSrc.value) return;
    src.value = resolvedSrc.value;
    clearFallbackTimer();
};

const resetImage = () => {
    isLoaded.value = false;
    src.value = defaultSrc;
    clearFallbackTimer();
};

const getPreloadOffset = () => {
    const firstMargin = resolvedRootMargin.value.trim().split(/\s+/)[0] ?? '0px';
    const value = Number.parseInt(firstMargin.replace('px', ''), 10);
    return Number.isNaN(value) ? 0 : value;
};

const getScrollParent = (element: HTMLElement | null): HTMLElement | null => {
    if (!element) return null;
    let parent = element.parentElement;
    while (parent) {
        const style = window.getComputedStyle(parent);
        const overflowY = style.overflowY;
        const canScrollY = overflowY === 'auto' || overflowY === 'scroll' || overflowY === 'overlay';
        if (canScrollY && parent.scrollHeight > parent.clientHeight) {
            return parent;
        }
        parent = parent.parentElement;
    }
    return null;
};

const tryLoadByPosition = () => {
    if (!(imgEle.value instanceof HTMLImageElement)) return;
    const rect = imgEle.value.getBoundingClientRect();
    const offset = getPreloadOffset();
    const root = observerRoot.value;

    let inRange = false;
    if (root) {
        const rootRect = root.getBoundingClientRect();
        inRange = rect.bottom >= rootRect.top - offset && rect.top <= rootRect.bottom + offset;
    } else {
        const viewportHeight = window.innerHeight || document.documentElement.clientHeight || 0;
        inRange = rect.bottom >= -offset && rect.top <= viewportHeight + offset;
    }

    if (inRange) {
        loadImage();
        cleanupObserver();
    }
};

const setupObserver = () => {
    if (!(imgEle.value instanceof HTMLImageElement)) return;
    cleanupObserver();
    observerRoot.value = getScrollParent(imgEle.value);

    if (resolvedMode.value === 'direct') {
        loadImage();
        return;
    }

    io.value = new IntersectionObserver((entries) => {
        const entry = entries[0];
        if (!entry) return;
        if (entry.isIntersecting || entry.intersectionRatio > 0) {
            loadImage();
            cleanupObserver();
        }
    }, {
        root: observerRoot.value,
        rootMargin: resolvedRootMargin.value,
        threshold: resolvedThreshold.value,
    });

    io.value.observe(imgEle.value);
    // 兜底：虚拟列表里偶发 observer 漏触发时，主动检查一次并延迟再检查一次
    requestAnimationFrame(() => {
        tryLoadByPosition();
    });
    fallbackTimer.value = window.setTimeout(() => {
        tryLoadByPosition();
    }, 120);
};

const onLoad = () => {
    isLoaded.value = true;
};

watch(
    () => [resolvedSrc.value, resolvedMode.value],
    async () => {
        if (resolvedMode.value === 'direct') {
            cleanupObserver();
            loadImage();
            return;
        }

        resetImage();
        await nextTick();
        setupObserver();
    },
    { immediate: true }
);

onMounted(() => {
    setupObserver();
});

onUnmounted(() => {
    cleanupObserver();
    clearFallbackTimer();
});
</script>