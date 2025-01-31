<template>
  <div class="min-h-screen bg-gray-50">
    <div class="container mx-auto px-4 py-4">
      <nav class="bg-gradient-to-r from-purple-500 to-pink-500 rounded-lg shadow-lg mb-6">
        <div class="flex flex-wrap items-center justify-center md:justify-start space-x-2 md:space-x-6 p-4">
          <router-link 
            v-for="(link, index) in navLinks" 
            :key="index"
            :to="link.path"
            class="nav-link px-3 py-2 rounded-md transition-all duration-300 hover:bg-white/20"
          >
            {{ link.name }}
          </router-link>
        </div>
      </nav>

      <main class="bg-white rounded-lg shadow-lg p-6">
        <router-view></router-view>
      </main>
    </div>
  </div>
</template>

<script lang="ts" setup>
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

@media (max-width: 640px) {
  .nav-link {
    @apply text-sm py-1;
  }
}
</style>