'use client';

import * as React from 'react';
import { createPortal } from 'react-dom';
import { cn } from '@/lib/utils';
import { useSidebar } from '@/components/ui/sidebar';
import Link from 'next/link';

interface HoverMenuItem {
  title: string;
  url: string;
}

interface SidebarHoverMenuProps {
  items: HoverMenuItem[];
  children: React.ReactNode;
  className?: string;
}

export function SidebarHoverMenu({ items, children, className }: SidebarHoverMenuProps) {
  const { state } = useSidebar();
  const [isHovered, setIsHovered] = React.useState(false);
  const [buttonRect, setButtonRect] = React.useState<DOMRect | null>(null);
  const timeoutRef = React.useRef<number | undefined>(undefined);
  const buttonRef = React.useRef<HTMLDivElement>(null);

  const handleMouseEnter = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    if (buttonRef.current) {
      setButtonRect(buttonRef.current.getBoundingClientRect());
    }
    setIsHovered(true);
  };

  const handleMouseLeave = () => {
    timeoutRef.current = window.setTimeout(() => {
      setIsHovered(false);
    }, 150);
  };

  React.useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  if (state !== 'collapsed') {
    return <>{children}</>;
  }

  const menuContent = isHovered && items.length > 0 && buttonRect && (
    <div
      className={cn(
        'fixed z-50 min-w-[200px] rounded-md border bg-secondary p-2 shadow-lg',
        'animate-in fade-in-0 zoom-in-95 duration-200',
        className
      )}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{
        position: 'fixed',
        left: `${buttonRect.right + 8}px`,
        top: `${buttonRect.top}px`,
        zIndex: 9999
      }}
    >
      <div className="space-y-1">
        {items.map((item, index) => (
          <React.Fragment key={item.title}>
            <div className="px-2 py-1.5 rounded-sm hover:bg-background hover:text-muted-foreground transition-colors cursor-pointer">
              <Link href={item.url} className="block text-sm font-medium">
                {item.title}
              </Link>
            </div>
            {index < items.length - 1 && <div className="border-t border-muted my-1"></div>}
          </React.Fragment>
        ))}
      </div>
    </div>
  );

  return (
    <>
      <div
        ref={buttonRef}
        className="relative"
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
      >
        {children}
      </div>
      {typeof window !== 'undefined' && menuContent && createPortal(menuContent, document.body)}
    </>
  );
}
