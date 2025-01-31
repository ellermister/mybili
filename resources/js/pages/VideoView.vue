<template>
    <h1 class="my-8 text-2xl">
        <RouterLink to="/">🌸</RouterLink> my fav {{ $route.params.id }}
    </h1>

    <div v-if="notfound">
        <div class="text-center text-3xl my-8">Not found</div>
    </div>

    <div class="flex flex-col md:flex-row gap-4 w-full">
        <div id="dplayer" ref="dplayerEle" class="w-full md:flex-1"></div>
        <div class="w-full md:w-80 md:shrink-0" v-if="videoInfo && videoInfo.parts.length > 1">
            <div class="parts bg-[#f1f2f3] rounded-lg p-4">
                <h3 class="text-xl mb-4 font-medium">Parts</h3>
                <div class="flex flex-col gap-2">
                    <button v-for="part in videoInfo?.parts" :key="part.id" @click="playPart(part.id)"
                        class="px-4 py-2 text-left rounded hover:text-sky-500 transition-colors"
                        :class="{ 'bg-white': currentPart === part.id, 'text-sky-500': currentPart === part.id }">
                        {{ part.title }}
                    </button>
                </div>
            </div>
        </div>
    </div>


    <div class="flex flex-col" v-if="videoInfo != null">
        <h2 class="text-2xl my-4">{{ videoInfo.title }}</h2>
        <p>{{ videoInfo.intro }}</p>
        <p>pub time: {{ formatTimestamp(videoInfo.pubtime, "yyyy-mm-dd hh:ii:ss") }}</p>
        <a class="text-base text-green-800 hover:text-green-600" :href="bilibiliUrl(videoInfo.bvid)"
            target="_blank">👉打开到哔哩哔哩</a>
    </div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { formatTimestamp } from '../lib/helper';
import DPlayer from 'dplayer';
/**
 *  player 1.27 版本问题巨多，弹幕显示不出来或者弹幕速度有问题，也不用倒退到1.25，1.26没有css也能够正确显示。
 */

const dplayerEle = ref()

const route = useRoute()
const id = route.params.id

interface VideoType {
    title: string
    id: number
    link: string
    intro: string
    pubtime: number
    fav_time: string
    bvid: string
    attr: number
    page: number
    parts: {
        id: number
        url: string
        title: string
    }[]
    _metas: {
        cover: string
    }
}

const bilibiliUrl = (bvid: string) => {
    return `https://www.bilibili.com/video/${bvid}`
}

const videoInfo = ref<VideoType | null>()
const notfound = ref(false)

const currentPart = ref(null)

const playPart = (partId: number) => {
    console.log('playPart', partId)
    const part = videoInfo.value?.parts.find(part => part.id === partId)
    if (part) {
        // p1 视频, p2 弹幕
        dp.value.switchVideo({
            id: part.id,
            url: part.url,
            type: 'mp4',
        }, {
            id: part.id,
            api: '/api/danmaku/',
        })

        setTimeout(() => {
            dp.value.play()
        }, 1000)

        currentPart.value = partId
    }
}
const dp = ref<DPlayer | null>(null)

onMounted(() => {
    fetch(`/api/video/${id}`).then(async (rsp) => {
        if (!rsp.ok) {
            notfound.value = true
        } else {
            const jsonData = await rsp.json()
            videoInfo.value = jsonData

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
            const part = jsonData.parts[0]
            if (part) {
                options.video.url = part.url
                options.danmaku.id = part.id
            }
            dp.value = new DPlayer(options);

            window.dp = dp.value

            // 添加初始化后的处理
            dp.value.on('loadedmetadata', () => {
                // 确保容器尺寸正确
                const container = dplayerEle.value;
                const containerWidth = container.clientWidth;

                // 更新弹幕容器样式
                const danmakuContainer = container.querySelector('.dplayer-danmaku');
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
        }
    })
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