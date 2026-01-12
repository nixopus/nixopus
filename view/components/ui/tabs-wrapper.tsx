'use client';

import React, { createContext, useContext } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { cn } from '@/lib/utils';
import type { LucideIcon } from 'lucide-react';

export interface TabItem {
  value: string;
  label: string;
  icon?: LucideIcon;
  content: React.ReactNode;
  disabled?: boolean;
}

interface TabsWrapperContextValue {
  tabs: TabItem[];
  tabsListClassName?: string;
  tabsTriggerClassName?: string;
  tabsContentClassName?: string;
  showTabsCondition: boolean;
}

const TabsWrapperContext = createContext<TabsWrapperContextValue | null>(null);

export function useTabsWrapper() {
  const context = useContext(TabsWrapperContext);
  if (!context) {
    throw new Error('useTabsWrapper must be used within TabsWrapper');
  }
  return context;
}

export interface TabsWrapperProps {
  value: string;
  onValueChange: (value: string) => void;
  tabs: TabItem[];
  className?: string;
  tabsListClassName?: string;
  tabsTriggerClassName?: string;
  tabsContentClassName?: string;
  showTabsCondition?: boolean;
  defaultContent?: React.ReactNode;
  children: React.ReactNode;
}

export function TabsWrapper({
  value,
  onValueChange,
  tabs,
  className,
  tabsListClassName,
  tabsTriggerClassName,
  tabsContentClassName,
  showTabsCondition = true,
  defaultContent,
  children
}: TabsWrapperProps) {
  const shouldShowTabs = showTabsCondition && tabs.length > 0;

  const contextValue: TabsWrapperContextValue = {
    tabs,
    tabsListClassName,
    tabsTriggerClassName,
    tabsContentClassName,
    showTabsCondition: shouldShowTabs
  };

  const renderTabsContent = () => {
    if (!shouldShowTabs) {
      return defaultContent ? (
        <div className={cn('mt-6', tabsContentClassName)}>{defaultContent}</div>
      ) : null;
    }

    return (
      <>
        {tabs.map((tab) => (
          <TabsContent
            key={tab.value}
            value={tab.value}
            className={cn('mt-6', tabsContentClassName)}
          >
            {tab.content}
          </TabsContent>
        ))}
      </>
    );
  };

  return (
    <TabsWrapperContext.Provider value={contextValue}>
      <Tabs value={value} onValueChange={onValueChange} className={cn('w-full', className)}>
        {children}
        {renderTabsContent()}
      </Tabs>
    </TabsWrapperContext.Provider>
  );
}

export interface TabsWrapperListProps {
  className?: string;
}

export function TabsWrapperList({ className }: TabsWrapperListProps) {
  const { tabs, tabsListClassName, tabsTriggerClassName, showTabsCondition } = useTabsWrapper();

  if (!showTabsCondition) return null;

  return (
    <TabsList className={cn('w-fit justify-start', tabsListClassName, className)}>
      {tabs.map((tab) => {
        const Icon = tab.icon;
        return (
          <TabsTrigger
            key={tab.value}
            value={tab.value}
            disabled={tab.disabled}
            className={tabsTriggerClassName}
          >
            {Icon && <Icon className="mr-2 h-4 w-4" />}
            {tab.label}
          </TabsTrigger>
        );
      })}
    </TabsList>
  );
}

export default TabsWrapper;
