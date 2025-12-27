import type { FeatureOption, ValidationError } from './types'

export const createError = (message: string, type: ValidationError['type'] = 'error'): ValidationError => ({
  message,
  type,
})

export const hasValue = (features: FeatureOption[], id: string): boolean => {
  const feature = features.find(f => f.id === id)
  return !!(feature?.enabled && feature?.value)
}

export const getFeature = (features: FeatureOption[], id: string): FeatureOption | undefined =>
  features.find(f => f.id === id)

export const validateNetwork = (features: FeatureOption[], errors: ValidationError[]): void => {
  const apiDomain = getFeature(features, 'apiDomain')
  const viewDomain = getFeature(features, 'viewDomain')
  const hostIp = getFeature(features, 'hostIp')
  
  const hasApiDomain = hasValue(features, 'apiDomain')
  const hasViewDomain = hasValue(features, 'viewDomain')
  const hasHostIp = hasValue(features, 'hostIp')

  // Domain mode vs IP mode (mutually exclusive)
  if ((hasApiDomain || hasViewDomain) && hasHostIp) {
    errors.push(createError(
      'Domain mode and Host IP mode are mutually exclusive. Use domains for HTTPS or Host IP for HTTP, not both.',
      'error'
    ))
  }

  // Both domains required together
  if (hasApiDomain && !hasViewDomain) {
    errors.push(createError(
      'API Domain requires App Domain. Both domains are mandatory for HTTPS setup with Let\'s Encrypt.',
      'error'
    ))
  }

  if (hasViewDomain && !hasApiDomain) {
    errors.push(createError(
      'App Domain requires API Domain. Both domains are mandatory for HTTPS setup with Let\'s Encrypt.',
      'error'
    ))
  }

  // Domain format validation
  if (apiDomain?.enabled && apiDomain?.value) {
    const domainRegex = /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/
    if (!domainRegex.test(apiDomain.value)) {
      errors.push(createError('API Domain must be a valid domain (e.g., api.example.com)', 'error'))
    }
    if (apiDomain.value.includes('localhost') || /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/.test(apiDomain.value)) {
      errors.push(createError(
        'API Domain cannot be localhost or IP address. Use Host IP mode for local/IP-based setup.',
        'error'
      ))
    }
  }

  if (viewDomain?.enabled && viewDomain?.value) {
    const domainRegex = /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/
    if (!domainRegex.test(viewDomain.value)) {
      errors.push(createError('App Domain must be a valid domain (e.g., app.example.com)', 'error'))
    }
    if (viewDomain.value.includes('localhost') || /\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/.test(viewDomain.value)) {
      errors.push(createError(
        'App Domain cannot be localhost or IP address. Use Host IP mode for local/IP-based setup.',
        'error'
      ))
    }
  }

  // Same domain check
  if (hasApiDomain && hasViewDomain && apiDomain?.value && viewDomain?.value && apiDomain.value === viewDomain.value) {
    errors.push(createError(
      'API Domain and App Domain must be different (e.g., api.example.com and app.example.com)',
      'error'
    ))
  }

  // Host IP validation
  if (hostIp?.enabled && hostIp?.value) {
    const ipRegex = /^(\d{1,3})\.(\d{1,3})\.(\d{1,3})\.(\d{1,3})$/
    const match = hostIp.value.match(ipRegex)
    
    if (!match) {
      errors.push(createError('Host IP must be a valid IPv4 address (e.g., 192.168.1.100)', 'error'))
    } else {
      const octets = match.slice(1, 5).map(Number)
      if (octets.some(o => o > 255)) {
        errors.push(createError('Host IP octets must be 0-255', 'error'))
      }
      
      const isPrivate = (octets[0] === 10) || 
                       (octets[0] === 172 && octets[1] >= 16 && octets[1] <= 31) ||
                       (octets[0] === 192 && octets[1] === 168)
      
      if (!isPrivate) {
        errors.push(createError('Host IP appears to be public. Ensure firewall rules are configured.', 'warning'))
      }
      
      if (hostIp.value === '127.0.0.1' || hostIp.value.startsWith('127.')) {
        errors.push(createError('Localhost (127.x.x.x) will only be accessible from this machine.', 'info'))
      }
    }
  }

  // Auto-detection info
  if (!hasApiDomain && !hasViewDomain && !hasHostIp) {
    errors.push(createError(
      'Using default HTTP mode. Host IP will be auto-detected. Add domains for HTTPS or specify Host IP.',
      'info'
    ))
  }
}

