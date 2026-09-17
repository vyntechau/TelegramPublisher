import React, { useState } from 'react';
import { User, UserRole } from '../../types';
import { API_ENDPOINTS, apiUrl } from '../../constants';
import { RefreshCw } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface AdminUsersTabProps {
  users: User[];
  total: number;
  onRefresh: () => void;
}

export const AdminUsersTab: React.FC<AdminUsersTabProps> = ({ users, total, onRefresh }) => {
  const { t } = useTranslation();
  const [updatingId, setUpdatingId] = useState<number | null>(null);

  const handleRoleChange = async (telegramId: number, newRole: UserRole) => {
    try {
      setUpdatingId(telegramId);
      await fetch(apiUrl(`${API_ENDPOINTS.USERS}/${telegramId}`), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({ role: newRole }),
      });
      onRefresh();
    } catch (e) {
      console.error(e);
    } finally {
      setUpdatingId(null);
    }
  };

  const handleStatusChange = async (telegramId: number, newStatus: string) => {
    try {
      setUpdatingId(telegramId);
      await fetch(apiUrl(`${API_ENDPOINTS.USERS}/${telegramId}`), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({ status: newStatus }),
      });
      onRefresh();
    } catch (e) {
      console.error(e);
    } finally {
      setUpdatingId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight">{t('users_mgmt.title', 'User Account Management')}</h2>
          <p className="text-xs text-slate-400">{t('users_mgmt.desc', { total }) || `Manage user roles, VIP subscriptions, and account permissions (${total} total users)`}</p>
        </div>

        <button
          onClick={onRefresh}
          className="liquid-pill px-3 py-2 rounded-xl text-xs font-semibold text-slate-200 hover:text-white flex items-center gap-1.5 self-start sm:self-auto"
        >
          <RefreshCw size={13} />
          <span>{t('common.refresh', 'Refresh')}</span>
        </button>
      </div>

      <div className="liquid-glass-card rounded-3xl overflow-hidden border border-white/10 shadow-xl">
        <div className="overflow-x-auto">
          <table className="w-full text-start text-xs">
            <thead className="bg-white/[0.04] text-slate-400 font-semibold border-b border-white/10 uppercase tracking-wider text-[10px]">
              <tr>
                <th className="p-4 text-start">{t('users_mgmt.col_user', 'User')}</th>
                <th className="p-4 text-start">{t('users_mgmt.col_tg_id', 'Telegram ID')}</th>
                <th className="p-4 text-start">{t('users_mgmt.col_role', 'Role')}</th>
                <th className="p-4 text-start">{t('users_mgmt.col_status', 'Status')}</th>
                <th className="p-4 text-end">{t('users_mgmt.col_actions', 'Actions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {users.map((u) => (
                <tr key={u.id || u.telegram_id} className="hover:bg-white/[0.02] transition-colors">
                  <td className="p-4">
                    <div className="flex items-center gap-2.5">
                      <div className="w-8 h-8 rounded-xl bg-gradient-to-tr from-cyan-500/20 to-blue-600/20 border border-cyan-500/30 flex items-center justify-center font-bold text-cyan-300 shrink-0">
                        {u.first_name ? u.first_name.charAt(0).toUpperCase() : 'U'}
                      </div>
                      <div className="text-start">
                        <div className="font-bold text-white">{u.first_name || t('users_mgmt.anonymous', 'Anonymous')}</div>
                        <div className="text-[10px] text-slate-400" dir="ltr">{u.username ? `@${u.username}` : t('users_mgmt.no_username', 'No username')}</div>
                      </div>
                    </div>
                  </td>
                  <td className="p-4 font-mono text-slate-300 text-start" dir="ltr">{u.telegram_id}</td>
                  <td className="p-4 text-start">
                    <select
                      value={u.role}
                      disabled={updatingId === u.telegram_id}
                      onChange={(e) => handleRoleChange(u.telegram_id, e.target.value as UserRole)}
                      className="liquid-input px-2.5 py-1 rounded-xl text-xs capitalize text-slate-200"
                    >
                      <option value="user">{t('users_mgmt.role_user', 'User')}</option>
                      <option value="author">{t('users_mgmt.role_author', 'Author')}</option>
                      <option value="admin">{t('users_mgmt.role_admin', 'Admin')}</option>
                      <option value="owner">{t('users_mgmt.role_owner', 'Owner')}</option>
                    </select>
                  </td>
                  <td className="p-4 text-start">
                    <span className={`px-2 py-0.5 rounded-lg text-[10px] font-bold uppercase ${
                      u.status === 'active' ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' : 'bg-red-500/15 text-red-400 border border-red-500/30'
                    }`}>
                      {u.status === 'active' ? t('users_mgmt.status_active', 'active') : (u.status === 'banned' ? t('users_mgmt.status_banned', 'banned') : (u.status || 'active'))}
                    </span>
                  </td>
                  <td className="p-4 text-end">
                    <button
                      onClick={() => handleStatusChange(u.telegram_id, u.status === 'banned' ? 'active' : 'banned')}
                      disabled={updatingId === u.telegram_id}
                      className={`px-3 py-1 rounded-xl text-[11px] font-semibold transition-all ${
                        u.status === 'banned'
                          ? 'bg-emerald-500/20 text-emerald-300 hover:bg-emerald-500/30'
                          : 'bg-red-500/20 text-red-300 hover:bg-red-500/30'
                      }`}
                    >
                      {u.status === 'banned' ? t('users_mgmt.btn_unban', 'Unban') : t('users_mgmt.btn_ban', 'Ban')}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
