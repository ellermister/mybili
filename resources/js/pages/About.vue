<template>
  <div class="max-w-2xl mx-auto px-4 py-16">
    <div class="text-center space-y-4">
      <h1 class="text-3xl font-bold">mybili</h1>
      <p class="text-lg text-gray-600">{{ t('about.subtitle') }}</p>
      
      <div class="mt-8">
        <p class="text-gray-700 mb-4">
          {{ t('about.description') }}
        </p>
        
        <div class="space-y-2">
          <h2 class="text-xl font-semibold">{{ t('about.mainFeatures') }}</h2>
          <ul class="text-gray-600 list-disc list-inside">
            <li>{{ t('about.features.syncFavorites') }}</li>
            <li>{{ t('about.features.autoDownload') }}</li>
            <li>{{ t('about.features.onlinePlayback') }}</li>
            <li>{{ t('about.features.danmakuDownload') }}</li>
          </ul>
        </div>
      </div>

      <div class="mt-8" v-if="!loading">
        <div class="flex flex-col sm:flex-row gap-4 justify-center">
          <a 
            href="https://github.com/ellermister/mybili" 
            target="_blank"
            class="inline-flex items-center px-4 py-2 bg-gray-800 hover:bg-gray-700 text-white rounded-lg transition"
          >
            <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path fill-rule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" clip-rule="evenodd" />
            </svg>
            {{ t('about.viewOnGitHub') }}
          </a>
          
          <a 
            href="https://t.me/+SqAyFZfxF7dmNjk1" 
            target="_blank"
            class="inline-flex items-center px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white rounded-lg transition"
          >
            <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M11.944 0A12 12 0 0 0 0 12a12 12 0 0 0 12 12 12 12 0 0 0 12-12A12 12 0 0 0 12 0a12 12 0 0 0-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 0 1 .171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.48.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
            </svg>
            {{ t('about.joinTelegramGroup') }}
          </a>
        </div>
      </div>

      <!-- 赞助支持区域 -->
      <div class="mt-12 border-t pt-8">
        <h2 class="text-xl font-semibold mb-6 text-center">💓 {{ t('about.supportProject') }}</h2>
        <p class="text-gray-600 text-center mb-6">
          {{ t('about.supportDescription') }}
        </p>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-2xl mx-auto">
          <!-- Buy me a coffee -->
          <a 
            href="https://buymeacoffee.com/ellermister" 
            target="_blank"
            class="flex items-center p-4 bg-yellow-50 hover:bg-yellow-100 border border-yellow-200 rounded-lg transition-colors"
          >
            <div class="flex-shrink-0 w-10 h-10 bg-yellow-500 rounded-full flex items-center justify-center mr-4">
              <span class="text-white text-lg">☕</span>
            </div>
            <div>
              <h3 class="font-medium text-gray-900">{{ t('about.buyMeCoffee') }}</h3>
              <p class="text-sm text-gray-600">{{ t('about.buyMeCoffee') }}</p>
            </div>
          </a>

          <!-- 爱发电 -->
          <a 
            href="https://afdian.com/a/eller" 
            target="_blank"
            class="flex items-center p-4 bg-purple-50 hover:bg-purple-100 border border-purple-200 rounded-lg transition-colors"
          >
            <div class="flex-shrink-0 w-10 h-10 bg-purple-500 rounded-full flex items-center justify-center mr-4">
              <span class="text-white text-lg">💜</span>
            </div>
            <div>
              <h3 class="font-medium text-gray-900">{{ t('about.aifadian') }}</h3>
              <p class="text-sm text-gray-600">{{ t('about.goSupport') }}</p>
            </div>
          </a>

          <!-- USDT -->
          <div class="flex items-center p-4 bg-green-50 border border-green-200 rounded-lg">
            <div class="flex-shrink-0 w-10 h-10 bg-green-500 rounded-full flex items-center justify-center mr-4">
              <span class="text-white text-lg">₮</span>
            </div>
            <div class="flex-1">
              <h3 class="font-medium text-gray-900">{{ t('about.usdtTrc20') }}</h3>
              <p class="text-sm text-gray-600 font-mono break-all">TRjWTbPfQBhHawCD8DrfLGa8ECbhPP6F3b</p>
            </div>
            <button 
              @click="copyToClipboard('TRjWTbPfQBhHawCD8DrfLGa8ECbhPP6F3b')"
              class="ml-2 px-3 py-1 text-xs bg-green-500 hover:bg-green-600 text-white rounded transition-colors"
            >
              {{ t('about.copy') }}
            </button>
          </div>

          <!-- LTC -->
          <div class="flex items-center p-4 bg-gray-50 border border-gray-200 rounded-lg">
            <div class="flex-shrink-0 w-10 h-10 bg-gray-500 rounded-full flex items-center justify-center mr-4">
              <span class="text-white text-lg">Ł</span>
            </div>
            <div class="flex-1">
              <h3 class="font-medium text-gray-900">{{ t('about.ltcLitecoin') }}</h3>
              <p class="text-sm text-gray-600 font-mono break-all">LdH6SxbAq3No9P4zaNR2aGgH9Kr9yfuGHi</p>
            </div>
            <button 
              @click="copyToClipboard('LdH6SxbAq3No9P4zaNR2aGgH9Kr9yfuGHi')"
              class="ml-2 px-3 py-1 text-xs bg-gray-500 hover:bg-gray-600 text-white rounded transition-colors"
            >
              {{ t('about.copy') }}
            </button>
          </div>
        </div>
        
        <p class="text-center text-gray-500 text-sm mt-4">
          {{ t('about.supportThankYou') }}
        </p>
      </div>

      <!-- 系统信息区域 -->
      <div class="mt-12 border-t pt-8">
        <h2 class="text-xl font-semibold mb-6">{{ t('about.systemInfo') }}</h2>
        
        <div v-if="loading" class="flex justify-center items-center py-8">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-gray-800"></div>
        </div>
        
        <div v-else class="grid grid-cols-2 gap-4 text-left">
          <div class="space-y-4">
            <div>
              <h3 class="text-lg font-medium mb-2">{{ t('about.versionInfo') }}</h3>
              <div class="space-y-2 text-gray-600">
                <p>{{ t('about.appVersion') }}：{{ systemInfo.app_version }}</p>
                <p>{{ t('about.phpVersion') }}：{{ systemInfo.php_version }}</p>
                <p>{{ t('about.laravelVersion') }}：{{ systemInfo.laravel_version }}</p>
                <p>{{ t('about.databaseVersion') }}：{{ systemInfo.database_version }}</p>
              </div>
            </div>
            
            <div>
              <h3 class="text-lg font-medium mb-2">{{ t('about.timeInfo') }}</h3>
              <div class="space-y-2 text-gray-600">
                <p>{{ t('about.timezone') }}：{{ systemInfo.timezone }}</p>
                <p>{{ t('about.currentTime') }}：{{ systemInfo.time_now }}</p>
              </div>
            </div>
          </div>

          <div>
            <h3 class="text-lg font-medium mb-2">{{ t('about.databaseUsage') }}</h3>
            <div class="space-y-2 text-gray-600">
              <p>{{ t('about.favoriteLists') }}：{{ systemInfo.database_usage.favorite_lists }} {{ t('about.units.count') }}</p>
              <p>{{ t('about.videos') }}：{{ systemInfo.database_usage.videos }} {{ t('about.units.count') }}</p>
              <p>{{ t('about.videoParts') }}：{{ systemInfo.database_usage.video_parts }} {{ t('about.units.count') }}</p>
              <p>{{ t('about.danmaku') }}：{{ systemInfo.database_usage.danmaku.toLocaleString() }} {{ t('about.units.danmaku') }}</p>
              <p>{{ t('about.databaseSize') }}：{{ (systemInfo.database_usage.db_size / 1024 / 1024).toFixed(2) }} {{ t('about.units.mb') }}</p>
              <p>{{ t('about.mediaVideosUsage') }}：{{ systemInfo.media_usage.videos_size }}</p>
              <p>{{ t('about.mediaImagesUsage') }}：{{ systemInfo.media_usage.images_size }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { getSystemInfo } from '@/api/system'

const { t } = useI18n();

interface DatabaseUsage {
  favorite_lists: number
  videos: number
  video_parts: number
  danmaku: number
  db_size: number
}

interface MediaUsage {
  videos_size: string
  images_size: string
}

interface SystemInfo {
  app_version: string
  php_version: string
  laravel_version: string
  database_version: string
  timezone: string
  time_now: string
  database_usage: DatabaseUsage
  media_usage: MediaUsage
}

const loading = ref(true)
const systemInfo = ref<SystemInfo>({
  app_version: '',
  php_version: '',
  laravel_version: '',
  database_version: '',
  timezone: '',
  time_now: '',
  database_usage: {
    favorite_lists: 0,
    videos: 0,
    video_parts: 0,
    danmaku: 0,
    db_size: 0
  },
  media_usage:{
    videos_size: '',
    images_size: ''
  }
})

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    // 这里可以添加一个提示，比如 toast 通知
    alert(t('about.copiedToClipboard'))
  } catch (err) {
    console.error('复制失败:', err)
    // 降级方案：使用传统的复制方法
    const textArea = document.createElement('textarea')
    textArea.value = text
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    alert(t('about.copiedToClipboard'))
  }
}

onMounted(async () => {
  try {
    const res = await getSystemInfo()
    systemInfo.value = res
  } catch (error) {
    console.error('获取系统信息失败:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
</style>