import React, { useState } from 'react';
import { Report } from '../../types';
import { API_ENDPOINTS } from '../../constants';
import { Clock, RefreshCw } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface AdminReportsTabProps {
  reports: Report[];
  onRefresh: () => void;
}

export const AdminReportsTab: React.FC<AdminReportsTabProps> = ({ reports, onRefresh }) => {
  const { t } = useTranslation();
  const [resolvingId, setResolvingId] = useState<number | null>(null);
  const [note, setNote] = useState('');

  const handleUpdateStatus = async (reportId: number, status: 'resolved' | 'rejected') => {
    try {
      setResolvingId(reportId);
      await fetch(`${API_ENDPOINTS.REPORTS}/${reportId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({
          status,
          resolution_note: note || `Marked as ${status} by admin`,
        }),
      });
      setNote('');
      onRefresh();
    } catch (e) {
      console.error(e);
    } finally {
      setResolvingId(null);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight">{t('reports.title', 'Broken File Reports & Tickets')}</h2>
          <p className="text-xs text-slate-400">{t('reports.desc', 'Manage user reports for expired, broken, or copyright-flagged media files')}</p>
        </div>

        <button
          onClick={onRefresh}
          className="liquid-pill px-3 py-2 rounded-xl text-xs font-semibold text-slate-200 hover:text-white flex items-center gap-1.5 self-start sm:self-auto"
        >
          <RefreshCw size={13} />
          <span>{t('common.refresh', 'Refresh')}</span>
        </button>
      </div>

      <div className="space-y-3">
        {reports.length === 0 ? (
          <div className="liquid-glass p-8 rounded-3xl text-center text-slate-400 text-xs">
            {t('reports.no_reports', 'No active reports in queue. Everything is running smoothly!')}
          </div>
        ) : (
          reports.map((rep) => (
            <div key={rep.id} className="liquid-glass-card p-5 rounded-3xl space-y-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <span className={`px-2 py-0.5 rounded-lg text-[10px] font-bold uppercase ${
                    rep.status === 'resolved'
                      ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                      : rep.status === 'rejected'
                      ? 'bg-red-500/15 text-red-400 border border-red-500/30'
                      : 'bg-amber-500/15 text-amber-400 border border-amber-500/30'
                  }`}>
                    {rep.status}
                  </span>
                  <span className="text-xs font-bold text-white">{t('reports.ticket_num', { id: rep.id }) || `Ticket #${rep.id}`}</span>
                  <span className="text-xs text-slate-400">• {t('reports.post_num', { id: rep.post_id }) || `Post ID #${rep.post_id}`}</span>
                </div>

                <span className="text-[11px] text-slate-400 flex items-center gap-1">
                  <Clock size={12} />
                  {rep.created_at ? new Date(rep.created_at).toLocaleDateString() : t('reports.recent', 'Recent')}
                </span>
              </div>

              <div className="p-3.5 rounded-2xl bg-white/[0.02] border border-white/5 text-xs text-slate-300 text-start">
                <strong>{t('reports.reason_label', 'Reason:')}</strong> {rep.reason || t('reports.default_reason', 'User reported media file unavailable or corrupted.')}
              </div>

              {rep.status === 'pending' && (
                <div className="flex flex-col sm:flex-row items-center gap-2 pt-1">
                  <input
                    type="text"
                    placeholder={t('reports.resolution_ph', 'Resolution note (e.g. Re-uploaded media)...')}
                    value={resolvingId === rep.id ? note : ''}
                    onChange={(e) => {
                      setResolvingId(rep.id);
                      setNote(e.target.value);
                    }}
                    className="liquid-input flex-1 px-3 py-2 rounded-xl text-xs text-white placeholder-slate-500 w-full text-start"
                  />
                  <div className="flex items-center gap-2 w-full sm:w-auto">
                    <button
                      onClick={() => handleUpdateStatus(rep.id, 'resolved')}
                      disabled={resolvingId === rep.id}
                      className="liquid-button px-4 py-2 rounded-xl text-xs font-bold text-white shadow-md flex-1 sm:flex-initial"
                    >
                      {t('reports.btn_resolve', 'Resolve')}
                    </button>
                    <button
                      onClick={() => handleUpdateStatus(rep.id, 'rejected')}
                      disabled={resolvingId === rep.id}
                      className="liquid-pill px-4 py-2 rounded-xl text-xs font-semibold text-red-400 hover:bg-red-500/10 flex-1 sm:flex-initial"
                    >
                      {t('reports.btn_reject', 'Reject')}
                    </button>
                  </div>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
};
