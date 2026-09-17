import React, { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { API_ENDPOINTS, apiUrl } from '../constants';

interface SetupGuardProps {
  children: React.ReactNode;
}

export const SetupGuard: React.FC<SetupGuardProps> = ({ children }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [checking, setChecking] = useState<boolean>(() => {
    // If user is already heading to setup or about, skip blocking check
    if (
      location.pathname === '/setup' ||
      location.pathname === '/onboarding' ||
      location.pathname === '/about'
    ) {
      return false;
    }
    // If localStorage already marked it complete, don't show full-page loader initially
    return localStorage.getItem('tp_onboarding_completed') !== 'true';
  });

  useEffect(() => {
    // If on setup or about page, no redirect needed
    if (
      location.pathname === '/setup' ||
      location.pathname === '/onboarding' ||
      location.pathname === '/about'
    ) {
      return;
    }

    let isMounted = true;

    const checkSetupStatus = async () => {
      try {
        const resp = await fetch(apiUrl(API_ENDPOINTS.SETUP_STATUS));
        if (resp.ok) {
          const data = await resp.json();
          if (data.completed) {
            localStorage.setItem('tp_onboarding_completed', 'true');
            if (isMounted) setChecking(false);
          } else {
            localStorage.removeItem('tp_onboarding_completed');
            navigate('/setup', { replace: true });
          }
        } else {
          // If setup status returns non-200 and localStorage is not completed, redirect to setup
          if (localStorage.getItem('tp_onboarding_completed') !== 'true') {
            navigate('/setup', { replace: true });
          } else {
            if (isMounted) setChecking(false);
          }
        }
      } catch (err) {
        // Network error or backend offline - if not verified, guide user to setup wizard
        if (localStorage.getItem('tp_onboarding_completed') !== 'true') {
          navigate('/setup', { replace: true });
        } else {
          if (isMounted) setChecking(false);
        }
      }
    };

    checkSetupStatus();

    return () => {
      isMounted = false;
    };
  }, [location.pathname, navigate]);

  if (checking) {
    return (
      <div className="min-h-[50vh] flex flex-col items-center justify-center space-y-3 text-slate-400">
        <div className="w-8 h-8 rounded-full border-2 border-cyan-400/20 border-t-cyan-400 animate-spin" />
        <span className="text-xs font-medium">Verifying platform configuration...</span>
      </div>
    );
  }

  return <>{children}</>;
};
