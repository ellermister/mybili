<template>
    <h1 class="my-8 text-2xl">
        <RouterLink to="/">🌸</RouterLink> my fav {{ $route.params.id }}
    </h1>

    <div v-if="notfound">
        <div class="text-center text-3xl my-8">Not found</div>
    </div>

    <div id="dplayer" ref="dplayerEle"></div>
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
    _metas: {
        cover: string
    }
}

const videoURL = (id: number) => {
    return `/storage/videos/${id}.mp4`
}

const bilibiliUrl = (bvid: string) => {
    return `https://www.bilibili.com/video/${bvid}`
}

const videoInfo = ref<VideoType | null>()
const notfound = ref(false)



onMounted(() => {
    fetch(`/api/video/${id}`).then(async (rsp) => {
        if (!rsp.ok) {
            notfound.value = true
        } else {
            const jsonData = await rsp.json()
            videoInfo.value = jsonData

            console.log('dplayer.value', dplayerEle.value)
            const options = {
                container: dplayerEle.value,
                video: {
                    url: videoURL(jsonData.id),
                },
            }
            const dp = new DPlayer(options);
            console.log('dp', dp)
        }
    })

})
</script>
<style scoped>
#dplayer {
    /* max-height: 600px; */
    height: 600px;
}
</style>