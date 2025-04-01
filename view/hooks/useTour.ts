import { useState, useEffect } from 'react';

const TOUR_SEEN_KEY = 'nixopus_tour_seen';

export const useTour = () => {
  const [hasSeenTour, setHasSeenTour] = useState(() => {
    if (typeof window !== 'undefined') {
      return localStorage.getItem(TOUR_SEEN_KEY) === 'true';
    }
    return false;
  });

  const startTour = () => {
    setHasSeenTour(false);
    localStorage.removeItem(TOUR_SEEN_KEY);
  };

  const stopTour = () => {
    setHasSeenTour(true);
    localStorage.setItem(TOUR_SEEN_KEY, 'true');
  };

  return {
    hasSeenTour,
    startTour,
    stopTour
  };
};
