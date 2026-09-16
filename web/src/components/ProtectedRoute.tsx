import React from 'react';
import { Navigate, Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { useTranslation } from '../context/LanguageContext';
import { ShieldAlert, ArrowLeft, Lock, Sparkles, CheckCircle2 } from 'lucide-react';
import { VynTechLogo } from './VynTechLogo';

interface ProtectedRouteProps {
  allowedRoles: Array<'user' | 'author' | 'admin' | 'owner'>;
  children: React.ReactNode;
}

export const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ allowedRoles, children }) => {
  const { role, setRole, isLoading } = useAuth();
  const { t } = useTranslation();

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="liquid-glass p-8 rounded-3xl border border-white/10 flex flex-col items-center gap-4 text-center">
          <div className="w-12 h-12 rounded-2xl bg-cyan-500/10 border border-cyan-500/30 flex items-center justify-center animate-spin">
            <Sparkles size={24} className="text-cyan-400" />
          </div>
          <div className="text-sm font-semibold text-slate-200">{t('auth.verifying', 'Verifying session permissions...')}</div>
        </div>
      </div>
    );
  }

  const isAuthorized = allowedRoles.includes(role);

  if (!isAuthorized) {
    return (
      <div className="flex items-center justify-center min-h-[70vh] px-4">
        <div className="liquid-glass-card max-w-lg w-full p-8 md:p-10 rounded-3xl border border-red-500/20 text-center relative overflow-hidden shadow-2xl">
          {/* Ambient Glow */}
          <div className="ambient-orb w-64 h-64 bg-red-600/15 -top-20 -right-20 pointer-events-none" />
          <div className="ambient-orb w-64 h-64 bg-blue-600/10 -bottom-20 -left-20 pointer-events-none" />

          {/* VynTech Logo Badge */}
          <div className="flex justify-center mb-6">
            <VynTechLogo size="sm" showWordmark={false} />
          </div>

          {/* Shield Icon */}
          <div className="w-16 h-16 rounded-2xl bg-gradient-to-tr from-red-500/20 via-orange-500/10 to-red-500/5 border border-red-500/30 flex items-center justify-center mx-auto mb-5 shadow-lg shadow-red-500/10">
            <ShieldAlert size={32} className="text-red-400 animate-pulse" />
          </div>

          <h2 className="text-2xl font-black text-white mb-2 tracking-tight">{t('auth.restricted_title', 'Access Restricted')}</h2>
          <p className="text-slate-300 text-sm mb-6 leading-relaxed">
            {t('auth.restricted_desc', { role })}
          </p>

          <div className="liquid-glass p-4 rounded-2xl border border-white/10 text-start mb-6">
            <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2 flex items-center gap-1.5">
              <Lock size={12} className="text-amber-400" />
              <span>{t('auth.required_privileges', 'Required Privileges')}</span>
            </div>
            <div className="flex flex-wrap gap-2">
              {allowedRoles.map((r) => (
                <span
                  key={r}
                  className="text-xs px-2.5 py-1 rounded-xl bg-white/5 border border-white/10 text-slate-200 capitalize font-medium flex items-center gap-1"
                >
                  <CheckCircle2 size={12} className="text-emerald-400" />
                  {r}
                </span>
              ))}
            </div>
          </div>

          {/* Sandbox Role Switcher for quick test/demo */}
          <div className="liquid-glass p-3.5 rounded-2xl border border-white/10 mb-6">
            <div className="text-[11px] font-semibold text-slate-400 mb-2">{t('auth.simulate_role', 'Simulate Authorized Role (Sandbox Mode):')}</div>
            <div className="grid grid-cols-3 gap-2">
              {allowedRoles.map((r) => (
                <button
                  key={r}
                  onClick={() => setRole(r)}
                  className="py-1.5 px-3 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 text-white text-xs font-semibold hover:brightness-110 active:scale-95 transition-all shadow-md shadow-blue-500/20 capitalize"
                >
                  {t('auth.switch_to', { role: r })}
                </button>
              ))}
            </div>
          </div>

          {/* Return Home Button */}
          <Link
            to="/catalog"
            className="liquid-button inline-flex items-center justify-center gap-2 w-full py-3 rounded-xl text-sm font-bold text-white shadow-lg"
          >
            <ArrowLeft size={16} className="rtl:rotate-180" />
            <span>{t('auth.return_catalog', 'Return to Media Catalog')}</span>
          </Link>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};
