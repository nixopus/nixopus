import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import DashboardPageHeader from '@/packages/components/utility-header';

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
