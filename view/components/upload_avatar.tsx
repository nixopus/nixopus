import React, { useState } from 'react';
import { Camera, Upload, X } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter
} from '@/components/ui/dialog';
import { Avatar, AvatarImage, AvatarFallback } from '@/components/ui/avatar';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';

interface UploadAvatarProps {
  initialImage?: string;
  onImageChange?: (imageUrl: string | null) => void;
  username?: string;
}

function UploadAvatar({ initialImage, onImageChange, username = 'User' }: UploadAvatarProps) {
  const [image, setImage] = useState<string | null>(initialImage || null);
  const [isOpen, setIsOpen] = useState(false);

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const result = e.target?.result as string;
        setImage(result);
        onImageChange?.(result);
        setIsOpen(false);
      };
      reader.readAsDataURL(file);
    }
  };

  const removeImage = () => {
    setImage(null);
    onImageChange?.(null);
  };

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
        <Avatar className="w-24 h-24 border-4  shadow-md">
          {image ? (
            <AvatarImage src={image} alt="Profile avatar" />
          ) : (
            <AvatarFallback className="bg-primary text-primary-foreground text-xl font-medium">
              {getInitials(username)}
            </AvatarFallback>
          )}
        </Avatar>

        <div className="absolute inset-0 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity bg-black/50 rounded-full">
          <div className="flex gap-1">
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Dialog open={isOpen} onOpenChange={setIsOpen}>
                    <DialogTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 rounded-full bg-background/80 text-foreground hover:bg-background"
                      >
                        <Camera size={16} />
                      </Button>
                    </DialogTrigger>
                    <DialogContent className="sm:max-w-md">
                      <DialogHeader>
                        <DialogTitle>Upload avatar</DialogTitle>
                      </DialogHeader>
                      <div className="flex flex-col items-center justify-center space-y-4 py-4">
                        <div className="flex items-center justify-center w-40 h-40 rounded-full bg-muted">
                          {image ? (
                            <img
                              src={image}
                              alt="Preview"
                              className="w-full h-full object-cover rounded-full"
                            />
                          ) : (
                            <Upload className="h-10 w-10 text-muted-foreground" />
                          )}
                        </div>
                        <div className="grid w-full gap-2">
                          <label htmlFor="avatar-upload" className="cursor-pointer">
                            <div className="flex items-center justify-center px-4 py-2 bg-primary text-primary-foreground hover:bg-primary/90 rounded-md text-sm font-medium">
                              Choose file
                            </div>
                            <input
                              id="avatar-upload"
                              type="file"
                              accept="image/*"
                              onChange={handleFileChange}
                              className="hidden"
                            />
                          </label>
                          <p className="text-xs text-muted-foreground text-center">
                            Recommended: Square JPG, PNG, or GIF, at least 300x300 pixels.
                          </p>
                        </div>
                      </div>
                      <DialogFooter className="sm:justify-start">
                        <Button type="button" variant="outline" onClick={() => setIsOpen(false)}>
                          Cancel
                        </Button>
                      </DialogFooter>
                    </DialogContent>
                  </Dialog>
                </TooltipTrigger>
                <TooltipContent>
                  <p>Change avatar</p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>

            {image && (
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8 rounded-full bg-background/80 text-red-500 hover:bg-background hover:text-red-600"
                      onClick={removeImage}
                    >
                      <X size={16} />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Remove avatar</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            )}
          </div>
        </div>
      </div>
      <p className="text-sm font-medium">{username}</p>
    </div>
  );
}

export default UploadAvatar;