export const validateDatabase = (features: FeatureOption[], errors: ValidationError[]): void => {
  const externalDb = getFeature(features, 'externalDb')
  const dbPort = getFeature(features, 'dbPort')
  const hasExternalDb = hasValue(features, 'externalDb')

  if (hasExternalDb) {
    if (!externalDb?.value) {
      errors.push(createError('External Database URL is required when enabled', 'error'))
    } else {
      if (!externalDb.value.startsWith('postgresql://') && !externalDb.value.startsWith('postgres://')) {
        errors.push(createError('External Database URL must start with postgresql:// or postgres://', 'error'))
      }
      
      const dbUrlPattern = /^postgres(ql)?:\/\/([^:]+):([^@]+)@([^:\/]+):?(\d+)?\/(.+)$/
      if (!dbUrlPattern.test(externalDb.value)) {
        errors.push(createError(
          'Invalid database URL format. Expected: postgresql://user:password@host:port/database',
          'error'
        ))
      }
    }

    if (dbPort?.enabled && dbPort?.value) {
      errors.push(createError(
        'Database Port is ignored when using External Database. Remove Database Port or disable External Database.',
        'error'
      ))
    }
  }
}

export const validatePorts = (features: FeatureOption[], errors: ValidationError[]): void => {
  const portFields = [
    { id: 'apiPort', name: 'API Port', default: 8443 },
    { id: 'viewPort', name: 'Dashboard Port', default: 7443 },
    { id: 'dbPort', name: 'Database Port', default: 5432 },
    { id: 'redisPort', name: 'Redis Port', default: 6379 },
    { id: 'caddyAdminPort', name: 'Caddy Admin Port', default: 2019 },
    { id: 'caddyHttpPort', name: 'Caddy HTTP Port', default: 80 },
    { id: 'caddyHttpsPort', name: 'Caddy HTTPS Port', default: 443 },
    { id: 'supertokensPort', name: 'SuperTokens Port', default: 3567 },
  ]

  const enabledPorts: Array<{ name: string; value: number; id: string }> = []
  const wellKnownPorts: Record<number, string> = {
    22: 'SSH', 80: 'HTTP', 443: 'HTTPS', 3306: 'MySQL',
    5432: 'PostgreSQL', 6379: 'Redis', 27017: 'MongoDB',
  }

  portFields.forEach(({ id, name, default: defaultPort }) => {
    const field = getFeature(features, id)
    if (field?.enabled && field?.value) {
      const port = parseInt(field.value)
      
      if (isNaN(port) || port < 1 || port > 65535) {
        errors.push(createError(`${name} must be between 1-65535`, 'error'))
      } else {
        enabledPorts.push({ name, value: port, id })
        
        if (port < 1024 && port !== 443 && port !== 80) {
          errors.push(createError(
            `${name} ${port} requires root/sudo privileges. Recommended: ${defaultPort}`,
            'warning'
          ))
        }
        
        if (wellKnownPorts[port] && !['dbPort', 'redisPort'].includes(id)) {
          errors.push(createError(
            `${name} ${port} is commonly used by ${wellKnownPorts[port]}. May cause conflicts.`,
            'warning'
          ))
        }
      }
    }
  })

  // Port conflict detection
  const portMap = new Map<number, string[]>()
  enabledPorts.forEach(p => {
    if (!portMap.has(p.value)) portMap.set(p.value, [])
    portMap.get(p.value)!.push(p.name)
  })

  portMap.forEach((services, port) => {
    if (services.length > 1) {
      errors.push(createError(
        `Port ${port} conflict: ${services.join(' and ')} cannot use the same port`,
        'error'
      ))
    }
  })

  // Domain mode + custom ports warning
  if (hasValue(features, 'apiDomain') && hasValue(features, 'viewDomain')) {
    const apiPort = getFeature(features, 'apiPort')
    const viewPort = getFeature(features, 'viewPort')
    
    if (apiPort?.enabled && apiPort?.value) {
      const port = parseInt(apiPort.value)
      if (port !== 443 && port !== 8443) {
        errors.push(createError(
          `API Port ${port} with HTTPS domains. Ensure DNS and firewall allow this port.`,
          'info'
        ))
      }
    }
    
    if (viewPort?.enabled && viewPort?.value) {
      const port = parseInt(viewPort.value)
      if (port !== 443 && port !== 7443) {
        errors.push(createError(
          `Dashboard Port ${port} with HTTPS domains. Ensure DNS and firewall allow this port.`,
          'info'
        ))
      }
    }
  }
}

