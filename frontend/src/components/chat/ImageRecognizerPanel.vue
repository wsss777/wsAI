<template>
  <section class="tool-panel">
    <div class="tool-panel__header">
      <div>
        <p class="eyebrow">图片识别</p>
        <h2>上传并识别图片</h2>
      </div>
      <span class="tool-chip">已登录</span>
    </div>

    <p class="tool-panel__copy">
      这里直接调用后端图片识别接口，请求会以 <code>multipart/form-data</code> 方式上传，字段名为
      <code>image</code>。
    </p>

    <label class="upload-dropzone" :class="{ 'upload-dropzone--filled': previewUrl }">
      <input type="file" accept="image/*" @change="onFileChange" />
      <template v-if="previewUrl">
        <img :src="previewUrl" alt="图片预览" class="upload-preview" />
      </template>
      <template v-else>
        <strong>选择图片</strong>
        <span>支持浏览器可上传的常见图片格式</span>
      </template>
    </label>

    <div class="tool-actions">
      <button class="secondary-button" type="button" :disabled="!file" @click="reset">
        清空
      </button>
      <button class="primary-button" type="button" :disabled="!file || loading" @click="submit">
        {{ loading ? '识别中...' : '开始识别' }}
      </button>
    </div>

    <div class="result-card" :class="{ 'result-card--ready': status === 'success', 'result-card--error': status === 'error' }">
      <span class="result-card__label">识别结果</span>
      <strong>{{ resultText }}</strong>
      <p class="result-card__hint">{{ helperText }}</p>
    </div>
  </section>
</template>

<script setup>
import { onBeforeUnmount, ref } from 'vue'
import { recognizeImage } from '../../services/api'

const emit = defineEmits(['error', 'success'])

const file = ref(null)
const previewUrl = ref('')
const loading = ref(false)
const status = ref('idle')
const resultText = ref('暂未识别')
const helperText = ref('识别成功后，结果会固定显示在这里；识别失败时也会在这里显示错误信息。')

function revokePreview() {
  if (previewUrl.value) {
    URL.revokeObjectURL(previewUrl.value)
    previewUrl.value = ''
  }
}

function onFileChange(event) {
  const nextFile = event.target.files?.[0]
  revokePreview()
  status.value = 'idle'
  resultText.value = '暂未识别'
  helperText.value = '识别成功后，结果会固定显示在这里；识别失败时也会在这里显示错误信息。'

  if (!nextFile) {
    file.value = null
    return
  }

  file.value = nextFile
  previewUrl.value = URL.createObjectURL(nextFile)
}

function reset() {
  file.value = null
  status.value = 'idle'
  resultText.value = '暂未识别'
  helperText.value = '识别成功后，结果会固定显示在这里；识别失败时也会在这里显示错误信息。'
  revokePreview()
}

async function submit() {
  if (!file.value || loading.value) {
    return
  }

  loading.value = true
  status.value = 'idle'
  resultText.value = '识别中...'
  helperText.value = '正在等待后端返回识别结果。'

  try {
    const data = await recognizeImage(file.value)
    const className = data.class_name || '未识别'
    status.value = 'success'
    resultText.value = className
    helperText.value = '后端已成功返回识别类别。'
    emit('success', `图片识别完成：${className}`)
  } catch (error) {
    status.value = 'error'
    resultText.value = error.message || '识别失败'
    helperText.value = '如果后端日志显示 HTTP 200，但这里仍报错，请检查接口返回的 status_code 和 status_msg。'
    emit('error', resultText.value)
  } finally {
    loading.value = false
  }
}

onBeforeUnmount(() => {
  revokePreview()
})
</script>
