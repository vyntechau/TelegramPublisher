import { User } from './auth';
import { Post, Report } from './media';

export interface AnalyticsSummary {
  total_users: number;
  active_users_24h: number;
  vip_subscribers: number;
  total_posts: number;
  total_views: number;
  total_reports: number;
  pending_reports: number;
  total_revenue_usd: number;
  top_posts?: Post[];
  recent_users?: User[];
}

export interface Channel {
  id: number;
  telegram_id: number | string;
  title: string;
  username?: string;
  invite_link?: string;
  is_active: boolean;
  member_count?: number;
  created_at?: string;
}

export interface BroadcastPayload {
  message: string;
  target_audience: 'all' | 'free' | 'vip';
  include_media_preview?: boolean;
  inline_buttons?: Array<{ text: string; url: string }>;
}

export interface PaymentPlan {
  id: string;
  name: string;
  duration_days: number;
  price_usd: number;
  features: string[];
  is_popular?: boolean;
}
