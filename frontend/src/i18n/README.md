# i18n System Documentation

This is a modular, scalable internationalization (i18n) system that keeps all translation logic completely separate from UI components.

## Architecture

The i18n system is organized into clear, separate modules:

- **`types.ts`** - TypeScript type definitions for translations and language support
- **`translations/`** - Translation files for each language (en.ts, ru.ts, etc.)
- **`context.tsx`** - React context provider for language state management
- **`hooks.ts`** - React hooks for accessing translations in components
- **`index.ts`** - Main export file

## Adding a New Language

To add a new language, follow these simple steps:

### 1. Add the language code to types

Edit `types.ts` and add the new language to `SupportedLanguage`:

```typescript
export type SupportedLanguage = "en" | "ru" | "de"; // Add "de" for German
```

### 2. Create a translation file

Create a new file in `translations/` directory, e.g., `de.ts`:

```typescript
import type { Translations } from "../types";

export const de: Translations = {
  common: {
    play: "SPIELEN",
    install: "INSTALLIEREN...",
    // ... copy structure from en.ts and translate
  },
  // ... rest of translations
};
```

### 3. Register the translation

Edit `translations/index.ts` and add your new language:

```typescript
import { de } from "./de";

export const translations: Record<SupportedLanguage, Translations> = {
  en,
  ru,
  de, // Add here
};
```

### 4. Add language name (optional)

Edit `components/LanguageSwitcher.tsx` and add the display name:

```typescript
const languageNames: Record<SupportedLanguage, string> = {
  en: "English",
  ru: "Русский",
  de: "Deutsch", // Add here
};
```

That's it! The new language will automatically appear in the language switcher.

## Using Translations in Components

### Basic Usage

```typescript
import { useTranslation } from "../i18n";

function MyComponent() {
  const { t } = useTranslation();
  
  return <button>{t.common.play}</button>;
}
```

### Changing Language

```typescript
import { useTranslation } from "../i18n";

function MyComponent() {
  const { t, language, setLanguage } = useTranslation();
  
  return (
    <button onClick={() => setLanguage("ru")}>
      Switch to Russian
    </button>
  );
}
```

## Principles

1. **Separation of Concerns**: All translation logic is in the `i18n/` directory, completely separate from UI
2. **Type Safety**: Full TypeScript support ensures all translation keys are valid
3. **Scalability**: Easy to add new languages by following the 4-step process above
4. **Performance**: Translations are memoized and only re-render when language changes
5. **Persistence**: Language choice is saved to localStorage automatically

## File Structure

```
i18n/
├── types.ts              # Type definitions
├── context.tsx           # React context provider
├── hooks.ts              # React hooks
├── index.ts              # Main exports
├── translations/
│   ├── index.ts          # Translation registry
│   ├── en.ts             # English translations
│   └── ru.ts             # Russian translations
└── README.md             # This file
```

