import { ref, computed } from 'vue'
import type { FeatureOption, ValidationError } from '../utils/types'
import { DEFAULT_FEATURES, CATEGORIES, COMMANDS, FLAG_MAPPINGS, BOOLEAN_FLAGS } from '../utils/constants'
import { validateAll, hasValue } from '../utils/validation'

export function useInstallGenerator() {
  const features = ref<FeatureOption[]>([...DEFAULT_FEATURES])
  const copied = ref<'download' | 'run' | null>(null)
  const showAdvanced = ref(false)
  const activeTooltip = ref<string | null>(null)
  const dismissedInfoMessages = ref<Set<string>>(new Set())
  const draggedId = ref<string | null>(null)
  const dragOverId = ref<string | null>(null)

  const findFeature = (id: string): FeatureOption | undefined =>
    features.value.find(f => f.id === id)

  const isFeatureEnabled = (id: string): boolean =>
    findFeature(id)?.enabled ?? false

  const isHostIpDisabled = computed(() =>
    isFeatureEnabled('apiDomain') || isFeatureEnabled('viewDomain')
  )

  const areDomainsDisabled = computed(() =>
    isFeatureEnabled('hostIp')
  )

  const isDbPortDisabled = computed(() =>
    isFeatureEnabled('externalDb')
  )

  const isRedisPortDisabled = computed(() =>
    false
  )

  const isHealthCheckTimeoutDisabled = computed(() =>
    false
  )

  const isAdminEmailDisabled = computed(() =>
    false
  )

  const isAdminPasswordDisabled = computed(() =>
    false
  )

  const isForceDisabled = computed(() =>
    isFeatureEnabled('dryRun')
  )

  const isFeatureDisabled = (feature: FeatureOption): boolean => {
    switch (feature.id) {
      case 'hostIp':
        return isHostIpDisabled.value || false
      case 'apiDomain':
      case 'viewDomain':
        return areDomainsDisabled.value || false
      case 'dbPort':
        return isDbPortDisabled.value || false
      case 'redisPort':
        return isRedisPortDisabled.value || false
      case 'healthCheckTimeout':
        return isHealthCheckTimeoutDisabled.value || false
      case 'adminEmail':
        return isAdminEmailDisabled.value || false
      case 'adminPassword':
        return isAdminPasswordDisabled.value || false
      case 'force':
        return isForceDisabled.value || false
      default:
        return false
    }
  }

  const toggleFeature = (feature: FeatureOption) => {
    if (isFeatureDisabled(feature)) return

    if (!feature.enabled) {
      if (feature.id === 'hostIp') {
        const apiDomain = findFeature('apiDomain')
        const viewDomain = findFeature('viewDomain')
        if (apiDomain?.enabled || viewDomain?.enabled) return
      }

      if (feature.id === 'apiDomain' || feature.id === 'viewDomain') {
        const hostIp = findFeature('hostIp')
        if (hostIp?.enabled) return
      }

      if (feature.id === 'dbPort') {
        const externalDb = findFeature('externalDb')
        if (externalDb?.enabled) return
      }

      if (feature.id === 'force') {
        const dryRun = findFeature('dryRun')
        if (dryRun?.enabled) return
      }
    }

    feature.enabled = !feature.enabled
    if (!feature.enabled) {
      feature.value = ''

      if (feature.id === 'dryRun') {
        const force = findFeature('force')
        if (force?.enabled) {
          force.enabled = false
        }
      }
    }
  }

  const hasCustomizations = computed(() =>
    features.value.some(f => f.enabled)
  )

  const activeCount = computed(() =>
    features.value.filter(f => f.enabled).length
  )

  const baseCommand = computed(() => COMMANDS.curl)

  const runCommand = computed(() => {
    const flags: string[] = []

    features.value.forEach(f => {
      if (!f.enabled) return

      if (BOOLEAN_FLAGS[f.id]) {
        flags.push(BOOLEAN_FLAGS[f.id])
        return
      }

      if (FLAG_MAPPINGS[f.id] && f.value) {
        flags.push(FLAG_MAPPINGS[f.id](f.value))
      }
    })

    if (flags.length === 0) {
      return baseCommand.value
    }

    return `${COMMANDS.curl} -s -- ${flags.join(' ')}`
  })

  const validationErrors = computed<ValidationError[]>(() => {
    return validateAll(features.value, dismissedInfoMessages.value, hasCustomizations.value)
  })

  const hasErrors = computed(() =>
    validationErrors.value.some(e => e.type === 'error')
  )

  const hasWarnings = computed(() =>
    validationErrors.value.some(e => e.type === 'warning')
  )

  const hasInfo = computed(() =>
    validationErrors.value.some(e => e.type === 'info')
  )

  const getFeaturesByCategory = (category: string) =>
    features.value.filter(f => f.category === category)

  const copyToClipboard = async (text: string, type: 'download' | 'run') => {
    try {
      await navigator.clipboard.writeText(text)
      copied.value = type
      setTimeout(() => { copied.value = null }, 2000)
    } catch {
      const el = document.createElement('textarea')
      el.value = text
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      document.body.removeChild(el)
      copied.value = type
      setTimeout(() => { copied.value = null }, 2000)
    }
  }

  const handleDragStart = (featureId: string) => {
    draggedId.value = featureId
  }

  const handleDragOver = (e: DragEvent, featureId: string) => {
    e.preventDefault()
    dragOverId.value = featureId
  }

  const handleDragLeave = () => {
    dragOverId.value = null
  }

  const handleDrop = (e: DragEvent, dropFeatureId: string) => {
    e.preventDefault()
    
    if (draggedId.value === null || draggedId.value === dropFeatureId) {
      draggedId.value = null
      dragOverId.value = null
      return
    }
    
    const draggedIndex = features.value.findIndex(f => f.id === draggedId.value)
    const dropIndex = features.value.findIndex(f => f.id === dropFeatureId)
    
    if (draggedIndex === -1 || dropIndex === -1) {
      draggedId.value = null
      dragOverId.value = null
      return
    }
    
    const items = [...features.value]
    const [draggedItem] = items.splice(draggedIndex, 1)
    items.splice(dropIndex, 0, draggedItem)
    
    features.value = items
    draggedId.value = null
    dragOverId.value = null
  }

  const handleDragEnd = () => {
    draggedId.value = null
    dragOverId.value = null
  }

  const dismissInfoMessage = (message: string) => {
    dismissedInfoMessages.value.add(message)
  }

  const resetAll = () => {
    features.value.forEach(f => {
      f.enabled = false
      f.value = ''
    })
  }

  return {
    // State
    features,
    copied,
    showAdvanced,
    activeTooltip,
    dismissedInfoMessages,
    draggedId,
    dragOverId,
    
    // Computed
    hasCustomizations,
    activeCount,
    baseCommand,
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
    
    // Actions
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
  }
}

