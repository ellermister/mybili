<template>
    <div class="min-h-screen bg-gray-100 p-6">
        
        <div class="max-w-4xl mx-auto">
            <h1 class="text-3xl font-bold mb-8 text-gray-800 text-center">{{ t('settings.title') }}</h1>

            <!-- 功能开关分组 -->
            <div class="bg-white rounded-lg shadow-lg p-6 mb-6">
                <h2 class="text-xl font-semibold mb-6 text-gray-800 border-b border-gray-200 pb-3">
                    <span class="inline-flex items-center">
                        <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        </svg>
                        {{ t('settings.featureSwitches') }}
                    </span>
                </h2>
                
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <!-- Favorite Sync -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.features.favoriteSync') }}</label>
                        <div class="space-y-2">
                            <label class="flex items-center space-x-3">
                                <input v-model="favoriteSyncEnabled" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.off') }}</span>
                            </label>
                            <label class="flex items-center space-x-3">
                                <input v-model="favoriteSyncEnabled" type="radio" value="on"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.on') }}</span>
                            </label>
                        </div>
                    </div>

                    <!-- Video Download -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.features.videoDownload') }}</label>
                        <div class="space-y-2">
                            <label class="flex items-center space-x-3">
                                <input v-model="videoDownloadEnabled" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.off') }}</span>
                            </label>
                            <label class="flex items-center space-x-3">
                                <input v-model="videoDownloadEnabled" type="radio" value="on"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.on') }}</span>
                            </label>
                        </div>
                    </div>

                    <!-- Danmaku Download -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.features.danmakuDownload') }}</label>
                        <div class="space-y-2">
                            <label class="flex items-center space-x-3">
                                <input v-model="danmakuDownloadEnabled" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.off') }}</span>
                            </label>
                            <label class="flex items-center space-x-3">
                                <input v-model="danmakuDownloadEnabled" type="radio" value="on"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.on') }}</span>
                            </label>
                        </div>
                    </div>

                    <!-- Multi-Partition Download -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.features.multiPartitionDownload') }}</label>
                        <div class="space-y-2">
                            <label class="flex items-center space-x-3">
                                <input v-model="multiPartitionDownloadEnabled" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.off') }}</span>
                            </label>
                            <label class="flex items-center space-x-3">
                                <input v-model="multiPartitionDownloadEnabled" type="radio" value="on"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.on') }}</span>
                            </label>
                        </div>
                    </div>

                    <!-- Human Readable Name -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.features.humanReadableName') }}</label>
                        <div class="space-y-2">
                            <label class="flex items-center space-x-3">
                                <input v-model="humanReadableNameEnabled" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.off') }}</span>
                            </label>
                            <label class="flex items-center space-x-3">
                                <input v-model="humanReadableNameEnabled" type="radio" value="on"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.on') }}</span>
                            </label>
                        </div>
                    </div>

                    <!-- Usage Analytics -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.features.usageAnalytics') }}</label>
                        <div class="space-y-2">
                            <label class="flex items-center space-x-3">
                                <input v-model="usageAnalyticsEnabled" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.off') }}</span>
                            </label>
                            <label class="flex items-center space-x-3">
                                <input v-model="usageAnalyticsEnabled" type="radio" value="on"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.on') }}</span>
                            </label>
                        </div>
                        <p class="text-xs text-gray-500 mt-2">{{ t('settings.features.usageAnalyticsDescription') }}</p>
                    </div>
                </div>
            </div>

            <!-- 过滤设置分组 -->
            <div class="bg-white rounded-lg shadow-lg p-6 mb-6" v-if="availableCollections">
                <h2 class="text-xl font-semibold mb-6 text-gray-800 border-b border-gray-200 pb-3">
                    <span class="inline-flex items-center">
                        <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z"></path>
                        </svg>
                        {{ t('settings.filterSettings') }}
                    </span>
                </h2>

                <div class="space-y-6">
                    <!-- Filter by Name -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.filters.byName') }}</label>
                        <div class="space-y-3">
                            <label class="flex items-center space-x-3">
                                <input v-model="nameExclude.type" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.noFilter') }}</span>
                            </label>

                            <label class="flex items-center space-x-3">
                                <input v-model="nameExclude.type" type="radio" value="contains"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.containsKeyword') }}</span>
                            </label>
                            <input v-if="nameExclude.type === 'contains'" v-model="nameExclude.contains" type="text"
                                class="ml-8 w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                                :placeholder="t('settings.placeholders.enterKeyword')" />

                            <label class="flex items-center space-x-3">
                                <input v-model="nameExclude.type" type="radio" value="regex"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.regexPattern') }}</span>
                            </label>
                            <input v-if="nameExclude.type === 'regex'" v-model="nameExclude.regex" type="text"
                                class="ml-8 w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                                :placeholder="t('settings.placeholders.enterRegex')" />
                        </div>
                    </div>

                    <!-- Filter by Video Size -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.filters.bySize') }}</label>
                        <div class="space-y-3">
                            <label class="flex items-center space-x-3">
                                <input v-model="sizeExclude.type" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.noFilter') }}</span>
                            </label>

                            <label class="flex items-center space-x-3">
                                <input v-model="sizeExclude.type" type="radio" value="1GB"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.largerThan1GB') }}</span>
                            </label>

                            <label class="flex items-center space-x-3">
                                <input v-model="sizeExclude.type" type="radio" value="2GB"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.largerThan2GB') }}</span>
                            </label>

                            <label class="flex items-center space-x-3">
                                <input v-model="sizeExclude.type" type="radio" value="custom"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.customSize') }}</span>
                            </label>
                            <input v-if="sizeExclude.type === 'custom'" v-model="sizeExclude.custom_size" type="number"
                                class="ml-8 w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                                :placeholder="t('settings.placeholders.enterSizeMB')" />
                        </div>
                    </div>

                    <!-- Filter by Collections -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.filters.byFavorites') }}</label>
                        <div class="space-y-3">
                            <label class="flex items-center space-x-3">
                                <input v-model="favExclude.enabled" type="checkbox"
                                    class="form-checkbox h-5 w-5 text-purple-600 rounded border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('settings.filters.enableFavoritesFilter') }}</span>
                            </label>

                            <div v-if="favExclude.enabled" class="ml-8">
                                <p class="text-sm text-gray-600 mb-3">{{ t('settings.filters.selectExcludedFavorites') }}</p>
                                <div class="flex flex-wrap gap-2">
                                    <label v-for="collection in availableCollections" :key="collection.id"
                                        class="inline-flex items-center px-3 py-1.5 bg-white hover:bg-gray-50 rounded-full cursor-pointer transition duration-200 border border-gray-200">
                                        <input type="checkbox" v-model="favExclude.selected" :value="collection.id"
                                            class="form-checkbox h-4 w-4 text-purple-600 rounded border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                        <span class="ml-2 text-sm text-gray-700">{{ collection.name }}</span>
                                    </label>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>


            <!-- 通知设置分组 -->
            <div class="bg-white rounded-lg shadow-lg p-6 mb-6">
                <h2 class="text-xl font-semibold mb-6 text-gray-800 border-b border-gray-200 pb-3">
                    <span class="inline-flex items-center">
                        <svg class="w-5 h-5 mr-2 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-5 5v-5z"></path>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z"></path>
                        </svg>
                        {{ t('settings.notificationSettings') }}
                    </span>
                </h2>

                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <!-- Telegram Bot 启用 -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.notifications.telegramBot') }}</label>
                        <div class="space-y-2">
                            <label class="flex items-center space-x-3">
                                <input v-model="telegramBotEnabled" type="radio" value="off"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.off') }}</span>
                            </label>
                            <label class="flex items-center space-x-3">
                                <input v-model="telegramBotEnabled" type="radio" value="on"
                                    class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="text-gray-700">{{ t('common.on') }}</span>
                            </label>
                        </div>
                        <p class="text-xs text-gray-500 mt-2">{{ t('settings.notifications.enableDescription') }}</p>
                    </div>

                    <!-- Telegram Bot Token -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.notifications.botToken') }}</label>
                        <input v-model="telegramBotToken" type="text"
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                            :placeholder="t('settings.placeholders.enterBotToken')" />
                        <p class="text-xs text-gray-500 mt-2">{{ t('settings.notifications.botTokenDescription') }}</p>
                    </div>

                    <!-- Telegram Chat ID -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.notifications.chatId') }}</label>
                        <input v-model="telegramChatId" type="text"
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                            :placeholder="t('settings.placeholders.enterChatId')" />
                        <p class="text-xs text-gray-500 mt-2">{{ t('settings.notifications.chatIdDescription') }}</p>
                    </div>

                    <!-- 测试连接按钮 -->
                    <div class="bg-gray-50 rounded-lg p-4">
                        <label class="block text-sm font-medium text-gray-700 mb-3">{{ t('settings.notifications.connectionTest') }}</label>
                        <button @click="testTelegramConnection"
                            class="w-full px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-200">
                            {{ t('settings.notifications.testConnection') }}
                        </button>
                        <p class="text-xs text-gray-500 mt-2">{{ t('settings.notifications.testConnectionDescription') }}</p>
                    </div>
                </div>
            </div>


            <!-- Save Button -->
            <div class="text-center">
                <button @click="saveSettingHandler"
                    class="bg-gradient-to-r from-purple-500 to-pink-500 text-white py-3 px-8 rounded-lg hover:from-purple-600 hover:to-pink-600 focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 transition duration-200 text-lg font-medium shadow-lg">
                    {{ t('settings.saveSettings') }}
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { getFavList } from '@/api/fav';
import { getSettings, saveSettings, testTelegramConnection as testTelegramAPI } from '@/api/settings';
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

