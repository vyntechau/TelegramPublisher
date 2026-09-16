import React, { useState } from 'react';
import { API_ENDPOINTS } from '../../constants';
import { AlertTriangle, CheckCircle2, X, Send } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface ReportModalProps {
  postId: number;
  isOpen: boolean;
  onClose: () => void;
}

export const ReportModal: React.FC<ReportModalProps> = ({ postId, isOpen, onClose }) => {
  const { t } = useTranslation();
  const [reason, setReason] = useState('broken_stream');
  const [details, setDetails] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [loading, setLoading] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      setLoading(true);
      await fetch(API_ENDPOINTS.REPORTS, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          post_id: postId,
          reason: `${reason}: ${details}`,
          status: 'pending',
        }),
      });
      setSubmitted(true);
      setTimeout(() => {
        setSubmitted(false);
        onClose();
      }, 1500);
    } catch (err) {
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-md">
      <div className="liquid-glass-card max-w-md w-full p-6 sm:p-8 rounded-3xl border border-white/10 space-y-4 shadow-2xl relative text-start">
        <button
          onClick={onClose}
          className="absolute top-4 ltr:right-4 rtl:left-4 p-2 text-slate-400 hover:text-white rounded-xl"
        >
          <X size={16} />
        </button>

        <div className="flex items-center gap-2.5 text-red-400 font-bold text-sm">
          <AlertTriangle size={18} />
          <span>{t('report_modal.title', 'Report Broken Media or File')}</span>
        </div>

        {submitted ? (
          <div className="p-6 text-center space-y-2">
            <CheckCircle2 size={36} className="text-emerald-400 mx-auto animate-bounce" />
            <div className="text-sm font-bold text-white">{t('report_modal.ticket_submitted', 'Ticket Submitted')}</div>
            <div className="text-xs text-slate-400">{t('report_modal.ticket_desc', 'Our moderators and the author have been notified.')}</div>
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">{t('report_modal.issue_category', 'Issue Category')}</label>
              <select
                value={reason}
                onChange={(e) => setReason(e.target.value)}
                className="liquid-input w-full px-3 py-2 rounded-xl text-xs text-white"
              >
                <option value="broken_stream">{t('report_modal.cat_broken_stream', "Video Won't Play / Broken Stream")}</option>
                <option value="expired_link">{t('report_modal.cat_expired_link', 'Link Expired Ahead of Time')}</option>
                <option value="corrupted_file">{t('report_modal.cat_corrupted_file', 'Corrupted Video File')}</option>
                <option value="copyright">{t('report_modal.cat_copyright', 'Copyright Issue')}</option>
                <option value="other">{t('report_modal.cat_other', 'Other Issue')}</option>
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="text-xs font-semibold text-slate-300">{t('report_modal.additional_details', 'Additional Details')}</label>
              <textarea
                rows={3}
                required
                value={details}
                onChange={(e) => setDetails(e.target.value)}
                placeholder={t('report_modal.details_ph', 'Describe what happened when you tried to watch...')}
                className="liquid-input w-full p-3 rounded-xl text-xs text-white resize-none text-start"
              />
            </div>

            <div className="flex items-center justify-end gap-2 pt-2">
              <button
                type="button"
                onClick={onClose}
                className="liquid-pill px-4 py-2 rounded-xl text-xs font-semibold text-slate-300"
              >
                {t('common.cancel', 'Cancel')}
              </button>
              <button
                type="submit"
                disabled={loading}
                className="liquid-button px-5 py-2 rounded-xl text-xs font-bold text-white shadow-md flex items-center gap-1.5"
              >
                <Send size={13} className="rtl:rotate-180" />
                <span>{loading ? t('report_modal.btn_submitting', 'Submitting...') : t('report_modal.btn_submit', 'Submit Ticket')}</span>
              </button>
            </div>
          </form>
        )}
      </div>
    </div>
  );
};
