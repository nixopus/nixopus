'use client';
import React from 'react';
import ListRepositories from '../components/github-repositories/list-repositories';

function page() {
  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <ListRepositories />
    </div>
  );
}

export default page;
