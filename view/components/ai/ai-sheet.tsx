'use client';

import React from 'react';
import { Sheet, SheetContent, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useTranslation } from '@/hooks/use-translation';
import { Sparkles } from 'lucide-react';
import { AIContent } from './ai-content';

interface AISheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AISheet({ open, onOpenChange }: AISheetProps) {
  const { t } = useTranslation();

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        side="right"
        className="sm:max-w-xl w-full p-0 flex flex-col gap-0 bg-background/95 backdrop-blur-sm h-full max-h-screen overflow-hidden"
      >
        <SheetHeader className="px-6 py-4 border-b border-border/50 shrink-0">
          <SheetTitle className="flex items-center gap-3 text-lg">
            <div className="flex items-center justify-center size-8 rounded-lg bg-primary/10">
              <Sparkles className="size-4 text-primary" />
            </div>
            <span>{t('ai.title')}</span>
          </SheetTitle>
        </SheetHeader>

        <AIContent open={open} className="flex-1 min-h-0 overflow-hidden" />
      </SheetContent>
    </Sheet>
  );
}
