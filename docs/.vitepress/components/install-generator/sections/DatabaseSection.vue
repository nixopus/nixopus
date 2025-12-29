<script setup lang="ts">
import { CATEGORIES } from '../utils'

defineProps<{
  activeTooltip: string | null
  draggedId: string | null
  dragOverId: string | null
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
  <!-- Database -->
  <div class="config-category">
    <div class="category-header">
      <h3>{{ CATEGORIES.database.title }}</h3>
      <span class="category-desc">{{ CATEGORIES.database.description }}</span>
    </div>
    <div class="features-list">
      <div 
        v-for="(feature, index) in getFeaturesByCategory('database')" 
        :key="feature.id"
        class="feature-item"
        :class="{ 
          active: feature.enabled,
          dragging: draggedId === feature.id,
          'drag-over': dragOverId === feature.id
        }"
        draggable="true"
        @dragstart="handleDragStart(feature.id)"
        @dragover="handleDragOver($event, feature.id)"
        @dragleave="handleDragLeave"
        @drop="handleDrop($event, feature.id)"
        @dragend="handleDragEnd"
      >
        <div class="feature-toggle" @click="toggleFeature(feature)">
          <div class="feature-info">
            <div class="feature-label-row">
              <span class="feature-label">{{ feature.label }}</span>
              <button
                v-if="feature.tooltip"
                class="tooltip-icon"
                @mouseenter="$emit('update:activeTooltip', feature.id)"
                @mouseleave="$emit('update:activeTooltip', null)"
                @click.stop
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"></circle>
                  <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"></path>
                  <line x1="12" y1="17" x2="12.01" y2="17"></line>
                </svg>
              </button>
            </div>
            <span class="feature-desc">{{ feature.description }}</span>
            <div v-if="activeTooltip === feature.id && feature.tooltip" class="tooltip-content">
              {{ feature.tooltip }}
            </div>
          </div>
          <div class="toggle-switch" :class="{ on: feature.enabled }">
            <div class="toggle-knob"></div>
          </div>
        </div>
        <input
          v-if="feature.enabled && feature.inputType !== 'toggle'"
          :type="feature.inputType === 'number' ? 'number' : 'text'"
          v-model="feature.value"
          :placeholder="feature.placeholder"
          class="feature-input"
          @click.stop
        />
      </div>
    </div>
    <div class="category-note">
      Leave disabled to use local Docker database
    </div>
  </div>
</template>

