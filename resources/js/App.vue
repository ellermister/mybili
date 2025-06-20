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
              {{ link.name }}
            </router-link>
          </div>

          <!-- Mobile Menu Button -->
          <div class="md:hidden">
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
          <h2 class="text-xl font-bold mb-6">Menu</h2>
          <nav class="flex flex-col space-y-4">
            <router-link
              v-for="(link, index) in navLinks"
              :key="`drawer-${index}`"
              :to="link.path"
              @click="isDrawerOpen = false"
              class="drawer-link text-gray-700 hover:text-purple-600 p-2 rounded-md"
            >
              {{ link.name }}
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
import { ref } from 'vue';

const isDrawerOpen = ref(false);

const navLinks = [
  { name: 'Home', path: '/' },
  { name: 'Progress', path: '/progress' },
  { name: 'Cookie', path: '/cookie' },
  { name: 'Settings', path: '/settings' },
  { name: 'About', path: '/about' }
]
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