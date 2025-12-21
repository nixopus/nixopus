export interface FeatureOption {
  id: string
  label: string
  description: string
  enabled: boolean
  inputType?: 'toggle' | 'text' | 'number'
  value?: string
  placeholder?: string
  category: 'network' | 'database' | 'ports' | 'options' | 'developer'
  tooltip?: string
}

export interface ValidationError {
  message: string
  type: 'error' | 'warning' | 'info'
}

