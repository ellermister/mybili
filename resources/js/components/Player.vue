<template>
    <div id="dplayer" ref="dplayerEle" class="w-full md:flex-1"></div>
</template>
<script lang="ts" setup>
import { onMounted, ref, nextTick } from 'vue';
import DPlayer from 'dplayer';
/**
 *  player 1.27 版本问题巨多，弹幕显示不出来或者弹幕速度有问题，也不用倒退到1.25，1.26没有css也能够正确显示。
 * https://github.com/DIYgod/DPlayer/blob/master/src/js/danmaku.js
 * 官方文档有大大debug，直接看源码比较好
 */

// 声明 DPlayer 类型
interface DPlayerInstance {
    switchVideo: (video: any, danmaku: any) => void;
    play: () => void;
    on: (event: string, callback: () => void) => void;
    danmaku: {
        draw: (options: any) => void;
    };
}

// 声明 window.dp 属性
declare global {
    interface Window {
        dp: DPlayerInstance;
    }
}

const dplayerEle = ref<HTMLDivElement | null>(null)
const dp = ref<DPlayerInstance | null>(null)
const isReady = ref(false)

// 定义事件
const emit = defineEmits<{
    ready: []
}>()

const switchVideo = (param: { url: string, type: string, danmaku_id: string }) => {
    if (dp.value && isReady.value) {
        dp.value.switchVideo({
            url: param.url,
            type: param.type,
        }, {
            id: param.danmaku_id,
            api: '/api/danmaku/',
        })
        // p1 视频, p2 弹幕
        dp.value?.play()
    } else {
        console.warn('Player not ready yet')
    }
}

onMounted(async () => {
    // await nextTick()
    
    const options = {
        container: dplayerEle.value,
        video: {
            url: '',
            type: 'mp4',
        },
        danmaku: {
            id: "",
            api: '/api/danmaku/',
        },
    }

    dp.value = new (DPlayer as any)(options);

    if (dp.value) {
        window.dp = dp.value

        // 添加初始化后的处理
        dp.value.on('loadedmetadata', () => {
            // 确保容器尺寸正确
            const container = dplayerEle.value;
            if (!container) return;

            const containerWidth = container.clientWidth;

            // 更新弹幕容器样式
            const danmakuContainer = container.querySelector('.dplayer-danmaku') as HTMLElement;
            if (danmakuContainer) {
                danmakuContainer.style.width = `${containerWidth}px`;
            }

            // 测试弹幕
            // dp.value.danmaku.draw({
            //     text: '测试弹幕',
            //     color: '#fff',
            //     type: 'right'
            // });

            // 输出调试信息
            console.log('容器宽度:', containerWidth);
            console.log('弹幕容器:', danmakuContainer);
        });

        isReady.value = true
        console.log('Player is ready')
        emit('ready')
       
    }
})

// 暴露方法给父组件
defineExpose({
    switchVideo,
    isReady
})
</script>
<style scoped>
#dplayer {
    height: 600px;
    position: relative;
    overflow: hidden;
    width: 100%;
    /* 确保容器有明确的宽度 */
}

@media (max-width: 768px) {
    #dplayer {
        height: 300px;
    }
}
</style>
<style scoped>
/* 弹幕速度 通过css控制，dplayer 的js初始化参数控制不了，默认5s, 这里设置为10s */
:deep(.dplayer-danmaku .dplayer-danmaku-right.dplayer-danmaku-move) {
    animation-duration: 10s !important;
}
</style>