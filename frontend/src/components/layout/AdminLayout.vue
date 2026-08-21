<template>
  <div class="library-shell">
    <aside class="library-sidebar">
      <RouterLink class="library-brand" to="/chat"><i class="pulse" /> AI 应用服务平台</RouterLink>
      <nav><RouterLink to="/chat">✦ 智能会话</RouterLink><RouterLink to="/documents">▤ 管理文档</RouterLink></nav>
      <button class="account-card" @click="logout"><i>{{ initials }}</i><span><b>{{ authStore.username.value }}</b><small>退出登录</small></span><em>···</em></button>
    </aside>
    <main class="library-workspace"><slot /></main>
  </div>
</template>
<script setup>
import { computed } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { authStore } from '../../stores/auth'
const router = useRouter()
const initials = computed(() => authStore.username.value.slice(0, 2).toUpperCase())
function logout() { authStore.clear(); router.push('/auth') }
</script>
<style scoped>
.library-shell{display:grid;grid-template-columns:268px minmax(0,1fr);min-height:100vh;background:#fff;color:#1b2230}.library-sidebar{display:flex;flex-direction:column;padding:22px 14px;background:#f8f9fa;border-right:1px solid #e2e5e9}.library-brand{display:flex;align-items:center;gap:9px;margin:5px 10px 28px;color:#182033;font:700 19px/1 "Playfair Display",serif;text-decoration:none}.library-brand .pulse{width:15px;height:15px}.library-sidebar nav{display:grid;gap:3px}.library-sidebar nav a{padding:11px 12px;border-radius:8px;color:#313b4b;text-decoration:none;font-size:13px;font-weight:700}.library-sidebar nav a:hover,.library-sidebar nav a.router-link-active{background:#e7ebf1}.account-card{display:grid;grid-template-columns:33px minmax(0,1fr) auto;align-items:center;gap:9px;margin-top:auto;padding:10px;border:0;border-radius:9px;background:transparent;color:#1b2230;text-align:left}.account-card:hover{background:#e7ebf1}.account-card i{display:grid;place-items:center;width:31px;height:31px;border-radius:50%;background:#d9dfc2;color:#526000;font-style:normal;font-size:10px;font-weight:800}.account-card span{display:grid;gap:2px;min-width:0}.account-card b{overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-size:12px}.account-card small{color:#7d8898;font-size:10px}.account-card em{font-style:normal;font-weight:700}.library-workspace{min-width:0}@media(max-width:760px){.library-shell{grid-template-columns:1fr}.library-sidebar{display:none}}
</style>
