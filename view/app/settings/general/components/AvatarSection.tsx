import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import UploadAvatar from '@/components/ui/upload_avatar';
import { User } from '@/redux/types/user';
import { useTranslation } from '@/hooks/use-translation';

interface AvatarSectionProps {
  onImageChange: (imageUrl: string | null) => void;
  user: User;
}

function AvatarSection({ onImageChange, user }: AvatarSectionProps) {
  const { t } = useTranslation();

  return (
    <div className="col-span-1">
      <Card>
        <CardHeader className="pb-2">
          <CardTitle>{t('settings.account.avatar.title')}</CardTitle>
          <CardDescription>{t('settings.account.avatar.description')}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-center pt-6">
          <UploadAvatar
            onImageChange={onImageChange}
            username={user?.username}
            initialImage={user?.avatar}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export default AvatarSection;
