import { createRouter, createWebHistory } from 'vue-router'
import AuthView from '../views/AuthView.vue'
import ChatView from '../views/ChatView.vue'
import { authStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: () => (authStore.isAuthenticated.value ? '/chat' : '/auth')
    },
    {
      path: '/auth',
      name: 'auth',
      component: AuthView,
      meta: { guestOnly: true }
    },
    {
      path: '/chat',
      name: 'chat',
      component: ChatView,
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !authStore.isAuthenticated.value) {
    return '/auth'
  }

  if (to.meta.guestOnly && authStore.isAuthenticated.value) {
    return '/chat'
  }

  return true
})

export default router
