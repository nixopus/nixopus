<script setup lang="ts">
import { CATEGORIES } from '../utils'

defineProps<{
  showAdvanced: boolean
  activeTooltip: string | null
  draggedId: string | null
  dragOverId: string | null
  isDbPortDisabled: boolean
  isRedisPortDisabled: boolean
  getFeaturesByCategory: (category: string) => any[]
  toggleFeature: (feature: any) => void
  handleDragStart: (featureId: string) => void
  handleDragOver: (e: DragEvent, featureId: string) => void
  handleDragLeave: () => void
  handleDrop: (e: DragEvent, featureId: string) => void
  handleDragEnd: () => void
}>()

const emit = defineEmits<{
  'update:showAdvanced': [value: boolean]
  'update:activeTooltip': [value: string | null]
}>()
</script>

<template>
  <!-- Advanced Toggle -->
  <button class="advanced-toggle" @click="$emit('update:showAdvanced', !showAdvanced)">
    <span class="toggle-icon">{{ showAdvanced ? '−' : '+' }}</span>
    {{ showAdvanced ? 'Hide' : 'Show' }} Advanced Options
    <span class="advanced-badge" v-if="!showAdvanced">Ports, Developer</span>
  </button>

  <!-- Advanced Options -->
  <div v-if="showAdvanced" class="advanced-section">
    <div class="config-grid">
      
      <!-- Ports -->
      <div class="config-category compact">
        <div class="category-header">
          <h3>{{ CATEGORIES.ports.title }}</h3>
          <span class="category-desc">{{ CATEGORIES.ports.description }}</span>
        </div>
        <div class="ports-list">
          <div 
            v-for="(feature, index) in getFeaturesByCategory('ports')" 
            :key="feature.id"
            class="port-feature-item"
            :class="{ 
              active: feature.enabled,
              disabled: (feature.id === 'dbPort' && isDbPortDisabled) || 
                       (feature.id === 'redisPort' && isRedisPortDisabled),
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
            <div class="port-toggle" @click="toggleFeature(feature)">
              <div class="port-info">
                <div class="feature-label-row">
                  <span class="port-label">{{ feature.label }}</span>
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
                <span class="port-desc">{{ feature.description }}</span>
                <div v-if="activeTooltip === feature.id && feature.tooltip" class="tooltip-content">
                  {{ feature.tooltip }}
                </div>
              </div>
              <div class="toggle-switch small" :class="{ on: feature.enabled }">
                <div class="toggle-knob"></div>
              </div>
            </div>
            <input
              v-if="feature.enabled"
              type="number"
              v-model="feature.value"
              :placeholder="feature.placeholder"
              class="port-input"
              @click.stop
            />
          </div>
        </div>
      </div>

      <!-- Developer -->
      <div class="config-category compact">
        <div class="category-header">
          <h3>{{ CATEGORIES.developer.title }}</h3>
        </div>
        <div class="features-list compact">
          <div 
            v-for="(feature, index) in getFeaturesByCategory('developer')" 
            :key="feature.id"
            class="feature-item compact"
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
              <div class="feature-info-compact">
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
                <div v-if="activeTooltip === feature.id && feature.tooltip" class="tooltip-content">
                  {{ feature.tooltip }}
                </div>
              </div>
              <div class="toggle-switch small" :class="{ on: feature.enabled }">
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
      </div>

    </div>
  </div>
</template>

