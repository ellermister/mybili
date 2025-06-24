<template>
  <div class="min-h-screen bg-gray-50">
    <div class="container mx-auto px-4 py-4">
      <nav class="bg-gradient-to-r from-purple-500 to-pink-500 rounded-lg shadow-lg mb-6">
        <div class="flex items-center justify-between p-4">
          <!-- Logo -->
          <router-link to="/" class="flex items-center">
            <img src="/assets/images/logo-white.svg" alt="logo" class="w-8 h-8">
            <span class="ml-2 text-white font-bold text-xl">Mybili</span>
          </router-link>

          <!-- Desktop Nav Links -->
          <div class="hidden md:flex items-center space-x-2 md:space-x-6">
            <router-link 
              v-for="(link, index) in navLinks" 
              :key="index"
              :to="link.path"
              class="nav-link px-3 py-2 rounded-md transition-all duration-300 hover:bg-white/20"
            >
              {{ t(`navigation.${link.key}`) }}
            </router-link>
            
            <!-- Language Selector -->
            <div class="relative">
              <button 
                @click="isLanguageDropdownOpen = !isLanguageDropdownOpen"
                class="nav-link px-3 py-2 rounded-md transition-all duration-300 hover:bg-white/20 flex items-center"
              >
                <span class="mr-1">🌐</span>
                {{ currentLanguageName }}
                <svg class="w-4 h-4 ml-1 transition-transform" :class="{ 'rotate-180': isLanguageDropdownOpen }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
                </svg>
              </button>
              
              <!-- Language Dropdown -->
              <div v-if="isLanguageDropdownOpen" class="absolute right-0 mt-2 w-32 bg-white rounded-md shadow-lg z-50">
                <div class="py-1">
                  <button 
                    v-for="lang in availableLanguages" 
                    :key="lang.code"
                    @click="changeLanguage(lang.code)"
                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors"
                    :class="{ 'bg-purple-100 text-purple-700': currentLocale === lang.code }"
                  >
                    {{ lang.name }}
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- Mobile Menu Button -->
          <div class="md:hidden flex items-center space-x-2">
            <!-- Mobile Language Selector -->
            <div class="relative">
              <button 
                @click="isMobileLanguageDropdownOpen = !isMobileLanguageDropdownOpen"
                class="text-white p-2 rounded-md hover:bg-white/20 transition-all duration-300"
              >
                <span class="text-lg">🌐</span>
              </button>
              
              <!-- Mobile Language Dropdown -->
              <div v-if="isMobileLanguageDropdownOpen" class="absolute right-0 mt-2 w-32 bg-white rounded-md shadow-lg z-50">
                <div class="py-1">
                  <button 
                    v-for="lang in availableLanguages" 
                    :key="lang.code"
                    @click="changeLanguage(lang.code)"
                    class="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 transition-colors"
                    :class="{ 'bg-purple-100 text-purple-700': currentLocale === lang.code }"
                  >
                    {{ lang.name }}
                  </button>
                </div>
              </div>
            </div>
            
            <button @click="isDrawerOpen = true">
              <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"></path></svg>
            </button>
          </div>
        </div>
      </nav>

      <!-- Drawer Overlay -->
      <transition name="fade">
        <div v-if="isDrawerOpen" @click="isDrawerOpen = false" class="fixed inset-0 bg-black bg-opacity-50 z-40 md:hidden"></div>
      </transition>

      <!-- Drawer Menu -->
      <transition name="slide">
        <div v-if="isDrawerOpen" class="fixed top-0 left-0 h-full w-64 bg-white z-50 shadow-lg p-6 md:hidden">
          <h2 class="text-xl font-bold mb-6">{{ t('navigation.menu') }}</h2>
          <nav class="flex flex-col space-y-4">
            <router-link
              v-for="(link, index) in navLinks"
              :key="`drawer-${index}`"
              :to="link.path"
              @click="isDrawerOpen = false"
              class="drawer-link text-gray-700 hover:text-purple-600 p-2 rounded-md"
            >
              {{ t(`navigation.${link.key}`) }}
            </router-link>
          </nav>
        </div>
      </transition>

      <main class="bg-white rounded-lg shadow-lg p-6">
        <router-view></router-view>
      </main>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';

const { t, locale } = useI18n();
const isDrawerOpen = ref(false);
const isLanguageDropdownOpen = ref(false);
const isMobileLanguageDropdownOpen = ref(false);

const navLinks = [
  { key: 'home', path: '/' },
  { key: 'progress', path: '/progress' },
  { key: 'cookie', path: '/cookie' },
  { key: 'settings', path: '/settings' },
  { key: 'about', path: '/about' }
];

const availableLanguages = [
  { code: 'zh-CN', name: '中文' },
  { code: 'en-US', name: 'English' }
];

const currentLocale = computed(() => locale.value);

const currentLanguageName = computed(() => {
  const lang = availableLanguages.find(l => l.code === currentLocale.value);
  return lang ? lang.name : '中文';
});

const changeLanguage = (langCode: string) => {
  locale.value = langCode;
  localStorage.setItem('locale', langCode);
  isLanguageDropdownOpen.value = false;
  isMobileLanguageDropdownOpen.value = false;
};

// 点击外部关闭下拉菜单
const handleClickOutside = (event: Event) => {
  const target = event.target as HTMLElement;
  if (!target.closest('.relative')) {
    isLanguageDropdownOpen.value = false;
    isMobileLanguageDropdownOpen.value = false;
  }
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  // 从localStorage恢复语言设置
  const savedLocale = localStorage.getItem('locale');
  if (savedLocale && availableLanguages.some(lang => lang.code === savedLocale)) {
    locale.value = savedLocale;
  }
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
});
</script>

<style scoped>
.nav-link {
  @apply text-white font-medium;
}

.nav-link.router-link-active {
  @apply bg-white/20;
}

.drawer-link.router-link-active {
  @apply bg-purple-100 text-purple-700 font-semibold;
}

/* Transitions for drawer */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s ease-in-out;
}

.slide-enter-from,
.slide-leave-to {
  transform: translateX(-100%);
}

@media (max-width: 640px) {
  .nav-link {
    @apply text-sm py-1;
  }
}
</style>