'use client';

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
import { useUpdateLanguageMutation } from '@/redux/services/users/userApi';
import { UserSettings } from '@/redux/types/user';

interface LanguageSwitcherProps {
  handleLanguageChange: (language: string) => void;
  isUpdatingLanguage: boolean;
  userSettings: UserSettings;
}

export function LanguageSwitcher({ handleLanguageChange, isUpdatingLanguage, userSettings }: LanguageSwitcherProps) {
  const { t } = useTranslation();
  const [currentLocale, setCurrentLocale] = useState(userSettings?.language || 'en');
  const [updateLanguage] = useUpdateLanguageMutation();

  useEffect(() => {
    if (userSettings?.language) {
      setCurrentLocale(userSettings.language);
    }
  }, [userSettings?.language]);

  const handleChange = async (locale: string) => {
    try {
      await updateLanguage({ language: locale }).unwrap();
      document.cookie = `NEXT_LOCALE=${locale}; path=/; max-age=31536000; SameSite=Lax`;
      setCurrentLocale(locale);
      handleLanguageChange(locale);
      window.location.reload();
    } catch (error) {
      console.error('Failed to update language:', error);
    }
  };

  return (
    <Select value={currentLocale} onValueChange={handleChange} disabled={isUpdatingLanguage}>
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
