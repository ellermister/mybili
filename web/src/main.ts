// import './bootstrap';

import { createApp } from "vue";
import { createI18n } from 'vue-i18n'
import language from './language'
const i18n = createI18n(language)
import router from './router';

import App from "./App.vue";

import './css/app.css';
import './css/button.css';
import './css/dark.css';

const app = createApp(App);
app.use(router);
app.use(i18n)
app.mount("#app");
