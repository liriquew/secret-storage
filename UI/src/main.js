import { createApp } from 'vue'
import App from './App.vue'
import router from "@/router/router.js";

const config = {
    "api_url": "http://localhost:8080",
};
export default config

const app = createApp(App);

app
    .use(router)
    .mount('#app');
