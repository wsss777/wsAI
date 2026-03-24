import { computed, ref } from 'vue'

const TOKEN_KEY = 'wsai-token'
const API_BASE_KEY = 'wsai-api-base'
const DEFAULT_API_BASE = import.meta.env.VITE_API_BASE_URL || 'http://127.0.0.1:9091/api/v1'

const token = ref(localStorage.getItem(TOKEN_KEY) || '')
const apiBaseUrl = ref(localStorage.getItem(API_BASE_KEY) || DEFAULT_API_BASE)
const username = ref(extractUsername(token.value))

function extractUsername(currentToken) {
  if (!currentToken) {
    return ''
  }

  try {
    const payload = currentToken.split('.')[1]
    const decoded = JSON.parse(atob(payload.replace(/-/g, '+').replace(/_/g, '/')))
    return decoded.username || decoded.Username || decoded.sub || ''
  } catch {
    return ''
  }
}

function setToken(nextToken) {
  token.value = nextToken
  username.value = extractUsername(nextToken)

  if (nextToken) {
    localStorage.setItem(TOKEN_KEY, nextToken)
  } else {
    localStorage.removeItem(TOKEN_KEY)
  }
}

function setApiBaseUrl(nextApiBase) {
  apiBaseUrl.value = nextApiBase.trim().replace(/\/$/, '')
  localStorage.setItem(API_BASE_KEY, apiBaseUrl.value)
}

export const authStore = {
  token,
  username,
  apiBaseUrl,
  isAuthenticated: computed(() => Boolean(token.value)),
  setToken,
  setApiBaseUrl,
  clear() {
    setToken('')
  }
}
