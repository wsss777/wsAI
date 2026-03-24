<template>
  <section ref="listRef" class="message-list">
    <MessageBubble
      v-for="(message, index) in messages"
      :key="`${message.role}-${index}`"
      :message="message"
    />

    <div v-if="!messages.length" class="message-empty">
      选择左侧会话，或直接发送第一条消息开始新的对话。
    </div>
  </section>
</template>

<script setup>
import { nextTick, ref, watch } from 'vue'
import MessageBubble from './MessageBubble.vue'

const props = defineProps({
  messages: {
    type: Array,
    default: () => []
  }
})

const listRef = ref(null)

watch(
  () => props.messages,
  async () => {
    await nextTick()
    if (listRef.value) {
      listRef.value.scrollTop = listRef.value.scrollHeight
    }
  },
  { deep: true }
)
</script>
