import { useState, useEffect } from 'react';

const LAST_ACTIVE_NAV_KEY = 'last_active_nav';

export function useNavigationState() {
  const [activeNav, setActiveNav] = useState<string>(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem(LAST_ACTIVE_NAV_KEY) || '/dashboard';
    }
    return '/dashboard';
  });

  useEffect(() => {
    localStorage.setItem(LAST_ACTIVE_NAV_KEY, activeNav);
  }, [activeNav]);

  return { activeNav, setActiveNav };
}
