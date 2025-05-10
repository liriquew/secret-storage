import { createRouter, createWebHistory } from "vue-router";
import AuthPage from "@/pages/AuthPage.vue"
import KeysPage from "@/pages/KeysPage.vue";
import UnsealPage from "@/pages/UnsealPage.vue"
import MasterPage from "@/pages/MasterPage.vue";

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
    },
    {
        path: '/master',
        component: MasterPage, 
    }
];

const router = createRouter({
    history: createWebHistory(),
    routes: routes,
});

export default router;