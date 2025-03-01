'use client';
import { useAppSelector } from '@/redux/hooks';
import React from 'react';

function page() {
  const user = useAppSelector((state) => state.auth.user);

  return <h1>Dashboard</h1>;
}

export default page;
