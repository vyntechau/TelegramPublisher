import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ADMIN_TABS, API_ENDPOINTS, apiUrl } from '../constants';
import { useTranslation } from '../context/LanguageContext';
import { AnalyticsSummary, User, Channel, Report, SettingsMap } from '../types';
import { AdminOverviewTab } from '../components/admin/AdminOverviewTab';
import { AdminBroadcastTab } from '../components/admin/AdminBroadcastTab';
import { AdminUsersTab } from '../components/admin/AdminUsersTab';
import { AdminChannelsTab } from '../components/admin/AdminChannelsTab';
import { AdminReportsTab } from '../components/admin/AdminReportsTab';
import { AdminSettingsTab } from '../components/admin/AdminSettingsTab';
import { RefreshCw } from 'lucide-react';

export const AdminDashboard: React.FC = () => {
  const { tab = 'overview' } = useParams<{ tab?: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();

  // State
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [usersTotal, setUsersTotal] = useState<number>(0);
  const [channels, setChannels] = useState<Channel[]>([]);
  const [reports, setReports] = useState<Report[]>([]);
  const [settings, setSettings] = useState<SettingsMap>({});
  const [isLoading, setIsLoading] = useState<boolean>(true);

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
  }, []);

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
