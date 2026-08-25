import { authStore } from '../stores/auth'
async function request(path, options = {}, retried = false) { const response = await fetch(`${authStore.apiBaseUrl.value}${path}`, { ...options, credentials: 'include', headers: { ...(options.body instanceof FormData ? {} : { 'Content-Type': 'application/json' }), ...(authStore.token.value ? { Authorization: `Bearer ${authStore.token.value}` } : {}), ...options.headers } }); const data = await response.json().catch(() => ({})); if (!retried && path !== '/user/refresh' && data.status_code === 2006) { const refreshed = await fetch(`${authStore.apiBaseUrl.value}/user/refresh`, { method: 'POST', credentials: 'include' }); const refreshData = await refreshed.json().catch(() => ({})); if (refreshed.ok && refreshData.status_code === 1000 && refreshData.token) { authStore.setToken(refreshData.token); return request(path, options, true) } authStore.clear() } if (!response.ok || data.status_code && data.status_code !== 1000) throw new Error(data.status_msg || `HTTP ${response.status}`); return data }
export const login = payload => request('/user/login', { method: 'POST', body: JSON.stringify(payload) })
export const loginWithEmail = payload => request('/user/email-login', { method: 'POST', body: JSON.stringify(payload) })
export const register = payload => request('/user/users', { method: 'POST', body: JSON.stringify(payload) })
export const sendCaptcha = email => request('/user/captcha', { method: 'POST', body: JSON.stringify({ email }) })
export const logout = () => request('/user/logout', { method: 'POST' })
export const refresh = () => request('/user/refresh', { method: 'POST' })
export const fetchDocuments = () => request('/rag/documents')
export const fetchDocument = id => request(`/rag/documents/${id}`)
export const deleteDocument = id => request(`/rag/documents/${id}`, { method: 'DELETE' })
export const uploadDocument = file => { const data = new FormData(); data.append('file', file); return request('/rag/documents', { method: 'POST', body: data }) }
export const bindDocuments = (sessionId, documentIds) => request(`/rag/sessions/${sessionId}/documents`, { method: 'POST', body: JSON.stringify({ document_ids: documentIds }) })
export const fetchSessionDocuments = sessionId => request(`/rag/sessions/${sessionId}/documents`)
export const unbindDocument = (sessionId, documentId) => request(`/rag/sessions/${sessionId}/documents/${documentId}`, { method: 'DELETE' })
export const recognizeImage = file => { const data = new FormData(); data.append('image', file); return request('/image/recognize', { method: 'POST', body: data }) }
export const fetchSessions = () => request('/AI/chatMessage/sessions')
export const fetchHistory = id => request(`/AI/chatMessage/sessions/${id}/messages`)
let selectedChatModelType = 'openai'
export function setChatModelType(modelType) { selectedChatModelType = modelType === 'zhipu' ? 'zhipu' : 'openai' }
export async function streamSessionMessage({ sessionId, question, documentIds = [], modelType = selectedChatModelType }, handlers = {}) {
	const path = sessionId ? `/AI/chatMessage/sessions/${sessionId}/messages/stream` : '/AI/chatMessage/sessions/stream'
	const body = sessionId ? { sessionId, question, modelType } : { question, modelType, document_ids: documentIds }
  const response = await fetch(`${authStore.apiBaseUrl.value}${path}`, { method: 'POST', headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${authStore.token.value}` }, body: JSON.stringify(body) })
  if (!response.ok || !response.body) throw new Error(`HTTP ${response.status}`)
  const reader = response.body.getReader(), decoder = new TextDecoder(); let buffer = ''
  while (true) {
    const { value, done } = await reader.read()
    if (done) break
    buffer += decoder.decode(value, { stream: true })
    const blocks = buffer.split('\n\n')
    buffer = blocks.pop() || ''
    for (const block of blocks) {
      const lines = block.split('\n')
      const event = lines.find(line => line.startsWith('event:'))?.replace(/^event:\s?/, '') || 'message'
      const data = lines.find(line => line.startsWith('data:'))?.replace(/^data:\s?/, '')
      if (!data || data === '[DONE]') continue

      let parsed
      try { parsed = JSON.parse(data) } catch {}
      if (event === 'citations') {
        if (Array.isArray(parsed)) handlers.onCitations?.(parsed)
        continue
      }
      if (parsed?.sessionId) { handlers.onSession?.(parsed.sessionId); continue }
      if (event === 'error') throw new Error(parsed?.message || 'Stream request failed')
      handlers.onChunk?.(data)
    }
  }
  handlers.onDone?.()
}
