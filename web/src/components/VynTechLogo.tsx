import React from 'react';
import { useLanguage } from '../context/LanguageContext';

interface VynTechLogoProps {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  showWordmark?: boolean;
  className?: string;
  animated?: boolean;
  localized?: boolean;
}

/**
 * Official VynTech Brand Logo & Icon
 * Exact vector geometry from D:/ws/vyntech/web codebase
 * ViewBox: 327.57 451.64 198.91 175.55 (clean, borderless, shadowless)
 */
export const VynTechLogo: React.FC<VynTechLogoProps> = ({
  size = 'md',
  showWordmark = true,
  className = '',
  animated = false,
  localized = true,
}) => {
  let t = (key: string, fallback: string) => fallback;
  let isRtl = false;
  try {
    const lang = useLanguage();
    if (lang) {
      t = lang.t;
      isRtl = lang.isRtl;
    }
  } catch {
    // In case logo is rendered outside of LanguageProvider
  }

  const sizeMap = {
    sm: { icon: 'w-6 h-6', text: 'text-xs', badge: 'text-[9px] px-1.5' },
    md: { icon: 'w-8 h-8', text: 'text-sm', badge: 'text-[10px] px-2' },
    lg: { icon: 'w-11 h-11', text: 'text-lg', badge: 'text-xs px-2.5' },
    xl: { icon: 'w-16 h-16', text: 'text-2xl', badge: 'text-xs px-3' },
  };

  const current = sizeMap[size];

  const brandName = localized ? t('brand.name', 'TelegramPublisher') : 'TelegramPublisher';
  const brandCompany = localized ? t('brand.company', 'VynTech Cloud') : 'VynTech Cloud';
  const brandVersion = localized ? t('brand.version', 'v1.0') : 'v1.0';

  return (
    <div className={`flex items-center gap-2.5 sm:gap-3 select-none text-start ${className}`}>
      {/* Clean Vector SVG Icon without border or shadow */}
      <svg
        viewBox="327.57 451.64 198.91 175.55"
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className={`${current.icon} shrink-0 ${animated ? 'transition-transform duration-300 hover:scale-105' : ''}`}
        aria-hidden="true"
      >
        <defs>
          <linearGradient id="vyn_tg_clean_1" x1="430.18" y1="516.59" x2="463.77" y2="562.12" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor="#0071e3" />
            <stop offset="100%" stopColor="#464b9e" />
          </linearGradient>
          <linearGradient id="vyn_tg_clean_2" x1="394.49" y1="508.56" x2="479.85" y2="518.68" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor="#00a8e8" />
            <stop offset="100%" stopColor="#464b9e" />
          </linearGradient>
          <linearGradient id="vyn_tg_clean_3" x1="449.9" y1="497.32" x2="450.12" y2="448.67" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor="#0071e3" />
            <stop offset="100%" stopColor="#464b9e" />
          </linearGradient>
          <linearGradient id="vyn_tg_clean_4" x1="399.49" y1="496.26" x2="360.96" y2="518" gradientUnits="userSpaceOnUse">
            <stop offset="0%" stopColor="#00a8e8" />
            <stop offset="100%" stopColor="#464b9e" />
          </linearGradient>
        </defs>

        <path
          fill="url(#vyn_tg_clean_1)"
          d="M444.72 504.44l-25.84 44.42c-5.01 8.43-7.7 17.98-9.14 26.11-1.9 10.73.04 21.78 5.42 31.26l11.87 20.96 73.62-129.94c-8.87 4.26-18.68 6.56-28.93 6.73L444.72 504.44z"
        />
        <polygon fill="url(#vyn_tg_clean_2)" points="426.09,521.15 435.72,504.6 416.16,504.93" />
        <path
          fill="url(#vyn_tg_clean_3)"
          d="M526.48 451.65l-13.57 23.94c-3.1 5.49-7.78 9.91-13.43 12.68-7.49 3.68-18.36 8.01-27.9 8.01l-22.32.39.01-.04-.02.04h.02l-.02.15-37.25.49-.49.01-11.82-19.32c-6.59-10.77-15.54-19.81-25.94-26.35H526.48z"
        />
        <path
          fill="url(#vyn_tg_clean_4)"
          d="M400.72 580.75c1.12-12.44 5.04-24.67 11.5-35.77l9.46-16.27-28.59-46.69c-6.51-11.1-18.2-19.72-26.66-24.89-5.86-3.58-12.58-5.49-19.45-5.49h-19.41L400.72 580.75z"
        />
      </svg>

      {/* Wordmark and Subtitle */}
      {showWordmark && (
        <div className="flex flex-col text-start">
          <div className={`font-extrabold tracking-tight text-white flex items-center gap-1.5 sm:gap-2 ${current.text}`}>
            <span className="truncate">{brandName}</span>
            <span className={`liquid-pill text-cyan-300 rounded-full font-semibold border border-cyan-400/30 whitespace-nowrap shrink-0 ${current.badge}`}>
              {brandVersion}
            </span>
          </div>
          <div className="text-[10px] sm:text-[11px] text-slate-400 font-medium tracking-tight">
            <span>{brandCompany}</span>
          </div>
        </div>
      )}
    </div>
  );
};
