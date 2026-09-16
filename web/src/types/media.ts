export interface Post {
  id: number;
  slug: string;
  title: string;
  description: string;
  thumbnail_url?: string;
  video_file_id?: string;
  video_url?: string;
  duration?: number;
  is_vip: boolean;
  is_published: boolean;
  views_count: number;
  likes_count: number;
  dislikes_count: number;
  author_id: number;
  created_at: string;
}

export interface Report {
  id: number;
  post_id: number;
  user_id: number;
  reason: string;
  status: 'pending' | 'resolved' | 'rejected';
  resolution_note?: string;
  resolver_id?: number;
  created_at: string;
  updated_at?: string;
  post_title?: string;
  post_slug?: string;
}

export interface PostReactionPayload {
  post_id: number;
  reaction: 'like' | 'dislike';
}
