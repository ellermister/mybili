<template>
    <Teleport to="body">
        <div
            class="pointer-events-none fixed inset-x-0 bottom-0 z-[100] p-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] md:inset-auto md:bottom-auto md:left-auto md:right-4 md:top-4 md:max-w-md md:p-0"
            role="region"
            aria-live="polite"
            :aria-label="regionLabel"
        >
            <TransitionGroup
                name="toast"
                tag="div"
                class="flex max-h-[50vh] flex-col-reverse gap-2 overflow-y-auto md:max-h-[min(80vh,32rem)] md:flex-col"
            >
                <div
                    v-for="item in toasts"
                    :key="item.id"
                    class="pointer-events-auto flex w-full items-start gap-3 rounded-xl border px-4 py-3 shadow-lg backdrop-blur-sm"
                    :class="toastClass(item.type)"
                >
                    <span class="mt-0.5 shrink-0 text-lg leading-none" aria-hidden="true">{{ toastIcon(item.type) }}</span>
                    <p class="min-w-0 flex-1 text-sm leading-relaxed">
                        {{ item.message }}
                    </p>
                    <button
                        type="button"
                        class="shrink-0 rounded-md p-1 text-current opacity-60 transition hover:bg-black/5 hover:opacity-100"
                        :aria-label="closeLabel"
                        @click="dismiss(item.id)"
                    >
                        <span class="block text-lg leading-none">×</span>
                    </button>
                </div>
            </TransitionGroup>
        </div>
    </Teleport>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useToast, type ToastType } from '@/composables/toast';

const { toasts, dismiss } = useToast();
const { t } = useI18n();

const closeLabel = computed(() => t('common.close'));
const regionLabel = computed(() => t('common.notifications'));

function toastClass(type: ToastType): string {
    switch (type) {
        case 'success':
            return 'border-emerald-200/90 bg-emerald-50/95 text-emerald-950';
        case 'error':
            return 'border-red-200/90 bg-red-50/95 text-red-950';
        default:
            return 'border-slate-200/90 bg-white/95 text-slate-900';
    }
}

function toastIcon(type: ToastType): string {
    switch (type) {
        case 'success':
            return '✓';
        case 'error':
            return '!';
        default:
            return 'ⓘ';
    }
}
</script>

<style scoped>
.toast-move,
.toast-enter-active,
.toast-leave-active {
    transition: opacity 0.25s ease, transform 0.25s ease;
}

.toast-enter-from {
    opacity: 0;
    transform: translateY(8px);
}

@media (min-width: 768px) {
    .toast-enter-from {
        transform: translateX(12px);
    }
}

.toast-leave-to {
    opacity: 0;
    transform: scale(0.98);
}
</style>
