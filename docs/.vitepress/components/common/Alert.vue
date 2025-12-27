<script setup lang="ts">
import { IconError, IconWarning, IconInfo } from './icons'

type AlertType = 'error' | 'warning' | 'info' | 'success'

defineProps<{
  type?: AlertType
  message: string
  dismissible?: boolean
}>()

defineEmits<{
  dismiss: []
}>()
</script>

<template>
  <div class="alert" :class="type || 'info'">
    <IconError v-if="type === 'error'" :size="16" />
    <IconWarning v-else-if="type === 'warning'" :size="16" />
    <IconInfo v-else :size="16" />

    <span class="alert-message">{{ message }}</span>

    <button
      v-if="dismissible"
      class="alert-dismiss"
      @click="$emit('dismiss')"
      aria-label="Dismiss"
    >
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <line x1="18" y1="6" x2="6" y2="18"></line>
        <line x1="6" y1="6" x2="18" y2="18"></line>
      </svg>
    </button>
  </div>
</template>

<style scoped>
.alert {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  border-radius: 4px;
  font-size: 0.875rem;
}

.alert.error {
  background: var(--vp-c-danger-soft);
  color: var(--vp-c-danger);
}

.alert.warning {
  background: var(--vp-c-warning-soft);
  color: var(--vp-c-warning);
}

.alert.info {
  background: var(--vp-c-tip-soft);
  color: var(--vp-c-tip);
}

.alert-message {
  flex: 1;
}

.alert-dismiss {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  opacity: 0.6;
  transition: opacity 0.2s;
}

.alert-dismiss:hover {
  opacity: 1;
}
</style>