interface Collection {
    id: number;
    name: string;
}

const nameExclude = ref({
    type: 'off',
    contains: '',
    regex: ''
});

const sizeExclude = ref({
    type: 'off',
    custom_size: 0
});

const multiPartitionDownloadEnabled = ref('on'); // off, on
const danmakuDownloadEnabled = ref('off'); // off, on
const videoDownloadEnabled = ref('off'); // off, on
const favoriteSyncEnabled = ref('off'); // off, on
const humanReadableNameEnabled = ref('off'); // off, on
const usageAnalyticsEnabled = ref('on'); // on, off

// Telegram Bot 设置相关的响应式数据
const telegramBotEnabled = ref('off'); // off, on
const telegramBotToken = ref('');
const telegramChatId = ref('');

// 添加合集过滤相关的响应式数据
const favExclude = ref({
    enabled: false,
    selected: [] as number[]
});

const availableCollections = ref<Collection[]>([]);

const saveSettingHandler = () => {
    console.log('Settings saved:', {
        name_exclude: nameExclude.value,
        size_exclude: sizeExclude.value,
        fav_exclude: favExclude.value,
        multi_partition_download_enabled: multiPartitionDownloadEnabled.value,
        danmaku_download_enabled: danmakuDownloadEnabled.value,
        video_download_enabled: videoDownloadEnabled.value,
        favorite_sync_enabled: favoriteSyncEnabled.value,
        human_readable_name_enabled: humanReadableNameEnabled.value,
        usage_analytics_enabled: usageAnalyticsEnabled.value,
        telegram_bot_enabled: telegramBotEnabled.value,
        telegram_bot_token: telegramBotToken.value,
        telegram_chat_id: telegramChatId.value,
    });

    saveSettings({
        name_exclude: nameExclude.value,
        size_exclude: sizeExclude.value,
        fav_exclude: favExclude.value,
        multi_partition_download_enabled: multiPartitionDownloadEnabled.value,
        danmaku_download_enabled: danmakuDownloadEnabled.value,
        video_download_enabled: videoDownloadEnabled.value,
        favorite_sync_enabled: favoriteSyncEnabled.value,
        human_readable_name_enabled: humanReadableNameEnabled.value,
        usage_analytics_enabled: usageAnalyticsEnabled.value,
        telegram_bot_enabled: telegramBotEnabled.value,
        telegram_bot_token: telegramBotToken.value,
        telegram_chat_id: telegramChatId.value,
    }).then(()=>{
        alert(t('settings.settingsSaved'));
    });
};

