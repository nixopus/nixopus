export const defaultLocale = 'en';
export const locales = ['en', 'es', 'fr', 'kn'] as const;
export type Locale = (typeof locales)[number];

export const languageNames: Record<Locale, string> = {
  en: 'English',
  es: 'Español',
  fr: 'Français',
  kn: 'ಕನ್ನಡ'
};
