<template>
  <header class="chat-header">
    <div>
      <p class="eyebrow">对话区</p>
      <h1>{{ title }}</h1>
      <p class="chat-header__copy">
        AI 回复会按 Markdown 进行渲染，段落、列表、代码块和表格都会更清晰。
      </p>
    </div>
    <div class="chat-badges">
      <span class="badge">{{ modelType }}</span>
      <span class="badge">{{ messageCount }} 条消息</span>
      <span class="badge" :class="{ 'badge--busy': isStreaming }">
        {{ isStreaming ? '生成中' : '空闲' }}
      </span>
    </div>
  </header>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    default: '新会话'
  },
  modelType: {
    type: String,
    default: 'openai'
  },
  isStreaming: {
    type: Boolean,
    default: false
  },
  messageCount: {
    type: Number,
    default: 0
  }
})
</script>
