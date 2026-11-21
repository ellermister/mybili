<template>
    <div class="min-h-screen bg-white p-4 md:p-6">
        <div class="max-w-4xl mx-auto">
            <h1 class="text-2xl font-bold mb-4 text-gray-800">{{ t('cookie.title') }}</h1>
            <section class="bg-white rounded-xl shadow-lg border border-gray-200 p-4 md:p-6">
                <h2 class="flex items-center text-lg font-semibold text-gray-800 border-b border-gray-200 pb-3 mb-2">
                    <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"></path>
                    </svg>
                    {{ t('cookie.title') }}
                </h2>
                <div class="cookie-row">
                    <div class="min-w-0">
                        <div class="text-sm text-gray-500 mb-2">{{ t('cookie.fileStatus') }}</div>
                        <span class="status" :class="badgeClass(state.exist)">{{ existText }}</span>
                    </div>
                    <button type="button" class="action" @click="uploadFile">{{ t('cookie.upload') }}</button>
                </div>
                <div class="cookie-row">
                    <div class="min-w-0">
                        <div class="text-sm text-gray-500 mb-2">{{ t('cookie.cookieStatus') }}</div>
                        <span class="status" :class="badgeClass(state.valid)">{{ validText }}</span>
                    </div>
                    <button type="button" class="action" @click="checkValid">{{ t('cookie.check') }}</button>
                </div>
            </section>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const state = ref<{ exist: boolean | null; valid: boolean | null }>({
    exist: null,
    valid: null,
})

const existText = computed(() => {
    if (state.value.exist === null) return '…'
    return state.value.exist ? t('cookie.exist') : t('cookie.notExist')
})
const validText = computed(() => {
    if (state.value.valid === null) return '…'
    return state.value.valid ? t('cookie.valid') : t('cookie.invalid')
})

function badgeClass(value: boolean | null) {
    if (value === null) return 'pending'
    return value ? 'ok' : 'bad'
}

const checkFile = () => {
    fetch(`/api/cookie/exist`).then(async (rsp) => {
        if (rsp.ok) {
            const data = await rsp.json()
            state.value.exist = data.exist
        }
    })
}
const checkValid = () => {
    fetch(`/api/cookie/status`).then(async (rsp) => {
        if (rsp.ok) {
            const data = await rsp.json()
            state.value.valid = data.logged
        }
    })
}

const uploadFile = () => {
    const input: HTMLInputElement = document.createElement('input')
    input.type = "file"
    input.onchange = () => {
        if (!input.files || input.files.length <= 0) {
            return
        }
        const data = new FormData()
        data.append('file', input.files[0])
        fetch(`/api/cookie/upload`, {
            method: 'POST',
            body: data
        }).then(async () => {
            checkFile()
            checkValid()
        })
    }
    input.click()
}

checkFile()
checkValid()
</script>

<style scoped>
.cookie-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    padding: 16px 0;
    border-top: 1px solid #e5e7eb;
}
.cookie-row:first-of-type { border-top: 0; }
.status {
    display: inline-flex;
    align-items: center;
    min-width: 4.5rem;
    justify-content: center;
    padding: 2px 10px;
    border-radius: 999px;
    font-size: 13px;
    font-weight: 600;
}
.status.ok { background: #dcfce7; color: #166534; }
.status.bad { background: #fee2e2; color: #991b1b; }
.status.pending { background: #f3f4f6; color: #6b7280; }
.action {
    flex: none;
    height: 36px;
    padding: 0 16px;
    border: 0;
    border-radius: 8px;
    background: #7c3aed;
    color: #fff;
    font-size: 14px;
    cursor: pointer;
}
.action:hover { background: #6d28d9; }
@media (max-width: 640px) {
    .cookie-row { align-items: flex-start; flex-direction: column; }
}
</style>
