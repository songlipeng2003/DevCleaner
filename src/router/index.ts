import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue')
  },
  {
    path: '/scan',
    name: 'scan',
    component: () => import('@/views/ScanView.vue')
  },
  {
    path: '/settings',
    name: 'settings',
    component: () => import('@/views/SettingsView.vue')
  },
  // v0.2.0 新增路由
  {
    path: '/projects',
    name: 'projects',
    component: () => import('@/views/ProjectCleanView.vue')
  },
  {
    path: '/history',
    name: 'history',
    component: () => import('@/views/HistoryView.vue')
  },
  // v0.3.0 新增路由
  {
    path: '/analysis',
    name: 'analysis',
    component: () => import('@/views/DiskAnalysisView.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
