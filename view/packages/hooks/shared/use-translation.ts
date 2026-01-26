'use client';

import { useEffect, useState } from 'react';
import { defaultLocale } from '@/lib/i18n/config';
import { loadTranslations } from '@/lib/i18n/load-translations';

// Import all domain files for type inference
import common from '@/lib/i18n/locales/en/common.json';
import containers from '@/lib/i18n/locales/en/containers.json';
import auth from '@/lib/i18n/locales/en/auth.json';
import settings from '@/lib/i18n/locales/en/settings.json';
import activities from '@/lib/i18n/locales/en/activities.json';
import languages from '@/lib/i18n/locales/en/languages.json';
import dashboard from '@/lib/i18n/locales/en/dashboard.json';
import fileManager from '@/lib/i18n/locales/en/fileManager.json';
import selfHost from '@/lib/i18n/locales/en/selfHost.json';
import terminal from '@/lib/i18n/locales/en/terminal.json';
import extensions from '@/lib/i18n/locales/en/extensions.json';
import navigation from '@/lib/i18n/locales/en/navigation.json';
import layout from '@/lib/i18n/locales/en/layout.json';
import user from '@/lib/i18n/locales/en/user.json';
import toasts from '@/lib/i18n/locales/en/toasts.json';

// Merge all domain translations for type inference
// Each domain file exports { domainName: { ... } }, so we merge them
// Using a utility type to properly merge object types
type Merge<T> = {
  [K in keyof T]: T[K];
};

type EnTranslations = Merge<
  typeof common &
    typeof containers &
    typeof auth &
    typeof settings &
    typeof activities &
    typeof languages &
    typeof dashboard &
    typeof fileManager &
    typeof selfHost &
    typeof terminal &
    typeof extensions &
    typeof navigation &
    typeof layout &
    typeof user &
    typeof toasts
>;

// Recursive way to infer types from nested json keys
type DeepKeyOf<T> = {
  [K in keyof T & string]: T[K] extends Record<string, any>
    ? `${K}` | `${K}.${DeepKeyOf<T[K]>}`
    : `${K}`;
}[keyof T & string];

// This will help us to make sure whatever keys that are entered in the t(``) string are correct,
// and enable autocompletion in editors
export type translationKey = DeepKeyOf<EnTranslations>;

// Pre merge English translations synchronously for immediate fallback
const defaultTranslations: Record<string, any> = {
  ...common,
  ...containers,
  ...auth,
  ...settings,
  ...activities,
  ...languages,
  ...dashboard,
  ...fileManager,
  ...selfHost,
  ...terminal,
  ...extensions,
  ...navigation,
  ...layout,
  ...user,
  ...toasts
};

export function useTranslation() {
  // Initialize with English translations immediately to avoid showing keys
  const [translations, setTranslations] = useState<Record<string, any>>(defaultTranslations);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    const getLocale = () => {
      if (typeof window === 'undefined') return defaultLocale;
      const cookie = document.cookie.split('; ').find((row) => row.startsWith('NEXT_LOCALE='));
      return cookie ? cookie.split('=')[1] : defaultLocale;
    };

    const load = async () => {
      const locale = getLocale();

      // If locale is English, we already have translations loaded, skip async loading
      if (locale === defaultLocale) {
        return;
      }

      try {
        // Load translations for non-English locale
        // English is already loaded as fallback, so we can show it immediately
        const data = await loadTranslations(locale);
        setTranslations(data);
      } catch (error) {
        // If loading fails, keep English translations (already loaded)
        console.warn(`Failed to load translations for locale ${locale}, using default:`, error);
      }
    };

    load();
  }, []);

  const t = (key: translationKey, params?: Record<string, string>): string => {
    const keys = key.split('.');
    let value: any = translations;

    for (const k of keys) {
      if (value && typeof value === 'object') {
        value = value[k];
      } else {
        // Fallback to key if translation not found
        return key;
      }
    }

    if (typeof value === 'string') {
      if (params) {
        return Object.entries(params).reduce(
          (str, [key, val]) => str.replace(`{${key}}`, val),
          value
        );
      }
      return value;
    }

    return key;
  };

  return { t, isLoading };
}
