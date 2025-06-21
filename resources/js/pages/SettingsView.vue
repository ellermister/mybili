<template>
    <div class="min-h-screen bg-gray-100 p-6">
        <div class="max-w-md mx-auto bg-white rounded-lg shadow-lg p-6">
            <h1 class="text-2xl font-bold mb-6 text-gray-800">Settings</h1>

            <!-- Filter by Name -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Exclude by Name</label>
                <div class="space-y-2">
                    <!-- Off (Default) -->
                    <label class="flex items-center space-x-3">
                        <input v-model="nameExclude.type" type="radio" value="off"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Off (Default)</span>
                    </label>

                    <!-- Name Contains -->
                    <label class="flex items-center space-x-3">
                        <input v-model="nameExclude.type" type="radio" value="contains"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Name Contains</span>
                        <input v-if="nameExclude.type === 'contains'" v-model="nameExclude.contains" type="text"
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                            placeholder="Enter text" />
                    </label>

                    <!-- Name Matches Regex -->
                    <label class="flex items-center space-x-3">
                        <input v-model="nameExclude.type" type="radio" value="regex"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Name Matches Regex</span>
                        <input v-if="nameExclude.type === 'regex'" v-model="nameExclude.regex" type="text"
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                            placeholder="Enter regex" />
                    </label>
                </div>
            </div>

            <!-- Filter by Video Size -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Exclude by Size</label>
                <div class="space-y-2">
                    <!-- Off (Default) -->
                    <label class="flex items-center space-x-3">
                        <input v-model="sizeExclude.type" type="radio" value="off"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Off (Default)</span>
                    </label>

                    <!-- Videos Larger Than 1GB -->
                    <label class="flex items-center space-x-3">
                        <input v-model="sizeExclude.type" type="radio" value="1GB"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Videos Larger Than 1GB</span>
                    </label>

                    <!-- Videos Larger Than 2GB -->
                    <label class="flex items-center space-x-3">
                        <input v-model="sizeExclude.type" type="radio" value="2GB"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Videos Larger Than 2GB</span>
                    </label>

                    <!-- Custom Size -->
                    <label class="flex items-center space-x-3">
                        <input v-model="sizeExclude.type" type="radio" value="custom"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Custom</span>
                        <input v-if="sizeExclude.type === 'custom'" v-model="sizeExclude.custom_size" type="number"
                            class="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition duration-200"
                            placeholder="Enter size in MB" />
                    </label>
                </div>
            </div>

            <!-- Filter by Collections -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Exclude by Favorites</label>
                <div class="space-y-2">
                    <!-- Off (Default) -->
                    <label class="flex items-center space-x-3">
                        <input v-model="favExclude.enabled" type="checkbox"
                            class="form-checkbox h-5 w-5 text-purple-600 rounded border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Enable Favorites Filter</span>
                    </label>

                    <!-- Collection Selection -->
                    <div v-if="favExclude.enabled" class="pl-8 space-y-2">
                        <div class="flex flex-wrap gap-2">
                            <label v-for="collection in availableCollections" :key="collection.id"
                                class="inline-flex items-center px-3 py-1.5 bg-gray-100 hover:bg-gray-200 rounded-full cursor-pointer transition duration-200">
                                <input type="checkbox" v-model="favExclude.selected" :value="collection.id"
                                    class="form-checkbox h-4 w-4 text-purple-600 rounded border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                                <span class="ml-2 text-sm text-gray-700">{{ collection.name }}</span>
                            </label>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Favorite Sync -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Favorite Sync</label>
                <div class="space-y-2">
                    <!-- Off (Default) -->
                    <label class="flex items-center space-x-3">
                        <input v-model="favoriteSyncEnabled" type="radio" value="off"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Off (Default)</span>
                    </label>
                    
                         <!-- On (Default) -->
                         <label class="flex items-center space-x-3">
                        <input v-model="favoriteSyncEnabled" type="radio" value="on"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">On </span>
                    </label>
                </div>
            </div>

            <!-- Multi-Partition Download -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Multi-Partition Download</label>
                <div class="space-y-2">
                    <!-- Off -->
                    <label class="flex items-center space-x-3">
                        <input v-model="multiPartitionDownloadEnabled" type="radio" value="off"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Off (Default)</span>
                    </label>

                    <!-- On (Default) -->
                    <label class="flex items-center space-x-3">
                        <input v-model="multiPartitionDownloadEnabled" type="radio" value="on"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">On </span>
                    </label>
                </div>
            </div>

            <!-- Danmaku Download -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Danmaku Download</label>
                <div class="space-y-2">
                    <!-- Off (Default) -->
                    <label class="flex items-center space-x-3">
                        <input v-model="danmakuDownloadEnabled" type="radio" value="off"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Off (Default)</span>
                    </label>

                    <!-- On -->
                    <label class="flex items-center space-x-3">
                        <input v-model="danmakuDownloadEnabled" type="radio" value="on"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">On</span>
                    </label>
                </div>
            </div>

            <!-- Video Download -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Video Download</label>
                <div class="space-y-2">
                    <!-- Off (Default) -->
                    <label class="flex items-center space-x-3">
                        <input v-model="videoDownloadEnabled" type="radio" value="off"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Off (Default)</span>
                    </label>
                </div>

                <!-- On -->
                <label class="flex items-center space-x-3">
                    <input v-model="videoDownloadEnabled" type="radio" value="on"
                        class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                    <span class="text-gray-700">On</span>
                </label>
            </div>


            <!-- Human Readable Name -->
            <div class="mb-6" v-if="availableCollections">
                <label class="block text-sm font-medium text-gray-700 mb-2">Human Readable Name</label>
                <div class="space-y-2">
                    <label class="flex items-center space-x-3">
                        <input v-model="humanReadableNameEnabled" type="radio" value="off"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">Off (Default)</span>
                    </label>
                    <label class="flex items-center space-x-3">
                        <input v-model="humanReadableNameEnabled" type="radio" value="on"
                            class="form-radio h-5 w-5 text-purple-600 rounded-full border-2 border-gray-300 focus:ring-2 focus:ring-purple-500 transition duration-200" />
                        <span class="text-gray-700">On</span>
                    </label>
                </div>
            </div>

            <!-- Save Button -->
            <button @click="saveSettingHandler"
                class="w-full bg-gradient-to-r from-purple-500 to-pink-500 text-white py-2 px-4 rounded-lg hover:from-purple-600 hover:to-pink-600 focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 transition duration-200">
                Save Settings
            </button>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { getFavList } from '@/api/fav';
import { getSettings, saveSettings } from '@/api/settings';
import { onMounted, ref } from 'vue';

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

// 添加合集过滤相关的响应式数据
const favExclude = ref({
    enabled: false,
    selected: []
});

const availableCollections = ref([]);

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
    }).then(()=>{
        alert('Settings saved successfully!');
    });
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