import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { API_ENDPOINTS, apiUrl } from '../../constants';
import { Upload, RefreshCw } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface StudioUploadTabProps {
  onSuccess: () => void;
}

export const StudioUploadTab: React.FC<StudioUploadTabProps> = ({ onSuccess }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [slug, setSlug] = useState('');
  const [title, setTitle] = useState('');
  const [fileID, setFileID] = useState('');
  const [fileUniqueID, setFileUniqueID] = useState('');
  const [caption, setCaption] = useState('');
  const [ttl, setTTL] = useState(120);
  const [isVip, setIsVip] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleCreatePost = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!fileID && !title) return;
    setIsSubmitting(true);

    try {
      const autoSlug = slug || 'p_' + Math.random().toString(36).substring(2, 9);
      const resp = await fetch(apiUrl(API_ENDPOINTS.POSTS), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({
          slug: autoSlug,
          title: title || autoSlug,
          file_id: fileID,
          file_unique_id: fileUniqueID,
          file_type: 'video',
          caption: caption,
          duration: ttl,
          is_vip: isVip,
          is_published: true,
        }),
      });

      if (resp.ok) {
        setSlug('');
        setTitle('');
        setFileID('');
        setFileUniqueID('');
        setCaption('');
        onSuccess();
        navigate('/studio/posts');
      }
    } catch (e) {
      console.error('Failed to create post', e);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="space-y-6 max-w-3xl mx-auto">
      <div>
        <h2 className="text-xl font-bold text-white tracking-tight">{t('studio_upload.title', 'Upload & Register Media Release')}</h2>
        <p className="text-xs text-slate-400">{t('studio_upload.desc', 'Register Telegram video file IDs or stream URLs with automatic 2-minute expiration')}</p>
      </div>

      <form onSubmit={handleCreatePost} className="liquid-glass-card p-6 sm:p-8 rounded-3xl space-y-5 border border-white/10 shadow-2xl">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">{t('studio_upload.label_title', 'Release Title')}</label>
            <input
              type="text"
              required
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder={t('studio_upload.ph_title', 'e.g. Cyberpunk Episode 1 [1080p]')}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs sm:text-sm text-white text-start"
            />
          </div>

          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">{t('studio_upload.label_slug', 'Custom URL Slug (Optional)')}</label>
            <input
              type="text"
              value={slug}
              onChange={(e) => setSlug(e.target.value)}
              placeholder={t('studio_upload.ph_slug', 'e.g. cyberpunk-ep1')}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs sm:text-sm font-mono text-white text-start"
              dir="ltr"
            />
          </div>
        </div>

        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-slate-300">{t('studio_upload.label_file_id', 'Telegram File ID or CDN Video URL')}</label>
          <input
            type="text"
            required
            value={fileID}
            onChange={(e) => setFileID(e.target.value)}
            placeholder={t('studio_upload.ph_file_id', 'BAACAgIAAxkBAAI...')}
            className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs sm:text-sm font-mono text-white text-start"
            dir="ltr"
          />
          <p className="text-[10px] text-slate-400">
            {t('studio_upload.hint_file_id', 'Forward your video file to the bot or input direct streaming media URL.')}
          </p>
        </div>

        <div className="space-y-1.5">
          <label className="text-xs font-semibold text-slate-300">{t('studio_upload.label_caption', 'Caption / Summary')}</label>
          <textarea
            rows={3}
            value={caption}
            onChange={(e) => setCaption(e.target.value)}
            placeholder={t('studio_upload.ph_caption', 'Release description, tags, and audio details...')}
            className="liquid-input w-full p-3.5 rounded-xl text-xs sm:text-sm text-white resize-none text-start"
          />
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-1">
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-slate-300">{t('studio_upload.label_ttl', 'Auto-Cleanup Duration (Seconds)')}</label>
            <input
              type="number"
              value={ttl}
              onChange={(e) => setTTL(Number(e.target.value))}
              className="liquid-input w-full px-3.5 py-2.5 rounded-xl text-xs font-mono text-white text-start"
              dir="ltr"
            />
          </div>

          <div className="flex items-center justify-between p-3.5 rounded-xl liquid-glass border border-white/10 self-end">
            <span className="text-xs font-semibold text-slate-300">{t('studio_upload.require_vip', 'Require VIP Subscription')}</span>
            <input
              type="checkbox"
              checked={isVip}
              onChange={(e) => setIsVip(e.target.checked)}
              className="w-4 h-4 accent-amber-500 rounded ms-2"
            />
          </div>
        </div>

        <button
          type="submit"
          disabled={isSubmitting}
          className="liquid-button w-full py-3 rounded-2xl text-sm font-bold text-white shadow-lg flex items-center justify-center gap-2"
        >
          {isSubmitting ? <RefreshCw size={15} className="animate-spin" /> : <Upload size={15} />}
          <span>{isSubmitting ? t('studio_upload.btn_publishing', 'Registering Release...') : t('studio_upload.btn_publish', 'Publish Release Now')}</span>
        </button>
      </form>
    </div>
  );
};
