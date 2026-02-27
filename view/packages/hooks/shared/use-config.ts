import { useEffect, useState } from 'react';
import { getPasswordLoginEnabled } from '@/redux/conf';

export function usePasswordLoginEnabled() {
  const [passwordLoginEnabled, setPasswordLoginEnabled] = useState<boolean | null>(null);

  useEffect(() => {
    getPasswordLoginEnabled().then(setPasswordLoginEnabled);
  }, []);

  return passwordLoginEnabled;
}
