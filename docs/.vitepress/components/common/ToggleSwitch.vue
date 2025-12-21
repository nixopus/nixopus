<script setup lang="ts">
defineProps<{
  enabled: boolean
  size?: 'default' | 'small'
}>()
</script>

<template>
  <div class="toggle-switch" :class="[{ on: enabled }, size === 'small' ? 'small' : '']">
    <div class="toggle-knob"></div>
  </div>
</template>
