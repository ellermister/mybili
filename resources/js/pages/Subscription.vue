<template>
    <Breadcrumbs :items="breadcrumbItems">
        <template #actions>
            <div class="flex items-center gap-3">
                <!-- 视图切换 -->
                <div class="flex items-center gap-2">
                    <span class="text-sm text-gray-600">{{ t('subscription.viewSwitch') }}:</span>
                    <div class="flex bg-gray-100 rounded-lg p-1">
                        <button @click="currentView = 'unified'" 
                                :class="currentView === 'unified' ? 'bg-white text-purple-600 shadow-sm' : 'text-gray-600 hover:text-gray-800'"
                                class="px-3 py-1 rounded-md text-sm font-medium transition-all duration-200">
                            {{ t('subscription.unifiedView') }}
                        </button>
                        <button @click="currentView = 'categorized'" 
                                :class="currentView === 'categorized' ? 'bg-white text-purple-600 shadow-sm' : 'text-gray-600 hover:text-gray-800'"
                                class="px-3 py-1 rounded-md text-sm font-medium transition-all duration-200">
                            {{ t('subscription.categorizedView') }}
                        </button>
                        <button @click="currentView = 'mixed'" 
                                :class="currentView === 'mixed' ? 'bg-white text-purple-600 shadow-sm' : 'text-gray-600 hover:text-gray-800'"
                                class="px-3 py-1 rounded-md text-sm font-medium transition-all duration-200">
                            {{ t('subscription.mixedView') }}
                        </button>
                    </div>
                </div>
                
                <!-- 创建订阅按钮 -->
                <button @click="showCreateModal = true" 
                        class="bg-purple-600 hover:bg-purple-700 text-white px-4 py-2 rounded-lg transition-colors duration-200 flex items-center gap-2">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
                    </svg>
                    {{ t('subscription.createSubscription') }}
                </button>
            </div>
        </template>
    </Breadcrumbs>

    <!-- 统一卡片样式视图 -->
    <div v-if="currentView === 'unified'" class="mb-12">
        <h2 class="text-2xl font-bold mb-6 text-gray-800 border-b border-gray-200 pb-3 flex items-center gap-3">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
            </svg>
            {{ t('subscription.unifiedCardStyle') }}
        </h2>
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 w-full gap-6">
            <div v-for="subscription in subscriptionList" :key="subscription.id" 
                 class="bg-white rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow duration-300">
                
                <!-- 图片区域 - 统一尺寸 -->
                <div class="relative">
                    <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                               class="block hover:opacity-90 transition-opacity duration-200">
                        <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'" 
                               :title="subscription.name"
                               class="w-full h-48 object-cover select-none pointer-events-none" />
                    </RouterLink>
                    
                    <!-- 状态标签 -->
                    <div class="absolute top-2 right-2">
                        <button @click="toggleStatus(subscription.id, subscription.status)"
                                :class="subscription.status === 1 ? 'bg-green-500 hover:bg-green-600' : 'bg-red-500 hover:bg-red-600'" 
                                class="text-white text-xs px-2 py-1 rounded-full select-none cursor-pointer transition-colors duration-200">
                            {{ subscription.status === 1 ? t('subscription.enabled') : t('subscription.disabled') }}
                        </button>
                    </div>
                    
                    <!-- 类型标签 -->
                    <div class="absolute top-2 left-2">
                        <span :class="subscription.type === 'up' ? 'bg-blue-500' : 'bg-orange-500'" 
                              class="text-white text-xs px-2 py-1 rounded-full select-none">
                            {{ subscription.type === 'up' ? t('subscription.upMaster') : t('subscription.'+ subscription.type) }}
                        </span>
                    </div>
                    
                    <!-- 视频数量 -->
                    <div class="absolute bottom-2 right-2">
                        <span class="bg-black bg-opacity-70 text-white text-sm px-2 py-1 rounded select-none">
                            {{ subscription.total }} {{ t('subscription.videoCount') }}
                        </span>
                    </div>
                </div>
                
                <!-- 内容区域 -->
                <div class="p-4">
                    <h3 class="font-semibold text-gray-800 mb-2 line-clamp-2 min-h-[2.5rem] flex items-center" :title="subscription.name">
                        <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                   class="hover:text-purple-600 transition-colors duration-200">
                            {{ subscription.name }}
                        </RouterLink>
                    </h3>
                    
                    <p v-if="subscription.description" class="text-sm text-gray-600 mb-3 line-clamp-2 min-h-[2.5rem]">
                        {{ subscription.description }}
                    </p>
                    <p v-else class="text-sm text-gray-600 mb-3 min-h-[2.5rem]">
                        &nbsp;
                    </p>
                    
                                            <div class="text-xs text-gray-500 space-y-1">
                            <div class="flex justify-between">
                                <span>{{ t('subscription.upMasterId') }}:</span>
                                <span>{{ subscription.mid }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span>{{ t('subscription.lastCheck') }}:</span>
                                <span>{{ subscription.last_check_at }}</span>
                            </div>
                        </div>
                    
                    <!-- 操作按钮 -->
                    <div class="mt-4 flex gap-2">
                        <button
                                @click="deleteSubscription(subscription.id)"
                                class="flex-1 bg-red-500 hover:bg-red-600 text-white px-3 py-2 rounded text-sm transition-colors duration-200">
                            {{ t('subscription.delete') }}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- 分类分区展示视图 -->
    <div v-if="currentView === 'categorized'" class="mb-12">
        <h2 class="text-2xl font-bold mb-6 text-gray-800 border-b border-gray-200 pb-3 flex items-center gap-3">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
            {{ t('subscription.categorizedDisplay') }}
        </h2>
        
        <!-- Seasons 分区 -->
        <div class="mb-8">
            <h3 class="text-xl font-semibold mb-4 text-gray-700 flex items-center gap-2">
                <span class="w-3 h-3 bg-orange-500 rounded-full"></span>
                {{ t('subscription.seriesSubscriptions') }}
            </h3>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 w-full gap-6">
                <div v-for="subscription in seasonsSubscriptions" :key="subscription.id" 
                     class="bg-white rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow duration-300">
                    
                    <!-- 横板封面 -->
                    <div class="relative">
                        <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                   class="block hover:opacity-90 transition-opacity duration-200">
                            <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'" 
                                   :title="subscription.name"
                                   class="w-full h-32 object-cover select-none pointer-events-none" />
                        </RouterLink>
                        
                        <!-- 状态标签 -->
                        <div class="absolute top-2 right-2">
                            <button @click="toggleStatus(subscription.id, subscription.status)"
                                    :class="subscription.status === 1 ? 'bg-green-500 hover:bg-green-600' : 'bg-red-500 hover:bg-red-600'" 
                                    class="text-white text-xs px-2 py-1 rounded-full select-none cursor-pointer transition-colors duration-200">
                                {{ subscription.status === 1 ? t('subscription.enabled') : t('subscription.disabled') }}
                            </button>
                        </div>
                        
                        <!-- 视频数量 -->
                        <div class="absolute bottom-2 right-2">
                            <span class="bg-black bg-opacity-70 text-white text-sm px-2 py-1 rounded select-none">
                                {{ subscription.total }} {{ t('subscription.videoCount') }}
                            </span>
                        </div>
                    </div>
                    
                    <!-- 内容区域 -->
                    <div class="p-4">
                        <h3 class="font-semibold text-gray-800 mb-2 line-clamp-2 min-h-[2.5rem] flex items-center" :title="subscription.name">
                            <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                       class="hover:text-purple-600 transition-colors duration-200">
                                {{ subscription.name }}
                            </RouterLink>
                        </h3>
                        
                        <p v-if="subscription.description" class="text-sm text-gray-600 mb-3 line-clamp-2 min-h-[2.5rem]">
                            {{ subscription.description }}
                        </p>
                        <p v-else class="text-sm text-gray-600 mb-3 min-h-[2.5rem]">
                            &nbsp;
                        </p>
                        
                        <div class="text-xs text-gray-500 space-y-1">
                            <div class="flex justify-between">
                                <span>{{ t('subscription.upMasterId') }}:</span>
                                <span>{{ subscription.mid }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span>{{ t('subscription.lastCheck') }}:</span>
                                <span>{{ subscription.last_check_at }}</span>
                            </div>
                        </div>
                        
                        <!-- 操作按钮 -->
                        <div class="mt-4 flex gap-2">
                            <button
                                    @click="deleteSubscription(subscription.id)"
                                    class="flex-1 bg-red-500 hover:bg-red-600 text-white px-3 py-2 rounded text-sm transition-colors duration-200">
                                {{ t('subscription.delete') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- UP主分区 -->
        <div>
            <h3 class="text-xl font-semibold mb-4 text-gray-700 flex items-center gap-2">
                <span class="w-3 h-3 bg-blue-500 rounded-full"></span>
                {{ t('subscription.upSubscriptions') }}
            </h3>
            <div class="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 w-full gap-6">
                <div v-for="subscription in upSubscriptions" :key="subscription.id" 
                     class="bg-white rounded-lg shadow-lg overflow-hidden hover:shadow-xl transition-shadow duration-300 text-center">
                    
                    <!-- 圆形头像 -->
                    <div class="relative p-4">
                        <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                   class="block hover:opacity-90 transition-opacity duration-200">
                            <div class="mx-auto w-24 h-24 rounded-full overflow-hidden border-4 border-gray-200">
                                <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'" 
                                       :title="subscription.name"
                                       class="w-full h-full object-cover select-none pointer-events-none" />
                            </div>
                        </RouterLink>
                        
                        <!-- 状态标签 -->
                        <div class="absolute top-6 right-6">
                            <button @click="toggleStatus(subscription.id, subscription.status)"
                                    :class="subscription.status === 1 ? 'bg-green-500 hover:bg-green-600' : 'bg-red-500 hover:bg-red-600'" 
                                    class="text-white text-xs px-2 py-1 rounded-full select-none cursor-pointer transition-colors duration-200">
                                {{ subscription.status === 1 ? t('subscription.enabled') : t('subscription.disabled') }}
                            </button>
                        </div>
                    </div>
                    
                    <!-- 内容区域 -->
                    <div class="px-4 pb-4">
                        <h3 class="font-semibold text-gray-800 mb-2 line-clamp-2 min-h-[2.5rem] flex items-center" :title="subscription.name">
                            <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                       class="hover:text-purple-600 transition-colors duration-200">
                                {{ subscription.name }}
                            </RouterLink>
                        </h3>
                        
                        <p v-if="subscription.description" class="text-xs text-gray-600 mb-3 line-clamp-2 min-h-[2.5rem]">
                            {{ subscription.description }}
                        </p>
                        <p v-else class="text-xs text-gray-600 mb-3 min-h-[2.5rem]">
                            &nbsp;
                        </p>
                        
                        <div class="text-xs text-gray-500 space-y-1 mb-3">
                            <div class="flex justify-between">
                                <span>ID:</span>
                                <span>{{ subscription.mid }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span>{{ t('subscription.videoCount') }}:</span>
                                <span>{{ subscription.total }} {{ t('subscription.videoCount') }}</span>
                            </div>
                            <div class="flex justify-between">
                                <span>{{ t('subscription.lastCheck') }}:</span>
                                <span>{{ subscription.last_check_at }}</span>
                            </div>
                        </div>
                        
                        <!-- 操作按钮 -->
                        <div class="flex gap-2">
                            <button 
                                    @click="deleteSubscription(subscription.id)"
                                    class="flex-1 bg-red-500 hover:bg-red-600 text-white px-2 py-1 rounded text-xs transition-colors duration-200">
                                {{ t('subscription.delete') }}
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- 混合视图：统一卡片 + 分类分区 -->
    <div v-if="currentView === 'mixed'" class="mb-12">
        <h2 class="text-2xl font-bold mb-6 text-gray-800 border-b border-gray-200 pb-3 flex items-center gap-3">
            <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
            </svg>
            {{ t('subscription.mixedViewTitle') }}
        </h2>
        
        <!-- 上半部分：统一卡片样式 -->
        <div class="mb-8">
            <h3 class="text-lg font-semibold mb-4 text-gray-700">{{ t('subscription.allSubscriptionsOverview') }}</h3>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 w-full gap-4">
                <div v-for="subscription in subscriptionList.slice(0, 8)" :key="subscription.id" 
                     class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-lg transition-shadow duration-300">
                    
                    <div class="relative">
                        <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                   class="block hover:opacity-90 transition-opacity duration-200">
                            <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'" 
                                   :title="subscription.name"
                                   class="w-full h-32 object-cover select-none pointer-events-none" />
                        </RouterLink>
                        
                        <div class="absolute top-2 right-2">
                            <button @click="toggleStatus(subscription.id, subscription.status)"
                                    :class="subscription.status === 1 ? 'bg-green-500 hover:bg-green-600' : 'bg-red-500 hover:bg-red-600'" 
                                    class="text-white text-xs px-2 py-1 rounded-full select-none cursor-pointer transition-colors duration-200">
                                {{ subscription.status === 1 ? t('subscription.enabled') : t('subscription.disabled') }}
                            </button>
                        </div>
                        
                        <div class="absolute top-2 left-2">
                            <span :class="subscription.type === 'up' ? 'bg-blue-500' : 'bg-orange-500'" 
                                  class="text-white text-xs px-2 py-1 rounded-full select-none">
                                {{ subscription.type === 'up' ? t('subscription.upMaster') : t('subscription.'+ subscription.type) }}
                            </span>
                        </div>
                    </div>
                    
                    <div class="p-3">
                        <h4 class="font-medium text-gray-800 mb-1 line-clamp-1 text-sm min-h-[1.25rem] flex items-center" :title="subscription.name">
                            <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                       class="hover:text-purple-600 transition-colors duration-200">
                                {{ subscription.name }}
                            </RouterLink>
                        </h4>
                        <div class="text-xs text-gray-500 flex justify-between">
                            <span>{{ subscription.total }} {{ t('subscription.videoCount') }}</span>
                            <span>{{ subscription.last_check_at }}</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
        
        <!-- 下半部分：分类展示 -->
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <!-- Seasons 分区 -->
            <div>
                <h3 class="text-lg font-semibold mb-4 text-gray-700 flex items-center gap-2">
                    <span class="w-3 h-3 bg-orange-500 rounded-full"></span>
                    {{ t('subscription.seriesSubscriptions') }}
                </h3>
                <div class="space-y-3">
                    <div v-for="subscription in seasonsSubscriptions.slice(0, 6)" :key="subscription.id" 
                         class="bg-white rounded-lg shadow-md p-3 flex items-center gap-3 hover:shadow-lg transition-shadow duration-300">
                        
                        <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                   class="block hover:opacity-90 transition-opacity duration-200">
                            <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'" 
                                   :title="subscription.name"
                                   class="w-16 h-12 rounded object-cover flex-shrink-0 select-none pointer-events-none" />
                        </RouterLink>
                        
                        <div class="flex-1 min-w-0">
                            <h4 class="font-medium text-gray-800 mb-1 line-clamp-1 min-h-[1.25rem] flex items-center" :title="subscription.name">
                                <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                           class="hover:text-purple-600 transition-colors duration-200">
                                    {{ subscription.name }}
                                </RouterLink>
                            </h4>
                            <p v-if="subscription.description" class="text-xs text-gray-500 mb-1 line-clamp-1 min-h-[1.25rem]">
                                {{ subscription.description }}
                            </p>
                            <p v-else class="text-xs text-gray-500 mb-1 min-h-[1.25rem]">
                                &nbsp;
                            </p>
                            <div class="text-xs text-gray-500 flex justify-between">
                                <span>{{ subscription.total }} {{ t('subscription.videoCount') }}</span>
                                <span>{{ subscription.last_check_at }}</span>
                            </div>
                        </div>
                        
                        <button v-if="subscription.total === 0" 
                                @click="deleteSubscription(subscription.id)"
                                class="bg-red-500 hover:bg-red-600 text-white px-2 py-1 rounded text-xs transition-colors duration-200">
                            {{ t('subscription.delete') }}
                        </button>
                    </div>
                </div>
            </div>
            
            <!-- UP主分区 -->
            <div>
                <h3 class="text-lg font-semibold mb-4 text-gray-700 flex items-center gap-2">
                    <span class="w-3 h-3 bg-blue-500 rounded-full"></span>
                    {{ t('subscription.upSubscriptions') }}
                </h3>
                <div class="space-y-3">
                    <div v-for="subscription in upSubscriptions.slice(0, 6)" :key="subscription.id" 
                         class="bg-white rounded-lg shadow-md p-3 flex items-center gap-3 hover:shadow-lg transition-shadow duration-300">
                        
                        <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                   class="block hover:opacity-90 transition-opacity duration-200">
                            <div class="w-12 h-12 rounded-full overflow-hidden flex-shrink-0">
                                <Image :src="subscription.cover_info?.image_url || '/assets/images/notfound.webp'" 
                                       :title="subscription.name"
                                       class="w-full h-full object-cover select-none pointer-events-none" />
                            </div>
                        </RouterLink>
                        
                        <div class="flex-1 min-w-0">
                            <h4 class="font-medium text-gray-800 mb-1 line-clamp-1 min-h-[1.25rem] flex items-center" :title="subscription.name">
                                <RouterLink :to="{ name: 'subscription-id', params: { id: subscription.id } }" 
                                           class="hover:text-purple-600 transition-colors duration-200">
                                    {{ subscription.name }}
                                </RouterLink>
                            </h4>
                            <p v-if="subscription.description" class="text-xs text-gray-500 mb-1 line-clamp-1 min-h-[1.25rem]">
                                {{ subscription.description }}
                            </p>
                            <p v-else class="text-xs text-gray-500 mb-1 min-h-[1.25rem]">
                                &nbsp;
                            </p>
                            <div class="text-xs text-gray-500 flex justify-between">
                                <span>{{ subscription.total }} {{ t('subscription.videoCount') }}</span>
                                <span>{{ subscription.last_check_at }}</span>
                            </div>
                        </div>
                        
                        <button v-if="subscription.total === 0" 
                                @click="deleteSubscription(subscription.id)"
                                class="bg-red-500 hover:bg-red-600 text-white px-2 py-1 rounded text-xs transition-colors duration-200">
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
                    <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('subscription.subscriptionType') }}</label>
                    <select v-model="newSubscription.type" 
                            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent">
                        <option value="up">{{ t('subscription.upMaster') }}</option>
                        <option value="series">{{ t('subscription.series') }}</option>
                        <option value="seasons">{{ t('subscription.seasons') }}</option>
                    </select>
                </div>
                
                <div class="mb-6">
                    <label class="block text-sm font-medium text-gray-700 mb-2">{{ t('subscription.subscriptionLink') }}</label>
                    <input v-model="newSubscription.url" 
                           type="url" 
                           required
                           :placeholder="t('subscription.enterSubscriptionLink')"
                           class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-purple-500 focus:border-transparent" />
                </div>
                
                <div class="flex gap-3">
                    <button type="button" 
                            @click="showCreateModal = false"
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
const currentView = ref<'unified' | 'categorized' | 'mixed'>('unified');
const newSubscription = ref({
    type: 'up' as 'up' | 'seasons',
    url: ''
});

// 面包屑导航
const breadcrumbItems = computed(() => [
    { text: t('navigation.home'), to: '/' },
    { text: t('subscription.title') }
]);

// 分类订阅数据
const seasonsSubscriptions = computed(() => 
    subscriptionList.value.filter(sub => sub.type === 'seasons' || sub.type === 'series')
);

const upSubscriptions = computed(() => 
    subscriptionList.value.filter(sub => sub.type === 'up')
);

// 创建订阅
const createSubscription = async() => {
    if (!newSubscription.value.url.trim()) {
        alert(t('subscription.enterSubscriptionLink'));
        return;
    }
    try{
        await createNewSubscription(newSubscription.value);
    }catch(error){
        console.error(error);
        alert(t('subscription.createSubscriptionFailed'));
    }


    refreshSubscriptionList();
    // 重置表单并关闭模态框
    newSubscription.value = { type: 'up', url: '' };
    showCreateModal.value = false;
};

// 删除订阅
const deleteSubscription = async(id: number) => {
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


const refreshSubscriptionList = async() => {
    try{
        const data = await getSubscriptionList();
        subscriptionList.value = data;
    }catch(error){
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
