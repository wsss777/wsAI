<template>
  <div class="rich-message">
    <div v-if="role === 'user'" class="markdown-body markdown-body--user">
      <p>{{ content }}</p>
    </div>
    <div v-else class="markdown-body" v-html="renderedHtml"></div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import { marked } from 'marked'

marked.setOptions({
  gfm: true,
  breaks: true
})

const props = defineProps({
  content: {
    type: String,
    default: ''
  },
  role: {
    type: String,
    default: 'assistant'
  }
})

const renderedHtml = computed(() => {
  const raw = props.content || ''
  const html = marked.parse(raw)
  return DOMPurify.sanitize(html)
})
</script>
