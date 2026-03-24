<template>
  <aside class="sidebar">
    <div class="sidebar-top">
      <div>
        <p class="eyebrow">会话列表</p>
        <h2>聊天记录</h2>
      </div>
      <button class="secondary-button" type="button" @click="$emit('new-session')">
        新建
      </button>
    </div>

    <div class="sidebar-user">
      <div>
        <strong>{{ username || '未识别用户' }}</strong>
        <p>当前模型：通义千问</p>
      </div>
      <button type="button" class="ghost-link" @click="$emit('logout')">退出登录</button>
    </div>

    <div class="sidebar-config">
      <label class="field">
        <span>接口地址</span>
        <input
          :value="apiBaseUrl"
          type="text"
          placeholder="http://127.0.0.1:9091/api/v1"
          @change="$emit('update-api-base', $event.target.value)"
        />
      </label>
    </div>

    <div class="session-list">
      <button
        v-for="session in sessions"
        :key="session.sessionId"
        type="button"
        class="session-card"
        :class="{ 'session-card--active': activeSessionId === session.sessionId }"
        @click="$emit('select', session.sessionId)"
      >
        <strong>{{ session.title || '未命名会话' }}</strong>
        <span class="session-time">最近更新：{{ formatTime(session.updatedAt) }}</span>
      </button>

      <div v-if="!sessions.length" class="session-empty">
        还没有历史会话。发送第一条消息后会自动创建。
      </div>
    </div>
  </aside>
</template>

<script setup>
defineProps({
  username: {
    type: String,
    default: ''
  },
  apiBaseUrl: {
    type: String,
    default: ''
  },
  sessions: {
    type: Array,
    default: () => []
  },
  activeSessionId: {
    type: String,
    default: ''
  }
})

defineEmits(['select', 'new-session', 'logout', 'update-api-base'])

function formatTime(value) {
  if (!value) {
    return '暂无'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return '暂无'
  }

  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>
