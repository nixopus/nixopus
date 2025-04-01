import { useState, useEffect } from 'react';

export const useFontSelection = () => {
  const [selectedFont, setSelectedFont] = useState('outfit');

  useEffect(() => {
    const savedFont = localStorage.getItem('selectedFont');
    if (savedFont) {
      setSelectedFont(savedFont);
      applyFont(savedFont);
    } else {
      applyFont('outfit');
    }
  }, []);

  const applyFont = (font: string) => {
    document.documentElement.style.setProperty('--font-sans', font);
    document.documentElement.style.setProperty(
      '--font-mono',
      font === 'geist' ? 'var(--font-geist-mono)' : font
    );
  };

  const handleFontChange = (value: string) => {
    setSelectedFont(value);
    localStorage.setItem('selectedFont', value);
    applyFont(value);
  };

  return {
    selectedFont,
    handleFontChange
  };
};
