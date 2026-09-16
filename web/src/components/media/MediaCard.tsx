import React from 'react';
import { Post } from '../../types';
import { Play, Eye, ThumbsUp, Crown, Clock } from 'lucide-react';
import { useTranslation } from '../../context/LanguageContext';

interface MediaCardProps {
  post: Post;
  isSelected?: boolean;
  onSelect: (post: Post) => void;
}

export const MediaCard: React.FC<MediaCardProps> = ({ post, isSelected = false, onSelect }) => {
  const { t } = useTranslation();

  return (
    <div
      onClick={() => onSelect(post)}
      className={`liquid-glass-card rounded-2xl sm:rounded-3xl p-3 sm:p-4 cursor-pointer transition-all duration-300 group flex flex-col justify-between ${
        isSelected ? 'border-cyan-500/50 shadow-lg shadow-cyan-500/10' : ''
      }`}
    >
      {/* Thumbnail Aspect Box */}
      <div className="relative aspect-video w-full rounded-xl sm:rounded-2xl bg-gradient-to-tr from-slate-900 via-indigo-950/40 to-slate-900 overflow-hidden mb-3 flex items-center justify-center border border-white/5 group-hover:border-white/10 transition-colors">
        <div className="w-10 h-10 sm:w-12 sm:h-12 rounded-2xl bg-cyan-500/20 backdrop-blur-md border border-cyan-500/30 flex items-center justify-center text-cyan-300 group-hover:scale-110 transition-transform shadow-lg">
          <Play size={18} className="fill-current ms-0.5 rtl:rotate-180" />
        </div>

        {/* VIP / Protection Badge */}
        <div className="absolute top-2.5 ltr:left-2.5 rtl:right-2.5 flex items-center gap-1.5">
          {post.is_vip && (
            <span className="px-2 py-0.5 rounded-lg bg-amber-500/20 backdrop-blur-md text-amber-300 border border-amber-500/30 text-[10px] font-extrabold flex items-center gap-1">
              <Crown size={10} />
              {t('media_player.vip_badge', 'VIP')}
            </span>
          )}
          <span className="px-2 py-0.5 rounded-lg bg-black/40 backdrop-blur-md text-slate-300 border border-white/10 text-[10px] font-mono flex items-center gap-1">
            <Clock size={10} />
            {t('media_player.ttl_badge', '2m TTL')}
          </span>
        </div>
      </div>

      {/* Title & Metadata */}
      <div className="space-y-2 text-start">
        <h3 className="text-xs sm:text-sm font-bold text-white group-hover:text-cyan-300 transition-colors line-clamp-1">
          {post.title || post.slug}
        </h3>

        <div className="flex items-center justify-between text-[11px] text-slate-400 font-medium">
          <div className="flex items-center gap-1">
            <Eye size={12} className="text-cyan-400" />
            <span>{t('media_player.views_count', { count: post.views_count || 0 }) || `${post.views_count || 0} views`}</span>
          </div>

          <div className="flex items-center gap-1">
            <ThumbsUp size={12} className="text-slate-400" />
            <span>{post.likes_count || 0}</span>
          </div>
        </div>
      </div>
    </div>
  );
};
