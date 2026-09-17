import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { STUDIO_TABS, API_ENDPOINTS, apiUrl } from '../constants';
import { useTranslation } from '../context/LanguageContext';
import { Post, Report } from '../types';
import { StudioPostsTab } from '../components/studio/StudioPostsTab';
import { StudioUploadTab } from '../components/studio/StudioUploadTab';
import { StudioReportsTab } from '../components/studio/StudioReportsTab';
import { Clapperboard } from 'lucide-react';

export const AuthorStudio: React.FC = () => {
  const { tab = 'posts' } = useParams<{ tab?: string }>();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const [posts, setPosts] = useState<Post[]>([]);
  const [reports, setReports] = useState<Report[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const fetchStudioData = async () => {
    try {
      setIsLoading(true);
      const [pResp, rResp] = await Promise.all([
        fetch(apiUrl(`${API_ENDPOINTS.POSTS}?limit=50`)),
        fetch(apiUrl(`${API_ENDPOINTS.REPORTS}?status=pending`)),
      ]);

      if (pResp.ok) {
        const pData = await pResp.json();
        setPosts(pData.posts || []);
      }
      if (rResp.ok) {
        const rData = await rResp.json();
        setReports(rData.reports || []);
      }
    } catch (e) {
      console.warn('Failed to load studio data:', e);
    } finally {
      setIsLoading(false);
    }
  };

  const handleDeletePost = async (id: number) => {
    try {
      await fetch(apiUrl(`${API_ENDPOINTS.POSTS}/${id}`), {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('tp_token')}`,
        },
      });
      fetchStudioData();
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => {
    fetchStudioData();
  }, []);

  return (
    <div className="space-y-6 sm:space-y-8">
      {/* Studio Sub-Navigation Bar */}
      <div className="flex items-center gap-1.5 p-1.5 rounded-2xl liquid-glass border border-white/10 overflow-x-auto no-scrollbar shadow-inner">
        {STUDIO_TABS.map((tItem) => {
          const isActive = tab === tItem.id;
          const labelMap: Record<string, string> = {
            posts: t('studio.tab_posts', 'Published Media'),
            upload: t('studio.tab_upload', 'Upload & Publish'),
            reports: t('studio.tab_reports', 'Content Reports'),
          };
          const label = labelMap[tItem.id] || tItem.label;

          return (
            <button
              key={tItem.id}
              onClick={() => navigate(tItem.path)}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold whitespace-nowrap transition-all ${
                isActive
                  ? 'liquid-pill-active text-white shadow-md'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/[0.04]'
              }`}
            >
              <span>{label}</span>
            </button>
          );
        })}
      </div>

      {/* Tab Content */}
      {tab === 'posts' && (
        <StudioPostsTab posts={posts} onRefresh={fetchStudioData} onDelete={handleDeletePost} />
      )}

      {tab === 'upload' && <StudioUploadTab onSuccess={fetchStudioData} />}

      {tab === 'reports' && (
        <StudioReportsTab reports={reports} onRefresh={fetchStudioData} />
      )}
    </div>
  );
};
