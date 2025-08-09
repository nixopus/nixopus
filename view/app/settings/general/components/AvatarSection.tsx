import React from 'react';
import { Card, CardContent, CardHeader } from '@/components/ui/card';
import UploadAvatar from '@/components/ui/upload_avatar';
import { User } from '@/redux/types/user';
import { useTranslation } from '@/hooks/use-translation';
import { RBACGuard } from '@/components/rbac/RBACGuard';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';

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
          <TypographySmall>{t('settings.account.avatar.title')}</TypographySmall>
          <TypographyMuted>{t('settings.account.avatar.description')}</TypographyMuted>
        </CardHeader>
        <CardContent className="flex flex-col items-center pt-6">
          <RBACGuard resource="user" action="update">
            <UploadAvatar
              onImageChange={onImageChange}
              username={user?.username}
              initialImage={user?.avatar}
            />
          </RBACGuard>
        </CardContent>
      </Card>
    </div>
  );
}

export default AvatarSection;
