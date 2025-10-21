import React from 'react';
import { useTranslation } from '@/hooks/use-translation';
import { TypographyMuted } from '@/components/ui/typography';

function FileManagerHeader() {
  const { t } = useTranslation();

  return (
    <div>
      <div className="">
        <h1 className="text-md font-bold capitalize leading-normal text-primary sm:text-lg md:text-xl lg:text-3xl">
          {t('fileManager.header.title')}
        </h1>
        <TypographyMuted>{t('fileManager.header.subtitle')}</TypographyMuted>
      </div>
    </div>
  );
}

export default FileManagerHeader;
