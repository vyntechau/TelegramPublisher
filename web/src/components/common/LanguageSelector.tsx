import React, { useState, useRef, useEffect } from 'react';
import { useLanguage } from '../../context/LanguageContext';
import { Globe, ChevronDown, Check } from 'lucide-react';

interface LanguageSelectorProps {
  variant?: 'compact' | 'full' | 'dropdown';
  className?: string;
}

export const LanguageSelector: React.FC<LanguageSelectorProps> = ({
  variant = 'dropdown',
  className = '',
}) => {
  const { language, setLanguage, currentLanguageMeta, supportedLanguages, isRtl } = useLanguage();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (variant === 'compact') {
    return (
      <div className={`relative ${className}`} ref={dropdownRef}>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="liquid-pill px-2.5 py-1.5 rounded-xl text-slate-300 hover:text-white flex items-center gap-1.5 text-xs transition-all hover:scale-105 border border-white/[0.08]"
          title="Change Language"
        >
          <img
            src={`https://flagcdn.com/w40/${currentLanguageMeta.countryCode}.png`}
            alt={currentLanguageMeta.name}
            className="w-4 h-3 object-cover rounded-sm border border-white/20 shadow-xs"
            loading="lazy"
          />
          <span className="font-semibold">{currentLanguageMeta.code.toUpperCase()}</span>
          <ChevronDown size={11} className={`transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
        </button>

        {isOpen && (
          <div
            className={`absolute top-full mt-2 w-48 liquid-glass-card rounded-2xl border border-white/[0.12] shadow-2xl p-1.5 z-50 animate-in fade-in zoom-in-95 duration-150 backdrop-blur-3xl ${
              isRtl ? 'left-0' : 'right-0'
            }`}
          >
            {supportedLanguages.map((l) => {
              const active = l.code === language;
              return (
                <button
                  key={l.code}
                  type="button"
                  onClick={() => {
                    setLanguage(l.code);
                    setIsOpen(false);
                  }}
                  className={`w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs font-medium transition-all ${
                    active
                      ? 'bg-cyan-500/20 text-cyan-300 font-semibold'
                      : 'text-slate-300 hover:bg-white/[0.06] hover:text-white'
                  }`}
                >
                  <div className="flex items-center gap-2.5">
                    <img
                      src={`https://flagcdn.com/w40/${l.countryCode}.png`}
                      alt={l.name}
                      className="w-4 h-3 object-cover rounded-sm border border-white/20 shadow-xs shrink-0"
                      loading="lazy"
                    />
                    <span>{l.nativeName}</span>
                  </div>
                  {active && <Check size={13} className="text-cyan-400" />}
                </button>
              );
            })}
          </div>
        )}
      </div>
    );
  }

  // Full / Dropdown default
  return (
    <div className={`relative ${className}`} ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="liquid-glass px-3 py-1.5 rounded-xl text-slate-300 hover:text-white flex items-center gap-2 text-xs transition-all hover:border-cyan-500/40 border border-white/[0.08] shadow-sm"
      >
        <Globe size={13} className="text-cyan-400 shrink-0" />
        <img
          src={`https://flagcdn.com/w40/${currentLanguageMeta.countryCode}.png`}
          alt={currentLanguageMeta.name}
          className="w-4 h-3 object-cover rounded-sm border border-white/20 shadow-xs shrink-0"
          loading="lazy"
        />
        <span className="font-medium hidden sm:inline">{currentLanguageMeta.nativeName}</span>
        <ChevronDown size={12} className={`text-slate-400 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`} />
      </button>

      {isOpen && (
        <div
          className={`absolute top-full mt-2 w-52 liquid-glass-card rounded-2xl border border-white/[0.12] shadow-2xl p-1.5 z-50 backdrop-blur-3xl animate-in fade-in zoom-in-95 duration-150 ${
            isRtl ? 'left-0' : 'right-0'
          }`}
        >
          <div className="px-2 py-1 text-[10px] uppercase font-bold tracking-wider text-slate-400 border-b border-white/[0.06] mb-1">
            Languages / زبان‌ها
          </div>
          <div className="max-h-60 overflow-y-auto space-y-0.5 custom-scrollbar">
            {supportedLanguages.map((l) => {
              const active = l.code === language;
              return (
                <button
                  key={l.code}
                  type="button"
                  onClick={() => {
                    setLanguage(l.code);
                    setIsOpen(false);
                  }}
                  className={`w-full flex items-center justify-between px-3 py-2 rounded-xl text-xs transition-all text-start ${
                    active
                      ? 'bg-cyan-500/20 text-cyan-300 font-semibold shadow-inner'
                      : 'text-slate-300 hover:bg-white/[0.06] hover:text-white'
                  }`}
                >
                  <div className="flex items-center gap-2.5">
                    <img
                      src={`https://flagcdn.com/w40/${l.countryCode}.png`}
                      alt={l.name}
                      className="w-4 h-3 object-cover rounded-sm border border-white/20 shadow-xs shrink-0"
                      loading="lazy"
                    />
                    <div className="flex flex-col">
                      <span className="leading-tight">{l.nativeName}</span>
                      <span className="text-[10px] text-slate-400 leading-tight">{l.name}</span>
                    </div>
                  </div>
                  {active && <Check size={14} className="text-cyan-400 shrink-0 ms-1" />}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};
