<template>
    <div ref="hostEl" class="image-layer-root w-full h-full">
        <img
            ref="baseImgEl"
            v-bind="forwardedAttrs"
            :class="[imageClass]"
            :src="baseSrc"
            :data-src="resolvedOriginalSrc"
            :alt="props.title"
            :loading="resolvedMode === 'direct' ? 'eager' : 'lazy'"
            decoding="async"
            @load="onBaseLoad"
        >

        <img
            v-if="showOriginalOverlay"
            ref="originalImgEl"
            v-bind="forwardedAttrs"
            :class="[imageClass, 'image-original-overlay']"
            :src="originalOverlaySrc"
            :alt="props.title"
            loading="eager"
            decoding="async"
            :style="{ opacity: originalVisible ? 1 : 0 }"
            @load="onOriginalLoad"
            @error="onOriginalError"
        >
    </div>
</template>
<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, useAttrs, watch } from 'vue';
defineOptions({ inheritAttrs: false });

const defaultSrc = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7';

const props = defineProps<{
    // 兼容老调用：src 作为原图
    src?: string;
    // 新参数：原图（推荐）
    originalSrc?: string;
    // 新参数：缩略图（可选）
    thumbSrc?: string;
    title: string;
    directLoad?: boolean;
    mode?: 'observer' | 'observer-preload' | 'direct';
    preloadMargin?: string;
    threshold?: number;
}>();

const attrs = useAttrs();
const hostEl = ref<HTMLElement | null>(null);
const baseImgEl = ref<HTMLImageElement | null>(null);
const originalImgEl = ref<HTMLImageElement | null>(null);
const isLoaded = ref(false);
const hasRequestedOriginal = ref(false);
const originalVisible = ref(false);
const io = ref<IntersectionObserver | null>(null);
const fallbackCheckTimer = ref<number | null>(null);
const scrollStopTimer = ref<number | null>(null);

const resolvedMode = computed(() => {
    if (props.mode) return props.mode;
    return props.directLoad ? 'direct' : 'observer';
});

const resolvedOriginalSrc = computed(() => props.originalSrc ?? props.src ?? '');
const resolvedThumbSrc = computed(() => props.thumbSrc ?? '');
const hasThumb = computed(() => Boolean(resolvedThumbSrc.value));

const forwardedAttrs = computed(() => {
    const { class: _class, style: _style, ...rest } = attrs;
    return rest;
});

const imageClass = computed(() => attrs.class);

const baseSrc = computed(() => {
    if (hasThumb.value) return resolvedThumbSrc.value || defaultSrc;
    return hasRequestedOriginal.value ? (resolvedOriginalSrc.value || defaultSrc) : defaultSrc;
});

const originalOverlaySrc = computed(() => {
    if (!hasRequestedOriginal.value) return '';
    return resolvedOriginalSrc.value || defaultSrc;
});

const showOriginalOverlay = computed(() => hasThumb.value && hasRequestedOriginal.value);

const resolvedRootMargin = computed(() => {
    if (props.preloadMargin) return props.preloadMargin;
    return resolvedMode.value === 'observer-preload' ? '800px 0px' : '0px';
});

const resolvedThreshold = computed(() => {
    if (typeof props.threshold === 'number') return props.threshold;
    return resolvedMode.value === 'observer-preload' ? 0 : 0.1;
});

const cleanupObserver = () => {
    if (io.value) {
        io.value.disconnect();
        io.value = null;
    }
};

const getPreloadOffset = () => {
    const first = resolvedRootMargin.value.trim().split(/\s+/)[0] ?? '0px';
    const parsed = Number.parseInt(first.replace('px', ''), 10);
    return Number.isNaN(parsed) ? 0 : parsed;
};

const isElementNearViewport = (el: HTMLElement) => {
    const rect = el.getBoundingClientRect();
    const offset = getPreloadOffset();
    const viewportHeight = window.innerHeight || document.documentElement.clientHeight || 0;
    return rect.bottom >= -offset && rect.top <= viewportHeight + offset;
};

