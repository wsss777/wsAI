import { authStore } from '../stores/auth'

function buildAuthHeaders(extraHeaders = {}, withJson = true) {
  const headers = {
    ...extraHeaders
  }

  if (withJson && !headers['Content-Type']) {
    headers['Content-Type'] = 'application/json'
  }

  if (authStore.token.value) {
    headers.Authorization = `Bearer ${authStore.token.value}`
  }

  return headers
}

async function parseJsonResponse(response) {
  const text = await response.text()
  return text ? JSON.parse(text) : {}
}

async function request(path, options = {}) {
  const response = await fetch(`${authStore.apiBaseUrl.value}${path}`, {
    ...options,
    headers: buildAuthHeaders(options.headers)
  })

  const data = await parseJsonResponse(response)

  if (!response.ok) {
    throw new Error(data?.status_msg || `HTTP ${response.status}`)
  }

  if (data.status_code && data.status_code !== 1000) {
    if (data.status_code === 2006 || data.status_code === 2007) {
      authStore.clear()
    }
    throw new Error(data.status_msg || '请求失败')
  }

  return data
}

function parseSseEvent(block) {
  const lines = block.split('\n')
  const event = { type: 'message', data: '' }
  let hasData = false

  for (const line of lines) {
    if (line.startsWith('event:')) {
      event.type = line.slice(6).trim()
    }
    if (line.startsWith('data:')) {
      const value = line.startsWith('data: ') ? line.slice(6) : line.slice(5)
      if (hasData) {
        event.data += '\n'
      }
      event.data += value
      hasData = true
    }
  }

  return event
}

export async function login(payload) {
  return request('/user/login', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export async function loginWithEmail(payload) {
  return request('/user/email-login', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export async function register(payload) {
  return request('/user/users', {
    method: 'POST',
    body: JSON.stringify(payload)
  })
}

export async function sendCaptcha(email) {
  return request('/user/captcha', {
    method: 'POST',
    body: JSON.stringify({ email })
  })
}

export async function logout() {
  return request('/user/logout', {
    method: 'POST'
  })
}

export async function fetchSessions() {
  return request('/AI/chatMessage/sessions', {
    method: 'GET'
  })
}

export async function fetchHistory(sessionId) {
  return request(`/AI/chatMessage/sessions/${sessionId}/messages`, {
    method: 'GET'
  })
}

export async function recognizeImage(file) {
  const formData = new FormData()
  formData.append('image', file)

  const response = await fetch(`${authStore.apiBaseUrl.value}/image/recognize`, {
    method: 'POST',
    headers: buildAuthHeaders({}, false),
    body: formData
  })

  const data = await parseJsonResponse(response)

  if (!response.ok) {
    throw new Error(data?.status_msg || `HTTP ${response.status}`)
  }

  if (data.status_code && data.status_code !== 1000) {
    throw new Error(data.status_msg || '图片识别失败')
  }

  return data
}

export async function streamSessionMessage({ sessionId, question, modelType }, handlers) {
  const isNewSession = !sessionId
  const path = isNewSession
    ? '/AI/chatMessage/sessions/stream'
    : `/AI/chatMessage/sessions/${sessionId}/messages/stream`

  const body = isNewSession
    ? { question, modelType }
    : { sessionId, question, modelType }

  const response = await fetch(`${authStore.apiBaseUrl.value}${path}`, {
    method: 'POST',
    headers: buildAuthHeaders(),
    body: JSON.stringify(body)
  })

  if (!response.ok || !response.body) {
    throw new Error(`流式请求失败：HTTP ${response.status}`)
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder('utf-8')
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()

    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const parts = buffer.split('\n\n')
    buffer = parts.pop() || ''

    for (const part of parts) {
      const event = parseSseEvent(part)
      if (!event.data) {
        continue
      }

      if (event.type === 'error') {
        let message = '流式响应失败'
        try {
          message = JSON.parse(event.data).message || message
        } catch {
          message = event.data
        }
        throw new Error(message)
      }

      if (event.data === '[DONE]') {
        handlers.onDone?.()
        return
      }

      try {
        const payload = JSON.parse(event.data)
        if (payload.sessionId) {
          handlers.onSession?.(payload.sessionId)
          continue
        }
      } catch {
      }

      handlers.onChunk?.(event.data)
    }
  }

  handlers.onDone?.()
}
