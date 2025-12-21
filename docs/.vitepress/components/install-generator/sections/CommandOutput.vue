<script setup lang="ts">
import type { ValidationError } from '../utils/types'
import { ValidationMessages } from '../components'

defineProps<{
  runCommand: string
  activeCount: number
  hasCustomizations: boolean
  copied: 'download' | 'run' | null
  validationErrors: ValidationError[]
}>()

defineEmits<{
  copyToClipboard: [text: string, type: 'run']
  resetAll: []
  dismissInfoMessage: [message: string]
}>()
</script>

<template>
  <div class="command-section">
    <div class="section-header">
      <div class="header-left">
        <h3>Your Install Command</h3>
        <span v-if="activeCount > 0" class="config-count">
          {{ activeCount }} option{{ activeCount > 1 ? 's' : '' }}
        </span>
      </div>
      <button class="reset-btn" @click="$emit('resetAll')" v-if="hasCustomizations">
        Reset
      </button>
    </div>

    <div class="command-notice">
      <div class="command-content">
        <code>{{ runCommand }}</code>
      </div>
      <button
        class="copy-mini"
        :class="{ copied: copied === 'run' }"
        @click="$emit('copyToClipboard', runCommand, 'run')"
        :title="copied === 'run' ? 'Copied!' : 'Copy command'"
      >
        <svg v-if="copied === 'run'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
        <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
        </svg>
      </button>
    </div>

    <div class="whats-included">
      <span class="included-label">Installs:</span>
      <div class="included-list">
        <span class="included-item">Nixopus API</span>
        <span class="included-item">Dashboard</span>
        <span class="included-item">PostgreSQL</span>
        <span class="included-item">Redis</span>
        <span class="included-item">Caddy</span>
        <span class="included-item">Docker</span>
      </div>
    </div>

    <!-- Validation Messages -->
    <ValidationMessages
      :validation-errors="validationErrors"
      @dismiss-info-message="$emit('dismissInfoMessage', $event)"
    />
  </div>
</template>

