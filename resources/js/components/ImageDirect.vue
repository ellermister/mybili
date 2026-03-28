<template>
    <Image
        :original-src="props.originalSrc ?? props.src"
        :title="props.title"
        mode="direct"
    />
</template>

<script lang="ts" setup>
import Image from './Image.vue';

const props = defineProps<{
    src?: string;
    originalSrc?: string;
    title: string;
}>();
</script>
