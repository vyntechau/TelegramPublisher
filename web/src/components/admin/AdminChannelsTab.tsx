import React, { useState } from 'react';
import { Channel } from '../../types';
import { API_ENDPOINTS, apiUrl } from '../../constants';
import { Send, Plus, Trash2 } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface AdminChannelsTabProps {
  channels: Channel[];
  onRefresh: () => void;
}

export const AdminChannelsTab: React.FC<AdminChannelsTabProps> = ({ channels, onRefresh }) => {
  const { t } = useTranslation();
  const [newTgID, setNewTgID] = useState('');
  const [newTitle, setNewTitle] = useState('');
  const [newLink, setNewLink] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleAddChannel = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTgID.trim() || !newTitle.trim()) return;

    try {
      setIsSubmitting(true);
      await fetch(apiUrl(API_ENDPOINTS.CHANNELS), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({
          telegram_id: newTgID,
          title: newTitle,
          invite_link: newLink,
          is_active: true,
        }),
      });
      setNewTgID('');
      setNewTitle('');
      setNewLink('');
      onRefresh();
    } catch (e) {
      console.error(e);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleRemoveChannel = async (id: number) => {
    try {
      await fetch(apiUrl(`${API_ENDPOINTS.CHANNELS}/${id}`), {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
      });
      onRefresh();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-xl font-bold text-white tracking-tight">{t('channels.title', 'Connected Broadcast Channels')}</h2>
        <p className="text-xs text-slate-400">{t('channels.desc', 'Telegram channels configured for automatic post distribution and forced subscription gates')}</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Add Channel Form */}
        <form onSubmit={handleAddChannel} className="liquid-glass-card p-5 sm:p-6 rounded-3xl space-y-4">
          <div className="text-xs font-bold text-white flex items-center gap-1.5 uppercase tracking-wider">
            <Plus size={14} className="text-cyan-400" />
            <span>{t('channels.connect_card', 'Connect Channel')}</span>
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">{t('channels.channel_title_label', 'Channel Title / Name')}</label>
            <input
              type="text"
              required
              value={newTitle}
              onChange={(e) => setNewTitle(e.target.value)}
              placeholder={t('channels.channel_title_ph', 'e.g. Cinema Hub HD')}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs text-white text-start"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">{t('channels.tg_id_label', 'Telegram ID or Username')}</label>
            <input
              type="text"
              required
              value={newTgID}
              onChange={(e) => setNewTgID(e.target.value)}
              placeholder={t('channels.tg_id_ph', 'e.g. @cinemahub or -100123456789')}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white text-start"
              dir="ltr"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">{t('channels.invite_link_label', 'Invite Link (Optional)')}</label>
            <input
              type="text"
              value={newLink}
              onChange={(e) => setNewLink(e.target.value)}
              placeholder={t('channels.invite_link_ph', 'https://t.me/+joinlink')}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs text-white text-start"
              dir="ltr"
            />
          </div>

          <button
            type="submit"
            disabled={isSubmitting}
            className="liquid-button w-full py-2.5 rounded-xl text-xs font-bold text-white shadow-md flex items-center justify-center gap-2"
          >
            <Plus size={14} />
            <span>{isSubmitting ? t('channels.btn_adding', 'Adding...') : t('channels.btn_add', 'Add Channel')}</span>
          </button>
        </form>

        {/* Channels List */}
        <div className="lg:col-span-2 space-y-3">
          {channels.length === 0 ? (
            <div className="liquid-glass p-8 rounded-3xl text-center text-slate-400 text-xs">
              {t('channels.no_channels', 'No channels connected yet. Add one using the form.')}
            </div>
          ) : (
            channels.map((ch) => (
              <div key={ch.id} className="liquid-glass-card p-4 rounded-2xl flex items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-cyan-500/10 border border-cyan-500/20 flex items-center justify-center text-cyan-400 shrink-0">
                    <Send size={18} />
                  </div>
                  <div className="text-start">
                    <div className="text-sm font-bold text-white">{ch.title}</div>
                    <div className="text-xs font-mono text-slate-400" dir="ltr">{ch.telegram_id}</div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <button
                    onClick={() => handleRemoveChannel(ch.id)}
                    className="p-2 rounded-xl text-red-400 hover:bg-red-500/10 transition-colors"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};
