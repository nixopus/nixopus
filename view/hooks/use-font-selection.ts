import { useState, useEffect } from 'react';
import { useUpdateFontMutation } from '@/redux/services/users/userApi';
import { useAppSelector } from '@/redux/hooks';
import { RootState } from '@/redux/store';

export const useFontSelection = () => {
  const [updateFont] = useUpdateFontMutation();
  const userSettings = useAppSelector((state: RootState) => state.auth.userSettings);
  const [selectedFont, setSelectedFont] = useState(userSettings?.font_family || 'outfit');

  useEffect(() => {
    if (userSettings?.font_family) {
      setSelectedFont(userSettings.font_family);
      applyFont(userSettings.font_family);
    }
  }, [userSettings?.font_family]);

  const applyFont = (font: string) => {
    document.documentElement.style.setProperty('--font-sans', font);
    document.documentElement.style.setProperty(
      '--font-mono',
      font === 'geist' ? 'var(--font-geist-mono)' : font
    );
  };

  const handleFontChange = async (value: string) => {
    try {
      await updateFont({
        font_family: value,
        font_size: userSettings?.font_size || 16
      }).unwrap();
      setSelectedFont(value);
      applyFont(value);
    } catch (error) {
      console.error('Failed to update font:', error);
    }
  };

  return {
    selectedFont,
    handleFontChange
  };
};
