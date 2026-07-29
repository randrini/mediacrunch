import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
    },
    {
      path: '/instances',
      name: 'instances',
      component: () => import('../views/InstancesView.vue'),
    },
    {
      path: '/instances/:id/media',
      name: 'media',
      component: () => import('../views/MediaView.vue'),
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue'),
    },
    {
      path: '/instances/:id/settings',
      name: 'instance-settings',
      component: () => import('../views/SettingsView.vue'),
    },
    {
      path: '/logs',
      name: 'logs',
      component: () => import('../views/LogsView.vue'),
    },
  ],
})

export default router
