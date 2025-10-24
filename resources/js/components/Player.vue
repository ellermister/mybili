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
    const fontSize = document.documentElement.clientWidth > 500 ? 25 : 14
    console.log('fontSize', fontSize)
    art.value = new Artplayer({
        container: $container.value as HTMLDivElement,
        fullscreen: true,
        fullscreenWeb: true,
        autoOrientation: true,
        url: props.url,
        setting: true,
        volume: 0.5,
        flip: true,
        playbackRate: true,
        theme: "#e749a0",
        miniProgressBar: true,
  
        plugins: [
            artplayerPluginDanmuku({
                danmuku: props.danmaku,
                speed: 7,
                antiOverlap: true,
                synchronousPlayback: false,
                fontSize: fontSize,
                theme: "light",
            }),
            artplayerPluginAutoThumbnail({
            }),
        ]
    })
    console.log('art.value', art.value)
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