import { useState, useEffect } from 'react';

const COLLAPSIBLE_STATE_KEY = 'nav_collapsible_state';

export function useCollapsibleState() {
  const [collapsedItems, setCollapsedItems] = useState<Record<string, boolean>>(() => {
    if (typeof window !== 'undefined') {
      const savedState = localStorage.getItem(COLLAPSIBLE_STATE_KEY);
      return savedState ? JSON.parse(savedState) : {};
    }
    return {};
  });

  useEffect(() => {
    localStorage.setItem(COLLAPSIBLE_STATE_KEY, JSON.stringify(collapsedItems));
  }, [collapsedItems]);

  const toggleItem = (itemId: string) => {
    setCollapsedItems((prev) => ({
      ...prev,
      [itemId]: !prev[itemId]
    }));
  };

  const isItemCollapsed = (itemId: string) => {
    return collapsedItems[itemId] ?? false;
  };

  return { isItemCollapsed, toggleItem };
}
