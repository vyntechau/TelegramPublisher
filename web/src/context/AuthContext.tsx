import React, { createContext, useContext, useState, useEffect } from 'react';
import { User, Subscription, UserRole, AuthContextType } from '../types';
import { API_ENDPOINTS, apiUrl } from '../constants';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [token, setToken] = useState<string | null>(localStorage.getItem('tp_token'));
  const [isLoading, setIsLoading] = useState(true);

  // Default detected role
  const [role, setRoleState] = useState<UserRole>('owner');

  const setRole = (newRole: UserRole) => {
    setRoleState(newRole);
    if (user) {
      setUser({ ...user, role: newRole });
    }
  };

  const loginWithTelegram = async () => {
    try {
      setIsLoading(true);
      const tg = (window as any).Telegram?.WebApp;
      const initData = tg?.initData || '';

      const resp = await fetch(apiUrl(API_ENDPOINTS.AUTH_TELEGRAM), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ init_data: initData }),
      });

      if (resp.ok) {
        const data = await resp.json();
        setToken(data.token);
        localStorage.setItem('tp_token', data.token);
        setUser(data.user);
        setRoleState(data.user.role || 'user');
      } else {
        const mockUser: User = {
          id: 1,
          telegram_id: 123456789,
          username: 'demo_admin',
          first_name: 'VynTech Admin',
          role: role,
          status: 'active',
        };
        setUser(mockUser);
      }
    } catch (e) {
      console.warn('Telegram auth fallback:', e);
      setUser({
        id: 1,
        telegram_id: 123456789,
        username: 'demo_admin',
        first_name: 'VynTech Admin',
        role: role,
        status: 'active',
      });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loginWithTelegram();
  }, []);

  const logout = () => {
    localStorage.removeItem('tp_token');
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        subscription,
        token,
        role: user?.role || role,
        setRole,
        isLoading,
        loginWithTelegram,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
