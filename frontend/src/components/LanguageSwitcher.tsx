/**
 * Language Switcher Component
 * UI component for changing language
 * All translation logic is in i18n module
 */

import React, { useState, useRef, useEffect } from "react";
import { Globe, Check } from "lucide-react";
import { useTranslation, getAvailableLanguages, type SupportedLanguage } from "../i18n";

interface LanguageSwitcherProps {
  className?: string;
}

const languageNames: Record<SupportedLanguage, string> = {
  en: "English",
  ru: "Русский",
};

export const LanguageSwitcher: React.FC<LanguageSwitcherProps> = ({ className = "" }) => {
  const { language, setLanguage, t } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);
  const availableLanguages = getAvailableLanguages();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("keydown", handleEscape);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("keydown", handleEscape);
    };
  }, [isOpen]);

  const handleLanguageChange = (lang: SupportedLanguage) => {
    setLanguage(lang);
    setIsOpen(false);
  };

  return (
    <div className="relative pointer-events-auto" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        style={{ WebkitAppRegion: "no-drag" } as React.CSSProperties}
        className="
          w-[48px] h-[48px]
          bg-[#090909]/55 backdrop-blur-[12px]
          rounded-[14px] border border-[#7C7C7C]/[0.10]
          cursor-pointer
          flex items-center justify-center
          hover:bg-[#090909]/70
          transition-all duration-150
          z-50 pointer-events-auto
        "
        title={languageNames[language]}
      >
        <Globe size={18} className="text-[#CCD9E0]/[0.90]" />
      </button>

      {isOpen && (
        <div
          className="
            absolute top-[56px] left-0
            w-[180px]
            bg-[#090909]/[0.75] backdrop-blur-[12px]
            rounded-[20px]
            border border-[#7C7C7C]/[0.10]
            overflow-hidden
            z-[100]
            shadow-xl
          "
        >
          {availableLanguages.map((lang, idx) => (
            <button
              key={lang}
              onClick={() => handleLanguageChange(lang)}
              className={`
                w-full h-[56px] px-[18px]
                flex items-center justify-between
                text-[#CCD9E0]/[0.90] text-[16px] font-[Mazzard]
                hover:bg-white/[0.05]
                cursor-pointer transition
                ${idx !== availableLanguages.length - 1 ? "border-b border-white/10" : ""}
              `}
            >
              <span>{languageNames[lang]}</span>
              {lang === language && <Check size={16} className="text-[#FFA845]" />}
            </button>
          ))}
        </div>
      )}
    </div>
  );
};