const testTelegramConnection = async () => {
    if (!telegramBotToken.value || !telegramChatId.value) {
        alert(t('settings.notifications.fillRequiredFields'));
        return;
    }

    try {
        const response = await testTelegramAPI(telegramBotToken.value, telegramChatId.value);
        
        if (response.result === true) {
            alert(t('settings.notifications.connectionSuccess'));
        } else {
            alert(t('settings.notifications.connectionFailed'));
        }
    } catch (error) {
        console.error('Connect Telegram Bot failed:', error);
        alert(t('settings.notifications.connectionError'));
    }
};

onMounted(()=>{
    getFavList().then((data)=>{
        availableCollections.value = data.map((item:any)=>({
            id: item.id,
            name: item.title,
        }));
    })
    getSettings().then((data)=>{
        console.log(data);

        nameExclude.value = data.name_exclude;
        sizeExclude.value = data.size_exclude;
        favExclude.value = data.fav_exclude;
        multiPartitionDownloadEnabled.value = data.multi_partition_download_enabled;
        danmakuDownloadEnabled.value = data.danmaku_download_enabled;
        videoDownloadEnabled.value = data.video_download_enabled;
        favoriteSyncEnabled.value = data.favorite_sync_enabled;
        humanReadableNameEnabled.value = data.human_readable_name_enabled;
        usageAnalyticsEnabled.value = data.usage_analytics_enabled;
        
        // 加载 Telegram Bot 设置
        telegramBotEnabled.value = data.telegram_bot_enabled || 'off';
        telegramBotToken.value = data.telegram_bot_token || '';
        telegramChatId.value = data.telegram_chat_id || '';
    })
})

