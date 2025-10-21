'use client';

import { useEffect, useState } from 'react';
import { defaultLocale } from '@/lib/i18n/config';
import en from '@/lib/i18n/locales/en.json';

// Recursive way to infer types from nested json keys
type DeepKeyOf<T> = {
  [K in keyof T & string]: T[K] extends Record<string, any>
    ? `${K}` | `${K}.${DeepKeyOf<T[K]>}`
    : `${K}`;
}[keyof T & string];

// This will help us to make sure whatever keys that are entered in the t(``) string are correct,
// and enable autocompletion in editors
export type translationKey = DeepKeyOf<typeof en>;

export function useTranslation() {
  const [translations, setTranslations] = useState<Record<string, any>>({});
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const getLocale = () => {
      if (typeof window === 'undefined') return defaultLocale;
      const cookie = document.cookie.split('; ').find((row) => row.startsWith('NEXT_LOCALE='));
      return cookie ? cookie.split('=')[1] : defaultLocale;
    };

    const loadTranslations = async () => {
      try {
        setIsLoading(true);
        const locale = getLocale();
        const data = await import(`@/lib/i18n/locales/${locale}.json`);
        setTranslations(data.default);
      } catch (error) {
        const data = await import(`@/lib/i18n/locales/${defaultLocale}.json`);
        setTranslations(data.default);
      } finally {
        setIsLoading(false);
      }
    };

    loadTranslations();
  }, []);

  const t = (key: translationKey, params?: Record<string, string>): string => {
    if (isLoading) return key;

    const keys = key.split('.');
    let value: any = translations;

    for (const k of keys) {
      if (value && typeof value === 'object') {
        value = value[k];
      } else {
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
