import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, useLocation } from 'react-router-dom';
import { API_ENDPOINTS } from '../constants';
import { Post } from '../types';
import { useTranslation } from '../context/LanguageContext';
import { MediaCard } from '../components/media/MediaCard';
import { VideoPlayerModal } from '../components/media/VideoPlayerModal';
import { ReportModal } from '../components/media/ReportModal';
import { VipModal } from '../components/media/VipModal';
import { Crown, Sparkles, RefreshCw, Smartphone } from 'lucide-react';

export const UserMiniApp: React.FC = () => {
  const { slug } = useParams<{ slug?: string }>();
  const location = useLocation();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const [posts, setPosts] = useState<Post[]>([]);
  const [selectedPost, setSelectedPost] = useState<Post | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [reportModalOpen, setReportModalOpen] = useState<boolean>(false);
  const [vipModalOpen, setVipModalOpen] = useState<boolean>(location.pathname === '/vip');
  const [activeFilter, setActiveFilter] = useState<'all' | 'vip'>('all');

  const fetchPosts = async () => {
    try {
      setIsLoading(true);
      const resp = await fetch(`${API_ENDPOINTS.POSTS}?limit=30`);
      if (resp.ok) {
        const data = await resp.json();
        const list: Post[] = data.posts || [];
        setPosts(list);

        if (slug) {
          const match = list.find((p) => p.slug === slug);
          if (match) setSelectedPost(match);
        } else if (list.length > 0 && !selectedPost) {
          setSelectedPost(list[0]);
        }
      }
    } catch (e) {
      console.warn('Failed to fetch media catalog:', e);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchPosts();
  }, [slug]);

  useEffect(() => {
    if (location.pathname === '/vip') {
      setVipModalOpen(true);
    }
  }, [location.pathname]);

  const handleSelectPost = (post: Post) => {
    setSelectedPost(post);
    navigate(`/watch/${post.slug}`);
  };

  const handleReaction = async (postId: number, reaction: 'like' | 'dislike') => {
    try {
      const resp = await fetch(API_ENDPOINTS.POST_REACT, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
        body: JSON.stringify({ post_id: postId, reaction }),
      });
      if (resp.ok) {
        const updated = await resp.json();
        setPosts((prev) => prev.map((p) => (p.id === postId ? updated : p)));
        if (selectedPost?.id === postId) {
          setSelectedPost(updated);
        }
      }
    } catch (e) {
      console.error('Failed to react:', e);
    }
  };

  const filteredPosts = posts.filter((p) => (activeFilter === 'vip' ? p.is_vip : true));

  return (
    <div className="space-y-6 sm:space-y-8">
      {/* Top Banner / Actions Bar */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <button
            onClick={() => setActiveFilter('all')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all ${
              activeFilter === 'all' ? 'liquid-pill-active text-white' : 'liquid-pill text-slate-400'
            }`}
          >
            {t('user.all_media', 'All Media Releases')}
          </button>
          <button
            onClick={() => setActiveFilter('vip')}
            className={`px-4 py-2 rounded-xl text-xs font-bold transition-all flex items-center gap-1.5 ${
              activeFilter === 'vip' ? 'liquid-pill-active text-white' : 'liquid-pill text-slate-400'
            }`}
          >
            <Crown size={13} className="text-amber-400" />
            <span>{t('user.vip_active_badge', 'VIP Catalog')}</span>
          </button>
        </div>

        <button
          onClick={() => setVipModalOpen(true)}
          className="liquid-button px-4 py-2 rounded-xl text-xs font-bold text-white shadow-md flex items-center gap-1.5 self-start sm:self-auto"
        >
          <Crown size={14} className="text-amber-300" />
          <span>{t('user.upgrade_vip', 'Get VIP Membership')}</span>
        </button>
      </div>

      {/* Selected Video Player Area (if on /watch/:slug or selected) */}
      {selectedPost && (
        <VideoPlayerModal
          post={selectedPost}
          onReact={handleReaction}
          onOpenReport={() => setReportModalOpen(true)}
        />
      )}

      {/* Media Catalog Grid */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-bold text-slate-200 uppercase tracking-wider">
            {t('user.explore_catalog', { count: filteredPosts.length }) || `Explore Media Catalog (${filteredPosts.length})`}
          </h2>
          <button onClick={fetchPosts} className="text-slate-400 hover:text-white text-xs">
            <RefreshCw size={13} className={isLoading ? 'animate-spin' : ''} />
          </button>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
          {filteredPosts.map((post) => (
            <MediaCard
              key={post.id}
              post={post}
              isSelected={selectedPost?.id === post.id}
              onSelect={handleSelectPost}
            />
          ))}
        </div>
      </div>

      {/* Report Broken Modal */}
      {selectedPost && (
        <ReportModal
          postId={selectedPost.id}
          isOpen={reportModalOpen}
          onClose={() => setReportModalOpen(false)}
        />
      )}

      {/* VIP Upgrade Modal */}
      <VipModal isOpen={vipModalOpen} onClose={() => setVipModalOpen(false)} />
    </div>
  );
};
