import React from 'react';
import { Link } from 'react-router-dom';
import { API_ENDPOINTS, VYNTECH_URLS } from '../../constants';
import { useTranslation } from '../../context/LanguageContext';
import { Github, BookOpen, Info, Sparkles } from 'lucide-react';

export const Footer: React.FC = () => {
  const { t } = useTranslation();

  return (
    <footer className="border-t border-white/[0.06] py-4 sm:py-5 liquid-glass text-xs text-slate-400 relative z-10 mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 flex flex-col sm:flex-row items-center justify-between gap-3 text-center sm:text-start">
        {/* Brand Copyright */}
        <div className="flex items-center gap-2">
          <span className="w-1.5 h-1.5 rounded-full bg-cyan-400" />
          <span className="text-slate-300 font-medium">{t('brand.company', 'VynTech Cloud')}</span>
          <span className="text-slate-600">•</span>
          <span className="text-slate-400">{t('brand.name', 'TelegramPublisher')} {t('brand.version', 'v1.0')}</span>
        </div>

        {/* Minimal Navigation Links */}
        <div className="flex flex-wrap items-center justify-center gap-3 sm:gap-5 text-xs">
          <Link
            to="/setup"
            className="text-slate-400 hover:text-cyan-300 transition-colors flex items-center gap-1.5"
          >
            <Sparkles size={13} className="text-cyan-400" />
            <span>{t('nav.setup', 'Setup Wizard')}</span>
          </Link>

          <Link
            to="/about"
            className="text-slate-400 hover:text-white transition-colors flex items-center gap-1.5"
          >
            <Info size={13} />
            <span>{t('nav.about', 'About')}</span>
          </Link>

          <a
            href={VYNTECH_URLS.GITHUB_REPO}
            target="_blank"
            rel="noreferrer"
            className="text-slate-400 hover:text-white transition-colors flex items-center gap-1.5"
          >
            <Github size={13} />
            <span>GitHub</span>
          </a>

          <a
            href={API_ENDPOINTS.DOCS_SCALAR}
            target="_blank"
            rel="noreferrer"
            className="text-slate-400 hover:text-white transition-colors flex items-center gap-1.5"
          >
            <BookOpen size={13} />
            <span>{t('about.btn_api_docs', 'API Docs')}</span>
          </a>
        </div>
      </div>
    </footer>
  );
};
