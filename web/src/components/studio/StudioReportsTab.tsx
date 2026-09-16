import React, { useState } from 'react';
import { Report } from '../../types';
import { API_ENDPOINTS } from '../../constants';
import { Clock, RefreshCw } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface StudioReportsTabProps {
  reports: Report[];
  onRefresh: () => void;
}

export const StudioReportsTab: React.FC<StudioReportsTabProps> = ({ reports, onRefresh }) => {
  const { t } = useTranslation();
  const [resolvingId, setResolvingId] = useState<number | null>(null);
  const [note, setNote] = useState('');

  const handleResolve = async (reportId: number) => {
    try {
      setResolvingId(reportId);
      await fetch(`${API_ENDPOINTS.REPORTS}/${reportId}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({
          status: 'resolved',
          resolution_note: note || 'Resolved by author',
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
          <h2 className="text-xl font-bold text-white tracking-tight">{t('studio_reports.title', 'Reported Files & Tickets')}</h2>
          <p className="text-xs text-slate-400">{t('studio_reports.desc', 'Broken media and expired link complaints filed by viewers on your posts')}</p>
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
            {t('studio_reports.no_tickets', 'No pending tickets for your media releases.')}
          </div>
        ) : (
          reports.map((rep) => (
            <div key={rep.id} className="liquid-glass-card p-5 rounded-3xl space-y-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-amber-400 animate-pulse" />
                  <span className="text-xs font-bold text-white">{t('reports.post_num', { id: rep.post_id }) || `Post ID #${rep.post_id}`}</span>
                  <span className="text-xs text-slate-400">• {t('reports.ticket_num', { id: rep.id }) || `Ticket #${rep.id}`}</span>
                </div>
                <span className="text-[11px] text-slate-400 flex items-center gap-1">
                  <Clock size={12} />
                  {rep.created_at ? new Date(rep.created_at).toLocaleDateString() : t('reports.recent', 'Recent')}
                </span>
              </div>

              <div className="p-3.5 rounded-2xl bg-white/[0.02] border border-white/5 text-xs text-slate-300 text-start">
                {rep.reason || t('studio_reports.default_reason', 'User reported file unavailable.')}
              </div>

              <div className="flex flex-col sm:flex-row items-center gap-2 pt-1">
                <input
                  type="text"
                  placeholder={t('studio_reports.resolution_ph', 'Resolution note (e.g. Re-uploaded media)...')}
                  value={resolvingId === rep.id ? note : ''}
                  onChange={(e) => {
                    setResolvingId(rep.id);
                    setNote(e.target.value);
                  }}
                  className="liquid-input flex-1 px-3 py-2 rounded-xl text-xs text-white placeholder-slate-500 w-full text-start"
                />
                <button
                  onClick={() => handleResolve(rep.id)}
                  disabled={resolvingId === rep.id}
                  className="liquid-button px-4 py-2 rounded-xl text-xs font-bold text-white shadow-md w-full sm:w-auto"
                >
                  {t('studio_reports.btn_resolve', 'Resolve Ticket')}
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
