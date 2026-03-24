<template>
  <main class="auth-layout">
    <StatusBanner :message="notice.message" :variant="notice.variant" @close="clearNotice" />

    <AuthHero />

    <section class="auth-card">
      <div class="api-panel">
        <label class="field">
          <span>后端接口地址</span>
          <input
            :value="authStore.apiBaseUrl.value"
            type="text"
            placeholder="http://127.0.0.1:9091/api/v1"
            @change="authStore.setApiBaseUrl($event.target.value)"
          />
        </label>
      </div>

      <AuthTabs v-model="mode" :items="tabItems" />

      <template v-if="mode === 'login'">
        <AuthTabs v-model="loginMethod" :items="loginMethodItems" />
        <LoginForm v-if="loginMethod === 'username'" :loading="loading" @submit="handleLoginByUsername" />
        <EmailLoginForm v-else :loading="loading" @submit="handleLoginByEmail" />
      </template>

      <RegisterForm
        v-else
        :loading="loading"
        :captcha-loading="captchaLoading"
        :countdown="countdown"
        @submit="handleRegister"
        @captcha="handleCaptcha"
      />
    </section>
  </main>
</template>

<script setup>
import { onBeforeUnmount, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import AuthHero from '../components/auth/AuthHero.vue'
import AuthTabs from '../components/auth/AuthTabs.vue'
import EmailLoginForm from '../components/auth/EmailLoginForm.vue'
import LoginForm from '../components/auth/LoginForm.vue'
import RegisterForm from '../components/auth/RegisterForm.vue'
import StatusBanner from '../components/common/StatusBanner.vue'
import { login, loginWithEmail, register, sendCaptcha } from '../services/api'
import { authStore } from '../stores/auth'

const router = useRouter()
const mode = ref('login')
const loginMethod = ref('username')
const loading = ref(false)
const captchaLoading = ref(false)
const countdown = ref(0)
const notice = reactive({ message: '', variant: 'info' })

const tabItems = [
  { label: '登录', value: 'login' },
  { label: '注册', value: 'register' }
]

const loginMethodItems = [
  { label: '用户名登录', value: 'username' },
  { label: '邮箱登录', value: 'email' }
]

let countdownTimer = null

function showNotice(message, variant = 'info') {
  notice.message = message
  notice.variant = variant
}

function clearNotice() {
  notice.message = ''
}

function applyToken(token, successText) {
  authStore.setToken(token)
  showNotice(successText, 'success')
  router.push('/chat')
}

async function handleLoginByUsername(payload) {
  loading.value = true
  clearNotice()

  try {
    const data = await login(payload)
    applyToken(data.token, '用户名登录成功，正在进入聊天页面。')
  } catch (error) {
    showNotice(error.message, 'error')
  } finally {
    loading.value = false
  }
}

async function handleLoginByEmail(payload) {
  loading.value = true
  clearNotice()

  try {
    const data = await loginWithEmail(payload)
    applyToken(data.token, '邮箱登录成功，正在进入聊天页面。')
  } catch (error) {
    showNotice(error.message, 'error')
  } finally {
    loading.value = false
  }
}

async function handleRegister(payload) {
  loading.value = true
  clearNotice()

  try {
    const data = await register(payload)
    applyToken(data.token, '注册成功，系统生成的用户名会发送到你的邮箱。')
  } catch (error) {
    showNotice(error.message, 'error')
  } finally {
    loading.value = false
  }
}

function startCountdown() {
  countdown.value = 60
  countdownTimer = window.setInterval(() => {
    countdown.value -= 1
    if (countdown.value <= 0) {
      window.clearInterval(countdownTimer)
      countdownTimer = null
    }
  }, 1000)
}

async function handleCaptcha(email) {
  if (!email) {
    showNotice('请先输入邮箱地址。', 'error')
    return
  }

  captchaLoading.value = true
  clearNotice()

  try {
    await sendCaptcha(email)
    showNotice('验证码已发送，请检查邮箱。', 'success')
    startCountdown()
  } catch (error) {
    showNotice(error.message, 'error')
  } finally {
    captchaLoading.value = false
  }
}

onBeforeUnmount(() => {
  if (countdownTimer) {
    window.clearInterval(countdownTimer)
  }
})
</script>
