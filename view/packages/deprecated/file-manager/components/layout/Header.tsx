import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import DashboardPageHeader from '@/components/layout/dashboard-page-header';

function FileManagerHeader() {
  const { t } = useTranslation();

  return (
    <DashboardPageHeader
      label={t('fileManager.header.title')}
      description={t('fileManager.header.subtitle')}
    />
  );
}

export default FileManagerHeader;
