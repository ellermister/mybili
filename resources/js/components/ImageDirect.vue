<template>
    <Image
        :original-src="props.originalSrc ?? props.src"
        :thumb-src="props.thumbSrc"
        :title="props.title"
        mode="direct"
        :image-class="props.imageClass"
    />
</template>

<script lang="ts" setup>
import Image from './Image.vue';

const props = defineProps<{
    src?: string;
    originalSrc?: string;
    thumbSrc?: string;
    title: string;
    imageClass?: string
}>();
</script>
