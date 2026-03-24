<template>
  <article class="message-card" :class="`message-card--${message.role}`">
    <span class="message-role">{{ message.role === 'user' ? '我' : 'AI' }}</span>
    <RichMessageContent :content="message.content" :role="message.role" />
  </article>
</template>

<script setup>
import RichMessageContent from './RichMessageContent.vue'

defineProps({
  message: {
    type: Object,
    required: true
  }
})
</script>
