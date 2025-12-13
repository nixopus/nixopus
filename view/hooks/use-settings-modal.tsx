'use client';

import { useState, createContext, useContext, ReactNode, useCallback } from 'react';

interface SettingsModalContextType {
  open: boolean;
  activeCategory: string;
  openSettings: (category?: string) => void;
  closeSettings: () => void;
  setActiveCategory: (category: string) => void;
}

const SettingsModalContext = createContext<SettingsModalContextType | undefined>(undefined);

export function SettingsModalProvider({ children }: { children: ReactNode }) {
  const [open, setOpen] = useState(false);
  const [activeCategory, setActiveCategory] = useState('general');

  const openSettings = useCallback((category?: string) => {
    if (category) setActiveCategory(category);
    setOpen(true);
  }, []);

  const closeSettings = useCallback(() => setOpen(false), []);

  return (
    <SettingsModalContext.Provider
      value={{ open, activeCategory, openSettings, closeSettings, setActiveCategory }}
    >
      {children}
    </SettingsModalContext.Provider>
  );
}

export function useSettingsModal() {
  const context = useContext(SettingsModalContext);
  if (!context) throw new Error('useSettingsModal must be used within SettingsModalProvider');
  return context;
}
