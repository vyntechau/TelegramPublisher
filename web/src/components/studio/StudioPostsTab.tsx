import React, { useState } from 'react';
import { Post } from '../../types';
import { Copy, Check, Eye, Trash2, Film, RefreshCw } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface StudioPostsTabProps {
  posts: Post[];
  onRefresh: () => void;
  onDelete?: (id: number) => void;
}

export const StudioPostsTab: React.FC<StudioPostsTabProps> = ({ posts, onRefresh, onDelete }) => {
  const { t } = useTranslation();
  const [copiedSlug, setCopiedSlug] = useState<string | null>(null);

  const copyToClipboard = (text: string, slugKey: string) => {
    navigator.clipboard.writeText(text);
    setCopiedSlug(slugKey);
    setTimeout(() => setCopiedSlug(null), 2000);
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 className="text-xl font-bold text-white tracking-tight">{t('studio_posts.title', 'Your Published Media Releases')}</h2>
          <p className="text-xs text-slate-400">{t('studio_posts.desc', 'Inventory of video assets, deep-links, and auto-cleanup protection timers')}</p>
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
                <th className="p-4 text-start">{t('studio_posts.col_title', 'Release Title / Slug')}</th>
                <th className="p-4 text-start">{t('studio_posts.col_deep_link', 'Deep Link')}</th>
                <th className="p-4 text-start">{t('studio_posts.col_views', 'Views')}</th>
                <th className="p-4 text-start">{t('studio_posts.col_access', 'Access Level')}</th>
                <th className="p-4 text-end">{t('studio_posts.col_actions', 'Actions')}</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5">
              {posts.length === 0 ? (
                <tr>
                  <td colSpan={5} className="p-8 text-center text-slate-400">
                    {t('studio_posts.no_releases', 'No releases uploaded yet. Click "Upload Release" above to start publishing.')}
                  </td>
                </tr>
              ) : (
                posts.map((p) => {
                  const botDeepLink = `https://t.me/YourBot?start=${p.slug}`;
                  const isCopied = copiedSlug === p.slug;

                  return (
                    <tr key={p.id} className="hover:bg-white/[0.02] transition-colors">
                      <td className="p-4">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400 shrink-0">
                            <Film size={16} />
                          </div>
                          <div className="text-start">
                            <div className="font-bold text-white truncate max-w-[200px]">{p.title || p.slug}</div>
                            <div className="text-[10px] font-mono text-slate-400" dir="ltr">{p.slug}</div>
                          </div>
                        </div>
                      </td>

                      <td className="p-4 text-start">
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-slate-300 text-[11px] truncate max-w-[150px]" dir="ltr">
                            {botDeepLink}
                          </span>
                          <button
                            onClick={() => copyToClipboard(botDeepLink, p.slug)}
                            className="p-1.5 rounded-lg bg-white/[0.04] hover:bg-white/[0.1] text-slate-300 transition-all"
                            title={t('studio_posts.copy_tooltip', 'Copy deep-link')}
                          >
                            {isCopied ? <Check size={12} className="text-emerald-400" /> : <Copy size={12} />}
                          </button>
                        </div>
                      </td>

                      <td className="p-4 font-mono text-slate-300 text-start">
                        <div className="flex items-center gap-1">
                          <Eye size={13} className="text-cyan-400" />
                          <span>{p.views_count || 0}</span>
                        </div>
                      </td>

                      <td className="p-4 text-start">
                        <span className={`px-2 py-0.5 rounded-lg text-[10px] font-bold uppercase ${
                          p.is_vip ? 'bg-amber-500/15 text-amber-400 border border-amber-500/30' : 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30'
                        }`}>
                          {p.is_vip ? t('studio_posts.tier_vip', 'VIP Tier') : t('studio_posts.tier_public', 'Public')}
                        </span>
                      </td>

                      <td className="p-4 text-end">
                        {onDelete && (
                          <button
                            onClick={() => onDelete(p.id)}
                            className="p-1.5 rounded-lg text-red-400 hover:bg-red-500/10 transition-colors"
                            title={t('studio_posts.delete_tooltip', 'Delete post')}
                          >
                            <Trash2 size={14} />
                          </button>
                        )}
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
