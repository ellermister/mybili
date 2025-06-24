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
              <p>{{ t('about.databaseSize') }}：{{ (systemInfo.database_usage.db_size / 1024).toFixed(2) }} {{ t('about.units.mb') }}</p>
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

interface SystemInfo {
  app_version: string
  php_version: string
  laravel_version: string
  database_version: string
  timezone: string
  time_now: string
  database_usage: DatabaseUsage
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
  }
})

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