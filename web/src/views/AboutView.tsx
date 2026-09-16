import React from 'react';
import { VynTechLogo } from '../components/VynTechLogo';
import { useTranslation } from '../context/LanguageContext';
import { 
  Coins, 
  Send, 
  Cpu, 
  Lock, 
  Code2, 
  ArrowUpRight, 
  Sparkles, 
  Server, 
  Cloud, 
  ShieldCheck, 
  Globe2, 
  Bot, 
  Zap, 
  ExternalLink,
  Terminal 
} from 'lucide-react';

export const AboutView: React.FC = () => {
  const { t } = useTranslation();

  return (
    <div className="max-w-4xl mx-auto space-y-10 py-4 pb-16">
      {/* Hero Section */}
      <div className="liquid-glass-card p-8 sm:p-12 rounded-3xl relative overflow-hidden text-center">
        <div className="flex justify-center mb-6">
          <VynTechLogo size="lg" showWordmark={true} />
        </div>

        <h1 className="text-3xl sm:text-4xl md:text-5xl font-extrabold text-white tracking-tight mb-4">
          {t('about.title', 'Open-Source Telegram Media & Publishing Platform')}
        </h1>

        <p className="text-slate-300 text-sm sm:text-base max-w-2xl mx-auto leading-relaxed mb-8">
          {t('about.subtitle', 'Enterprise-grade bot infrastructure and unified role-based web application for automated media distribution, copyright protection, crypto subscriptions, and multi-channel publishing.')}
        </p>

        <div className="flex flex-wrap items-center justify-center gap-3">
          <a
            href="https://github.com/vyntechau/TelegramPublisher"
            target="_blank"
            rel="noreferrer"
            className="liquid-button inline-flex items-center gap-2 px-5 py-2.5 rounded-xl text-xs sm:text-sm font-semibold text-white"
          >
            <Code2 size={15} />
            <span>{t('about.btn_github', 'GitHub Repository')}</span>
            <ArrowUpRight size={13} className="opacity-70 rtl:-scale-x-100" />
          </a>

          <a
            href="/docs"
            target="_blank"
            rel="noreferrer"
            className="liquid-pill px-5 py-2.5 rounded-xl text-xs sm:text-sm font-medium text-slate-300 hover:text-white transition-all inline-flex items-center gap-2"
          >
            <Server size={15} className="text-cyan-400" />
            <span>{t('about.btn_api_docs', 'Interactive API Docs')}</span>
          </a>
        </div>
      </div>

      {/* VynTech Cloud Platform Introduction & CTA Banner */}
      <div className="liquid-glass-card p-8 sm:p-10 rounded-3xl relative overflow-hidden border border-cyan-500/20 bg-gradient-to-b from-cyan-950/20 via-transparent to-indigo-950/20 shadow-xl">
        <div className="ambient-orb w-64 h-64 bg-cyan-500/10 -top-20 -right-20 pointer-events-none" />

        <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-6 relative z-10">
          <div className="space-y-3 max-w-2xl">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full liquid-pill text-cyan-300 text-xs font-semibold border border-cyan-400/20">
              <Cloud size={14} className="text-cyan-400" />
              <span>{t('about.cloud_intro', 'Powered by VynTech Cloud')}</span>
            </div>

            <h2 className="text-2xl sm:text-3xl font-black text-white tracking-tight">
              {t('about.cloud_title', 'Scale Your Digital Infrastructure with VynTech Cloud')}
            </h2>

            <p className="text-slate-300 text-sm leading-relaxed">
              {t('about.cloud_desc', 'VynTech Cloud delivers high-availability cloud hosting, intelligent Telegram bots, decentralized storage, edge streaming networks, and enterprise automation engineered for modern businesses.')}
            </p>

            {/* Capability Badges */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2.5 pt-2">
              <div className="flex items-center gap-2 text-xs text-slate-300 bg-white/[0.03] p-2 rounded-xl border border-white/5">
                <Globe2 size={14} className="text-cyan-400 shrink-0" />
                <span>{t('about.badge_cdn', 'Global Edge CDN')}</span>
              </div>
              <div className="flex items-center gap-2 text-xs text-slate-300 bg-white/[0.03] p-2 rounded-xl border border-white/5">
                <ShieldCheck size={14} className="text-emerald-400 shrink-0" />
                <span>{t('about.badge_security', 'Zero-Trust Security')}</span>
              </div>
              <div className="flex items-center gap-2 text-xs text-slate-300 bg-white/[0.03] p-2 rounded-xl border border-white/5">
                <Bot size={14} className="text-purple-400 shrink-0" />
                <span>{t('about.badge_automation', 'AI Automation')}</span>
              </div>
              <div className="flex items-center gap-2 text-xs text-slate-300 bg-white/[0.03] p-2 rounded-xl border border-white/5">
                <Zap size={14} className="text-amber-400 shrink-0" />
                <span>{t('about.badge_sla', '99.99% SLA Uptime')}</span>
              </div>
            </div>
          </div>

          {/* Action CTAs */}
          <div className="flex flex-col sm:flex-row md:flex-col gap-3 w-full md:w-auto shrink-0">
            <a
              href="https://www.vyntech.com.au/products/cloud"
              target="_blank"
              rel="noreferrer"
              className="liquid-button inline-flex items-center justify-center gap-2 px-6 py-3 rounded-2xl text-sm font-bold text-white shadow-lg text-center whitespace-nowrap"
            >
              <span>{t('about.btn_cloud_intro', 'Explore VynTech Cloud')}</span>
              <ExternalLink size={15} />
            </a>

            <a
              href="https://console.vyntech.com.au"
              target="_blank"
              rel="noreferrer"
              className="liquid-pill inline-flex items-center justify-center gap-2 px-6 py-3 rounded-2xl text-xs font-semibold text-slate-200 hover:text-white border border-white/10 hover:border-white/20 transition-all text-center whitespace-nowrap"
            >
              <Terminal size={14} className="text-cyan-400" />
              <span>{t('about.btn_cloud_console', 'Developer Console')}</span>
            </a>
          </div>
        </div>
      </div>

      {/* Core Highlights Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="liquid-glass-card p-6 rounded-2xl group text-start">
          <div className="w-10 h-10 rounded-xl bg-cyan-500/10 flex items-center justify-center text-cyan-400 mb-3 group-hover:scale-105 transition-transform">
            <Lock size={20} />
          </div>
          <h3 className="text-base font-bold text-white mb-1.5">{t('about_view.f_copyright_title', 'Copyright & Media Shield')}</h3>
          <p className="text-xs text-slate-400 leading-relaxed">
            {t('about_view.f_copyright_desc', 'Automatic 2-minute media message expiration and auto-cleanup queues keep direct downloads safe from unauthorized redistribution.')}
          </p>
        </div>

        <div className="liquid-glass-card p-6 rounded-2xl group text-start">
          <div className="w-10 h-10 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-400 mb-3 group-hover:scale-105 transition-transform">
            <Coins size={20} />
          </div>
          <h3 className="text-base font-bold text-white mb-1.5">{t('about_view.f_crypto_title', 'AZPays Crypto Gateway')}</h3>
          <p className="text-xs text-slate-400 leading-relaxed">
            {t('about_view.f_crypto_desc', 'Native VIP subscription checkout accepting USDT, BTC, ETH, and TON via official AZPays SDK with cryptographic webhook signatures.')}
          </p>
        </div>

        <div className="liquid-glass-card p-6 rounded-2xl group text-start">
          <div className="w-10 h-10 rounded-xl bg-indigo-500/10 flex items-center justify-center text-indigo-400 mb-3 group-hover:scale-105 transition-transform">
            <Send size={20} />
          </div>
          <h3 className="text-base font-bold text-white mb-1.5">{t('about_view.f_broadcast_title', 'Multi-Channel Broadcast')}</h3>
          <p className="text-xs text-slate-400 leading-relaxed">
            {t('about_view.f_broadcast_desc', 'Simultaneous auto-publishing to multiple Telegram channels with teaser buttons and deep-linked watch buttons driving user engagement.')}
          </p>
        </div>
      </div>

      {/* Technology & Architecture Section */}
      <div className="liquid-glass-card p-6 sm:p-8 rounded-3xl space-y-5 text-start">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-xl bg-purple-500/10 flex items-center justify-center text-purple-400">
            <Cpu size={18} />
          </div>
          <div>
            <h2 className="text-base font-bold text-white">{t('about_view.arch_title', 'Architecture & Engineering')}</h2>
            <p className="text-xs text-slate-400">{t('about_view.arch_desc', 'Built for scale, zero-downtime reconfiguration, and twelve-factor portability')}</p>
          </div>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
          <div className="p-3.5 rounded-xl bg-white/[0.03] space-y-1 text-start">
            <div className="text-[11px] font-semibold text-cyan-400 uppercase tracking-wider">{t('about_view.arch_backend_badge', 'Backend Core')}</div>
            <div className="text-xs font-bold text-white">{t('about_view.arch_backend_title', 'Go 1.24 Standard Library')}</div>
            <div className="text-[11px] text-slate-400 leading-snug">{t('about_view.arch_backend_desc', 'Zero third-party web frameworks, high-throughput concurrent handlers.')}</div>
          </div>

          <div className="p-3.5 rounded-xl bg-white/[0.03] space-y-1 text-start">
            <div className="text-[11px] font-semibold text-blue-400 uppercase tracking-wider">{t('about_view.arch_storage_badge', 'Storage Engine')}</div>
            <div className="text-xs font-bold text-white">{t('about_view.arch_storage_title', 'Twelve-Factor URL DSN')}</div>
            <div className="text-[11px] text-slate-400 leading-snug">{t('about_view.arch_storage_desc', 'SQLite (embedded), PostgreSQL, or MySQL with auto driver inference.')}</div>
          </div>

          <div className="p-3.5 rounded-xl bg-white/[0.03] space-y-1 text-start">
            <div className="text-[11px] font-semibold text-indigo-400 uppercase tracking-wider">{t('about_view.arch_web_badge', 'Web Frontend')}</div>
            <div className="text-xs font-bold text-white">{t('about_view.arch_web_title', 'React 18 + Vite + TS')}</div>
            <div className="text-[11px] text-slate-400 leading-snug">{t('about_view.arch_web_desc', 'Balanced Apple glass design tokens, standalone URLs & route middleware.')}</div>
          </div>

          <div className="p-3.5 rounded-xl bg-white/[0.03] space-y-1 text-start">
            <div className="text-[11px] font-semibold text-emerald-400 uppercase tracking-wider">{t('about_view.arch_config_badge', 'Runtime Config')}</div>
            <div className="text-xs font-bold text-white">{t('about_view.arch_config_title', 'Live DB Settings Table')}</div>
            <div className="text-[11px] text-slate-400 leading-snug">{t('about_view.arch_config_desc', 'Update bot buttons, channels, and keys instantly without restarting.')}</div>
          </div>
        </div>
      </div>

      {/* Project Metadata & Credits */}
      <div className="flex flex-col sm:flex-row items-center justify-between gap-3 text-xs text-slate-400 px-2">
        <div className="flex items-center gap-2 text-center sm:text-start">
          <Sparkles size={13} className="text-cyan-400 shrink-0" />
          <span>{t('about_view.credits_text', 'Maintained by VynTech Cloud under open-source license.')}</span>
        </div>
        <div className="flex items-center gap-3 text-[11px]">
          <a
            href="https://github.com/vyntechau/TelegramPublisher/blob/main/LICENSE"
            target="_blank"
            rel="noreferrer"
            className="hover:text-white transition-colors"
          >
            {t('about_view.mit_license', 'MIT License')}
          </a>
          <span>•</span>
          <a
            href="https://github.com/vyntechau/TelegramPublisher/releases"
            target="_blank"
            rel="noreferrer"
            className="hover:text-white transition-colors"
          >
            {t('about_view.release_version', 'Release v1.0.0')}
          </a>
        </div>
      </div>
    </div>
  );
};
