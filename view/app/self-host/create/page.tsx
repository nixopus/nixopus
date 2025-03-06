'use client';
import { useRouter } from 'next/navigation';
import React from 'react';

function page() {
  const router = useRouter();

  React.useEffect(() => {
    router.push('/dashboard');
  }, []);

  return <></>;
}

export default page;
