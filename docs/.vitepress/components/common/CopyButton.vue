<script setup lang="ts">
import { IconCheck, IconCopy } from './icons'

defineProps<{
  copied: boolean
  title?: string
}>()

defineEmits<{
  copy: []
}>()
</script>

<template>
  <button
    class="copy-btn"
    :class="{ copied }"
    @click="$emit('copy')"
    :title="title || (copied ? 'Copied!' : 'Copy')"
  >
    <IconCheck v-if="copied" :size="16" />
    <IconCopy v-else :size="16" />
  </button>
</template>

<style scoped>
.copy-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  color: var(--vp-c-text-2);
  transition: color 0.2s;
}

.copy-btn:hover {
  color: var(--vp-c-brand);
}

.copy-btn.copied {
  color: var(--vp-c-green);
}
</style>
