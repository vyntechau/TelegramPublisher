import React, { createContext, useContext, useState, useEffect } from 'react';
import { User, Subscription, UserRole, AuthContextType } from '../types';
import { API_ENDPOINTS, apiUrl } from '../constants';

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [token, setToken] = useState<string | null>(() => localStorage.getItem('tp_token'));
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [role, setRoleState] = useState<UserRole>('owner');

  const setRole = (newRole: UserRole) => {
    setRoleState(newRole);
    if (user) {
      setUser({ ...user, role: newRole });
    }
  };

  const validateToken = async (testToken: string): Promise<boolean> => {
    try {
      const resp = await fetch(apiUrl(API_ENDPOINTS.AUTH_ME), {
        headers: { Authorization: `Bearer ${testToken}` },
      });
      if (resp.ok) {
        const data = await resp.json();
        setUser(data.user);
        setRoleState(data.user.role || 'user');
        setSubscription(data.subscription || null);
        setToken(testToken);
        localStorage.setItem('tp_token', testToken);
        return true;
      }
      return false;
    } catch {
      return false;
    }
  };

  const loginWithToken = async (newToken: string): Promise<boolean> => {
    setIsLoading(true);
    try {
      const cleanToken = newToken.trim();
      const valid = await validateToken(cleanToken);
      if (!valid) {
        localStorage.removeItem('tp_token');
        setToken(null);
      }
      return valid;
    } finally {
      setIsLoading(false);
    }
  };

  const loginWithTelegram = async () => {
    try {
      setIsLoading(true);

      // 1. Check for one-click login token in URL (?token=... or ?auth_token=...)
      const urlParams = new URLSearchParams(window.location.search);
      const urlToken = urlParams.get('token') || urlParams.get('auth_token');
      if (urlToken) {
        urlParams.delete('token');
        urlParams.delete('auth_token');
        const newSearch = urlParams.toString();
        const cleanUrl = window.location.pathname + (newSearch ? `?${newSearch}` : '') + window.location.hash;
        window.history.replaceState({}, document.title, cleanUrl);

        const valid = await validateToken(urlToken);
        if (valid) return;
      }

      // 2. Check existing token in localStorage
      const savedToken = localStorage.getItem('tp_token');
      if (savedToken) {
        const valid = await validateToken(savedToken);
        if (valid) return;
        localStorage.removeItem('tp_token');
        setToken(null);
      }

      // 3. If inside Telegram WebApp, exchange initData
      const tg = (window as any).Telegram?.WebApp;
      const initData = tg?.initData || '';

      if (initData) {
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
          return;
        }
      }

      // 4. Default unauthenticated state
      setUser(null);
    } catch (e) {
      console.warn('Telegram auth initialization:', e);
      setUser(null);
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
        loginWithToken,
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
