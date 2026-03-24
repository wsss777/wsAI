<template>
  <form class="auth-panel" @submit.prevent="submit">
    <header class="panel-header">
      <h2>账号登录</h2>
      <p>请输入注册邮件中收到的用户名与密码。</p>
    </header>

    <label class="field">
      <span>用户名</span>
      <input v-model.trim="form.username" type="text" placeholder="请输入用户名" required />
    </label>

    <label class="field">
      <span>密码</span>
      <input v-model="form.password" type="password" placeholder="请输入密码" required />
    </label>

    <button class="primary-button" type="submit" :disabled="loading">
      {{ loading ? '登录中...' : '登录并进入聊天' }}
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
  username: '',
  password: ''
})

function submit() {
  emit('submit', { ...form })
}
</script>
