import AuthPage from "@/pages/AuthPage.vue"
import KeysPage from "@/pages/KeysPage.vue";
import {createRouter, createWebHistory} from "vue-router";

const routes = [
    {
        path: '/',
        component: AuthPage,
    },
    {
        path: '/keys/:path*',
        component: KeysPage,
    },
    {
        path: '/keys',
        component: KeysPage,
    }
];

const router = createRouter({
    history: createWebHistory(),
    routes: routes,
});

export default router;