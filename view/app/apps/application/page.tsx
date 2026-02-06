'use client';
import { useRouter } from 'next/navigation';
import React, { useEffect } from 'react';

function page() {
  const router = useRouter();

  useEffect(() => {
    router.push('/apps');
    return () => {};
  }, []);

  return;
  <></>;
}

export default page;
