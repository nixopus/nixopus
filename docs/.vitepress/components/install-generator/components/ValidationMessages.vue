<script setup lang="ts">
import type { ValidationError } from '../utils/types'
import ValidationMessage from './ValidationMessage.vue'

defineProps<{
  validationErrors: ValidationError[]
}>()

defineEmits<{
  dismissInfoMessage: [message: string]
}>()
</script>

<template>
  <div v-if="validationErrors.length > 0" class="validation-messages">
    <ValidationMessage
      v-for="(error, index) in validationErrors"
      :key="index"
      :error="error"
      @dismiss="$emit('dismissInfoMessage', $event)"
    />
  </div>
</template>

