/**
 * i18n Module - Main Export
 * Centralized exports for the i18n system
 */

export { I18nProvider, useI18n } from "./context";
export { useTranslation, useLanguage, useSetLanguage } from "./hooks";
export { translations, getDefaultLanguage, getAvailableLanguages } from "./translations";
export type { Translations, SupportedLanguage, I18nContextValue } from "./types";

