<template>
    <img class="bg-gray-200 " ref="imgEle" :src="src" :data-src="props.src" :alt="props.title" @visible="visible" @load="onLoad">
</template>
<script lang="ts" setup>
import { onMounted, onUnmounted, ref, watch } from 'vue';

const imgEle = ref<HTMLImageElement | null>(null)
const defaultSrc = 'data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7'
const src = ref(defaultSrc)
const isLoaded = ref(false)

import observer from "@/lib/obs"

const props = defineProps<{
    src: string,
    title: string
}>()


watch(props, ()=>{
    if(src.value != defaultSrc && props.src != src.value){
        isLoaded.value = false
        src.value = props.src
    }
})

const visible = () => {
    isLoaded.value = false
    src.value = props.src
    if (imgEle.value instanceof HTMLImageElement) {
        observer.unobserve(imgEle.value)
    }
}

const onLoad = () => {
    isLoaded.value = true
}
onMounted(() => {
    if (imgEle.value instanceof HTMLImageElement) {
        observer.observe(imgEle.value)
    }
})

onUnmounted(() => {
    if (imgEle.value instanceof HTMLImageElement) {
        observer.unobserve(imgEle.value)
    }
})
</script>