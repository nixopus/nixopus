import { useEffect, useState } from 'react';
import { getPasswordLoginEnabled, getAgentConfigured } from '@/redux/conf';

export function usePasswordLoginEnabled() {
  const [passwordLoginEnabled, setPasswordLoginEnabled] = useState<boolean | null>(null);

  useEffect(() => {
    getPasswordLoginEnabled().then(setPasswordLoginEnabled);
  }, []);

  return passwordLoginEnabled;
}

export function useAgentConfigured() {
  const [agentConfigured, setAgentConfigured] = useState<boolean | null>(null);

  useEffect(() => {
    getAgentConfigured().then(setAgentConfigured);
  }, []);

  return agentConfigured;
}
