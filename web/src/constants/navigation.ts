import { Server, Send, Radio, Coins, Sparkles, BarChart3, Users, MessageSquare, AlertCircle, Settings } from 'lucide-react';

export const ADMIN_TABS = [
  { id: 'overview', label: 'Overview', icon: BarChart3, path: '/admin/overview' },
  { id: 'broadcast', label: 'Broadcast', icon: MessageSquare, path: '/admin/broadcast' },
  { id: 'users', label: 'Users', icon: Users, path: '/admin/users' },
  { id: 'channels', label: 'Channels', icon: Send, path: '/admin/channels' },
  { id: 'reports', label: 'Reports', icon: AlertCircle, path: '/admin/reports' },
  { id: 'settings', label: 'Settings', icon: Settings, path: '/admin/settings' },
] as const;

export const STUDIO_TABS = [
  { id: 'posts', label: 'Posts & Inventory', path: '/studio/posts' },
  { id: 'upload', label: 'Upload Release', path: '/studio/upload' },
  { id: 'reports', label: 'Reports & Tickets', path: '/studio/reports' },
] as const;

export const ONBOARDING_STEPS = [
  { id: 1, title: 'API & Security', icon: Server, desc: 'Endpoint & Protection' },
  { id: 2, title: 'Bot & Channels', icon: Send, desc: 'Auto-Publishing' },
  { id: 3, title: 'Post Keyboards', icon: Radio, desc: 'Interactive Buttons' },
  { id: 4, title: 'Crypto Subscriptions', icon: Coins, desc: 'AzPays / Coinbase / NOWPayments' },
  { id: 5, title: 'Launch', icon: Sparkles, desc: 'Ready to Publish' },
] as const;
