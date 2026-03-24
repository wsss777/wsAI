<template>
  <form class="auth-panel" @submit.prevent="submit">
    <header class="panel-header">
      <h2>邮箱注册</h2>
      <p>后端会生成随机用户名并通过邮件发送，注册成功后会直接返回令牌。</p>
    </header>

    <label class="field">
      <span>邮箱</span>
      <input v-model.trim="form.email" type="email" placeholder="请输入邮箱地址" required />
    </label>

    <label class="field">
      <span>验证码</span>
      <div class="inline-field">
        <input v-model.trim="form.captcha" type="text" placeholder="请输入验证码" required />
        <button
          class="secondary-button"
          type="button"
          :disabled="captchaLoading || countdown > 0"
          @click="emit('captcha', form.email)"
        >
          {{ captchaLoading ? '发送中...' : countdown > 0 ? `${countdown} 秒` : '发送验证码' }}
        </button>
      </div>
    </label>

    <label class="field">
      <span>密码</span>
      <input v-model="form.password" type="password" placeholder="请输入密码" required />
    </label>

    <button class="primary-button" type="submit" :disabled="loading">
      {{ loading ? '注册中...' : '注册并进入聊天' }}
    </button>
  </form>
</template>

<script setup>
import { reactive } from 'vue'

defineProps({
  loading: {
    type: Boolean,
    default: false
  },
  captchaLoading: {
    type: Boolean,
    default: false
  },
  countdown: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['submit', 'captcha'])

const form = reactive({
  email: '',
  captcha: '',
  password: ''
})

function submit() {
  emit('submit', { ...form })
}
</script>