const stopFallbackChecks = () => {
    if (fallbackCheckTimer.value !== null) {
        window.clearTimeout(fallbackCheckTimer.value);
        fallbackCheckTimer.value = null;
    }
    if (scrollStopTimer.value !== null) {
        window.clearTimeout(scrollStopTimer.value);
        scrollStopTimer.value = null;
    }
    window.removeEventListener('resize', tryLoadByPosition);
    window.removeEventListener('resize', onGlobalScrollOrResize);
    // capture=true 监听任意滚动容器（包括虚拟列表容器）
    document.removeEventListener('scroll', tryLoadByPosition, true);
    document.removeEventListener('scroll', onGlobalScrollOrResize, true);
};

const startFallbackChecks = () => {
    stopFallbackChecks();
    window.addEventListener('resize', tryLoadByPosition, { passive: true });
    window.addEventListener('resize', onGlobalScrollOrResize, { passive: true });
    document.addEventListener('scroll', tryLoadByPosition, true);
    document.addEventListener('scroll', onGlobalScrollOrResize, true);
    fallbackCheckTimer.value = window.setTimeout(() => {
        tryLoadByPosition();
    }, 300);
    // 首屏兜底：即使没有滚动事件，也触发一次“滚动停止后加载原图”的流程
    onGlobalScrollOrResize();
};

const tryLoadByPosition = () => {
    if (resolvedMode.value === 'direct' || hasRequestedOriginal.value) {
        stopFallbackChecks();
        return;
    }
    const el = hostEl.value;
    if (!(el instanceof HTMLElement)) return;
    if (isElementNearViewport(el)) {
        maybeLoadOriginal('near-viewport');
    }
};

const onGlobalScrollOrResize = () => {
    if (hasRequestedOriginal.value) return;
    if (scrollStopTimer.value !== null) {
        window.clearTimeout(scrollStopTimer.value);
    }
    scrollStopTimer.value = window.setTimeout(() => {
        maybeLoadOriginal('scroll-stop');
    }, 120);
};

const loadOriginal = (force = false) => {
    if (!resolvedOriginalSrc.value) return;
    if (!force && hasRequestedOriginal.value) return;
    hasRequestedOriginal.value = true;
    cleanupObserver();
    stopFallbackChecks();
};

const maybeLoadOriginal = (reason: 'observer' | 'scroll-stop' | 'near-viewport') => {
    if (hasRequestedOriginal.value) return;
    if (!(hostEl.value instanceof HTMLElement)) return;

    const inViewport = isElementNearViewport(hostEl.value);
    if (!inViewport) return;

    if (resolvedThumbSrc.value) {
        // 缩略图模式下：滚动过程中保持缩略图，滚动停止后再加载原图
        if (reason !== 'scroll-stop') return;
    }

    loadOriginal();
};

const resetSource = () => {
    isLoaded.value = false;
    hasRequestedOriginal.value = false;
    originalVisible.value = false;
    stopFallbackChecks();
};

const setupObserver = () => {
    if (!(hostEl.value instanceof HTMLElement)) return;
    cleanupObserver();

    if (resolvedMode.value === 'direct') {
        loadOriginal(true);
        return;
    }

    io.value = new IntersectionObserver((entries) => {
        const entry = entries[0];
        if (!entry) return;
        if (entry.isIntersecting || entry.intersectionRatio > 0) {
            maybeLoadOriginal('observer');
        }
    }, {
        rootMargin: resolvedRootMargin.value,
        threshold: resolvedThreshold.value,
    });

    io.value.observe(hostEl.value);
    // 双保险：虚拟列表复用节点时，observer 偶发漏触发，主动检查一次位置
    requestAnimationFrame(() => {
        tryLoadByPosition();
    });
    startFallbackChecks();
};

const onBaseLoad = () => {
    if (!hasThumb.value) {
        isLoaded.value = true;
    }
};

const onOriginalLoad = () => {
    isLoaded.value = true;
    requestAnimationFrame(() => {
        originalVisible.value = true;
    });
};

const onOriginalError = () => {
    isLoaded.value = true;
    originalVisible.value = false;
};

watch(
    () => [resolvedOriginalSrc.value, resolvedThumbSrc.value, resolvedMode.value, resolvedRootMargin.value, resolvedThreshold.value],
    async () => {
        if (resolvedMode.value === 'direct') {
            cleanupObserver();
            loadOriginal(true);
            return;
        }
        resetSource();
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
    stopFallbackChecks();
});
</script>
<style scoped>
.image-layer-root {
    position: relative;
    display: block;
}

.image-original-overlay {
    position: absolute;
    inset: 0;
    transition: opacity 160ms ease-out;
    will-change: opacity;
}
</style>