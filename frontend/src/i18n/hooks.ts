/**
 * i18n Hooks
 * Convenience hooks for using translations
 * Keeps translation logic separate from UI components
 */

import type React from "react";
import { useI18n } from "./context";
import type { SupportedLanguage, Translations } from "./types";

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
export type PageConfigBase = {
  id: string;
  nameKey: keyof Translations["pages"];
  icon: React.ComponentType<{ size?: number | string }>;
  component: React.ComponentType;
  background?: React.ComponentType;
};
