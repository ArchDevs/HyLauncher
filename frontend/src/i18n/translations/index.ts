/**
 * Translations index
 * Centralized export for all translation files
 * Easy to add new languages by importing and adding to the map
 */

import { en } from "./en";
import { ru } from "./ru";
import type { Translations, SupportedLanguage } from "../types";

export const translations: Record<SupportedLanguage, Translations> = {
  en,
  ru,
};

/**
 * Get default language based on browser/system preferences
 */
export const getDefaultLanguage = (): SupportedLanguage => {
  if (typeof window === "undefined") return "en";
  
  const browserLang = navigator.language.toLowerCase();
  if (browserLang.startsWith("ru")) return "ru";
  return "en";
};

/**
 * Get available languages
 */
export const getAvailableLanguages = (): SupportedLanguage[] => {
  return Object.keys(translations) as SupportedLanguage[];
};

