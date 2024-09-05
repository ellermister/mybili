// import './bootstrap';

import { createApp } from "vue";

import router from './router';

import App from "./App.vue";


const app = createApp(App);

// app.component("example-counter", ExampleCounter);
// app.component("AppView", AppView);

app.use(router);

app.mount("#app");
