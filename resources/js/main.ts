// import './bootstrap';

import { createApp } from "vue";

// import ExampleCounter from "./components/ExampleCounter.vue";
import router from './router';
import Image from "@/components/Image.vue";


import App from "./App.vue";

console.log('Image', Image)

const app = createApp(App);

// app.component("example-counter", ExampleCounter);
// app.component("AppView", AppView);

app.use(router);

app.mount("#app");
