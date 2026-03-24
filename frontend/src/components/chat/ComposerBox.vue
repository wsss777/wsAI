<template>
  <form class="composer" @submit.prevent="submit">
    <label class="field composer-field">
      <span>输入问题</span>
      <textarea
        v-model="draft"
        rows="4"
        placeholder="请输入你的问题。按回车发送，按 Shift + 回车换行。"
        :disabled="disabled"
        @keydown.enter.exact.prevent="submit"
      />
    </label>

    <div class="composer-actions">
      <span class="composer-hint">
        {{ activeSessionId ? '消息会发送到当前会话。' : '发送后会自动创建新会话。' }}
      </span>
      <button class="primary-button" type="submit" :disabled="disabled || !draft.trim()">
        {{ disabled ? '生成中...' : '发送消息' }}
      </button>
    </div>
  </form>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  disabled: {
    type: Boolean,
    default: false
  },
  activeSessionId: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['submit'])
const draft = ref('')

function submit() {
  const question = draft.value.trim()
  if (!question || props.disabled) {
    return
  }

  emit('submit', question)
  draft.value = ''
}
</script>
