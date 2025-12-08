'use client';

import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { Button } from '@/components/ui/button';
import { ExternalLink } from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import Image from 'next/image';
import { TypographyH1, TypographyMuted } from '@/components/ui/typography';

interface ExtensionsHeroProps {
  isLoading?: boolean;
}

function ExtensionsHero({ isLoading = false }: ExtensionsHeroProps) {
  const { t } = useTranslation();

  if (isLoading) {
    return <ExtensionsHeroSkeleton />;
  }

  return (
    <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary/20 via-primary/10 to-secondary/20 px-4 py-1 md:px-6 md:py-1">
      <div className="relative z-10 flex flex-col items-start justify-between gap-2 md:flex-row md:items-center">
        <div className="flex-1 space-y-4">
          <div className="inline-flex items-center rounded-full bg-primary/10 px-2 py-1 text-xs font-medium text-primary">
            {t('extensions.beta')}
          </div>
          <TypographyH1>{t('extensions.title')}</TypographyH1>
          <TypographyMuted>{t('extensions.subtitle')}</TypographyMuted>
        </div>
        <div className="flex-1">
          <div className="relative mx-auto max-w-xs">
            <div className="aspect-square">
              <div className="flex h-full items-center justify-center">
                <div className="text-center">
                  <Image
                    src="/plugin.png"
                    alt="Extensions Hero"
                    className="w-full h-full text-white object-contain "
                    fill
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div className="absolute inset-0 bg-gradient-to-r from-transparent via-background/5 to-transparent" />
    </div>
  );
}

function ExtensionsHeroSkeleton() {
  return (
    <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary/20 via-primary/10 to-secondary/20 px-4 py-1 md:px-6 md:py-1">
      <div className="relative z-10 flex flex-col items-start justify-between gap-2 md:flex-row md:items-center">
        <div className="flex-1 space-y-1">
          <Skeleton className="h-5 w-12 rounded-full" />
          <Skeleton className="h-6 w-48 md:w-56 lg:w-64" />
          <Skeleton className="h-4 w-72 md:w-80" />
          <Skeleton className="h-8 w-32 mt-2" />
        </div>
        <div className="flex-1">
          <div className="relative mx-auto max-w-xs aspect-square">
            <div className="flex h-full items-center justify-center">
              <div className="text-center">
                <Image
                  src="/plugin.png"
                  alt="Extensions Hero"
                  className="w-full h-full text-white object-contain "
                  fill
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export default ExtensionsHero;
