<template>
  <main class="chat-layout">
    <StatusBanner :message="notice.message" :variant="notice.variant" @close="clearNotice" />

    <SessionSidebar
      :username="authStore.username.value"
      :api-base-url="authStore.apiBaseUrl.value"
      :sessions="sessions"
      :active-session-id="activeSessionId"
      @select="selectSession"
      @new-session="startNewSession"
      @logout="handleLogout"
      @update-api-base="authStore.setApiBaseUrl($event)"
    />

    <section class="chat-main">
      <MessageList :messages="messages" />
      <ComposerBox :disabled="isStreaming" :active-session-id="activeSessionId" @submit="sendMessage" />
    </section>

    <aside class="utility-rail">
      <ImageRecognizerPanel @error="showNotice($event, 'error')" @success="showNotice($event, 'success')" />
    </aside>
  </main>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import ComposerBox from '../components/chat/ComposerBox.vue'
import ImageRecognizerPanel from '../components/chat/ImageRecognizerPanel.vue'
import MessageList from '../components/chat/MessageList.vue'
import SessionSidebar from '../components/chat/SessionSidebar.vue'
import StatusBanner from '../components/common/StatusBanner.vue'
import { fetchHistory, fetchSessions, logout, streamSessionMessage } from '../services/api'
import { authStore } from '../stores/auth'

const router = useRouter()
const sessions = ref([])
const activeSessionId = ref('')
const messages = ref([])
const modelType = ref('openai')
const isStreaming = ref(false)
const notice = reactive({ message: '', variant: 'info' })

function showNotice(message, variant = 'info') {
  notice.message = message
  notice.variant = variant
}

function clearNotice() {
  notice.message = ''
}

async function loadSessions() {
  try {
    const data = await fetchSessions()
    sessions.value = data.sessions || []
  } catch (error) {
    showNotice(error.message, 'error')
  }
}

async function selectSession(sessionId) {
  activeSessionId.value = sessionId
  clearNotice()

  try {
    const data = await fetchHistory(sessionId)
    messages.value = (data.history || []).map((item) => ({
      role: item.is_user ? 'user' : 'assistant',
      content: item.content
    }))
  } catch (error) {
    messages.value = []
    showNotice(error.message, 'error')
  }
}

function startNewSession() {
  activeSessionId.value = ''
  messages.value = []
  clearNotice()
}

function insertSessionToTop(sessionId, title) {
  const existing = sessions.value.find((item) => item.sessionId === sessionId)
  const updatedAt = new Date().toISOString()

  sessions.value = [
    {
      sessionId,
      title: title.length > 24 ? `${title.slice(0, 24)}...` : title,
      updatedAt: existing?.updatedAt || updatedAt
    },
    ...sessions.value.filter((item) => item.sessionId !== sessionId)
  ]
}

function moveCurrentSessionToTop() {
  if (!activeSessionId.value) {
    return
  }

  const current = sessions.value.find((item) => item.sessionId === activeSessionId.value)
  if (!current) {
    return
  }

  sessions.value = [
    {
      ...current,
      updatedAt: new Date().toISOString()
    },
    ...sessions.value.filter((item) => item.sessionId !== activeSessionId.value)
  ]
}

async function sendMessage(question) {
  clearNotice()
  isStreaming.value = true

  if (activeSessionId.value) {
    moveCurrentSessionToTop()
  }

  const nextMessages = [...messages.value, { role: 'user', content: question }, { role: 'assistant', content: '' }]
  messages.value = nextMessages
  const assistantIndex = nextMessages.length - 1
  const pendingSessionTitle = question

  try {
    await streamSessionMessage(
      {
        sessionId: activeSessionId.value,
        question,
        modelType: modelType.value
      },
      {
        onSession: async (sessionId) => {
          activeSessionId.value = sessionId
          insertSessionToTop(sessionId, pendingSessionTitle)
          await loadSessions()
        },
        onChunk: (chunk) => {
          messages.value[assistantIndex].content += chunk
        },
        onDone: async () => {
          await loadSessions()
        }
      }
    )
  } catch (error) {
    messages.value[assistantIndex].content = messages.value[assistantIndex].content || '[流式回复失败]'
    showNotice(error.message, 'error')
  } finally {
    isStreaming.value = false
  }
}

async function handleLogout() {
  try {
    await logout()
  } catch {
  } finally {
    authStore.clear()
    router.push('/auth')
  }
}

onMounted(async () => {
  await loadSessions()
  if (sessions.value.length > 0) {
    await selectSession(sessions.value[0].sessionId)
  }
})
</script>
