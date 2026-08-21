import { createRouter, createWebHistory } from 'vue-router'
import AuthView from '../views/AuthView.vue'
import DocumentsView from '../views/DocumentsView.vue'
import ChatView from '../views/ChatView.vue'
import { authStore } from '../stores/auth'
const router=createRouter({history:createWebHistory(),routes:[{path:'/',redirect:'/chat'},{path:'/auth',component:AuthView},{path:'/dashboard',redirect:'/chat'},{path:'/documents',component:DocumentsView,meta:{auth:true}},{path:'/knowledge',redirect:'/documents'},{path:'/chat',component:ChatView,meta:{auth:true}}]})
router.beforeEach(to => to.path !== '/auth' && !authStore.isAuthenticated.value ? '/auth' : true)
export default router
