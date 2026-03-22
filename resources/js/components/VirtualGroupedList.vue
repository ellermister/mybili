<template>
    <virtualList
        ref="virtualListRef"
        :sortable="false"
        :draggable="''"
        :class="containerClass"
        v-model="groupedRows"
        data-key="'id'"
        :keeps="keeps"
        :size="size"
    >
        <template v-slot:item="{ record, index, dataKey }">
            <div class="virtual-grouped-item">
                <slot name="item" :record="record" :index="index" :data-key="dataKey"></slot>
            </div>
        </template>
    </virtualList>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
// @ts-ignore 缺少类型声明，按行忽略类型检查
import VirtualList from 'vue-virtual-sortable';

const props = withDefaults(defineProps<{
    items: any[];
    columns: number;
    size: number;
    keeps?: number;
    containerClass?: string;
}>(), {
    keeps: 60,
    containerClass: '',
});

const virtualListRef = ref<any>(null);

const groupedRows = computed(() => {
    const grouped: { id: string; videos: any[] }[] = [];
    const cols = Math.max(1, props.columns || 1);

    for (let i = 0; i < props.items.length; i += cols) {
        grouped.push({
            id: `row-${i}`,
            videos: props.items.slice(i, i + cols),
        });
    }

    return grouped;
});

const scrollToIndex = (index: number) => {
    virtualListRef.value?.scrollToIndex?.(index);
};

const scrollToOffset = (offset: number) => {
    virtualListRef.value?.scrollToOffset?.(offset);
};

defineExpose({
    scrollToIndex,
    scrollToOffset,
});
</script>

<style scoped>
.virtual-grouped-item {
    width: 100%;
}
</style>
