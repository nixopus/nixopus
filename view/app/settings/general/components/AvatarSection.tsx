import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import UploadAvatar from '@/components/upload_avatar';
import { User } from '@/redux/types/user';

interface AvatarSectionProps {
  onImageChange: (imageUrl: string | null) => void;
  user: User;
}

function AvatarSection({ onImageChange, user }: AvatarSectionProps) {
  return (
    <div className="col-span-1">
      <Card>
        <CardHeader className="pb-2">
          <CardTitle>Profile</CardTitle>
          <CardDescription>Manage your public profile</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col items-center pt-6">
          <UploadAvatar
            onImageChange={onImageChange}
            username={user.username}
            initialImage={user.avatar}
          />
        </CardContent>
      </Card>
    </div>
  );
}

export default AvatarSection;
