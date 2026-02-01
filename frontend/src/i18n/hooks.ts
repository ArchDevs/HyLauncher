/**
 * i18n Hooks
 * Convenience hooks for using translations
 * Keeps translation logic separate from UI components
 */

import { useI18n } from "./context";
import type { SupportedLanguage } from "./types";

/**
 * Main translation hook
 * Returns the translation function and current language
 */
export const useTranslation = () => {
  const { t, language, setLanguage } = useI18n();
  return { t, language, setLanguage };
};

/**
 * Hook to get current language only
 */
export const useLanguage = (): SupportedLanguage => {
  const { language } = useI18n();
  return language;
};

/**
 * Hook to change language
 */
export const useSetLanguage = () => {
  const { setLanguage } = useI18n();
  return setLanguage;
};

