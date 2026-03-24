<template>
  <div class="rich-message">
    <div v-if="role === 'user'" class="markdown-body markdown-body--user plain-text">
      {{ content }}
    </div>
    <template v-else>
      <template v-for="(block, index) in renderedBlocks" :key="`${block.type}-${index}`">
        <div v-if="block.type === 'text'" class="markdown-body" v-html="block.html"></div>
        <section v-else class="code-block">
          <header class="code-block__header">{{ block.language }}</header>
          <pre><code>{{ block.content }}</code></pre>
        </section>
      </template>
    </template>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import DOMPurify from 'dompurify'
import { marked } from 'marked'
import { buildMessageBlocks } from '../../utils/messageFormatter'

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

const renderedBlocks = computed(() => {
  return buildMessageBlocks(props.content, props.role).map((block) => {
    if (block.type !== 'text') {
      return block
    }

    return {
      ...block,
      html: DOMPurify.sanitize(marked.parse(block.content || ''))
    }
  })
})
</script>
