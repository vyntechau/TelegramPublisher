import React from 'react';
import { NavLink } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { useTranslation } from '../../context/LanguageContext';
import { VynTechLogo } from '../VynTechLogo';
import { API_ENDPOINTS } from '../../constants';
import { LanguageSelector } from './LanguageSelector';
import { BookOpen, Terminal } from 'lucide-react';

export const Header: React.FC = () => {
  const { role, user } = useAuth();
  const { t } = useTranslation();

  const roleLabels: Record<string, string> = {
    owner: t('nav.owner', 'Owner'),
    admin: t('nav.admin_role', 'Administrator'),
    author: t('nav.author_role', 'Author'),
    user: t('nav.user_role', 'Subscriber'),
  };

  const displayName = user?.first_name || roleLabels[role] || t('nav.guest', 'Guest');
  const displayUsername = user?.username ? `@${user.username}` : `@${role}_account`;
  const initial = displayName.charAt(0).toUpperCase();

  const roleColors: Record<string, { bg: string; text: string; border: string }> = {
    owner: { bg: 'bg-purple-500/15', text: 'text-purple-300', border: 'border-purple-500/30' },
    admin: { bg: 'bg-cyan-500/15', text: 'text-cyan-300', border: 'border-cyan-500/30' },
    author: { bg: 'bg-indigo-500/15', text: 'text-indigo-300', border: 'border-indigo-500/30' },
    user: { bg: 'bg-emerald-500/15', text: 'text-emerald-300', border: 'border-emerald-500/30' },
  };

  const currentRoleStyle = roleColors[role] || roleColors.user;

  return (
    <header className="sticky top-0 z-50 liquid-glass border-b border-white/[0.08] backdrop-blur-2xl">
      <div className="max-w-7xl mx-auto px-3 sm:px-6 h-16 flex items-center justify-between gap-2 sm:gap-4">
        {/* Logo & Brand */}
        <NavLink to="/" className="group transition-transform active:scale-95 shrink-0">
          <VynTechLogo size="md" />
        </NavLink>

        {/* Right Tools: Language Switcher, Interactive Docs & Real Account Info */}
        <div className="flex items-center gap-2 sm:gap-3">
          {/* Language Switcher */}
          <LanguageSelector variant="compact" />

          {/* Interactive Docs */}
          <div className="hidden md:flex items-center gap-2 text-xs">
            <a 
              href={API_ENDPOINTS.DOCS_SCALAR}
              target="_blank" 
              rel="noreferrer"
              className="liquid-pill px-3 py-1.5 rounded-xl text-slate-300 hover:text-white font-medium flex items-center gap-1.5 text-xs transition-all hover:scale-105"
            >
              <BookOpen size={13} className="text-cyan-400" /> 
              <span>{t('nav.scalar_api', 'Scalar API')}</span>
            </a>
            <a 
              href={API_ENDPOINTS.GRAPHQL_PLAYGROUND}
              target="_blank" 
              rel="noreferrer"
              className="liquid-pill px-3 py-1.5 rounded-xl text-slate-300 hover:text-white font-medium flex items-center gap-1.5 text-xs transition-all hover:scale-105"
            >
              <Terminal size={13} className="text-purple-400" /> 
              <span>{t('nav.graphql', 'GraphQL')}</span>
            </a>
          </div>

          {/* Account Information Card */}
          <div className="flex items-center gap-2 sm:gap-2.5 liquid-glass px-2.5 sm:px-3 py-1.5 rounded-2xl border border-white/[0.08] shadow-sm">
            <div className="relative">
              <div className="w-6 h-6 sm:w-7 sm:h-7 rounded-xl bg-gradient-to-tr from-cyan-500 to-blue-600 flex items-center justify-center font-bold text-xs text-white shadow-sm">
                {initial}
              </div>
              <span className="absolute -bottom-0.5 ltr:-right-0.5 rtl:-left-0.5 w-2 h-2 rounded-full bg-emerald-400 border border-[#060911]" />
            </div>

            <div className="flex flex-col text-start">
              <div className="text-[11px] sm:text-xs font-semibold text-white tracking-tight leading-tight max-w-[90px] sm:max-w-none truncate">
                {displayName}
              </div>
              <div className="text-[9px] sm:text-[10px] text-slate-400 leading-none max-w-[90px] sm:max-w-none truncate">
                {displayUsername}
              </div>
            </div>

            <div className={`text-[9px] sm:text-[10px] uppercase font-bold px-1.5 sm:px-2 py-0.5 rounded-lg border tracking-wider ms-0.5 sm:ms-1 ${currentRoleStyle.bg} ${currentRoleStyle.text} ${currentRoleStyle.border}`}>
              {role}
            </div>
          </div>
        </div>
      </div>
    </header>
  );
};
