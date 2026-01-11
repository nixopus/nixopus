'use client';

import { locales } from '@/lib/i18n/config';
import { SelectWrapper, SelectOption } from '@/components/ui/select-wrapper';
import { useTranslation } from '@/hooks/use-translation';
import { useEffect, useState } from 'react';
import { useUpdateLanguageMutation } from '@/redux/services/users/userApi';
import { UserSettings } from '@/redux/types/user';

interface LanguageSwitcherProps {
  handleLanguageChange: (language: string) => void;
  isUpdatingLanguage: boolean;
  userSettings: UserSettings;
}

export function LanguageSwitcher({
  handleLanguageChange,
  isUpdatingLanguage,
  userSettings
}: LanguageSwitcherProps) {
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

  const languageOptions: SelectOption[] = locales.map((locale) => ({
    value: locale,
    label: t(`languages.${locale}`)
  }));

  return (
    <SelectWrapper
      value={currentLocale}
      onValueChange={handleChange}
      options={languageOptions}
      placeholder={t('settings.preferences.language.select')}
      disabled={isUpdatingLanguage}
      className="w-[180px]"
    />
  );
}
