import { computed, ref } from 'vue'

const token = ref(localStorage.getItem('wsai-token') || '')
const username = ref(localStorage.getItem('wsai-username') || '未登录')
const apiBaseUrl = ref(localStorage.getItem('wsai-api-base-url') || import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:9091/api/v1')

export const authStore = {
  token,
  username,
  apiBaseUrl,
  isAuthenticated: computed(() => !!token.value),
  setToken(value) { token.value = value; localStorage.setItem('wsai-token', value) },
  setUser(value) { username.value = value; localStorage.setItem('wsai-username', value) },
  setApiBaseUrl(value) { apiBaseUrl.value = value.replace(/\/$/, ''); localStorage.setItem('wsai-api-base-url', apiBaseUrl.value) },
  clear() { token.value = ''; username.value = '未登录'; localStorage.removeItem('wsai-token'); localStorage.removeItem('wsai-username') }
}
