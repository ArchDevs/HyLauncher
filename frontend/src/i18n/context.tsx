/**
 * i18n Context Provider
 * Manages language state and provides translations to the app
 * Logic is completely separated from UI
 */

import React, { createContext, useContext, useState, useEffect, useMemo } from "react";
import type { SupportedLanguage, I18nContextValue } from "./types";
import { translations, getDefaultLanguage } from "./translations";

const I18nContext = createContext<I18nContextValue | undefined>(undefined);

const LANGUAGE_STORAGE_KEY = "hylauncher_language";

interface I18nProviderProps {
  children: React.ReactNode;
  defaultLanguage?: SupportedLanguage;
}

export const I18nProvider: React.FC<I18nProviderProps> = ({
  children,
  defaultLanguage,
}) => {
  const [language, setLanguageState] = useState<SupportedLanguage>(() => {
    // Try to get from localStorage first
    if (typeof window !== "undefined") {
      const stored = localStorage.getItem(LANGUAGE_STORAGE_KEY) as SupportedLanguage;
      if (stored && translations[stored]) {
        return stored;
      }
    }
    // Use provided default or system default
    return defaultLanguage || getDefaultLanguage();
  });

  // Persist language choice to localStorage
  useEffect(() => {
    if (typeof window !== "undefined") {
      localStorage.setItem(LANGUAGE_STORAGE_KEY, language);
    }
  }, [language]);

  const setLanguage = (lang: SupportedLanguage) => {
    if (translations[lang]) {
      setLanguageState(lang);
    }
  };

  const value = useMemo<I18nContextValue>(
    () => ({
      language,
      setLanguage,
      t: translations[language],
    }),
    [language]
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
};

/**
 * Hook to access i18n context
 * Throws error if used outside I18nProvider
 */
export const useI18n = (): I18nContextValue => {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error("useI18n must be used within an I18nProvider");
  }
  return context;
};

