// import './bootstrap';

import { createApp } from "vue";
import { createI18n } from 'vue-i18n'
import language from './language'
const i18n = createI18n(language)
import router from './router';

import App from "./App.vue";

import '../css/button.css';

const app = createApp(App);

// app.component("example-counter", ExampleCounter);
// app.component("AppView", AppView);

app.use(router);
app.use(i18n)
app.mount("#app");
