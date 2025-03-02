'use client';
import { useAppSelector } from '@/redux/hooks';
import { useRouter } from 'next/navigation';
import React from 'react';

function page() {
  const user = useAppSelector((state) => state.auth.user);
  const authenticated = useAppSelector((state) => state.auth.isAuthenticated);
  const router = useRouter();

  if (!user || !authenticated) {
    router.push('/login');
    return null;
  }

  return <h1>Dashboard</h1>;
}

export default page;
