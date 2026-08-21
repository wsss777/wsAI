<template>
  <main class="auth-page">
    <section class="auth-intro" aria-labelledby="auth-title">
      <RouterLink class="auth-brand" to="/auth"><i class="pulse" /> AI 应用服务平台</RouterLink>
      <div>
        <p class="auth-kicker">KNOWLEDGE · CONVERSATION · ACTION</p>
        <h1 id="auth-title">把资料变成<br /><em>可用的智能。</em></h1>
        <p class="auth-description">登录后即可管理知识档案，并与 AI 持续对话。</p>
      </div>
      <div class="auth-feature"><span>01</span><p>文档入库后，随时在对话中调用。</p></div>
    </section>
    <section class="auth-panel" aria-label="账户认证">
      <form class="auth-card" @submit.prevent="submit">
        <p class="form-eyebrow">账户中心</p>
        <h2>{{ mode === 'login' ? '欢迎回来' : '创建账户' }}</h2>
        <p class="form-hint">{{ mode === 'login' ? '选择一种方式登录服务平台。' : '通过邮箱验证码完成注册。' }}</p>
        <div class="auth-tabs" role="tablist" aria-label="认证方式">
          <button type="button" :class="{ active: mode === 'login' }" @click="switchMode('login')">登录</button>
          <button type="button" :class="{ active: mode === 'register' }" @click="switchMode('register')">注册</button>
        </div>
        <div v-if="mode === 'login'" class="login-methods" role="tablist" aria-label="登录方式">
          <button type="button" :class="{ active: loginMethod === 'username' }" @click="loginMethod = 'username'">用户名登录</button>
          <button type="button" :class="{ active: loginMethod === 'email' }" @click="loginMethod = 'email'">邮箱登录</button>
        </div>
        <label v-if="mode === 'login' && loginMethod === 'username'" class="auth-field"><span>用户名</span><input v-model.trim="form.username" autocomplete="username" placeholder="输入用户名" required /></label>
        <label v-else class="auth-field"><span>邮箱</span><input v-model.trim="form.email" type="email" autocomplete="email" placeholder="name@example.com" required /></label>
        <label class="auth-field"><span>密码</span><input v-model="form.password" type="password" :autocomplete="mode === 'login' ? 'current-password' : 'new-password'" minlength="6" placeholder="至少 6 位" required /></label>
        <label v-if="mode === 'register'" class="auth-field"><span>邮箱验证码</span><div class="captcha-row"><input v-model.trim="form.captcha" inputmode="numeric" autocomplete="one-time-code" placeholder="输入验证码" required /><button class="captcha-button" type="button" :disabled="captchaLoading || countdown > 0" @click="requestCaptcha">{{ captchaLoading ? '发送中…' : countdown ? `${countdown}s 后重试` : '获取验证码' }}</button></div></label>
        <p v-if="message" class="auth-message" :class="messageType" role="alert">{{ message }}</p>
        <button class="auth-submit" :disabled="loading" type="submit">{{ loading ? '请稍候…' : mode === 'login' ? '登录并进入平台' : '注册并进入平台' }}</button>
      </form>
    </section>
  </main>
</template>

<script setup>
import { onBeforeUnmount, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { login, loginWithEmail, register, sendCaptcha } from '../services/api'
import { authStore } from '../stores/auth'

const router = useRouter()
const mode = ref('login'), loginMethod = ref('username'), loading = ref(false), captchaLoading = ref(false), countdown = ref(0), message = ref(''), messageType = ref('error')
const form = reactive({ username: '', email: '', password: '', captcha: '' })
let timer
const notify = (text, type = 'error') => { message.value = text; messageType.value = type }
const switchMode = nextMode => { mode.value = nextMode; message.value = '' }
const normalizeError = error => error.message === 'Failed to fetch' ? '无法连接后端服务，请检查接口地址和服务状态。' : error.message

function finish(identity, token) { authStore.setUser(identity); authStore.setToken(token); router.push('/chat') }
async function submit() {
  loading.value = true; message.value = ''
  try {
    let result, identity
    if (mode.value === 'register') { result = await register({ email: form.email, password: form.password, captcha: form.captcha }); identity = form.email }
    else if (loginMethod.value === 'email') { result = await loginWithEmail({ email: form.email, password: form.password }); identity = form.email }
    else { result = await login({ username: form.username, password: form.password }); identity = form.username }
    finish(identity, result.token)
  } catch (error) { notify(normalizeError(error)) } finally { loading.value = false }
}
async function requestCaptcha() {
  if (!form.email) return notify('请先输入邮箱地址。')
  captchaLoading.value = true; message.value = ''
  try {
    await sendCaptcha(form.email); notify('验证码已发送，请检查邮箱。', 'success'); countdown.value = 60
    timer = window.setInterval(() => { countdown.value -= 1; if (countdown.value <= 0) { window.clearInterval(timer); timer = undefined } }, 1000)
  } catch (error) { notify(normalizeError(error)) } finally { captchaLoading.value = false }
}
onBeforeUnmount(() => { if (timer) window.clearInterval(timer) })
</script>

<style scoped>
.auth-card { width: min(390px, 100%); padding: 8px; border: 0; background: transparent; }
.auth-tabs { display: grid; grid-template-columns: 1fr 1fr; margin-bottom: 18px; border: 1px solid #ced6df; }
.auth-tabs button { padding: 12px; border: 0; border-radius: 0; background: transparent; color: #69758a; font-size: 13px; font-weight: 800; }
.auth-tabs button.active { background: #1769ff; color: #fff; }
.login-methods button { background: transparent; color: #69758a; }
.login-methods button.active { background: #fff; color: #182033; }
.auth-submit { border-radius: 0; }
</style>
