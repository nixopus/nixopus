import React from 'react';
import { useTranslation } from '@/hooks/use-translation';

function FileManagerHeader() {
  const { t } = useTranslation();

  return (
    <div>
      <div className="">
        <h1 className="text-md font-bold capitalize leading-normal text-primary sm:text-lg md:text-xl lg:text-3xl">
          {t('fileManager.header.title')}
        </h1>
        <h2 className="text-xs leading-normal text-muted-foreground sm:text-sm lg:text-xl">
          {t('fileManager.header.subtitle')}
        </h2>
      </div>
    </div>
  );
}

export default FileManagerHeader;
