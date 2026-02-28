'use client';

import { useParams, useRouter } from 'next/navigation';
import { useEffect } from 'react';

export default function WorkflowsIndexPage() {
  const params = useParams();
  const router = useRouter();
  const applicationId = params.id as string;

  useEffect(() => {
    router.replace(`/apps/application/${applicationId}?tab=workflows`);
  }, [applicationId, router]);

  return null;
}
