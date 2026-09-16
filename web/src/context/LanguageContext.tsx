import React, { createContext, useContext, useState, useEffect, useCallback, useMemo } from 'react';
import { ALL_LANGUAGES, LanguageMeta, translations, en } from '../locales';

interface LanguageContextType {
  language: string;
  setLanguage: (lang: string) => void;
  t: (path: string, fallbackOrParams?: any) => string;
  currentLanguageMeta: LanguageMeta;
  supportedLanguages: LanguageMeta[];
  setSupportedLanguagesCSV: (csv: string) => void;
  defaultLanguage: string;
  setDefaultLanguage: (lang: string) => void;
  isRtl: boolean;
  dir: 'rtl' | 'ltr';
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

const STORAGE_KEY = 'app_language';

export const LanguageProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [supportedCSV, setSupportedCSV] = useState<string>(() => {
    return localStorage.getItem('app_supported_languages') || 'en,fa,ar,ru,es,de,zh';
  });

  const [defaultLanguage, setDefaultLanguageState] = useState<string>(() => {
    return localStorage.getItem('app_default_language') || 'en';
  });

  const [language, setLanguageState] = useState<string>(() => {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (saved && translations[saved]) {
      return saved;
    }
    // Check browser languages
    const navLang = navigator.language?.slice(0, 2)?.toLowerCase();
    if (navLang && translations[navLang]) {
      return navLang;
    }
    return defaultLanguage;
  });

  const supportedLanguages = useMemo(() => {
    if (!supportedCSV) return ALL_LANGUAGES;
    const allowed = new Set(
      supportedCSV
        .split(',')
        .map((s) => s.trim().toLowerCase())
        .filter(Boolean)
    );
    const filtered = ALL_LANGUAGES.filter((l) => allowed.has(l.code));
    return filtered.length > 0 ? filtered : ALL_LANGUAGES;
  }, [supportedCSV]);

  const currentLanguageMeta = useMemo(() => {
    return ALL_LANGUAGES.find((l) => l.code === language) || ALL_LANGUAGES[0];
  }, [language]);

  const isRtl = currentLanguageMeta.isRtl;
  const dir: 'rtl' | 'ltr' = isRtl ? 'rtl' : 'ltr';

  // Apply direction and html lang dynamically
  useEffect(() => {
    document.documentElement.setAttribute('dir', dir);
    document.documentElement.setAttribute('lang', language);
    if (isRtl) {
      document.documentElement.classList.add('rtl');
    } else {
      document.documentElement.classList.remove('rtl');
    }
  }, [dir, language, isRtl]);

  const setLanguage = useCallback((newLang: string) => {
    if (translations[newLang]) {
      setLanguageState(newLang);
      localStorage.setItem(STORAGE_KEY, newLang);
    }
  }, []);

  const setSupportedLanguagesCSV = useCallback((csv: string) => {
    setSupportedCSV(csv);
    localStorage.setItem('app_supported_languages', csv);
  }, []);

  const setDefaultLanguage = useCallback((lang: string) => {
    if (translations[lang]) {
      setDefaultLanguageState(lang);
      localStorage.setItem('app_default_language', lang);
    }
  }, []);

  // Helper to retrieve deeply nested keys like 'nav.home' or 'admin.kpi_users'
  const t = useCallback(
    (path: string, fallbackOrParams?: any): string => {
      const keys = path.split('.');
      const currentDict = translations[language] || translations.en;
      const fallbackDict = translations.en;

      const getVal = (dict: any): any => {
        let curr = dict;
        for (const k of keys) {
          if (curr && typeof curr === 'object' && k in curr) {
            curr = curr[k];
          } else {
            return undefined;
          }
        }
        return curr;
      };

      let val = getVal(currentDict);
      if (val === undefined) {
        val = getVal(fallbackDict);
      }

      if (val === undefined || typeof val !== 'string') {
        if (typeof fallbackOrParams === 'string') {
          return fallbackOrParams;
        }
        return path;
      }

      // Handle simple parameter replacement if object passed { count: 5 }
      if (typeof fallbackOrParams === 'object' && fallbackOrParams !== null) {
        let formatted = val;
        Object.keys(fallbackOrParams).forEach((k) => {
          formatted = formatted.replace(new RegExp(`{${k}}`, 'g'), String(fallbackOrParams[k]));
        });
        return formatted;
      }

      return val;
    },
    [language]
  );

  return (
    <LanguageContext.Provider
      value={{
        language,
        setLanguage,
        t,
        currentLanguageMeta,
        supportedLanguages,
        setSupportedLanguagesCSV,
        defaultLanguage,
        setDefaultLanguage,
        isRtl,
        dir,
      }}
    >
      {children}
    </LanguageContext.Provider>
  );
};

export const useLanguage = (): LanguageContextType => {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};

export const useTranslation = () => {
  const { t, language, isRtl, dir } = useLanguage();
  return { t, language, isRtl, dir };
};
