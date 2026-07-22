import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('@/layouts/MainLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: '仪表板' }
      },
      {
        path: 'products',
        name: 'Products',
        component: () => import('@/views/Products.vue'),
        meta: { title: '商品管理', permission: 'product:view' }
      },
      {
        path: 'product-types',
        name: 'ProductTypes',
        component: () => import('@/views/ProductTypes.vue'),
        meta: { title: '产品模板', permission: 'product:view' }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('@/views/Orders.vue'),
        meta: { title: '订单管理', permission: 'order:view' }
      },
      {
        path: 'shipping',
        name: 'Shipping',
        component: () => import('@/views/Shipping.vue'),
        meta: { title: '物流管理', permission: 'shipping:view' }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('@/views/Users.vue'),
        meta: { title: '用户管理', permission: 'user:view' }
      },
      {
        path: 'content',
        name: 'Content',
        component: () => import('@/views/Content.vue'),
        meta: { title: '内容管理', permission: 'content:view' }
      },
      {
        path: 'faqs',
        name: 'FAQs',
        component: () => import('@/views/FAQs.vue'),
        meta: { title: 'FAQ管理', permission: 'faq:view' }
      },
      {
        path: 'galleries',
        name: 'Galleries',
        component: () => import('@/views/Galleries.vue'),
        meta: { title: '图库管理', permission: 'gallery:view' }
      },
      {
        path: 'subscriptions',
        name: 'Subscriptions',
        component: () => import('@/views/Subscriptions.vue'),
        meta: { title: '订阅管理', permission: 'subscription:view' }
      },
      {
        path: 'tickets',
        name: 'Tickets',
        component: () => import('@/views/Tickets.vue'),
        meta: { title: '工单管理', permission: 'ticket:view' }
      },
      {
        path: 'marketing',
        name: 'Marketing',
        component: () => import('@/views/Marketing.vue'),
        meta: { title: '营销管理', permission: 'marketing:view' }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: () => import('@/views/Settings.vue'),
        meta: { title: '系统设置', permission: 'settings:view' }
      },
      {
        path: 'audit-logs',
        name: 'AuditLogs',
        component: () => import('@/views/AuditLogs.vue'),
        meta: { title: '审计日志', permission: 'logs:view' }
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  const requiresAuth = to.meta.requiresAuth !== false

  if (requiresAuth && !authStore.initialized) {
    await authStore.initAuth()
  }
  
  if (requiresAuth && !authStore.isAuthenticated) {
    // 需要认证但未登录，跳转到登录页
    next({ name: 'Login', query: { redirect: to.fullPath } })
  } else if (to.name === 'Login' && authStore.isAuthenticated) {
    // 已登录访问登录页，跳转到首页
    next({ name: 'Dashboard' })
  } else if (to.meta.permission && !authStore.hasPermission(to.meta.permission)) {
    // 没有权限
    next({ name: 'Dashboard' })
  } else {
    next()
  }
})

export default router
