import React, { useEffect, useRef, useState } from 'react';
import { Avatar, AvatarFallback, AvatarImage } from '@nixopus/ui';
import { Button } from '@nixopus/ui';
import { Camera } from 'lucide-react';
import { toast } from 'sonner';

interface UploadAvatarProps {
  initialImage?: string;
  onImageChange?: (imageUrl: string | null) => void;
  username?: string;
}

function UploadAvatar({ initialImage, onImageChange, username = 'User' }: UploadAvatarProps) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [previewUrl, setPreviewUrl] = useState<string | undefined>(initialImage);
  const [isUploading, setIsUploading] = useState(false);

  useEffect(() => {
    if (initialImage) {
      setPreviewUrl(initialImage);
    }
  }, [initialImage]);

  const getInitials = (name: string) => {
    return name
      .split(' ')
      .map((part) => part[0])
      .join('')
      .toUpperCase()
      .substring(0, 2);
  };

  const handleFileChange = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    if (!file.type.startsWith('image/')) {
      toast.error('Invalid file type', {
        description: 'Please upload an image file (JPEG, PNG, or GIF)'
      });
      return;
    }

    if (file.size > 5 * 1024 * 1024) {
      toast.error('File too large', {
        description: 'Please upload an image smaller than 5MB'
      });
      return;
    }

    try {
      setIsUploading(true);
      const reader = new FileReader();
      reader.onloadend = () => {
        const base64String = reader.result as string;
        setPreviewUrl(base64String);
        onImageChange?.(base64String);
      };
      reader.readAsDataURL(file);
    } catch (error) {
      toast.error('Upload failed', {
        description: 'Failed to upload image. Please try again.'
      });
    } finally {
      setIsUploading(false);
    }
  };

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="flex flex-col items-center space-y-4">
      <div className="relative group">
        <Avatar className="w-24 h-24 border-1 shadow-md border-primary">
          <AvatarImage src={previewUrl} alt={username} />
          <AvatarFallback className="text-lg">{getInitials(username)}</AvatarFallback>
        </Avatar>
        <Button
          size="icon"
          variant="secondary"
          className="absolute bottom-0 right-0 rounded-full"
          onClick={handleClick}
          disabled={isUploading}
        >
          <Camera className="h-4 w-4" />
        </Button>
        <input
          ref={fileInputRef}
          type="file"
          accept="image/*"
          className="hidden"
          onChange={handleFileChange}
          disabled={isUploading}
        />
      </div>
      <p className="text-sm font-medium">{username}</p>
    </div>
  );
}

export default UploadAvatar;
