import { en } from './en';
import { fa } from './fa';
import { ar } from './ar';
import { ru } from './ru';
import { es } from './es';
import { de } from './de';
import { zh } from './zh';

export { en, fa, ar, ru, es, de, zh };


export interface LanguageMeta {
  code: string;
  countryCode: string;
  name: string;
  nativeName: string;
  flag: string;
  flagUrl: string;
  isRtl: boolean;
}

export const ALL_LANGUAGES: LanguageMeta[] = [
  { code: 'en', countryCode: 'us', name: 'English', nativeName: 'English', flag: '🇺🇸', flagUrl: 'https://flagcdn.com/w40/us.png', isRtl: false },
  { code: 'fa', countryCode: 'ir', name: 'Persian', nativeName: 'فارسی', flag: '🇮🇷', flagUrl: 'https://flagcdn.com/w40/ir.png', isRtl: true },
  { code: 'ar', countryCode: 'sa', name: 'Arabic', nativeName: 'العربية', flag: '🇸🇦', flagUrl: 'https://flagcdn.com/w40/sa.png', isRtl: true },
  { code: 'ru', countryCode: 'ru', name: 'Russian', nativeName: 'Русский', flag: '🇷🇺', flagUrl: 'https://flagcdn.com/w40/ru.png', isRtl: false },
  { code: 'es', countryCode: 'es', name: 'Spanish', nativeName: 'Español', flag: '🇪🇸', flagUrl: 'https://flagcdn.com/w40/es.png', isRtl: false },
  { code: 'de', countryCode: 'de', name: 'German', nativeName: 'Deutsch', flag: '🇩🇪', flagUrl: 'https://flagcdn.com/w40/de.png', isRtl: false },
  { code: 'zh', countryCode: 'cn', name: 'Chinese', nativeName: '中文', flag: '🇨🇳', flagUrl: 'https://flagcdn.com/w40/cn.png', isRtl: false },
];

export const translations: Record<string, typeof en> = {
  en,
  fa,
  ar,
  ru,
  es,
  de,
  zh,
};

export type LocaleKey = keyof typeof en;
export type TranslationsType = typeof en;
