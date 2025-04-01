'use client';

import { useRouter } from 'next/navigation';
import { locales } from '@/lib/i18n/config';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { useTranslation } from '@/hooks/use-translation';
import { useEffect, useState } from 'react';

export function LanguageSwitcher() {
  const router = useRouter();
  const { t } = useTranslation();
  const [currentLocale, setCurrentLocale] = useState('en');

  useEffect(() => {
    const cookie = document.cookie.split('; ').find((row) => row.startsWith('NEXT_LOCALE='));
    if (cookie) {
      const locale = cookie.split('=')[1];
      setCurrentLocale(locale);
    }
  }, []);

  const handleLanguageChange = async (locale: string) => {
    document.cookie = `NEXT_LOCALE=${locale}; path=/; max-age=31536000; SameSite=Lax`;
    setCurrentLocale(locale);
    window.location.reload();
  };

  return (
    <Select value={currentLocale} onValueChange={handleLanguageChange}>
      <SelectTrigger className="w-[180px]">
        <SelectValue placeholder={t('settings.preferences.language.select')} />
      </SelectTrigger>
      <SelectContent>
        {locales.map((locale) => (
          <SelectItem key={locale} value={locale}>
            {t(`languages.${locale}`)}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