</script>

<style scoped>
/* 自定义单选按钮样式 */
.form-radio {
    -webkit-appearance: none;
    -moz-appearance: none;
    appearance: none;
    border-radius: 50%;
    width: 1.25rem;
    height: 1.25rem;
    border: 2px solid #d1d5db;
    /* 默认边框颜色 */
    transition: all 0.2s ease;
    flex-shrink: 0;
    /* 防止单选按钮被挤压 */
}

.form-radio:checked {
    border-color: #9333ea;
    /* 选中时的边框颜色 */
    background-color: #9333ea;
    /* 选中时的背景颜色 */
    background-image: url("data:image/svg+xml,%3csvg viewBox='0 0 16 16' fill='white' xmlns='http://www.w3.org/2000/svg'%3e%3ccircle cx='8' cy='8' r='3'/%3e%3c/svg%3e");
    background-position: center;
    background-repeat: no-repeat;
    background-size: 60%;
}

.form-radio:focus {
    outline: none;
    box-shadow: 0 0 0 2px rgba(147, 51, 234, 0.5);
    /* 聚焦时的阴影 */
}

/* 添加复选框样式 */
.form-checkbox {
    -webkit-appearance: none;
    -moz-appearance: none;
    appearance: none;
    padding: 0;
    border-radius: 0.25rem;
    width: 1.25rem;
    height: 1.25rem;
    border: 2px solid #d1d5db;
    transition: all 0.2s ease;
    flex-shrink: 0;
}

.form-checkbox:checked {
    border-color: #9333ea;
    background-color: #9333ea;
    background-image: url("data:image/svg+xml,%3csvg viewBox='0 0 16 16' fill='white' xmlns='http://www.w3.org/2000/svg'%3e%3cpath d='M12.207 4.793a1 1 0 010 1.414l-5 5a1 1 0 01-1.414 0l-2-2a1 1 0 011.414-1.414L6.5 9.086l4.293-4.293a1 1 0 011.414 0z'/%3e%3c/svg%3e");
    background-position: center;
    background-repeat: no-repeat;
    background-size: 100% 100%;
}

.form-checkbox:focus {
    outline: none;
    box-shadow: 0 0 0 2px rgba(147, 51, 234, 0.5);
}
</style>