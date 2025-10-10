'use client';

import React from 'react';
import { TabsContent } from '@/components/ui/tabs';
import { useTranslation } from '@/hooks/use-translation';
import { TwoFactorSetup } from '@/app/settings/general/components/TwoFactorSetup';

interface SecuritySectionProps {}

function SecuritySection({}: SecuritySectionProps) {
  const { t } = useTranslation();

  return (
    <TabsContent value="security" className="space-y-4 mt-4">
      <TwoFactorSetup />
    </TabsContent>
  );
}

export default SecuritySection;
