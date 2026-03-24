<template>
  <div class="auth-tabs">
    <button
      v-for="item in items"
      :key="item.value"
      type="button"
      class="auth-tab"
      :class="{ 'auth-tab--active': modelValue === item.value }"
      @click="$emit('update:modelValue', item.value)"
    >
      {{ item.label }}
    </button>
  </div>
</template>

<script setup>
defineProps({
  modelValue: {
    type: String,
    required: true
  },
  items: {
    type: Array,
    required: true
  }
})

defineEmits(['update:modelValue'])
</script>
