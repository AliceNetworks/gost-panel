import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/Login.vue'),
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('../views/Register.vue'),
    },
    {
      path: '/verify-email',
      name: 'verify-email',
      component: () => import('../views/VerifyEmail.vue'),
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('../views/ForgotPassword.vue'),
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('../views/ResetPassword.vue'),
    },
    {
      path: '/',
      component: () => import('../views/Layout.vue'),
      children: [
        {
          path: '',
          name: 'dashboard',
          component: () => import('../views/Dashboard.vue'),
        },
        {
          path: 'nodes',
          name: 'nodes',
          component: () => import('../views/Nodes.vue'),
        },
        {
          path: 'clients',
          name: 'clients',
          component: () => import('../views/Clients.vue'),
        },
        {
          path: 'users',
          name: 'users',
          component: () => import('../views/Users.vue'),
        },
        {
          path: 'notify',
          name: 'notify',
          component: () => import('../views/Notify.vue'),
        },
        {
          path: 'port-forwards',
          name: 'port-forwards',
          component: () => import('../views/PortForwards.vue'),
        },
        {
          path: 'node-groups',
          name: 'node-groups',
          component: () => import('../views/NodeGroups.vue'),
        },
        {
          path: 'proxy-chains',
          name: 'proxy-chains',
          component: () => import('../views/ProxyChains.vue'),
        },
        {
          path: 'tunnels',
          name: 'tunnels',
          component: () => import('../views/Tunnels.vue'),
        },
        {
          path: 'settings',
          name: 'settings',
          component: () => import('../views/Settings.vue'),
        },
        {
          path: 'operation-logs',
          name: 'operation-logs',
          component: () => import('../views/OperationLogs.vue'),
        },
        {
          path: 'change-password',
          name: 'change-password',
          component: () => import('../views/ChangePassword.vue'),
        },
      ],
    },
  ],
})

// 公开页面（不需要登录）
const publicPages = ['login', 'register', 'verify-email', 'forgot-password', 'reset-password']

// 路由守卫
router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  const isPublicPage = publicPages.includes(to.name as string)

  if (!isPublicPage && !token) {
    next({ name: 'login' })
  } else {
    next()
  }
})

export default router
