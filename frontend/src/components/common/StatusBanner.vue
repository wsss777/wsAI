<template>
  <transition name="fade-slide">
    <div v-if="message" class="status-banner" :class="variantClass">
      <span>{{ message }}</span>
      <button type="button" class="status-close" @click="$emit('close')">关闭</button>
    </div>
  </transition>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  message: {
    type: String,
    default: ''
  },
  variant: {
    type: String,
    default: 'info'
  }
})

defineEmits(['close'])

const variantClass = computed(() => `status-banner--${props.variant}`)
</script>
