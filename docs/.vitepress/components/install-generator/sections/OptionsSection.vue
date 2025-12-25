<script setup lang="ts">
import { CATEGORIES } from '../utils'

defineProps<{
  activeTooltip: string | null
  draggedId: string | null
  dragOverId: string | null
  isFeatureDisabled: (feature: any) => boolean
  isHealthCheckTimeoutDisabled: boolean
  isAdminEmailDisabled: boolean
  isAdminPasswordDisabled: boolean
  getFeaturesByCategory: (category: string) => any[]
  toggleFeature: (feature: any) => void
  handleDragStart: (featureId: string) => void
  handleDragOver: (e: DragEvent, featureId: string) => void
  handleDragLeave: () => void
  handleDrop: (e: DragEvent, featureId: string) => void
  handleDragEnd: () => void
}>()

const emit = defineEmits<{
  'update:activeTooltip': [value: string | null]
}>()
</script>

<template>
  <!-- Options -->
  <div class="config-category options-full-width">
    <div class="category-header">
      <h3>{{ CATEGORIES.options.title }}</h3>
      <span class="category-desc">{{ CATEGORIES.options.description }}</span>
    </div>
    <div class="features-grid">
      <div
        v-for="(feature, index) in getFeaturesByCategory('options')"
        :key="feature.id"
        class="feature-chip"
        :class="{
          active: feature.enabled,
          disabled: isFeatureDisabled(feature),
          dragging: draggedId === feature.id,
          'drag-over': dragOverId === feature.id
        }"
        @click="toggleFeature(feature)"
        draggable="true"
        @dragstart="handleDragStart(feature.id)"
        @dragover="handleDragOver($event, feature.id)"
        @dragleave="handleDragLeave"
        @drop="handleDrop($event, feature.id)"
        @dragend="handleDragEnd"
        :title="feature.tooltip"
      >
        <span class="chip-label">{{ feature.label }}</span>
        <button
          v-if="feature.tooltip"
          class="tooltip-icon-chip"
          @mouseenter="$emit('update:activeTooltip', feature.id)"
          @mouseleave="$emit('update:activeTooltip', null)"
          @click.stop
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
            <line x1="12" y1="17" x2="12.01" y2="17"></line>
          </svg>
        </button>
      </div>
    </div>
    <div v-if="getFeaturesByCategory('options').some(f => activeTooltip === f.id)" class="tooltip-popover">
      <div v-for="feature in getFeaturesByCategory('options')" :key="feature.id">
        <div v-if="activeTooltip === feature.id" class="tooltip-content">
          <strong>{{ feature.label }}:</strong> {{ feature.tooltip }}
        </div>
      </div>
    </div>
    <div class="timeout-input" v-if="getFeaturesByCategory('options').find(f => f.id === 'timeout')?.enabled">
      <label>Timeout (seconds)</label>
      <input 
        type="number" 
        v-model="getFeaturesByCategory('options').find(f => f.id === 'timeout')!.value"
        placeholder="300"
      />
    </div>
    <div class="timeout-input" v-if="getFeaturesByCategory('options').find(f => f.id === 'healthCheckTimeout')?.enabled">
      <label>Health Check Timeout (seconds)</label>
      <input 
        type="number" 
        v-model="getFeaturesByCategory('options').find(f => f.id === 'healthCheckTimeout')!.value"
        placeholder="120"
        :disabled="isHealthCheckTimeoutDisabled"
      />
    </div>
    <div class="timeout-input" v-if="getFeaturesByCategory('options').find(f => f.id === 'configFile')?.enabled">
      <label>Config File Path</label>
      <input
        type="text"
        v-model="getFeaturesByCategory('options').find(f => f.id === 'configFile')!.value"
        placeholder="/path/to/config.yaml"
      />
    </div>
    <div class="timeout-input" v-if="getFeaturesByCategory('options').find(f => f.id === 'adminEmail')?.enabled" :class="{ disabled: isAdminEmailDisabled }">
      <label>Admin Email</label>
      <input
        type="text"
        v-model="getFeaturesByCategory('options').find(f => f.id === 'adminEmail')!.value"
        placeholder="admin@example.com"
        :disabled="isAdminEmailDisabled"
      />
    </div>
    <div class="timeout-input" v-if="getFeaturesByCategory('options').find(f => f.id === 'adminPassword')?.enabled" :class="{ disabled: isAdminPasswordDisabled }">
      <label>Admin Password</label>
      <input
        type="password"
        v-model="getFeaturesByCategory('options').find(f => f.id === 'adminPassword')!.value"
        placeholder="secure-password"
        :disabled="isAdminPasswordDisabled"
      />
    </div>
  </div>
</template>

