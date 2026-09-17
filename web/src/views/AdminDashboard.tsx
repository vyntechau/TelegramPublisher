import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ADMIN_TABS, API_ENDPOINTS, apiUrl } from '../constants';
import { useTranslation } from '../context/LanguageContext';
import { useAuth } from '../context/AuthContext';
import { AnalyticsSummary, User, Channel, Report, SettingsMap } from '../types';
import { AdminOverviewTab } from '../components/admin/AdminOverviewTab';
import { AdminBroadcastTab } from '../components/admin/AdminBroadcastTab';
import { AdminUsersTab } from '../components/admin/AdminUsersTab';
import { AdminChannelsTab } from '../components/admin/AdminChannelsTab';
import { AdminReportsTab } from '../components/admin/AdminReportsTab';
import { AdminSettingsTab } from '../components/admin/AdminSettingsTab';
import { ShieldAlert, KeyRound, ExternalLink, RefreshCw, CheckCircle2, ArrowRight } from 'lucide-react';

export const AdminDashboard: React.FC = () => {
  const { tab = 'overview' } = useParams<{ tab?: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const { token: authToken, loginWithToken } = useAuth();

  // State
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [usersTotal, setUsersTotal] = useState<number>(0);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [reports, setReports] = useState<Report[]>([]);
  const [settings, setSettings] = useState<SettingsMap>({});
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [isUnauthorized, setIsUnauthorized] = useState<boolean>(false);

  // Manual token input state
  const [tokenInput, setTokenInput] = useState<string>('');
  const [authError, setAuthError] = useState<string>('');
  const [authSubmitting, setAuthSubmitting] = useState<boolean>(false);

  // Normalize tab
  const activeTabId = tab === 'marketing' ? 'broadcast' : tab;

  const fetchAllData = async () => {
    try {
      setIsLoading(true);
      const token = localStorage.getItem('tp_token');
      const headers: Record<string, string> = {};
      if (token) headers['Authorization'] = `Bearer ${token}`;

      // Parallel fetching for performance
      const [ovResp, usrResp, chResp, repResp, stResp] = await Promise.all([
        fetch(apiUrl(API_ENDPOINTS.ANALYTICS_OVERVIEW), { headers }),
        fetch(apiUrl(`${API_ENDPOINTS.USERS}?limit=50`), { headers }),
        fetch(apiUrl(API_ENDPOINTS.CHANNELS), { headers }),
        fetch(apiUrl(API_ENDPOINTS.REPORTS), { headers }),
        fetch(apiUrl(API_ENDPOINTS.SETTINGS), { headers }),
      ]);

      if (ovResp.status === 401 || stResp.status === 401 || usrResp.status === 401) {
        setIsUnauthorized(true);
        return;
      }

      setIsUnauthorized(false);
      if (ovResp.ok) setSummary(await ovResp.json());
      if (usrResp.ok) {
        const uData = await usrResp.json();
        setUsers(uData.users || []);
        setUsersTotal(uData.total || 0);
      }
      if (chResp.ok) setChannels(await chResp.json());
      if (repResp.ok) {
        const rData = await repResp.json();
        setReports(rData.reports || []);
      }
      if (stResp.ok) setSettings(await stResp.json());
    } catch (e) {
      console.warn('Failed to load admin dataset:', e);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchAllData();
  }, [authToken]);

  const handleManualAuth = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!tokenInput.trim()) return;
    setAuthError('');
    setAuthSubmitting(true);
    try {
      const success = await loginWithToken(tokenInput.trim());
      if (success) {
        setTokenInput('');
        await fetchAllData();
      } else {
        setAuthError('Invalid or expired session token. Please generate a fresh token via Telegram bot /admin or /web.');
      }
    } catch {
      setAuthError('Connection failed while verifying token.');
    } finally {
      setAuthSubmitting(false);
    }
  };

  if (isUnauthorized) {
    return (
      <div className="max-w-2xl mx-auto py-8 space-y-6">
        <div className="liquid-glass-card p-6 sm:p-8 rounded-3xl border border-cyan-500/20 shadow-2xl relative overflow-hidden">
          <div className="ambient-orb w-64 h-64 bg-cyan-600/10 -top-20 -right-20 pointer-events-none" />

          <div className="flex items-center gap-3 mb-4">
            <div className="w-12 h-12 rounded-2xl bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-cyan-400">
              <KeyRound size={24} />
            </div>
            <div>
              <h2 className="text-lg sm:text-xl font-bold text-white tracking-tight">Admin Authentication Required</h2>
              <p className="text-xs text-slate-400">Your setup is complete. Connect your Telegram admin session to unlock operational data.</p>
            </div>
          </div>

          <div className="space-y-4 pt-2">
            {/* Method 1: Telegram Bot 1-Click Login */}
            <div className="p-4 rounded-2xl bg-white/[0.03] border border-white/10 space-y-2">
              <div className="text-xs font-bold text-cyan-300 flex items-center gap-2">
                <span>📱 Option 1: Generate 1-Click Login via Telegram Bot</span>
              </div>
              <ol className="text-xs text-slate-300 space-y-1 list-decimal list-inside leading-relaxed">
                <li>Open your Telegram bot chat</li>
                <li>Send command <code className="text-cyan-300 font-mono bg-white/5 px-1.5 py-0.5 rounded">/admin</code> or <code className="text-cyan-300 font-mono bg-white/5 px-1.5 py-0.5 rounded">/web</code></li>
                <li>Tap <strong className="text-white">🚀 Open Web Dashboard</strong> to sign in automatically</li>
              </ol>
            </div>

            {/* Method 2: Paste Session Token */}
            <form onSubmit={handleManualAuth} className="p-4 rounded-2xl bg-white/[0.03] border border-white/10 space-y-3">
              <div className="text-xs font-bold text-slate-200 flex items-center gap-2">
                <span>🔑 Option 2: Paste Admin Session Token</span>
              </div>
              <div className="space-y-2">
                <input
                  type="password"
                  value={tokenInput}
                  onChange={(e) => setTokenInput(e.target.value)}
                  placeholder="Paste JWT session token here..."
                  className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white placeholder:text-slate-600"
                />
                {authError && (
                  <p className="text-[11px] text-red-400 font-medium">{authError}</p>
                )}
                <div className="flex gap-2">
                  <button
                    type="submit"
                    disabled={authSubmitting || !tokenInput.trim()}
                    className="liquid-button px-5 py-2 rounded-xl text-xs font-bold text-white flex items-center gap-2 disabled:opacity-50"
                  >
                    {authSubmitting ? <RefreshCw size={13} className="animate-spin" /> : <ArrowRight size={13} />}
                    <span>{authSubmitting ? 'Verifying...' : 'Authenticate Session'}</span>
                  </button>
                  <button
                    type="button"
                    onClick={fetchAllData}
                    className="px-4 py-2 rounded-xl text-xs font-semibold text-slate-400 hover:text-white bg-white/5 hover:bg-white/10 transition-all"
                  >
                    Retry Connection
                  </button>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 sm:space-y-8">
      {/* Sub-Navigation Tab Bar for Admin Dashboard */}
      <div className="flex items-center gap-1.5 p-1.5 rounded-2xl liquid-glass border border-white/10 overflow-x-auto no-scrollbar shadow-inner">
        {ADMIN_TABS.map((tItem) => {
          const Icon = tItem.icon;
          const isActive = activeTabId === tItem.id;
          const labelMap: Record<string, string> = {
            overview: t('admin.tab_overview', 'Overview'),
            broadcast: t('admin.tab_broadcast', 'Broadcast'),
            users: t('admin.tab_users', 'User Directory'),
            channels: t('admin.tab_channels', 'Target Channels'),
            reports: t('admin.tab_reports', 'Content Reports'),
            settings: t('admin.tab_settings', 'Platform Settings'),
          };
          const label = labelMap[tItem.id] || tItem.label;

          return (
            <button
              key={tItem.id}
              onClick={() => navigate(tItem.path)}
              className={`flex items-center gap-2 px-3.5 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition-all ${
                isActive
                  ? 'liquid-pill-active text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04]'
              }`}
            >
              <Icon size={14} />
              <span>{label}</span>
            </button>
          );
        })}
      </div>

      {/* Tab Viewport */}
      {activeTabId === 'overview' && (
        <AdminOverviewTab summary={summary} isLoading={isLoading} onRefresh={fetchAllData} />
      )}

      {activeTabId === 'broadcast' && <AdminBroadcastTab />}

      {activeTabId === 'users' && (
        <AdminUsersTab users={users} total={usersTotal} onRefresh={fetchAllData} />
      )}

      {activeTabId === 'channels' && (
        <AdminChannelsTab channels={channels} onRefresh={fetchAllData} />
      )}

      {activeTabId === 'reports' && (
        <AdminReportsTab reports={reports} onRefresh={fetchAllData} />
      )}

      {activeTabId === 'settings' && (
        <AdminSettingsTab settings={settings} onRefresh={fetchAllData} />
      )}
    </div>
  );
};
