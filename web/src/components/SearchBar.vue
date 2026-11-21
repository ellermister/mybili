<template>
    <div>
        <div class="relative">
            <input
                ref="inputRef"
                :value="modelValue"
                type="text"
                :placeholder="placeholder"
                class="w-full pl-10 pr-10 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                @input="onInput"
                @keydown.enter.prevent="$emit('enter')"
                @keydown.esc="$emit('esc')"
            />
            <span class="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400">🔍</span>
            <button
                v-if="modelValue"
                @click="$emit('clear')"
                class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-400 hover:text-gray-600"
                title="清空搜索"
            >
                ✕
            </button>
        </div>

        <div v-if="modelValue && resultCount > 0" class="mt-2 text-sm text-gray-600 flex items-center gap-2 flex-wrap">
            <span>{{ resultFoundText }}</span>
            <span v-if="currentIndex >= 0" class="text-blue-600 font-semibold">
                ({{ currentIndex + 1 }}/{{ resultCount }})
            </span>
            <span v-if="navigateHintText" class="text-xs text-gray-500">{{ navigateHintText }}</span>
        </div>

        <div v-else-if="modelValue && resultCount === 0" class="mt-2 text-sm text-red-600">
            {{ noResultText }}
        </div>

        <div v-else-if="idleHintText" class="mt-2 text-sm text-gray-500">
            {{ idleHintText }}
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';

const props = withDefaults(defineProps<{
    modelValue: string;
    placeholder: string;
    resultCount: number;
    currentIndex: number;
    resultFoundText: string;
    noResultText: string;
    navigateHintText?: string;
    idleHintText?: string;
}>(), {
    navigateHintText: '',
    idleHintText: '',
});

const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
    (e: 'enter'): void;
    (e: 'esc'): void;
    (e: 'clear'): void;
}>();

const inputRef = ref<HTMLInputElement | null>(null);

const onInput = (event: Event) => {
    emit('update:modelValue', (event.target as HTMLInputElement).value);
};

const focusInput = () => {
    inputRef.value?.focus();
};

const selectInput = () => {
    inputRef.value?.select();
};

defineExpose({
    focusInput,
    selectInput,
});
</script>
