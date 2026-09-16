import React, { useState } from 'react';
import { API_ENDPOINTS } from '../../constants';
import { Send, RefreshCw, Radio } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

export const AdminBroadcastTab: React.FC = () => {
  const { t } = useTranslation();
  const [format, setFormat] = useState<'html' | 'raw_json'>('html');
  const [content, setContent] = useState('');
  const [rawJSON, setRawJSON] = useState(`{
  "text": "🔥 Special Announcement for TelegramPublisher Users!",
  "parse_mode": "HTML",
  "inline_keyboard": [
    [{"text": "🚀 Open Mini App", "web_app_url": "http://localhost:8080"}]
  ]
}`);
  const [filterRole, setFilterRole] = useState('');
  const [filterActiveWeek, setFilterActiveWeek] = useState(false);
  const [broadcasting, setBroadcasting] = useState(false);
  const [result, setResult] = useState<any>(null);

  const handleBroadcast = async () => {
    if (format === 'html' && !content.trim()) return;
    setBroadcasting(true);
    setResult(null);

    try {
      const payload = {
        format,
        content: format === 'html' ? content : rawJSON,
        filter_role: filterRole,
        filter_active_week: filterActiveWeek,
      };

      const resp = await fetch(API_ENDPOINTS.MARKETING_BROADCAST, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify(payload),
      });

      const data = await resp.json();
      setResult(data);
    } catch (err: any) {
      setResult({ error: err.message });
    } finally {
      setBroadcasting(false);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-bold text-white tracking-tight">{t('broadcast.title', 'Marketing & Broadcaster')}</h2>
        <p className="text-xs text-slate-400">{t('broadcast.desc', 'Dispatch announcements, updates, and custom interactive keyboards directly to all bot users')}</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Broadcast Form */}
        <div className="lg:col-span-2 space-y-4">
          <div className="flex items-center gap-2">
            <button
              onClick={() => setFormat('html')}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                format === 'html' ? 'liquid-pill-active text-white' : 'liquid-pill text-slate-400'
              }`}
            >
              {t('broadcast.rich_html', 'Rich HTML Editor')}
            </button>
            <button
              onClick={() => setFormat('raw_json')}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                format === 'raw_json' ? 'liquid-pill-active text-white' : 'liquid-pill text-slate-400'
              }`}
            >
              {t('broadcast.raw_json', 'Raw Telegram JSON (Keyboards)')}
            </button>
          </div>

          {format === 'html' ? (
            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-200">{t('broadcast.content_label', 'Message Content (HTML formatting allowed)')}</label>
              <textarea
                rows={6}
                value={content}
                onChange={(e) => setContent(e.target.value)}
                placeholder={t('broadcast.content_placeholder', '<b>📢 Attention Subscribers!</b>\n\nNew exclusive release uploaded! Tap below to watch.')}
                className="liquid-input w-full p-4 rounded-2xl text-xs sm:text-sm font-mono text-white placeholder-slate-500 resize-none text-start"
              />
            </div>
          ) : (
            <div className="space-y-2">
              <label className="text-xs font-semibold text-slate-200">{t('broadcast.payload_label', 'Raw Bot API Payload (JSON)')}</label>
              <textarea
                rows={8}
                value={rawJSON}
                onChange={(e) => setRawJSON(e.target.value)}
                className="liquid-input w-full p-4 rounded-2xl text-xs font-mono text-white placeholder-slate-500 resize-none text-start ltr:text-left rtl:text-left"
                dir="ltr"
              />
            </div>
          )}

          {/* Audience Filters */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-2">
            <div className="space-y-1">
              <label className="text-xs font-semibold text-slate-300">{t('broadcast.target_role', 'Target Role Audience')}</label>
              <select
                value={filterRole}
                onChange={(e) => setFilterRole(e.target.value)}
                className="liquid-input w-full px-3 py-2 rounded-xl text-xs text-white"
              >
                <option value="">{t('broadcast.all_users', 'All Users (Broadcast)')}</option>
                <option value="user">{t('broadcast.role_user', 'Free Regular Users')}</option>
                <option value="author">{t('broadcast.role_author', 'Authors Only')}</option>
                <option value="admin">{t('broadcast.role_admin', 'Administrators')}</option>
              </select>
            </div>

            <div className="flex items-center gap-2 pt-6">
              <input
                type="checkbox"
                id="activeWeek"
                checked={filterActiveWeek}
                onChange={(e) => setFilterActiveWeek(e.target.checked)}
                className="w-4 h-4 accent-cyan-500 rounded"
              />
              <label htmlFor="activeWeek" className="text-xs text-slate-300 cursor-pointer">
                {t('broadcast.active_week', 'Only users active in past 7 days')}
              </label>
            </div>
          </div>

          <button
            onClick={handleBroadcast}
            disabled={broadcasting}
            className="liquid-button px-6 py-3 rounded-2xl text-sm font-bold text-white shadow-lg flex items-center justify-center gap-2 w-full sm:w-auto"
          >
            {broadcasting ? <RefreshCw size={15} className="animate-spin" /> : <Send size={15} />}
            <span>{broadcasting ? t('broadcast.sending_btn', 'Broadcasting to Queue...') : t('broadcast.send_btn', 'Send Broadcast Campaign')}</span>
          </button>

          {result && (
            <div className={`p-4 rounded-2xl text-xs font-medium ${result.error ? 'bg-red-500/15 text-red-300 border border-red-500/30' : 'bg-emerald-500/15 text-emerald-300 border border-emerald-500/30'}`}>
              {result.error ? `${t('common.error', 'Error')}: ${result.error}` : (t('broadcast.success_msg', { count: result.total_queued || 0 }) || `Broadcast dispatched successfully! Queued: ${result.total_queued || 0} messages.`)}
            </div>
          )}
        </div>

        {/* Live Telegram Preview */}
        <div className="liquid-glass-card p-5 rounded-3xl space-y-3">
          <div className="text-xs font-bold text-slate-400 uppercase tracking-wider flex items-center gap-2">
            <Radio size={14} className="text-cyan-400" />
            <span>{t('broadcast.preview_title', 'Telegram Message Preview')}</span>
          </div>

          <div className="p-4 rounded-2xl bg-[#0e1626] border border-white/10 space-y-3 text-xs">
            <div className="flex items-center gap-2">
              <div className="w-7 h-7 rounded-full bg-cyan-500/20 text-cyan-400 font-bold flex items-center justify-center">
                TP
              </div>
              <div className="text-start">
                <div className="font-bold text-white">{t('broadcast.bot_name', 'TelegramPublisher Bot')}</div>
                <div className="text-[10px] text-slate-400">bot</div>
              </div>
            </div>

            <div className="text-slate-200 whitespace-pre-wrap leading-relaxed text-start">
              {content || t('broadcast.preview_placeholder', 'Your formatted announcement text will render here...')}
            </div>

            <div className="space-y-1.5 pt-2">
              <div className="w-full py-2 bg-blue-600/30 text-blue-300 text-center rounded-xl text-xs font-semibold border border-blue-500/30">
                {t('broadcast.open_miniapp', '🚀 Open Mini App')}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
