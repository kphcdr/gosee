import type { RouteRecordRaw } from 'vue-router'

export const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    component: () => import('@/views/login/Index.vue'),
    meta: { title: '登录', public: true },
  },
  {
    path: '/',
    component: () => import('@/layouts/BasicLayout.vue'),
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', component: () => import('@/views/dashboard/Index.vue'), meta: { title: '仪表盘' } },
      { path: 'servers', component: () => import('@/views/server/List.vue'), meta: { title: '服务器列表' } },
      { path: 'servers/:id', component: () => import('@/views/server/Detail.vue'), meta: { title: '服务器详情', hideInMenu: true } },
      { path: 'server-groups', component: () => import('@/views/serverGroup/Index.vue'), meta: { title: '服务器分组' } },
      { path: 'alert-rules', component: () => import('@/views/alert/Rule.vue'), meta: { title: '告警规则' } },
      { path: 'alert-events', component: () => import('@/views/alert/Event.vue'), meta: { title: '告警事件' } },
      { path: 'notification-channels', component: () => import('@/views/notificationChannel/Index.vue'), meta: { title: '通知通道' } },
    ],
  },
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' },
]
