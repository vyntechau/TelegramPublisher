export type UserRole = 'user' | 'author' | 'admin' | 'owner';

export interface User {
  id: number;
  telegram_id: number;
  username: string;
  first_name: string;
  role: UserRole;
  status: 'active' | 'banned' | 'pending';
  created_at?: string;
}

export interface Subscription {
  id: number;
  user_id?: number;
  telegram_id?: number;
  tier: string;
  status: 'active' | 'expired' | 'canceled';
  starts_at?: string;
  expires_at: string;
}

export interface AuthContextType {
  user: User | null;
  subscription: Subscription | null;
  token: string | null;
  role: UserRole;
  setRole: (role: UserRole) => void;
  isLoading: boolean;
  loginWithTelegram: () => Promise<void>;
  loginWithToken: (token: string) => Promise<boolean>;
  logout: () => void;
}
