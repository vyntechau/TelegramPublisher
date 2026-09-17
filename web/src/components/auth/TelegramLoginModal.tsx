import React, { useState } from 'react';
import { useAuth } from '../../context/AuthContext';
import { useTranslation } from '../../context/LanguageContext';
import { X, Send, KeyRound, RefreshCw, CheckCircle2, AlertCircle, Sparkles, ExternalLink } from 'lucide-react';

interface TelegramLoginModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSuccess?: () => void;
}

export const TelegramLoginModal: React.FC<TelegramLoginModalProps> = ({ isOpen, onClose, onSuccess }) => {
  const { loginWithToken } = useAuth();
  const { t } = useTranslation();
  const [tokenInput, setTokenInput] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [errorMsg, setErrorMsg] = useState<string>('');
  const [successMsg, setSuccessMsg] = useState<string>('');

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const cleanToken = tokenInput.trim();
    if (!cleanToken) return;

    setErrorMsg('');
    setSuccessMsg('');
    setIsLoading(true);

    try {
      const success = await loginWithToken(cleanToken);
      if (success) {
        setSuccessMsg(t('auth.login_success', 'Authenticated successfully! Redirecting...'));
        setTimeout(() => {
          onSuccess?.();
          onClose();
        }, 600);
      } else {
        setErrorMsg(t('auth.invalid_token', 'Invalid or expired session token. Please verify your token.'));
      }
    } catch {
      setErrorMsg(t('auth.conn_failed', 'Failed to communicate with authentication server.'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-md animate-in fade-in duration-200">
      <div 
        className="liquid-glass-card max-w-lg w-full p-6 sm:p-8 rounded-3xl border border-cyan-500/30 shadow-2xl relative overflow-hidden space-y-6"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Ambient Glow Orbs */}
        <div className="ambient-orb w-60 h-60 bg-[#229ED9]/15 -top-20 -right-20 pointer-events-none" />
        <div className="ambient-orb w-48 h-48 bg-indigo-600/10 -bottom-20 -left-20 pointer-events-none" />

        {/* Modal Header */}
        <div className="flex items-start justify-between gap-3 relative z-10">
          <div className="flex items-center gap-3">
            <div className="w-12 h-12 rounded-2xl bg-[#229ED9]/15 border border-[#229ED9]/40 flex items-center justify-center shadow-lg shadow-[#229ED9]/10">
              <Send size={22} className="text-[#229ED9] fill-[#229ED9] -rotate-12 translate-x-0.5" />
            </div>
            <div>
              <h2 className="text-lg sm:text-xl font-bold text-white tracking-tight flex items-center gap-2">
                <span>{t('auth.telegram_login_title', 'Login with Telegram')}</span>
                <span className="text-[10px] bg-cyan-500/15 text-cyan-300 px-2 py-0.5 rounded-full border border-cyan-500/30 font-medium">
                  Web Access
                </span>
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                {t('auth.telegram_login_desc', 'Connect your Telegram account to access administrator and author studio controls.')}
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-2 rounded-xl text-slate-400 hover:text-white hover:bg-white/5 transition-all"
            aria-label="Close"
          >
            <X size={18} />
          </button>
        </div>

        {/* Options Container */}
        <div className="space-y-4 relative z-10">
          {/* Option 1: 1-Click Link from Bot */}
          <div className="p-4 rounded-2xl bg-white/[0.03] border border-white/10 space-y-2.5">
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-cyan-300 flex items-center gap-1.5">
                <Sparkles size={14} className="text-cyan-400" />
                <span>{t('auth.option1_title', 'Option 1: Quick 1-Click Bot Login')}</span>
              </span>
              <span className="text-[10px] text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded-md border border-emerald-500/20 font-medium">
                {t('auth.recommended', 'Recommended')}
              </span>
            </div>

            <ol className="text-xs text-slate-300 space-y-1.5 list-decimal list-inside leading-relaxed">
              <li>{t('auth.step_open_bot', 'Open your Telegram bot chat')}</li>
              <li>{t('auth.step_send_cmd', 'Send')} <code className="text-cyan-300 font-mono bg-white/5 px-1.5 py-0.5 rounded">/admin</code> {t('auth.or', 'or')} <code className="text-cyan-300 font-mono bg-white/5 px-1.5 py-0.5 rounded">/web</code></li>
              <li>{t('auth.step_click_btn', 'Tap the')} <strong className="text-white">🚀 {t('auth.open_dashboard_btn', 'Open Web Dashboard')}</strong> {t('auth.step_auto_login', 'button to sign in instantly.')}</li>
            </ol>
          </div>

          {/* Option 2: Paste Session Token */}
          <form onSubmit={handleSubmit} className="p-4 rounded-2xl bg-white/[0.03] border border-white/10 space-y-3">
            <div className="flex items-center gap-1.5 text-xs font-bold text-slate-200">
              <KeyRound size={14} className="text-amber-400" />
              <span>{t('auth.option2_title', 'Option 2: Paste Session Token')}</span>
            </div>

            <div className="space-y-2">
              <input
                type="password"
                value={tokenInput}
                onChange={(e) => {
                  setTokenInput(e.target.value);
                  setErrorMsg('');
                }}
                placeholder={t('auth.token_placeholder', 'Paste JWT session token from bot...')}
                className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white placeholder:text-slate-500"
              />

              {errorMsg && (
                <div className="flex items-center gap-1.5 text-[11px] text-red-400 font-medium animate-in fade-in">
                  <AlertCircle size={13} className="shrink-0" />
                  <span>{errorMsg}</span>
                </div>
              )}

              {successMsg && (
                <div className="flex items-center gap-1.5 text-[11px] text-emerald-400 font-medium animate-in fade-in">
                  <CheckCircle2 size={13} className="shrink-0" />
                  <span>{successMsg}</span>
                </div>
              )}

              <div className="flex items-center justify-between pt-1">
                <button
                  type="submit"
                  disabled={isLoading || !tokenInput.trim()}
                  className="liquid-button bg-[#229ED9] hover:bg-[#1e8ec3] px-5 py-2.5 rounded-xl text-xs font-bold text-white flex items-center gap-2 shadow-lg shadow-[#229ED9]/15 disabled:opacity-50 transition-all hover:scale-[1.02] active:scale-[0.98]"
                >
                  {isLoading ? <RefreshCw size={14} className="animate-spin" /> : <Send size={13} className="fill-white -rotate-12" />}
                  <span>{isLoading ? t('common.verifying', 'Verifying...') : t('auth.btn_authenticate', 'Sign In with Token')}</span>
                </button>

                <button
                  type="button"
                  onClick={onClose}
                  className="text-xs text-slate-400 hover:text-white px-3 py-1.5 rounded-lg transition-colors"
                >
                  {t('common.cancel', 'Cancel')}
                </button>
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};
