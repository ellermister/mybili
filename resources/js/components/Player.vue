<template>
    <div id="artplayer" ref="$container" class="w-full md:flex-1 artplayer-app"></div>
</template>
<script lang="ts" setup>
import { onMounted, ref, shallowRef, onBeforeUnmount } from 'vue';
import Artplayer from 'artplayer'
import artplayerPluginDanmuku from 'artplayer-plugin-danmuku';
import artplayerPluginAutoThumbnail from 'artplayer-plugin-auto-thumbnail';
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

const switchVideo = (param: { url: string, danmaku: any[] }) => {
    if (art.value) {
        (art.value.plugins as any).artplayerPluginDanmuku.load(param.danmaku)
        art.value.url = param.url
        art.value.play()
    }
}

onMounted(async () => {
    const isMobile = document.documentElement.clientWidth < 768
    const fontSize = isMobile ? 14 : 25
    const volume = isMobile ? 1 : 0.5
    const danmakuMargin = isMobile ? [2, '75%'] as [number | `${number}%`, number | `${number}%`] : [2, '20%'] as [number | `${number}%`, number | `${number}%`]
    const danmakuSpeed = isMobile ? 4 : 7
    const fullscreenWeb = isMobile ? false : true

    console.log('player fontSize', fontSize)
    console.log('player volume', volume)
    console.log('player danmakuMargin', danmakuMargin)
    console.log('player danmakuSpeed', danmakuSpeed)
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

        plugins: [
            artplayerPluginDanmuku({
                danmuku: props.danmaku,
                speed: danmakuSpeed,
                antiOverlap: true,
                synchronousPlayback: false,
                fontSize: fontSize,
                theme: "light",
                margin: danmakuMargin,
            }),
            artplayerPluginAutoThumbnail({
            }),
        ]
    })
    emit('ready')
})
onBeforeUnmount(() => {
    art.value?.destroy(false)
})
defineExpose({
    switchVideo,
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