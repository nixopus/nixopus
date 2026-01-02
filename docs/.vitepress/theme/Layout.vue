<template>
  <Layout>
    <template #layout-top>
      <PostHog />
    </template>
  </Layout>
</template>

<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import PostHog from '../components/PostHog.vue'

const { Layout } = DefaultTheme
</script>

