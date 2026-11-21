<template>
    <div class="min-h-screen bg-white p-4 md:p-6">
        <div class="max-w-4xl mx-auto">
            <h1 class="text-2xl font-bold mb-4 text-gray-800">{{ t('settings.title') }}</h1>

            <nav class="settings-tabs">
                <button type="button" :class="{ on: tab === 'features' }" @click="tab = 'features'">{{ t('settings.featureSwitches') }}</button>
                <button type="button" :class="{ on: tab === 'sources' }" @click="tab = 'sources'">{{ t('settings.downloadSources') }}</button>
                <button type="button" :class="{ on: tab === 'filters' }" @click="tab = 'filters'">{{ t('settings.globalFilters') }}</button>
                <button type="button" :class="{ on: tab === 'notify' }" @click="tab = 'notify'">{{ t('settings.notificationSettings') }}</button>
            </nav>

            <section v-show="tab === 'features'" class="settings-panel shadow-lg border border-gray-200">
                <h2 class="settings-heading">
                    <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                    </svg>
                    {{ t('settings.featureSwitches') }}
                </h2>
                <div class="grid grid-cols-1 md:grid-cols-2 gap-x-8">
                    <div class="switch-row"><span>{{ t('settings.features.favoriteSync') }}</span><Toggle v-model="favoriteSyncEnabled" /></div>
                    <div class="switch-row"><span>{{ t('settings.features.videoDownload') }}</span><Toggle v-model="videoDownloadEnabled" /></div>
                    <div class="switch-row"><span>{{ t('settings.features.danmakuDownload') }}</span><Toggle v-model="danmakuDownloadEnabled" /></div>
                    <div class="switch-row">
                        <span class="inline-flex items-center">{{ t('settings.features.multiPartitionDownload') }}<Help :text="t('settings.features.multiPartitionHelp')" /></span>
                        <Toggle v-model="multiPartitionDownloadEnabled" />
                    </div>
                    <div class="switch-row">
                        <span class="inline-flex items-center">{{ t('settings.features.humanReadableName') }}<Help :text="t('settings.features.readableHelp')" /></span>
                        <Toggle v-model="humanReadableNameEnabled" />
                    </div>
                    <div class="switch-row">
                        <span class="inline-flex items-center">{{ t('settings.features.usageAnalytics') }}<Help :text="t('settings.features.usageAnalyticsDescription')" /></span>
                        <Toggle v-model="usageAnalyticsEnabled" />
                    </div>
                </div>
            </section>

            <section v-show="tab === 'sources'" class="settings-panel shadow-lg border border-gray-200">
                <h2 class="settings-heading">
                    <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"></path>
                    </svg>
                    {{ t('settings.downloadSources') }}
                </h2>
                <p class="text-xs text-gray-500 mb-4">{{ t('settings.sources.hint') }}</p>
                <div class="grid grid-cols-1 sm:grid-cols-[120px_minmax(0,1fr)] gap-2 sm:items-center mb-3">
                    <label class="text-sm text-gray-500">{{ t('settings.sources.defaultQuality') }}</label>
                    <select v-model="videoQuality" class="settings-select" :class="{ skip: videoQuality === 'skip' }">
                        <option v-for="item in qualityOptions" :key="item.value" :value="item.value">{{ t(item.label) }}</option>
                    </select>
                </div>
                <div class="flex flex-wrap items-center gap-2 mb-2">
                    <input v-model="sourceQuery" type="text" class="settings-select max-w-xs" :placeholder="t('settings.sources.search')">
                    <span class="text-xs text-gray-500 sm:ml-auto">{{ t('settings.sources.count', { count: visibleSources.length }) }}</span>
                </div>
                <div v-for="row in visibleSources" :key="row.id" class="source-row">
                    <div class="flex items-center gap-2 min-w-0">
                        <span class="source-tag" :class="{ sub: row.kind === 'subscription' }">{{ row.kind === 'subscription' ? t('settings.sources.subscription') : t('settings.sources.favorite') }}</span>
                        <span class="truncate">{{ row.name }}</span>
                    </div>
                    <select v-model="row.quality" class="settings-select" :class="{ skip: row.quality === 'skip' }">
                        <option value="0">{{ t('settings.features.followDefault') }}</option>
                        <option v-for="item in qualityOptions" :key="item.value" :value="item.value">{{ t(item.label) }}</option>
                    </select>
                </div>
            </section>

            <section v-show="tab === 'filters'" class="settings-panel shadow-lg border border-gray-200">
                <h2 class="settings-heading">
                    <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"></path>
                    </svg>
                    {{ t('settings.globalFilters') }}
                </h2>
                <p class="text-xs text-gray-500 mb-3">{{ t('settings.filters.globalHint') }}</p>
                <div class="filter-row">
                    <label>{{ t('settings.filters.byName') }}</label>
                    <div>
                        <select v-model="nameExclude.type" class="settings-select">
                            <option value="off">{{ t('settings.filters.noFilter') }}</option>
                            <option value="contains">{{ t('settings.filters.containsKeyword') }}</option>
                            <option value="regex">{{ t('settings.filters.regexPattern') }}</option>
                        </select>
                        <input v-if="nameExclude.type === 'contains'" v-model="nameExclude.contains" type="text" class="settings-select mt-2" :placeholder="t('settings.placeholders.enterKeyword')">
                        <input v-if="nameExclude.type === 'regex'" v-model="nameExclude.regex" type="text" class="settings-select mt-2" :placeholder="t('settings.placeholders.enterRegex')">
                    </div>
                </div>
                <div class="filter-row">
                    <label class="inline-flex items-center">{{ t('settings.filters.byDuration') }}<Help :text="t('settings.filters.byDurationDescription')" /></label>
                    <div>
                        <select v-model="durationVideoExclude.type" class="settings-select">
                            <option value="off">{{ t('settings.filters.noFilter') }}</option>
                            <option value="30min">{{ t('settings.filters.largerThan30min') }}</option>
                            <option value="60min">{{ t('settings.filters.largerThan60min') }}</option>
                            <option value="custom">{{ t('settings.filters.customDuration') }}</option>
                        </select>
                        <input v-if="durationVideoExclude.type === 'custom'" v-model="durationVideoExclude.custom_duration" type="number" class="settings-select mt-2" :placeholder="t('settings.placeholders.enterDuration')">
                    </div>
                </div>
                <div class="filter-row">
                    <label class="inline-flex items-center">{{ t('settings.filters.byDurationPart') }}<Help :text="t('settings.filters.partDurationHelp')" /></label>
                    <div>
                        <select v-model="durationVideoPartExclude.type" class="settings-select">
                            <option value="off">{{ t('settings.filters.noFilter') }}</option>
                            <option value="30min">{{ t('settings.filters.largerThan30min') }}</option>
                            <option value="60min">{{ t('settings.filters.largerThan60min') }}</option>
                            <option value="custom">{{ t('settings.filters.customDuration') }}</option>
                        </select>
                        <input v-if="durationVideoPartExclude.type === 'custom'" v-model="durationVideoPartExclude.custom_duration" type="number" class="settings-select mt-2" :placeholder="t('settings.placeholders.enterDuration')">
                    </div>
                </div>
                <div class="filter-row">
                    <label>{{ t('settings.filters.byFavoriteTime') }}</label>
                    <div>
                        <select v-model="favTimeExclude.type" class="settings-select">
                            <option value="off">{{ t('settings.filters.noFilter') }}</option>
                            <option value="1m">{{ t('settings.filters.before1Month') }}</option>
                            <option value="3m">{{ t('settings.filters.before3Month') }}</option>
                            <option value="6m">{{ t('settings.filters.before6Month') }}</option>
                            <option value="custom">{{ t('settings.filters.customDate') }}</option>
                        </select>
                        <input v-if="favTimeExclude.type === 'custom'" v-model="favTimeExclude.custom_date" type="date" class="settings-select mt-2">
                    </div>
                </div>
            </section>

            <section v-show="tab === 'notify'" class="settings-panel shadow-lg border border-gray-200">
                <h2 class="settings-heading">
                    <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-5 5v-5z"></path>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path>
                    </svg>
                    {{ t('settings.notificationSettings') }}
                </h2>
                <div class="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-6 border border-blue-100">
                    <div class="flex items-center justify-between gap-4">
                        <div class="flex items-center space-x-3 min-w-0">
                            <div class="p-2 bg-blue-500 rounded-lg">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6 text-white">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 12 3.269 3.125A59.769 59.769 0 0 1 21.485 12 59.768 59.768 0 0 1 3.27 20.875L5.999 12Zm0 0h7.5" />
                                </svg>
                            </div>
                            <div>
                                <h3 class="text-lg font-semibold text-gray-800">{{ t('settings.notifications.telegramBot') }}</h3>
                                <p class="text-sm text-gray-600">{{ t('settings.notifications.enableDescription') }}</p>
                            </div>
                        </div>
                        <Toggle v-model="telegramBotEnabled" />
                    </div>
                </div>

                <div v-if="telegramBotEnabled === 'on'" class="space-y-6 mt-6">
                    <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
                        <div class="flex items-start space-x-4">
                            <div class="p-2 bg-green-100 rounded-lg flex-shrink-0">
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6 text-green-600">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M12.75 3.03v.568c0 .334.148.65.405.864l1.068.89c.442.369.535 1.01.216 1.49l-.51.766a2.25 2.25 0 0 1-1.161.886l-.143.048a1.107 1.107 0 0 0-.57 1.664c.369.555.169 1.307-.427 1.605L9 13.125l.423 1.059a.956.956 0 0 1-1.652.928l-.679-.906a1.125 1.125 0 0 0-1.906.172L4.5 15.75l-.612.153M12.75 3.031a9 9 0 0 0-8.862 12.872M12.75 3.031a9 9 0 0 1 6.69 14.036m0 0-.177-.529A2.25 2.25 0 0 0 17.128 15H16.5l-.324-.324a1.453 1.453 0 0 0-2.328.377l-.036.073a1.586 1.586 0 0 1-.982.816l-.99.282c-.55.157-.894.702-.8 1.267l.073.438c.08.474.49.821.97.821.846 0 1.598.542 1.865 1.345l.215.643m5.276-3.67a9.012 9.012 0 0 1-5.276 3.67m0 0a9 9 0 0 1-10.275-4.835M15.75 9c0 .896-.393 1.7-1.016 2.25" />
                                </svg>
                            </div>
                            <div class="flex-1 min-w-0">
                                <label class="block text-sm font-semibold text-gray-800 mb-2">
                                    {{ t('settings.notifications.botApiUrl') }}
                                    <span class="text-gray-500 font-normal ml-1">({{ t('common.optional') }})</span>
                                </label>
                                <input v-model="telegramBotApiUrl" type="url" class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition duration-200 text-sm" :placeholder="t('settings.placeholders.enterBotApiUrl')">
                                <p class="text-xs text-gray-500 mt-2">{{ t('settings.notifications.botApiUrlDescription') }}</p>
                            </div>
                        </div>
                    </div>
                    <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
                        <div class="flex items-start space-x-4">
                            <div class="p-2 bg-yellow-100 rounded-lg flex-shrink-0">
                                <svg class="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"></path>
                                </svg>
                            </div>
                            <div class="flex-1 min-w-0">
                                <label class="block text-sm font-semibold text-gray-800 mb-2">
                                    {{ t('settings.notifications.botToken') }}
                                    <span class="text-red-500 ml-1">*</span>
                                </label>
                                <input v-model="telegramBotToken" type="text" class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition duration-200 text-sm font-mono" :placeholder="t('settings.placeholders.enterBotToken')">
                                <p class="text-xs text-gray-500 mt-2">{{ t('settings.notifications.botTokenDescription') }}</p>
                            </div>
                        </div>
                    </div>
                    <div class="bg-white border border-gray-200 rounded-xl p-6 shadow-sm">
                        <div class="flex items-start space-x-4">
                            <div class="p-2 bg-purple-100 rounded-lg flex-shrink-0">
                                <svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"></path>
                                </svg>
                            </div>
                            <div class="flex-1 min-w-0">
                                <label class="block text-sm font-semibold text-gray-800 mb-2">
                                    {{ t('settings.notifications.chatId') }}
                                    <span class="text-red-500 ml-1">*</span>
                                </label>
                                <input v-model="telegramChatId" type="text" class="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 transition duration-200 text-sm font-mono" :placeholder="t('settings.placeholders.enterChatId')">
                                <p class="text-xs text-gray-500 mt-2">{{ t('settings.notifications.chatIdDescription') }}</p>
                            </div>
                        </div>
                    </div>
                    <div class="bg-gradient-to-r from-blue-50 to-indigo-50 border border-blue-200 rounded-xl p-6">
                        <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between space-y-4 sm:space-y-0">
                            <div class="flex items-center space-x-3">
                                <div class="p-2 bg-blue-500 rounded-lg">
                                    <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                                    </svg>
                                </div>
                                <div>
                                    <h3 class="text-sm font-semibold text-gray-800">{{ t('settings.notifications.connectionTest') }}</h3>
                                    <p class="text-xs text-gray-600">{{ t('settings.notifications.testConnectionDescription') }}</p>
                                </div>
                            </div>
                            <button @click="testTelegramConnection" :disabled="!telegramBotToken || !telegramChatId" class="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-200 font-medium text-sm disabled:opacity-50 disabled:cursor-not-allowed min-w-max">
                                {{ t('settings.notifications.testConnection') }}
                            </button>
                        </div>
                    </div>
                </div>
            </section>

            <div class="text-center mt-4">
                <button @click="saveSettingHandler" class="bg-gradient-to-r from-purple-500 to-pink-500 text-white py-3 px-8 rounded-lg hover:from-purple-600 hover:to-pink-600 focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 transition duration-200 text-lg font-medium shadow-lg">
                    {{ t('settings.saveSettings') }}
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { getFavList } from '@/api/fav';
import { getSettings, saveSettings, testTelegramConnection as testTelegramAPI } from '@/api/settings';
import { getSubscriptionList } from '@/api/subscription';
import { computed, defineComponent, h, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const Help = defineComponent({
    props: { text: { type: String, required: true } },
    setup(props) {
        const open = ref(false);
        const toggle = (event: MouseEvent) => {
            event.stopPropagation();
            open.value = !open.value;
        };
        const close = () => { open.value = false; };
        onMounted(() => document.addEventListener('click', close));
        onUnmounted(() => document.removeEventListener('click', close));
        return () => h('button', {
            type: 'button',
            class: ['help', { open: open.value }],
            'aria-label': '说明',
            'aria-expanded': open.value ? 'true' : 'false',
            onClick: toggle,
        }, ['?', h('span', { class: 'tip' }, props.text)]);
    },
});

const Toggle = defineComponent({
    props: { modelValue: { type: String, required: true } },
    emits: ['update:modelValue'],
    setup(props, { emit }) {
        return () => h('label', { class: 'switch' }, [
            h('input', {
                type: 'checkbox',
                checked: props.modelValue === 'on',
                onChange: (event: Event) => emit('update:modelValue', (event.target as HTMLInputElement).checked ? 'on' : 'off'),
            }),
            h('i'),
        ]);
    },
});

type SourceRow = { id: number; kind: 'favorite' | 'subscription'; name: string; quality: string };

const qualityOptions = [
    { value: '1', label: 'settings.features.qualityBestShort' },
    { value: '2', label: 'settings.features.quality1080Short' },
    { value: '3', label: 'settings.features.qualityLowShort' },
    { value: '4', label: 'settings.features.qualityHighVideoShort' },
    { value: '5', label: 'settings.features.qualityHighAudioShort' },
    { value: 'skip', label: 'settings.features.doNotDownload' },
];

const tab = ref('features');
const sourceQuery = ref('');
const sources = ref<SourceRow[]>([]);
const sourcesReady = ref(false);

const nameExclude = ref({ type: 'off', contains: '', regex: '' });
const durationVideoExclude = ref({ type: 'off', custom_duration: 0 });
const durationVideoPartExclude = ref({ type: 'off', custom_duration: 0 });
const favTimeExclude = ref({ type: 'off', custom_date: '' });
const multiPartitionDownloadEnabled = ref('on');
const danmakuDownloadEnabled = ref('off');
const videoDownloadEnabled = ref('off');
const favoriteSyncEnabled = ref('off');
const humanReadableNameEnabled = ref('off');
const usageAnalyticsEnabled = ref('on');
const videoQuality = ref('1');
const telegramBotEnabled = ref('off');
const telegramBotToken = ref('');
const telegramChatId = ref('');
const telegramBotApiUrl = ref('');
const favExclude = ref({ enabled: false, selected: [] as number[] });

const visibleSources = computed(() => {
    const query = sourceQuery.value.trim().toLowerCase();
    if (!query) return sources.value;
    return sources.value.filter((row) => row.name.toLowerCase().includes(query));
});

const saveSettingHandler = () => {
    const skipped = sources.value.filter((row) => row.quality === 'skip').map((row) => row.id);
    const sourceQuality: Record<string, string> = {};
    sources.value.forEach((row) => {
        if (row.quality && row.quality !== '0') sourceQuality[String(row.id)] = row.quality;
    });
    const nextExclude = sourcesReady.value
        ? { enabled: skipped.length > 0, selected: skipped }
        : favExclude.value;
    saveSettings({
        name_exclude: nameExclude.value,
        video_quality: videoQuality.value,
        source_quality: sourcesReady.value ? sourceQuality : undefined,
        duration_video_exclude: durationVideoExclude.value,
        duration_video_part_exclude: durationVideoPartExclude.value,
        fav_time_exclude: favTimeExclude.value,
        fav_exclude: nextExclude,
        multi_partition_download_enabled: multiPartitionDownloadEnabled.value,
        danmaku_download_enabled: danmakuDownloadEnabled.value,
        video_download_enabled: videoDownloadEnabled.value,
        favorite_sync_enabled: favoriteSyncEnabled.value,
        human_readable_name_enabled: humanReadableNameEnabled.value,
        usage_analytics_enabled: usageAnalyticsEnabled.value,
        telegram_bot_enabled: telegramBotEnabled.value,
        telegram_bot_token: telegramBotToken.value,
        telegram_chat_id: telegramChatId.value,
        telegram_bot_api_url: telegramBotApiUrl.value,
    }).then(() => {
        favExclude.value = nextExclude;
        alert(t('settings.settingsSaved'));
    }).catch((err) => {
        alert(t('settings.settingsSaveFailed') + ' ' + err.message);
    });
};

const testTelegramConnection = async () => {
    if (!telegramBotToken.value || !telegramChatId.value) {
        alert(t('settings.notifications.fillRequiredFields'));
        return;
    }
    try {
        const response = await testTelegramAPI(telegramBotToken.value, telegramChatId.value, telegramBotApiUrl.value);
        alert(response.result === true ? t('settings.notifications.connectionSuccess') : t('settings.notifications.connectionFailed'));
    } catch (error) {
        console.error('Connect Telegram Bot failed:', error);
        alert(t('settings.notifications.connectionError'));
    }
};

function qualityFor(id: number, saved: Record<string, string>, excluded: Set<number>) {
    const value = saved[String(id)];
    if (value) return String(value);
    if (excluded.has(id)) return 'skip';
    return '0';
}

onMounted(() => {
    getSettings().then((data) => {
        nameExclude.value = data.name_exclude;
        videoQuality.value = data.video_quality == null || data.video_quality === '' ? '1' : String(data.video_quality);
        favExclude.value = data.fav_exclude ?? { enabled: false, selected: [] };
        durationVideoExclude.value = data.duration_video_exclude;
        durationVideoPartExclude.value = data.duration_video_part_exclude;
        favTimeExclude.value = data.fav_time_exclude;
        multiPartitionDownloadEnabled.value = data.multi_partition_download_enabled;
        danmakuDownloadEnabled.value = data.danmaku_download_enabled;
        videoDownloadEnabled.value = data.video_download_enabled;
        favoriteSyncEnabled.value = data.favorite_sync_enabled;
        humanReadableNameEnabled.value = data.human_readable_name_enabled;
        usageAnalyticsEnabled.value = data.usage_analytics_enabled;
        telegramBotEnabled.value = data.telegram_bot_enabled || 'off';
        telegramBotToken.value = data.telegram_bot_token || '';
        telegramChatId.value = data.telegram_chat_id || '';
        telegramBotApiUrl.value = data.telegram_bot_api_url || '';
        return data;
    }).then((data) => Promise.all([
        Promise.resolve(data),
        getFavList().catch(() => null),
        getSubscriptionList().catch(() => null),
    ])).then(([data, favs, subs]) => {
        if (!favs || !subs) return;
        const saved = data.source_quality && typeof data.source_quality === 'object' ? data.source_quality as Record<string, string> : {};
        const excluded = new Set<number>(favExclude.value.enabled ? (favExclude.value.selected ?? []).map((id) => Number(id)) : []);
        const rows: SourceRow[] = favs.map((item) => ({
            id: item.id,
            kind: 'favorite',
            name: item.title,
            quality: qualityFor(item.id, saved, excluded),
        }));
        subs.forEach((item) => {
            const id = -item.id;
            rows.push({ id, kind: 'subscription', name: item.name, quality: qualityFor(id, saved, excluded) });
        });
        sources.value = rows;
        sourcesReady.value = true;
    }).catch((error) => {
        console.error(error);
    });
});
</script>

<style scoped>
.settings-tabs {
    position: sticky;
    top: 0;
    z-index: 10;
    display: flex;
    gap: 6px;
    overflow-x: auto;
    margin-bottom: 12px;
    padding: 8px 0;
    background: #fff;
}
.settings-tabs button {
    flex: none;
    border: 0;
    border-radius: 999px;
    padding: 7px 14px;
    background: transparent;
    color: #6b7280;
    cursor: pointer;
}
.settings-tabs button.on {
    background: #f5f3ff;
    color: #7c3aed;
    font-weight: 600;
}
.settings-panel {
    background: #fff;
    border-radius: 12px;
    padding: 16px 18px 18px;
}
.settings-heading {
    display: flex;
    align-items: center;
    margin: 0 0 14px;
    padding-bottom: 12px;
    border-bottom: 1px solid #e5e7eb;
    font-size: 1.05rem;
    font-weight: 600;
    color: #1f2937;
}
.switch-row, .source-row, .filter-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-height: 44px;
}
.source-row, .filter-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) 220px;
    border-top: 1px solid #e5e7eb;
    padding: 8px 0;
}
.filter-row { grid-template-columns: 140px minmax(0, 1fr); align-items: start; }
.filter-row label { color: #6b7280; padding-top: 8px; }
.settings-select {
    width: 100%;
    height: 36px;
    padding: 0 10px;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    background: #fff;
    color: #1f2937;
}
.settings-select.skip { color: #b91c1c; }
.source-tag {
    flex: none;
    font-size: 12px;
    line-height: 20px;
    padding: 0 6px;
    border-radius: 999px;
    background: #f3f4f6;
    color: #6b7280;
}
.source-tag.sub { background: #fdf2f8; color: #be185d; }
:deep(.switch) { position: relative; width: 40px; height: 22px; flex: none; }
:deep(.switch input) { position: absolute; opacity: 0; }
:deep(.switch i) {
    position: absolute; inset: 0; border-radius: 999px; background: #d1d5db; transition: .15s;
}
:deep(.switch i:before) {
    content: ""; position: absolute; top: 2px; left: 2px; width: 18px; height: 18px;
    border-radius: 50%; background: #fff; transition: .15s;
}
:deep(.switch input:checked + i) { background: #7c3aed; }
:deep(.switch input:checked + i:before) { transform: translateX(18px); }
:deep(.help) {
    position: relative;
    width: 16px; height: 16px; margin-left: 6px; padding: 0;
    border: 1px solid #d1d5db; border-radius: 50%;
    background: #fff; color: #6b7280; font-size: 11px; line-height: 14px; cursor: pointer;
}
:deep(.help .tip) {
    display: none; position: absolute; z-index: 30; left: 0; top: calc(100% + 8px);
    width: 220px; padding: 8px 10px; border-radius: 8px;
    background: #111827; color: #fff; font-size: 12px; font-weight: 400; line-height: 1.45; text-align: left;
}
:deep(.help:hover .tip), :deep(.help.open .tip) { display: block; }
@media (max-width: 640px) {
    .source-row, .filter-row { grid-template-columns: 1fr; }
}
</style>
