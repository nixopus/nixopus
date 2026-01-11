import React from 'react';
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
    <div className="col-span-1 space-y-4 border border-border/50 rounded-lg p-6 h-fit">
      <div>
        <TypographySmall className="text-sm font-medium">
          {t('settings.account.avatar.title')}
        </TypographySmall>
        <TypographyMuted className="text-xs mt-1">
          {t('settings.account.avatar.description')}
        </TypographyMuted>
      </div>
      <RBACGuard resource="user" action="update">
        <UploadAvatar
          onImageChange={onImageChange}
          username={user?.username}
          initialImage={user?.avatar}
        />
      </RBACGuard>
    </div>
  );
}

export default AvatarSection;
