'use client';

import { useEffect, useState } from 'react';
import { defaultLocale, locales } from '@/lib/i18n/config';

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

  const t = (key: string, params?: Record<string, string>): string => {
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
