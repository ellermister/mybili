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
    on: (event: string, callback: (e: any|null) => void) => void;
    danmaku: {
        draw: (options: any) => void;
        speed: (rate: number) => void;
        dan: any[];
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


            // 设置弹幕速度
            dp.value.danmaku.speed(0.5)

            // 输出调试信息
            console.log('容器宽度:', containerWidth);
            console.log('弹幕容器:', danmakuContainer);
        });

        dp.value.on('danmaku_load_start', (e:any) => {
            console.log('弹幕加载开始:', e);
        });

        dp.value.on('danmaku_load_end', () => {
            console.log('弹幕加载完成, 数量：', dp.value.danmaku.dan.length);
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