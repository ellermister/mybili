<template>
    <img ref="imgEle" :src="src" :data-src="props.src" :alt="props.title" @visible="visible">
</template>
<script lang="ts" setup>
import { onMounted, onUnmounted, ref, watch } from 'vue';

const imgEle = ref<HTMLImageElement | null>(null)
const src = ref('')

import observer from "@/lib/obs"

const props = defineProps<{
    src: string,
    title: string
}>()


watch(props, ()=>{
    if(src.value != "" && props.src != src.value){
        src.value = props.src
    }
})

const visible = () => {
    src.value = props.src
    if (imgEle.value instanceof HTMLImageElement) {
        observer.unobserve(imgEle.value)
    }
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