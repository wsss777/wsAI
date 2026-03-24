<template>
  <form class="auth-panel" @submit.prevent="submit">
    <header class="panel-header">
      <h2>邮箱登录</h2>
      <p>使用注册账号时填写的邮箱和密码直接登录。</p>
    </header>

    <label class="field">
      <span>邮箱</span>
      <input v-model.trim="form.email" type="email" placeholder="请输入注册邮箱" required />
    </label>

    <label class="field">
      <span>密码</span>
      <input v-model="form.password" type="password" placeholder="请输入密码" required />
    </label>

    <button class="primary-button" type="submit" :disabled="loading">
      {{ loading ? '登录中...' : '邮箱登录并进入聊天' }}
    </button>
  </form>
</template>

<script setup>
import { reactive } from 'vue'

defineProps({
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['submit'])

const form = reactive({
  email: '',
  password: ''
})

function submit() {
  emit('submit', { ...form })
}
</script>
