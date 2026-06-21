import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: () => import('../pages/Home.vue'),
    meta: { public: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../pages/LoginPage.vue'),
    meta: { public: true }
  },
  {
    path: '/app',
    component: () => import('../layouts/MainLayout.vue'),
    redirect: '/app/dashboard',
    meta: { requiresAuth: true },
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('../pages/DashboardPage.vue'),
        meta: { title: '数据看板', adminOnly: true }
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('../pages/OrdersPage.vue'),
        meta: { title: '异常订单列表', adminOnly: true }
      },
      {
        path: 'operations',
        name: 'Operations',
        component: () => import('../pages/OperationsPage.vue'),
        meta: { title: '处理订单' }
      },
      {
        path: 'my-handled',
        name: 'MyHandled',
        component: () => import('../pages/MyHandledPage.vue'),
        meta: { title: '我的处理记录' }
      },
      {
        path: 'quality-check',
        name: 'QualityCheck',
        component: () => import('../pages/QualityCheckPage.vue'),
        meta: { title: '质检审核', adminOnly: true, inspectorAllowed: true }
      },
      {
        path: 'quality-review',
        name: 'QualityReview',
        component: () => import('../pages/QualityReviewPage.vue'),
        meta: { title: '质检复查', adminOnly: true, inspectorAllowed: true }
      },
      {
        path: 'stats',
        name: 'StatsView',
        component: () => import('../pages/StatsPage.vue'),
        meta: { title: '操作员统计', statsOnly: true }
      },
      {
        path: 'users',
        name: 'Users',
        component: () => import('../pages/UsersPage.vue'),
        meta: { title: '用户管理', adminOnly: true }
      },
      {
        path: 'teams',
        name: 'Teams',
        component: () => import('../pages/TeamsPage.vue'),
        meta: { title: '团队管理', adminOnly: true }
      }
    ]
  },
  { path: '/:pathMatch(.*)*', redirect: '/login' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  const userStr = localStorage.getItem('user')
  const user = userStr ? JSON.parse(userStr) : null

  const isInspector = user?.role === 'inspector'
  const isOperator = user?.role === 'operator'
  const isStatistician = user?.role === 'statistician'

  if (to.meta.requiresAuth && !token) {
    next('/login')
  } else if (to.path === '/login' && token) {
    if (isOperator || isInspector) next('/app/my-handled')
    else if (isStatistician) next('/app/stats')
    else next('/app/dashboard')
  } else if (to.path === '/app' || to.path === '/app/dashboard') {
    if (isOperator || isInspector) next('/app/my-handled')
    else if (isStatistician) next('/app/stats')
    else next()
  } else if (to.meta.adminOnly && (isOperator || isStatistician)) {
    isStatistician ? next('/app/stats') : next('/app/my-handled')
  } else if (to.meta.adminOnly && isInspector && !to.meta.inspectorAllowed) {
    next('/app/my-handled')
  } else if (to.meta.statsOnly && isOperator) {
    next('/app/my-handled')
  } else {
    next()
  }
})

export default router
