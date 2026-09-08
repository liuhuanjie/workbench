import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import MainLayout from '../layout/MainLayout.vue'

const routes = [
  { path: '/login', component: () => import('../views/Login.vue') },
  {
    path: '/',
    component: MainLayout,
    children: [
      { path: '', name: 'home', component: () => import('../views/Home.vue') },
      { path: 'news', name: 'news', component: () => import('../views/News.vue') },
      { path: 'debt', name: 'debt', component: () => import('../views/Debt.vue') },
      { path: 'house', name: 'house', component: () => import('../views/House.vue') },
      { path: 'favorite', name: 'favorite', component: () => import('../views/Favorite.vue') },
      { path: 'stats', name: 'stats', component: () => import('../views/Stats.vue') },
      { path: 'data', name: 'data', component: () => import('../views/DataManage.vue') },
      { path: 'settings', name: 'settings', component: () => import('../views/Settings.vue') },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// 路由守卫：未登录跳登录；强制改密期间仅可停留在设置页
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.path !== '/login' && !auth.token) return '/login'
  if (to.path === '/login' && auth.token) return '/'
  if (auth.mustChange && to.path !== '/settings') return '/settings'
})

export default router
