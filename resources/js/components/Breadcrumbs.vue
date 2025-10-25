<template>
    <div class="flex items-center justify-between my-8 mt-4 md:mt-0">
        <nav aria-label="Breadcrumb" class="min-w-0">
            <ol class="flex items-center space-x-1 md:space-x-2 overflow-hidden">
                <li v-for="(item, index) in items" :key="index" class="flex items-center">
                    <!-- Separator -->
                    <svg v-if="index > 0" class="w-3 h-3 text-gray-400 mx-1 flex-shrink-0" aria-hidden="true"
                        xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 6 10">
                        <path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="m1 9 4-4-4-4" />
                    </svg>
                    <!-- Link Item -->
                    <router-link v-if="item.to" :to="item.to"
                        class="inline-flex items-center text-sm font-medium text-gray-700 hover:text-purple-600 whitespace-nowrap">
                        <span v-if="index === 0"
                            class="text-xl mr-1  font-bold bg-gradient-to-r from-pink-500 to-purple-600 bg-clip-text text-transparent hover:from-pink-600 hover:to-purple-700 transition-all duration-300">🌸</span>
                        {{ item.text }}
                    </router-link>
                    <!-- Current Page Item -->
                    <span v-else class="ml-1 text-sm font-medium text-gray-500 md:ml-2 block truncate" :title="item.text">
                        {{ item.text }}
                    </span>
                </li>
            </ol>
        </nav>
        <!-- Actions Slot -->
        <div class="flex items-center gap-2 flex-shrink-0">
            <slot name="actions"></slot>
        </div>
    </div>
</template>

<script lang="ts" setup>
interface BreadcrumbItem {
    to?: string;
    text: string;
}

defineProps<{
    items: BreadcrumbItem[];
}>();
</script>

<style scoped></style>