export const validateOptions = (features: FeatureOption[], errors: ValidationError[]): void => {
  const dryRun = getFeature(features, 'dryRun')
  const force = getFeature(features, 'force')
  const timeout = getFeature(features, 'timeout')
  const configFile = getFeature(features, 'configFile')
  const noRollback = getFeature(features, 'noRollback')
  const healthCheckTimeout = getFeature(features, 'healthCheckTimeout')
  const adminEmail = getFeature(features, 'adminEmail')
  const adminPassword = getFeature(features, 'adminPassword')

  // Dry Run + Force
  if (dryRun?.enabled && force?.enabled) {
    errors.push(createError(
      'Dry Run previews changes without executing. Force flag will be ignored in dry run mode.',
      'warning'
    ))
  }

  // Timeout validation
  if (timeout?.enabled && timeout?.value) {
    const timeoutVal = parseInt(timeout.value)
    if (isNaN(timeoutVal) || timeoutVal < 1) {
      errors.push(createError('Timeout must be a positive number (seconds)', 'error'))
    } else if (timeoutVal < 60) {
      errors.push(createError(
        'Timeout < 60s may cause installation failure. Recommended: 300s (5 minutes)',
        'warning'
      ))
    } else if (timeoutVal > 3600) {
      errors.push(createError('Timeout > 3600s (1 hour) is unusually long', 'warning'))
    }
  }

  // Health check timeout validation
  if (healthCheckTimeout?.enabled && healthCheckTimeout?.value) {
    const healthTimeoutVal = parseInt(healthCheckTimeout.value)
    if (isNaN(healthTimeoutVal) || healthTimeoutVal < 1) {
      errors.push(createError('Health Check Timeout must be a positive number (seconds)', 'error'))
    } else if (healthTimeoutVal < 30) {
      errors.push(createError(
        'Health Check Timeout < 30s may cause false failures. Recommended: 120s (2 minutes)',
        'warning'
      ))
    } else if (healthTimeoutVal > 600) {
      errors.push(createError('Health Check Timeout > 600s (10 minutes) is unusually long', 'warning'))
    }
  }

  // Admin registration validation
  if ((adminEmail?.enabled && adminEmail?.value) || (adminPassword?.enabled && adminPassword?.value)) {
    if (adminEmail?.enabled && adminEmail?.value && !adminPassword?.enabled) {
      errors.push(createError(
        'Admin email provided but password is missing. Both are required for registration.',
        'warning'
      ))
    }
    
    if (adminPassword?.enabled && adminPassword?.value && !adminEmail?.enabled) {
      errors.push(createError(
        'Admin password provided but email is missing. Both are required for registration.',
        'warning'
      ))
    }
    
    if (adminEmail?.enabled && adminEmail?.value) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!emailRegex.test(adminEmail.value)) {
        errors.push(createError('Admin email must be a valid email address', 'error'))
      }
    }
    
    if (adminPassword?.enabled && adminPassword?.value) {
      if (adminPassword.value.length < 8) {
        errors.push(createError('Admin password should be at least 8 characters long', 'warning'))
      }
    }
  }

  // Config file validation
  if (configFile?.enabled && configFile?.value && !configFile.value.trim()) {
    errors.push(createError('Config file path cannot be empty', 'error'))
  }

  // No rollback warning
  if (noRollback?.enabled) {
    errors.push(createError(
      'Rollback disabled. Failed installations will not be automatically cleaned up.',
      'warning'
    ))
  }
}

