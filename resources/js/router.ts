import { createRouter, createWebHistory } from 'vue-router';
import Home from '@/pages/Home.vue';
import About from '@/pages/About.vue';
import FavlistView from './pages/FavlistView.vue';
import VideoView from './pages/VideoView.vue';
import ProgressView from './pages/ProgressView.vue';
import CookieView from './pages/CookieView.vue';
import SettingsView from './pages/SettingsView.vue';
import SubscriptionView from './pages/Subscription.vue';
import SubscriptionVideoView from './pages/SubscriptionVideo.vue';

const routes = [
  { path: '/', component: Home },
  { path: '/fav/:id', component: FavlistView, name:'favlist-id' },
  { path: '/fav/:id/video/:video_id', component: VideoView, name:'favlist-video-id' },
  { path: '/video/:id', component: VideoView, name:'video-id' },
  { path: '/about', component: About },
  { path: '/progress', component: ProgressView, name: 'progress' },
  { path: '/cookie', component: CookieView },
  { path: '/settings', component: SettingsView },
  { path: '/subscription', component: SubscriptionView, name: 'subscription' },
  { path: '/subscription/:id', component: SubscriptionVideoView, name: 'subscription-id' },
  { path: '/subscription/:id/video/:video_id', component: VideoView, name: 'subscription-video-id' },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;