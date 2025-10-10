'use client';

import React, { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { consumeCode, clearLoginAttemptInfo, getLoginAttemptInfo } from 'supertokens-web-js/recipe/passwordless';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Loader2, CheckCircle, AlertCircle } from 'lucide-react';

export default function OrganizationInvitePage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const orgId = searchParams.get('org_id');
  
  
  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<'loading' | 'success' | 'error' | 'intermediate'>('loading');
  const [message, setMessage] = useState('');
  const [isSameDevice, setIsSameDevice] = useState<boolean | null>(null);

  useEffect(() => {
    checkDeviceAndHandleMagicLink();
  }, []);

  const checkDeviceAndHandleMagicLink = async () => {
    try {
      // Check if this is the same browser/device that started the flow
      const loginAttemptInfo = await getLoginAttemptInfo();
      const sameDevice = loginAttemptInfo !== undefined;
      setIsSameDevice(sameDevice);

      if (sameDevice) {
        // Same device - can consume the magic link directly
        await handleMagicLinkConsumption();
      } else {
        // Different device - show intermediate step
        setStatus('intermediate');
        setMessage('Click the button below to complete your login on this device.');
      }
    } catch (error) {
      console.error('Error checking device:', error);
      setStatus('error');
      setMessage('An error occurred while processing your invitation.');
    }
  };

  const handleMagicLinkConsumption = async () => {
    setIsLoading(true);
    setStatus('loading');
    setMessage('Processing your invitation...');

    try {
      // Create userContext with organization data from URL parameters
      const userContext = orgId ? {
        organization_id: orgId,
        role: searchParams.get('role') || 'viewer', // Default role if not specified
        email: searchParams.get('email') || undefined
      } : undefined;


      // For magic links, consumeCode() automatically reads the link code from the URL
      // Pass the organization context to preserve invitation data and use preAPIHook to inject org data
      const response = await consumeCode({ 
        userContext,
        options: {
          preAPIHook: async (context) => {
            // Add organization data to the request body or headers
            if (orgId && userContext) {
              try {
                const requestBody = context.requestInit.body ? JSON.parse(context.requestInit.body as string) : {};
                requestBody.organization_id = orgId;
                requestBody.role = userContext.role;
                requestBody.email = userContext.email;
                
                context.requestInit.body = JSON.stringify(requestBody);
              } catch (error) {
                console.error('Error modifying request:', error);
              }
            }
            
            return context;
          }
        }
      });

      if (response.status === 'OK') {
        // Clear login attempt info since login was successful
        await clearLoginAttemptInfo();
        
        setStatus('success');
        setMessage('Welcome! You have successfully joined the organization.');
        
        // Redirect to dashboard after a short delay
        setTimeout(() => {
          router.push('/dashboard');
        }, 2000);
      } else {
        // Magic link expired, invalid, or denied
        await clearLoginAttemptInfo();
        setStatus('error');
        setMessage('This invitation link has expired or is invalid. Please request a new invitation.');
      }
    } catch (error: any) {
      console.error('Error consuming magic link:', error);
      await clearLoginAttemptInfo();
      
      if (error.isSuperTokensGeneralError === true) {
        setMessage(error.message);
      } else {
        setMessage('An error occurred while processing your invitation. Please try again.');
      }
      setStatus('error');
    } finally {
      setIsLoading(false);
    }
  };

  const handleIntermediateLogin = () => {
    handleMagicLinkConsumption();
  };

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
            <Button 
              onClick={() => router.push('/auth')} 
              variant="outline" 
              size="sm"
            >
              Back to Login
            </Button>
          </div>
        );

      case 'intermediate':
        return (
          <div className="flex flex-col items-center space-y-4">
            <AlertCircle className="h-8 w-8 text-accent-foreground" />
            <p className="text-sm text-center">{message}</p>
            <Button 
              onClick={handleIntermediateLogin}
              disabled={isLoading}
              className="w-full"
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Processing...
                </>
              ) : (
                'Complete Login on This Device'
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
        <CardContent>
          {renderContent()}
        </CardContent>
      </Card>
    </div>
  );
}
