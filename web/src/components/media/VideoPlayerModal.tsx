import React from 'react';
import { Post } from '../../types';
import { Play, ThumbsUp, ThumbsDown, AlertTriangle, ShieldCheck, Clock } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface VideoPlayerModalProps {
  post: Post;
  onReact: (postId: number, reaction: 'like' | 'dislike') => void;
  onOpenReport: () => void;
  onShare?: () => void;
}

export const VideoPlayerModal: React.FC<VideoPlayerModalProps> = ({
  post,
  onReact,
  onOpenReport,
}) => {
  const { t } = useTranslation();

  return (
    <div className="liquid-glass-card rounded-3xl p-5 sm:p-8 space-y-6 border border-white/10 shadow-2xl relative overflow-hidden text-start">
      {/* Video Viewport Container */}
      <div className="relative aspect-video w-full rounded-2xl bg-black overflow-hidden flex items-center justify-center border border-white/10 shadow-inner">
        <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-black/30 pointer-events-none" />

        {/* Big Play Button Overlay */}
        <div className="w-16 h-16 rounded-full bg-gradient-to-tr from-cyan-500 to-blue-600 flex items-center justify-center text-white shadow-xl shadow-cyan-500/30 cursor-pointer hover:scale-110 active:scale-95 transition-transform z-10">
          <Play size={28} className="fill-current ms-1 rtl:rotate-180" />
        </div>

        {/* Security / Copyright Watermark Banner */}
        <div className="absolute bottom-3 start-3 end-3 flex items-center justify-between text-xs text-slate-300 backdrop-blur-md bg-black/40 p-2.5 rounded-xl border border-white/10 z-10">
          <div className="flex items-center gap-1.5 text-[11px]">
            <ShieldCheck size={14} className="text-emerald-400" />
            <span>{t('media_player.encrypted_stream', 'Encrypted Stream & Auto-Delete Protection')}</span>
          </div>
          <div className="flex items-center gap-1 text-[11px] font-mono text-amber-300">
            <Clock size={12} />
            <span>{t('media_player.expires_in', 'Expires in 2m')}</span>
          </div>
        </div>
      </div>

      {/* Title & Metadata */}
      <div className="space-y-3">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <h1 className="text-xl sm:text-2xl font-black text-white tracking-tight">
            {post.title || post.slug}
          </h1>

          {/* Reaction & Action Pills */}
          <div className="flex items-center gap-2">
            <button
              onClick={() => onReact(post.id, 'like')}
              className="liquid-pill px-3 py-1.5 rounded-xl text-xs font-semibold text-slate-200 hover:text-white flex items-center gap-1.5 transition-all"
            >
              <ThumbsUp size={13} className="text-cyan-400" />
              <span>{post.likes_count || 0}</span>
            </button>

            <button
              onClick={() => onReact(post.id, 'dislike')}
              className="liquid-pill px-3 py-1.5 rounded-xl text-xs font-semibold text-slate-200 hover:text-white flex items-center gap-1.5 transition-all"
            >
              <ThumbsDown size={13} className="text-slate-400" />
              <span>{post.dislikes_count || 0}</span>
            </button>

            <button
              onClick={onOpenReport}
              className="liquid-pill px-3 py-1.5 rounded-xl text-xs font-semibold text-red-400 hover:bg-red-500/10 flex items-center gap-1.5 transition-all"
            >
              <AlertTriangle size={13} />
              <span className="hidden sm:inline">{t('media_player.report_broken', 'Report Broken')}</span>
            </button>
          </div>
        </div>

        {post.description && (
          <p className="text-xs sm:text-sm text-slate-300 leading-relaxed text-start">
            {post.description}
          </p>
        )}
      </div>
    </div>
  );
};
