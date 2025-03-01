'use client';
import { SerializedError } from '@reduxjs/toolkit';
import { FetchBaseQueryError } from '@reduxjs/toolkit/query';
import React, { useEffect } from 'react';
import { toast } from 'sonner';

interface ErrorBoundaryProps {
  errors: {
    error: SerializedError | FetchBaseQueryError | undefined;
    title?: string;
  }[];
}

export const ErrorBoundary: React.FC<ErrorBoundaryProps> = ({ errors }) => {
  useEffect(() => {
    const errorEntry = errors.find(({ error }) => error);
    if (errorEntry) {
      const { error, title } = errorEntry;

      toast(title || 'Error', {
        duration: 5000,
        description: (error as any)?.data?.error
      });
    }
  }, [errors, toast]);

  return null;
};
