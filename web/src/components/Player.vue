<template>
    <div id="artplayer" ref="$container" class="w-full md:flex-1 artplayer-app"></div>
</template>
<script lang="ts" setup>
import { onMounted, ref, shallowRef, onBeforeUnmount } from 'vue';
import Artplayer from 'artplayer'
import artplayerPluginDanmuku, { type Option } from 'artplayer-plugin-danmuku';
import { type Option as DanmakuOption } from  'artplayer-plugin-danmuku';

const DANMAKU_CONFIG_KEY = 'mybili_danmaku_config'

const art = shallowRef<Artplayer | null>(null)
const $container = ref<HTMLDivElement | null>(null)

// 定义事件
const emit = defineEmits<{
    ready: []
}>()

const props = defineProps<{
    danmaku: any[]
    url: string
}>()

/**
 * 从 localStorage 读取弹幕配置
 */
const loadDanmakuConfig = (): Partial<Option> | null => {
    try {
        const saved = localStorage.getItem(DANMAKU_CONFIG_KEY)
        if (saved) {
            return JSON.parse(saved)
        }
    } catch (error) {
        console.error('读取弹幕配置失败:', error)
    }
    return null
}

/**
 * 保存弹幕配置到 localStorage
 * 只保存预设中的 key
 */
const saveDanmakuConfig = (config: DanmakuOption) => {
    try {
        const configToSave: Partial<Option> = {
            speed: config.speed,
            antiOverlap: config.antiOverlap,
            synchronousPlayback: config.synchronousPlayback,
            fontSize: config.fontSize,
            theme: config.theme,
            margin: config.margin,
            modes: config.modes,
        }
        localStorage.setItem(DANMAKU_CONFIG_KEY, JSON.stringify(configToSave))
        console.log('弹幕配置已保存:', configToSave)
    } catch (error) {
        console.error('保存弹幕配置失败:', error)
    }
}

let danmakuToken = 0

const danmakuPlugin = () => (art.value?.plugins as any)?.artplayerPluginDanmuku

const sortDanmaku = (list: any[]) => {
    return list.slice().sort((prev, next) => (prev.time ?? 0) - (next.time ?? 0))
}

const playUrl = (url: string) => {
    if (!art.value || !url) {
        return
    }
    danmakuToken++
    art.value.url = url
    const played = art.value.play() as Promise<void> | void
    if (played && typeof played.catch === 'function') {
        played.catch(() => {})
    }
    const plugin = danmakuPlugin()
    if (plugin) {
        plugin.option.danmuku = []
        plugin.load()
    }
}

const seek = (time: number) => {
    if (!art.value || !Number.isFinite(time)) {
        return
    }
    art.value.currentTime = Math.max(0, time)
}

const setDanmaku = (list: any[]) => {
    const plugin = danmakuPlugin()
    if (!plugin) {
        return
    }
    const token = ++danmakuToken
    const comments = Array.isArray(list) ? list : []
    window.setTimeout(() => {
        if (token !== danmakuToken || !art.value) {
            return
        }
        plugin.option.danmuku = sortDanmaku(comments)
        plugin.load()
    }, 0)
}

const switchVideo = (param: { url: string, danmaku: any[] }) => {
    playUrl(param.url)
    setDanmaku(param.danmaku)
}

onMounted(async () => {
    const isMobile = document.documentElement.clientWidth < 768
    const volume = isMobile ? 1 : 0.5
    const fullscreenWeb = isMobile ? false : true

    // 默认配置
    const defaultDanmakuOption = {
        speed: isMobile ? 4 : 7.5,
        antiOverlap: true,
        synchronousPlayback: false,
        fontSize: isMobile ? 14 : 25,
        theme: "light",
        margin: isMobile ? [10, '75%'] as [number | `${number}%`, number | `${number}%`] : [10, 10] as [number | `${number}%`, number | `${number}%`],
        modes: [0, 1, 2],
    } as Option

    // 读取已保存的配置
    const savedConfig = loadDanmakuConfig()
    
    // 合并配置：优先使用已保存的配置，其次使用默认配置
    const presetDanmakuOption = {
        ...defaultDanmakuOption,
        ...savedConfig,
        danmuku: props.danmaku, // 弹幕数据始终使用 props
    } as Option

    console.log('使用的弹幕配置:', presetDanmakuOption)

    const plugins: any[] = [
        artplayerPluginDanmuku(presetDanmakuOption),
    ]
    art.value = new Artplayer({
        container: $container.value as HTMLDivElement,
        fullscreen: true,
        fullscreenWeb: fullscreenWeb,
        autoOrientation: true,
        url: props.url,
        setting: true,
        volume: volume,
        flip: true,
        playbackRate: true,
        theme: "#e749a0",
        miniProgressBar: true,

        plugins: plugins
    })
    // 监听弹幕配置变化并保存
    art.value?.on('artplayerPluginDanmuku:config', (...args: unknown[]) => {
        const option = args[0] as DanmakuOption
        console.info('弹幕配置变化:', option);
        saveDanmakuConfig(option)
    });
    
    emit('ready')
})
onBeforeUnmount(() => {
    art.value?.destroy(false)
})
defineExpose({
    switchVideo,
    playUrl,
    setDanmaku,
    seek,
})
</script>
<style scoped>
.artplayer-app {
    width: 100%;
    height: 600px;
    position: relative;
    overflow: hidden;
}

@media (max-width: 768px) {
    .artplayer-app {
        height: 300px;
    }
}
</style>