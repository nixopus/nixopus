'use client';

import React from 'react';
import { Button } from '@nixopus/ui';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@nixopus/ui';
import { Loader2, CheckCircle, AlertCircle } from 'lucide-react';
import useOrganizationInvite from '../../../packages/hooks/auth/use-org-invite';

export default function OrganizationInvitePage() {
  const { message, router, orgId, handleLoginAndAccept, isLoading, status } =
    useOrganizationInvite();

  const renderContent = () => {
    switch (status) {
      case 'loading':
        return (
          <div className="flex flex-col items-center space-y-4">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
            <p className="text-sm text-muted-foreground">{message}</p>
          </div>
        );

      case 'success':
        return (
          <div className="flex flex-col items-center space-y-4">
            <CheckCircle className="h-8 w-8 text-secondary" />
            <p className="text-sm text-secondary font-medium">{message}</p>
            <p className="text-xs text-muted-foreground">Redirecting to dashboard...</p>
          </div>
        );

      case 'error':
        return (
          <div className="flex flex-col items-center space-y-4">
            <AlertCircle className="h-8 w-8 text-destructive" />
            <p className="text-sm text-destructive font-medium">{message}</p>
            <Button onClick={() => router.push('/auth')} variant="outline" size="sm">
              Back to Login
            </Button>
          </div>
        );

      case 'needs-auth':
        return (
          <div className="flex flex-col items-center space-y-4">
            <AlertCircle className="h-8 w-8 text-accent-foreground" />
            <p className="text-sm text-center">{message}</p>
            <Button onClick={handleLoginAndAccept} disabled={isLoading} className="w-full">
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                'Sign In to Accept Invitation'
              )}
            </Button>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background py-12 px-4 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl font-bold">Organization Invitation</CardTitle>
          <CardDescription>
            {orgId ? `Joining organization: ${orgId}` : 'Processing your invitation...'}
          </CardDescription>
        </CardHeader>
        <CardContent>{renderContent()}</CardContent>
      </Card>
    </div>
  );
}
