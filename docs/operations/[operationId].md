---
aside: false
outline: false
title: vitepress-openapi
---

<script setup lang="ts">
import { useRoute, useData } from 'vitepress'
import { useOpenapi } from 'vitepress-openapi'

const route = useRoute()

const { isDark } = useData()

const openapi = useOpenapi()

const operationId = route.data.params.originalOperationId || route.data.params.operationId
</script>

<OAOperation :operationId="operationId" :isDark="isDark" />