import React from 'react';
import { AnalyticsSummary } from '../../types';
import { StatCard } from '../common/StatCard';
import { API_ENDPOINTS, apiUrl } from '../../constants';
import { Users, Crown, Eye, AlertCircle, DollarSign, Download, RefreshCw, Sparkles, Shield } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface AdminOverviewTabProps {
  summary: AnalyticsSummary | null;
  isLoading: boolean;
  onRefresh: () => void;
}

export const AdminOverviewTab: React.FC<AdminOverviewTabProps> = ({ summary, isLoading, onRefresh }) => {
  const { t } = useTranslation();

  return (
    <div className="space-y-6">
      {/* Header with Export & Refresh Tools */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            <span>{t('overview.title', 'Platform Overview & Analytics')}</span>
            <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
          </h2>
          <p className="text-xs text-slate-400">{t('overview.desc', 'Real-time Telegram bot metrics, audience growth, and media engagement')}</p>
        </div>

        <div className="flex items-center gap-2 flex-wrap">
          <button
            onClick={onRefresh}
            disabled={isLoading}
            className="liquid-pill px-3 py-2 rounded-xl text-xs font-semibold text-slate-200 hover:text-white flex items-center gap-1.5 transition-all"
          >
            <RefreshCw size={13} className={isLoading ? 'animate-spin text-cyan-400' : ''} />
            <span>{t('common.refresh', 'Refresh')}</span>
          </button>

          <a
            href={apiUrl(API_ENDPOINTS.ANALYTICS_EXPORT_USERS)}
            target="_blank"
            rel="noreferrer"
            className="liquid-pill px-3 py-2 rounded-xl text-xs font-semibold text-slate-200 hover:text-white flex items-center gap-1.5 transition-all"
          >
            <Download size={13} className="text-cyan-400" />
            <span>{t('overview.export_users', 'Users CSV')}</span>
          </a>

          <a
            href={apiUrl(API_ENDPOINTS.ANALYTICS_EXPORT_POSTS)}
            target="_blank"
            rel="noreferrer"
            className="liquid-pill px-3 py-2 rounded-xl text-xs font-semibold text-slate-200 hover:text-white flex items-center gap-1.5 transition-all"
          >
            <Download size={13} className="text-indigo-400" />
            <span>{t('overview.export_posts', 'Posts CSV')}</span>
          </a>
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3 sm:gap-4">
        <StatCard
          label={t('overview.kpi_users', 'Total Users')}
          value={summary?.total_users?.toLocaleString() || '0'}
          icon={Users}
          color="text-cyan-400"
          change="+12% wk"
        />

        <StatCard
          label={t('overview.kpi_active', 'Active 24h')}
          value={summary?.active_users_24h?.toLocaleString() || '0'}
          icon={Sparkles}
          color="text-emerald-400"
        />

        <StatCard
          label={t('overview.kpi_vip', 'VIP Subscribers')}
          value={summary?.vip_subscribers?.toLocaleString() || '0'}
          icon={Crown}
          color="text-amber-400"
          change="VIP"
        />

        <StatCard
          label={t('overview.kpi_views', 'Total Media Views')}
          value={summary?.total_views?.toLocaleString() || '0'}
          icon={Eye}
          color="text-blue-400"
        />
      </div>

      {/* Secondary Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="liquid-glass-card p-5 rounded-3xl space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">{t('overview.revenue_title', 'Revenue Gateway')}</span>
            <DollarSign size={16} className="text-emerald-400" />
          </div>
          <div className="text-2xl font-black text-white" dir="ltr">
            ${summary?.total_revenue_usd ? summary.total_revenue_usd.toFixed(2) : '0.00'}
          </div>
          <div className="text-xs text-slate-400">{t('overview.revenue_desc', 'Processed via AZPays cryptocurrency gateway')}</div>
        </div>

        <div className="liquid-glass-card p-5 rounded-3xl space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">{t('overview.releases_title', 'Published Releases')}</span>
            <Shield size={16} className="text-indigo-400" />
          </div>
          <div className="text-2xl font-black text-white">
            {summary?.total_posts || 0}
          </div>
          <div className="text-xs text-slate-400">{t('overview.releases_desc', 'Active media assets in catalog')}</div>
        </div>

        <div className="liquid-glass-card p-5 rounded-3xl space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-slate-400 uppercase tracking-wider">{t('overview.moderation_title', 'Ticket Moderation')}</span>
            <AlertCircle size={16} className="text-red-400" />
          </div>
          <div className="text-2xl font-black text-white">
            {summary?.pending_reports || 0} <span className="text-xs font-normal text-slate-400">{t('overview.moderation_pending', 'pending')}</span>
          </div>
          <div className="text-xs text-slate-400">{t('overview.moderation_desc', { count: summary?.total_reports || 0 }) || `${summary?.total_reports || 0} total tickets filed`}</div>
        </div>
      </div>
    </div>
  );
};