export const validateDeveloper = (features: FeatureOption[], errors: ValidationError[]): void => {
  const repo = getFeature(features, 'repo')
  const branch = getFeature(features, 'branch')
  const staging = getFeature(features, 'staging')
  const hasApiDomain = hasValue(features, 'apiDomain')
  const hasViewDomain = hasValue(features, 'viewDomain')

  // Custom repository
  if (repo?.enabled) {
    if (!repo?.value) {
      errors.push(createError('Repository URL required when enabled', 'error'))
    } else {
      if (!repo.value.startsWith('http://') && !repo.value.startsWith('https://')) {
        errors.push(createError('Repository URL must start with http:// or https://', 'error'))
      }
      
      if (!repo.value.includes('github.com') && !repo.value.includes('gitlab.com')) {
        errors.push(createError('Repository should be from GitHub or GitLab', 'warning'))
      }
      
      if (!branch?.enabled || !branch?.value) {
        errors.push(createError(
          'Branch name recommended with custom repository. Defaults to "main" if not specified.',
          'info'
        ))
      }
    }
  }

  // Branch without custom repo
  if (branch?.enabled && branch?.value && (!repo?.enabled || !repo?.value)) {
    errors.push(createError(
      'Branch specified without custom repository. Will use official Nixopus repo with this branch.',
      'info'
    ))
  }

  // Staging mode validation
  if (staging?.enabled) {
    if (repo?.enabled && repo?.value) {
      errors.push(createError(
        'Staging mode with custom repository. Ensure your fork has docker-compose-staging.yml file.',
        'warning'
      ))
    }
    
    if (hasApiDomain && hasViewDomain) {
      errors.push(createError(
        'Staging mode with production domains. Ensure this is intentional.',
        'warning'
      ))
    }
  }
}

export const validateEmptyFields = (features: FeatureOption[], errors: ValidationError[]): void => {
  const fieldsToCheck = [
    { id: 'apiDomain', name: 'API Domain' },
    { id: 'viewDomain', name: 'App Domain' },
    { id: 'hostIp', name: 'Host IP', isInfo: true },
    { id: 'externalDb', name: 'External Database' },
    { id: 'repo', name: 'Custom Repository' },
    { id: 'branch', name: 'Branch' },
    { id: 'configFile', name: 'Config File' },
    { id: 'healthCheckTimeout', name: 'Health Check Timeout' },
    { id: 'adminEmail', name: 'Admin Email' },
    { id: 'adminPassword', name: 'Admin Password' },
  ]

  fieldsToCheck.forEach(({ id, name, isInfo }) => {
    const feature = getFeature(features, id)
    if (feature?.enabled && !feature?.value) {
      const message = isInfo && id === 'hostIp'
        ? `${name} is enabled but empty. Will be auto-detected.`
        : `${name} is enabled but empty`
      errors.push(createError(message, isInfo ? 'info' : 'warning'))
    }
  })
}

export const validateAll = (
  features: FeatureOption[],
  dismissedInfoMessages: Set<string>,
  hasCustomizations: boolean
): ValidationError[] => {
  const errors: ValidationError[] = []
  
  validateNetwork(features, errors)
  validateDatabase(features, errors)
  validatePorts(features, errors)
  validateOptions(features, errors)
  validateDeveloper(features, errors)
  validateEmptyFields(features, errors)

  // Default installation info
  if (!hasCustomizations) {
    errors.push(createError(
      'Default installation: HTTP mode with auto-detected IP, local PostgreSQL and Redis. Toggle options to customize.',
      'info'
    ))
  }
  
  return errors.filter(error => !dismissedInfoMessages.has(error.message))
}

