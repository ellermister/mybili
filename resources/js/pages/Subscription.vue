<template>
    <div class="m-4">
        <Breadcrumbs :items="breadcrumbItems">
            <template #actions>
                <div class="flex items-center gap-3">
                    <!-- 删除模式切换按钮 -->
                    <button @click="deleteMode = !deleteMode"
                        :class="deleteMode ? 'bg-red-600 hover:bg-red-700' : 'bg-gray-600 hover:bg-gray-700'"
                        class="text-white px-3 md:px-4 py-2 rounded-lg transition-colors duration-200 flex items-center gap-2"
                        :title="deleteMode ? t('common.cancel') : t('subscription.editMode')">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path v-if="!deleteMode" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M6 18L18 6M6 6l12 12" />
                        </svg>
                        <span class="hidden md:inline">{{ deleteMode ? t('common.cancel') : t('subscription.editMode') }}</span>
                    </button>

                    <!-- 创建订阅按钮 -->
                    <button @click="showCreateModal = true"
                        class="bg-purple-600 hover:bg-purple-700 text-white px-3 md:px-4 py-2 rounded-lg transition-colors duration-200 flex items-center gap-2"
                        :title="t('subscription.createSubscription')">
                        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                                d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                        </svg>
                        <span class="hidden md:inline">{{ t('subscription.createSubscription') }}</span>
                    </button>
                </div>
            </template>
        </Breadcrumbs>

        <!-- 订阅列表 -->
        <div class="mb-12">
            <h2 class="text-2xl font-bold mb-6 text-gray-800 border-b border-gray-200 pb-3 flex items-center gap-3">
                <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M4 6h16M4 10h16M4 14h16M4 18h16" />
                </svg>
                {{ t('subscription.allSubscriptions') }}
            </h2>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 w-full gap-6">
                <div v-for="subscription in subscriptionList" :key="subscription.id"
                    class="bg-white rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow duration-300">

                    <!-- 图片/头像区域 - 统一高度 -->
                    <div class="relative h-48">
                        <!-- UP主类型 - 圆形头像 + 主色背景 -->
                        <div v-if="subscription.type === 'up'"
                            class="w-full h-full flex items-center justify-center transition-colors duration-300"
                            :style="{ backgroundColor: getDominantColor(subscription) }">
                            <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }"
                                class="block hover:opacity-90 transition-opacity duration-200">
                                <div class="w-32 h-32 rounded-full overflow-hidden border-4 border-white shadow-lg">
                                    <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'"
                                        :title="subscription.name"
                                        class="w-full h-full object-cover select-none pointer-events-none" />
                                </div>
                            </RouterLink>
                        </div>

                        <!-- 剧集/系列类型 - 横板封面 -->
                        <RouterLink v-else :to="{ name: 'subscription-id', params: { id: subscription.id } }"
                            class="block w-full h-full hover:opacity-90 transition-opacity duration-200">
                            <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'"
                                :title="subscription.name"
                                class="w-full h-full object-cover select-none pointer-events-none" />
                        </RouterLink>

                        <!-- 状态标签 - 统一位置 -->
                        <div class="absolute top-2 right-2">
                            <button @click="toggleStatus(subscription.id, subscription.status)"
                                :class="subscription.status === 1 ? 'bg-green-500 hover:bg-green-600' : 'bg-red-500 hover:bg-red-600'"
                                class="text-white text-xs px-2 py-1 rounded-full select-none cursor-pointer transition-colors duration-200 shadow-md">
                                {{ subscription.status === 1 ? t('subscription.enabled') : t('subscription.disabled') }}
                            </button>
                        </div>

                        <!-- 类型标签 - 统一位置 -->
                        <div class="absolute top-2 left-2">
                            <span :class="subscription.type === 'up' ? 'bg-blue-500' : 'bg-orange-500'"
                                class="text-white text-xs px-2 py-1 rounded-full select-none shadow-md">
                                {{ subscription.type === 'up' ? t('subscription.upMaster') : t('subscription.' + subscription.type) }}
                            </span>
                        </div>

                        <!-- 视频数量 - 统一位置 -->
                        <div class="absolute bottom-2 right-2">
                            <span class="bg-black bg-opacity-70 text-white text-sm px-2 py-1 rounded select-none shadow-md">
                                {{ subscription.total }} {{ t('subscription.videoCount') }}
                            </span>
                        </div>
                    </div>

                    <!-- 内容区域 - 统一样式 -->
                    <div class="p-4">
                        <h3 class="font-semibold text-gray-800 mb-2 line-clamp-2 min-h-[2.5rem] flex items-center"
                            :title="subscription.name">
                            <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }"
                                class="hover:text-purple-600 transition-colors duration-200">
                                {{ subscription.name }}
                            </RouterLink>
                        </h3>

                        <p v-if="subscription.description"
                            class="text-sm text-gray-600 mb-3 line-clamp-2 min-h-[2.5rem]">
                            {{ subscription.description }}
                        </p>
                        <p v-else class="text-sm text-gray-600 mb-3 min-h-[2.5rem]">
                            &nbsp;
                        </p>

                        <div class="text-xs text-gray-500 space-y-1">
                            <div class="flex justify-between">
                                <span>ID:</span>
                                <span>{{ subscription.mid }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span>{{ t('subscription.lastCheck') }}:</span>
                                <span>{{ subscription.last_check_at }}</span>
                            </div>
                        </div>

                        <!-- 操作按钮 - 只在删除模式下显示 -->
                        <div v-if="deleteMode" class="mt-4 flex gap-2">
                            <button @click="deleteSubscription(subscription.id)"
                                class="flex-1 bg-red-500 hover:bg-red-600 text-white px-3 py-2 rounded text-sm transition-colors duration-200 flex items-center justify-center gap-1">
                                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" 
                                        d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                </svg>
                                {{ t('subscription.delete') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- 创建订阅模态框 -->
        <div v-if="showCreateModal" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div class="bg-white rounded-lg p-6 w-full max-w-md mx-4">
                <h2 class="text-xl font-semibold mb-4 text-gray-800">{{ t('subscription.createNewSubscription') }}</h2>

                <form @submit.prevent="createSubscription">
                    <div class="mb-4">
                        <label class="block text-sm font-medium text-gray-700 mb-2">{{
                            t('subscription.subscriptionType')
                            }}</label>
                        <select v-model="newSubscription.type"
                            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent">
                            <option value="up">{{ t('subscription.upMaster') }}</option>
                            <option value="series">{{ t('subscription.series') }}</option>
                            <option value="seasons">{{ t('subscription.seasons') }}</option>
                        </select>
                    </div>

                    <div class="mb-6">
                        <label class="block text-sm font-medium text-gray-700 mb-2">{{
                            t('subscription.subscriptionLink')
                            }}</label>
                        <input v-model="newSubscription.url" type="url" required
                            :placeholder="t('subscription.enterSubscriptionLink')"
                            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent" />
                    </div>

                    <div class="flex gap-3">
                        <button type="button" @click="showCreateModal = false"
                            class="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition-colors duration-200">
                            {{ t('common.cancel') }}
                        </button>
                        <button type="submit"
                            class="flex-1 px-4 py-2 bg-purple-600 text-white rounded-lg hover:bg-purple-700 transition-colors duration-200">
                            {{ t('subscription.createNewSubscription') }}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    </div>

</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import Breadcrumbs from '@/components/Breadcrumbs.vue';
import Image from '@/components/Image.vue';
import { createNewSubscription, getSubscriptionList, removeSubscription, updateSubscription, type Subscription } from '@/api/subscription';
import { RouterLink } from 'vue-router';

const { t } = useI18n();

// 响应式数据
const subscriptionList = ref<Subscription[]>([]);
const showCreateModal = ref(false);
const deleteMode = ref(false);
const newSubscription = ref({
    type: 'up' as 'up' | 'seasons',
    url: ''
});

// 面包屑导航
const breadcrumbItems = computed(() => [
    { text: t('navigation.home'), to: '/' },
    { text: t('subscription.title') }
]);

// 创建订阅
const createSubscription = async () => {
    if (!newSubscription.value.url.trim()) {
        alert(t('subscription.enterSubscriptionLink'));
        return;
    }
    try {
        await createNewSubscription(newSubscription.value);
    } catch (error) {
        console.error(error);
        alert(t('subscription.createSubscriptionFailed'));
    }


    refreshSubscriptionList();
    // 重置表单并关闭模态框
    newSubscription.value = { type: 'up', url: '' };
    showCreateModal.value = false;
};

// 删除订阅
const deleteSubscription = async (id: number) => {
    if (confirm(t('subscription.confirmDelete'))) {
        try {
            await removeSubscription(id);

            // 从列表中移除
            const index = subscriptionList.value.findIndex(sub => sub.id === id);
            if (index > -1) {
                subscriptionList.value.splice(index, 1);
            }

            console.log('删除订阅成功:', id);
        } catch (error) {
            console.error('删除订阅失败:', error);
            alert(t('subscription.deleteSubscriptionFailed'));
        }
    }
};

// 切换订阅状态
const toggleStatus = async (id: number, currentStatus: number) => {
    try {
        const newStatus = currentStatus === 1 ? 0 : 1;
        await updateSubscription(id, { status: newStatus });

        // 更新本地状态
        const subscription = subscriptionList.value.find(sub => sub.id === id);
        if (subscription) {
            subscription.status = newStatus;
        }

        console.log('状态切换成功:', id, newStatus);
    } catch (error) {
        console.error('状态切换失败:', error);
    }
};

// 获取并处理主色调（增加亮度使其柔和）
const getDominantColor = (subscription: any): string => {
    const dominantColor = subscription.dominant_color;
    
    // 如果后端没有返回颜色，使用默认灰色
    if (!dominantColor) {
        return '#f3f4f6';
    }

    try {
        // 解析 hex 颜色
        const hex = dominantColor.replace('#', '');
        const r = parseInt(hex.substring(0, 2), 16);
        const g = parseInt(hex.substring(2, 4), 16);
        const b = parseInt(hex.substring(4, 6), 16);
        
        // 使颜色更柔和（增加亮度）
        const lighten = 0.5;
        const newR = Math.min(255, r + (255 - r) * lighten);
        const newG = Math.min(255, g + (255 - g) * lighten);
        const newB = Math.min(255, b + (255 - b) * lighten);
        
        return `rgb(${Math.round(newR)}, ${Math.round(newG)}, ${Math.round(newB)})`;
    } catch (error) {
        console.error('处理颜色失败:', error);
        return '#f3f4f6';
    }
};


const refreshSubscriptionList = async () => {
    try {
        const data = await getSubscriptionList();
        subscriptionList.value = data;
    } catch (error) {
        console.error(error);
    }
}

// 组件挂载时加载数据
onMounted(() => {
    refreshSubscriptionList();
});
</script>

<style scoped>
.line-clamp-1 {
    display: -webkit-box;
    -webkit-line-clamp: 1;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

.line-clamp-2 {
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
}

/* 响应式设计 */
@media (max-width: 768px) {
    .grid {
        grid-template-columns: repeat(1, 1fr);
    }
}

@media (min-width: 769px) and (max-width: 1024px) {
    .grid {
        grid-template-columns: repeat(2, 1fr);
    }
}

@media (min-width: 1025px) and (max-width: 1280px) {
    .grid {
        grid-template-columns: repeat(3, 1fr);
    }
}

@media (min-width: 1281px) {
    .grid {
        grid-template-columns: repeat(4, 1fr);
    }
}
</style>
