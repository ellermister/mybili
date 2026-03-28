<template>
    <Image
        :original-src="props.originalSrc ?? props.src"
        :thumb-src="props.thumbSrc"
        :title="props.title"
        mode="observer-preload"
        :preload-margin="props.preloadMargin"
        :image-class="props.imageClass"
    />
</template>

<script lang="ts" setup>
import Image from './Image.vue';

const props = withDefaults(defineProps<{
    src?: string;
    originalSrc?: string;
    thumbSrc?: string;
    title: string;
    preloadMargin?: string;
    imageClass?: string
}>(), {
    preloadMargin: '800px 0px',
});
</script>
