<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vitepress'

const router = useRouter()
const route = useRoute()

let posthogInstance: any = null

onMounted(() => {
  // Only initialize PostHog in browser environment
  if (typeof window === 'undefined') return

  // Get PostHog config from window (injected by VitePress config)
  const config = (window as any).__POSTHOG_CONFIG__
  
  // Don't initialize if API key is not configured
  if (!config || !config.apiKey) {
    if (import.meta.env.MODE === 'development') {
      console.warn('PostHog API key not configured. Set POSTHOG_API_KEY environment variable.')
    }
    return
  }

  import('posthog-js').then((posthogModule) => {
    const posthog = posthogModule.default

    // Initialize PostHog with config from environment
    posthog.init(config.apiKey, {
      api_host: config.host || 'https://eu.i.posthog.com',
      defaults: '2025-11-30',
      loaded: (posthog) => {
        if (import.meta.env.MODE === 'development') {
          console.log('PostHog loaded')
        }
      }
    })

    posthogInstance = posthog

    // Track initial pageview
    posthog.capture('$pageview', {
      $current_url: window.location.href
    })
  })
})

// Track pageviews on route changes
watch(() => route.path, (newPath, oldPath) => {
  if (posthogInstance && newPath !== oldPath) {
    posthogInstance.capture('$pageview', {
      $current_url: window.location.href
    })
  }
})
</script>

<template>
  <!-- This component doesn't render anything -->
</template>

