import { createRouter, createWebHistory } from "vue-router";
import AuthPage from "@/pages/AuthPage.vue"
import KeysPage from "@/pages/KeysPage.vue";
import UnsealPage from "@/pages/UnsealPage.vue"

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
    },
    {
        path: '/unseal',
        component: UnsealPage,
    }
];

const router = createRouter({
    history: createWebHistory(),
    routes: routes,
});

export default router;