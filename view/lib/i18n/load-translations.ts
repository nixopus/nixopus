// Helper function to load and merge domain-based translation files
// Using a mapping to ensure Next.js can statically analyze the imports
const domainLoaders: Record<string, Record<string, () => Promise<any>>> = {
  en: {
    common: () => import('@/lib/i18n/locales/en/common.json'),
    containers: () => import('@/lib/i18n/locales/en/containers.json'),
    auth: () => import('@/lib/i18n/locales/en/auth.json'),
    settings: () => import('@/lib/i18n/locales/en/settings.json'),
    activities: () => import('@/lib/i18n/locales/en/activities.json'),
    languages: () => import('@/lib/i18n/locales/en/languages.json'),
    dashboard: () => import('@/lib/i18n/locales/en/dashboard.json'),
    fileManager: () => import('@/lib/i18n/locales/en/fileManager.json'),
    selfHost: () => import('@/lib/i18n/locales/en/selfHost.json'),
    terminal: () => import('@/lib/i18n/locales/en/terminal.json'),
    extensions: () => import('@/lib/i18n/locales/en/extensions.json'),
    navigation: () => import('@/lib/i18n/locales/en/navigation.json'),
    layout: () => import('@/lib/i18n/locales/en/layout.json'),
    user: () => import('@/lib/i18n/locales/en/user.json'),
    toasts: () => import('@/lib/i18n/locales/en/toasts.json')
  },
  es: {
    common: () => import('@/lib/i18n/locales/es/common.json'),
    containers: () => import('@/lib/i18n/locales/es/containers.json'),
    auth: () => import('@/lib/i18n/locales/es/auth.json'),
    settings: () => import('@/lib/i18n/locales/es/settings.json'),
    activities: () => import('@/lib/i18n/locales/es/activities.json'),
    languages: () => import('@/lib/i18n/locales/es/languages.json'),
    dashboard: () => import('@/lib/i18n/locales/es/dashboard.json'),
    fileManager: () => import('@/lib/i18n/locales/es/fileManager.json'),
    selfHost: () => import('@/lib/i18n/locales/es/selfHost.json'),
    terminal: () => import('@/lib/i18n/locales/es/terminal.json'),
    extensions: () => import('@/lib/i18n/locales/es/extensions.json'),
    navigation: () => import('@/lib/i18n/locales/es/navigation.json'),
    layout: () => import('@/lib/i18n/locales/es/layout.json'),
    user: () => import('@/lib/i18n/locales/es/user.json'),
    toasts: () => import('@/lib/i18n/locales/es/toasts.json')
  },
  fr: {
    common: () => import('@/lib/i18n/locales/fr/common.json'),
    containers: () => import('@/lib/i18n/locales/fr/containers.json'),
    auth: () => import('@/lib/i18n/locales/fr/auth.json'),
    settings: () => import('@/lib/i18n/locales/fr/settings.json'),
    activities: () => import('@/lib/i18n/locales/fr/activities.json'),
    languages: () => import('@/lib/i18n/locales/fr/languages.json'),
    dashboard: () => import('@/lib/i18n/locales/fr/dashboard.json'),
    fileManager: () => import('@/lib/i18n/locales/fr/fileManager.json'),
    selfHost: () => import('@/lib/i18n/locales/fr/selfHost.json'),
    terminal: () => import('@/lib/i18n/locales/fr/terminal.json'),
    extensions: () => import('@/lib/i18n/locales/fr/extensions.json'),
    navigation: () => import('@/lib/i18n/locales/fr/navigation.json'),
    layout: () => import('@/lib/i18n/locales/fr/layout.json'),
    user: () => import('@/lib/i18n/locales/fr/user.json'),
    toasts: () => import('@/lib/i18n/locales/fr/toasts.json')
  },
  kn: {
    common: () => import('@/lib/i18n/locales/kn/common.json'),
    containers: () => import('@/lib/i18n/locales/kn/containers.json'),
    auth: () => import('@/lib/i18n/locales/kn/auth.json'),
    settings: () => import('@/lib/i18n/locales/kn/settings.json'),
    activities: () => import('@/lib/i18n/locales/kn/activities.json'),
    languages: () => import('@/lib/i18n/locales/kn/languages.json'),
    dashboard: () => import('@/lib/i18n/locales/kn/dashboard.json'),
    fileManager: () => import('@/lib/i18n/locales/kn/fileManager.json'),
    selfHost: () => import('@/lib/i18n/locales/kn/selfHost.json'),
    terminal: () => import('@/lib/i18n/locales/kn/terminal.json'),
    extensions: () => import('@/lib/i18n/locales/kn/extensions.json'),
    navigation: () => import('@/lib/i18n/locales/kn/navigation.json'),
    layout: () => import('@/lib/i18n/locales/kn/layout.json'),
    user: () => import('@/lib/i18n/locales/kn/user.json'),
    toasts: () => import('@/lib/i18n/locales/kn/toasts.json')
  },
  ml: {
    common: () => import('@/lib/i18n/locales/ml/common.json'),
    containers: () => import('@/lib/i18n/locales/ml/containers.json'),
    auth: () => import('@/lib/i18n/locales/ml/auth.json'),
    settings: () => import('@/lib/i18n/locales/ml/settings.json'),
    activities: () => import('@/lib/i18n/locales/ml/activities.json'),
    languages: () => import('@/lib/i18n/locales/ml/languages.json'),
    dashboard: () => import('@/lib/i18n/locales/ml/dashboard.json'),
    fileManager: () => import('@/lib/i18n/locales/ml/fileManager.json'),
    selfHost: () => import('@/lib/i18n/locales/ml/selfHost.json'),
    terminal: () => import('@/lib/i18n/locales/ml/terminal.json'),
    extensions: () => import('@/lib/i18n/locales/ml/extensions.json'),
    navigation: () => import('@/lib/i18n/locales/ml/navigation.json'),
    layout: () => import('@/lib/i18n/locales/ml/layout.json'),
    user: () => import('@/lib/i18n/locales/ml/user.json'),
    toasts: () => import('@/lib/i18n/locales/ml/toasts.json')
  }
};

export async function loadTranslations(locale: string) {
  const loaders = domainLoaders[locale] || domainLoaders.en;
  const translations: Record<string, any> = {};

  // Load all domain files and merge them
  // Each domain file exports { domainName: { ... } }, so we merge the objects
  for (const [domain, loader] of Object.entries(loaders)) {
    try {
      const module = await loader();
      // Each module.default is { domainName: { ... } }, so we merge it into translations
      Object.assign(translations, module.default);
    } catch (error) {
      console.warn(`Failed to load translation domain ${domain} for locale ${locale}:`, error);
    }
  }

  return translations;
}
