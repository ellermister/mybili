<template>
  <div class="w-auto">
    <div class="container  mx-auto justify-center px-4 md:px-16 ">
      <div class="mx-4">
        <nav class="bg-gradient-to-r from-purple-500 to-pink-500 rounded-b-lg  text-white w-full text-base ">
          <router-link to="/">Home</router-link>
          <router-link to="/progress">Progress</router-link>
          <router-link to="/cookie">Cookie</router-link>
          <router-link to="/about">About</router-link>
        </nav>
      </div>

      <div class="mx-4">
        <router-view></router-view>
      </div>

    </div>




  </div>
</template>

<script lang="ts" setup>
</script>

<style scoped>
/* 添加一些基本样式 */
nav {
  padding: 1rem;
}

nav a {
  margin-right: 10px;
  text-decoration: none;
}
</style>