'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { Card, CardDescription, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ExternalLink, Check } from 'lucide-react';

import { Extension } from '@/redux/types/extension';

interface ExtensionCardProps {
  extension: Extension;
  onInstall?: (extension: Extension) => void;
  onViewDetails?: (extension: Extension) => void;
}

export function ExtensionCard({ extension, onInstall, onViewDetails }: ExtensionCardProps) {
  const { t } = useTranslation();

  return (
    <Card className="group h-full transition-all duration-200 hover:shadow-lg bg-card border-border p-6">
      <div className="space-y-4">
        <div className="flex items-start gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-full bg-muted text-primary-foreground flex-shrink-0">
            <div className="text-muted-foreground text-lg font-bold">
              {extension.icon}
            </div>
          </div>
          <div className="flex-1 min-w-0">
            <CardTitle className="text-lg font-bold text-card-foreground mb-1">
              {extension.name}
            </CardTitle>
            <div className="flex items-center gap-2">
              <span className="text-sm text-muted-foreground">
                {t('extensions.madeBy')} {extension.author}
              </span>
              {extension.is_verified && (
                <div className="flex h-4 w-4 items-center justify-center rounded-full bg-primary">
                  <Check className="h-2.5 w-2.5 text-primary-foreground" />
                </div>
              )}
            </div>
          </div>
        </div>

        <CardDescription className="text-sm leading-relaxed text-muted-foreground">
          {extension.description}
        </CardDescription>
        <div className="flex gap-2 pt-6 justify-start">
          <Button
            className="font-medium min-w-[100px]"
            onClick={() => onInstall?.(extension)}
            // disabled={extension.is_installed}
          >
            {/* {extension.is_installed ? t('common.installed') : t('extensions.install')} */}
            {t('extensions.install')}
          </Button>
          <Button
            variant="ghost"
            onClick={() => onViewDetails?.(extension)}
            className="border-border hover:bg-accent text-card-foreground min-w-[100px] whitespace-nowrap"
          >
            {t('extensions.viewDetails')}
            <ExternalLink className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </div>
    </Card>
  );
}
