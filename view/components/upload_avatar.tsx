import React from 'react';

interface UploadAvatarProps {
  initialImage?: string;
  onImageChange?: (imageUrl: string | null) => void;
  username?: string;
}

function UploadAvatar({ initialImage, onImageChange, username = 'User' }: UploadAvatarProps) {
  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((part) => part[0])
      .join('')
      .toUpperCase()
      .substring(0, 2);
  };

  return (
    <div className="flex flex-col items-center space-y-4">
      <div className="relative group">
        <div className="w-24 h-24 border-4 shadow-md border-primary rounded-full text-foreground text-center flex items-center justify-center">
          {getInitials(username)}
        </div>
      </div>
      <p className="text-sm font-medium">{username}</p>
    </div>
  );
}

export default UploadAvatar;
