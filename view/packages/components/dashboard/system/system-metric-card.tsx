'use client';

import React from 'react';
import { LucideIcon } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { TypographySmall } from '@/components/ui/typography';

interface SystemMetricCardProps {
  title: string;
  icon: LucideIcon;
  isLoading?: boolean;
  children: React.ReactNode;
  skeletonContent?: React.ReactNode;
}

export const SystemMetricCard: React.FC<SystemMetricCardProps> = ({
  title,
  icon: Icon,
  isLoading = false,
  children,
  skeletonContent
}) => {
  const content = isLoading && skeletonContent ? skeletonContent : children;

  return (
    <Card className="overflow-hidden h-full flex flex-col">
      <CardHeader className="pb-2">
        <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
          <Icon className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2 text-muted-foreground" />
          <TypographySmall>{title}</TypographySmall>
        </CardTitle>
      </CardHeader>
      <CardContent className="flex-1">{content}</CardContent>
    </Card>
  );
};
