import { createRouter, createWebHistory } from 'vue-router'
import { getToken } from './sdk/api'
import { useToast } from './composables/useToast'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/chat'
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('./pages/Login.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('./pages/Register.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/chat',
      name: 'chat',
      component: () => import('./pages/ChatRoom.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to, from, next) => {
  const token = getToken()
  const requiresAuth = to.meta.requiresAuth
  const requiresGuest = to.meta.requiresGuest
  const { warning, info } = useToast()

  if (requiresAuth && !token) {
    warning('请先登录')
    next('/login')
  } else if (requiresGuest && token) {
    info('您已登录')
    next('/chat')
  } else {
    next()
  }
})

export default router
