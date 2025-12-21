<script setup lang="ts">
import { useInstallGenerator } from './install-generator/composables'
import { CATEGORIES } from './install-generator/utils'
import CommandOutput from './install-generator/sections/CommandOutput.vue'
import AdvancedSection from './install-generator/sections/AdvancedSection.vue'
import OptionsSection from './install-generator/sections/OptionsSection.vue'
import DatabaseSection from './install-generator/sections/DatabaseSection.vue'
import NetworkSection from './install-generator/sections/NetworkSection.vue'
import './install-generator/styles.css'

// Use the composable for all logic
const {
  features,
  copied,
  showAdvanced,
  activeTooltip,
  dismissedInfoMessages,
  draggedId,
  dragOverId,
  hasCustomizations,
  activeCount,
  runCommand,
  validationErrors,
  hasErrors,
  hasWarnings,
  hasInfo,
  isHostIpDisabled,
  areDomainsDisabled,
  isDbPortDisabled,
  isRedisPortDisabled,
  isHealthCheckTimeoutDisabled,
  isAdminEmailDisabled,
  isAdminPasswordDisabled,
  isForceDisabled,
  toggleFeature,
  isFeatureDisabled,
  getFeaturesByCategory,
  copyToClipboard,
  handleDragStart,
  handleDragOver,
  handleDragLeave,
  handleDrop,
  handleDragEnd,
  dismissInfoMessage,
  resetAll,
} = useInstallGenerator()
</script>

<template>
  <div class="install-generator">

    <!-- Command Output -->
    <CommandOutput
      :run-command="runCommand"
      :active-count="activeCount"
      :has-customizations="hasCustomizations"
      :copied="copied"
      :validation-errors="validationErrors"
      @copy-to-clipboard="copyToClipboard"
      @reset-all="resetAll"
      @dismiss-info-message="dismissInfoMessage"
    />

    <!-- Main Grid -->
    <div class="config-grid">
      
      <!-- Network -->
      <NetworkSection
        :active-tooltip="activeTooltip"
        :dragged-id="draggedId"
        :drag-over-id="dragOverId"
        :is-host-ip-disabled="isHostIpDisabled"
        :are-domains-disabled="areDomainsDisabled"
        :get-features-by-category="getFeaturesByCategory"
        :toggle-feature="toggleFeature"
        :handle-drag-start="handleDragStart"
        :handle-drag-over="handleDragOver"
        :handle-drag-leave="handleDragLeave"
        :handle-drop="handleDrop"
        :handle-drag-end="handleDragEnd"
        @update:active-tooltip="activeTooltip = $event"
      />

      <!-- Database -->
      <DatabaseSection
        :active-tooltip="activeTooltip"
        :dragged-id="draggedId"
        :drag-over-id="dragOverId"
        :get-features-by-category="getFeaturesByCategory"
        :toggle-feature="toggleFeature"
        :handle-drag-start="handleDragStart"
        :handle-drag-over="handleDragOver"
        :handle-drag-leave="handleDragLeave"
        :handle-drop="handleDrop"
        :handle-drag-end="handleDragEnd"
        @update:active-tooltip="activeTooltip = $event"
      />

      <!-- Options -->
      <OptionsSection
        :active-tooltip="activeTooltip"
        :dragged-id="draggedId"
        :drag-over-id="dragOverId"
        :is-feature-disabled="isFeatureDisabled"
        :is-health-check-timeout-disabled="isHealthCheckTimeoutDisabled"
        :is-admin-email-disabled="isAdminEmailDisabled"
        :is-admin-password-disabled="isAdminPasswordDisabled"
        :get-features-by-category="getFeaturesByCategory"
        :toggle-feature="toggleFeature"
        :handle-drag-start="handleDragStart"
        :handle-drag-over="handleDragOver"
        :handle-drag-leave="handleDragLeave"
        :handle-drop="handleDrop"
        :handle-drag-end="handleDragEnd"
        @update:active-tooltip="activeTooltip = $event"
      />

    </div>

    <!-- Advanced Section -->
    <AdvancedSection
      :show-advanced="showAdvanced"
      :active-tooltip="activeTooltip"
      :dragged-id="draggedId"
      :drag-over-id="dragOverId"
      :is-db-port-disabled="isDbPortDisabled"
      :is-redis-port-disabled="isRedisPortDisabled"
      :get-features-by-category="getFeaturesByCategory"
      :toggle-feature="toggleFeature"
      :handle-drag-start="handleDragStart"
      :handle-drag-over="handleDragOver"
      :handle-drag-leave="handleDragLeave"
      :handle-drop="handleDrop"
      :handle-drag-end="handleDragEnd"
      @update:show-advanced="showAdvanced = $event"
      @update:active-tooltip="activeTooltip = $event"
    />

  </div>
</template>
