import { createRouter, createWebHistory } from 'vue-router';
import Home from '@/pages/Home.vue';
import About from '@/pages/About.vue';
import FavlistView from './pages/FavlistView.vue';
import VideoView from './pages/VideoView.vue';
import ProgressView from './pages/ProgressView.vue';
import CookieView from './pages/CookieView.vue';

const routes = [
  { path: '/', component: Home },
  { path: '/fav/:id', component: FavlistView, name:'favlist-id' },
  { path: '/video/:id', component: VideoView, name:'video-id' },
  { path: '/about', component: About },
  { path: '/progress', component: ProgressView },
  { path: '/cookie', component: CookieView },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;