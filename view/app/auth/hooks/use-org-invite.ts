import { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import {
  consumeCode,
  clearLoginAttemptInfo,
  getLoginAttemptInfo
} from 'supertokens-web-js/recipe/passwordless';

function useOrganizationInvite() {
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
      const userContext = orgId
        ? {
            organization_id: orgId,
            role: searchParams.get('role') || 'viewer', // Default role if not specified
            email: searchParams.get('email') || undefined
          }
        : undefined;

      // For magic links, consumeCode() automatically reads the link code from the URL
      // Pass the organization context to preserve invitation data and use preAPIHook to inject org data
      const response = await consumeCode({
        userContext,
        options: {
          preAPIHook: async (context) => {
            // Add organization data to the request body or headers
            if (orgId && userContext) {
              try {
                const requestBody = context.requestInit.body
                  ? JSON.parse(context.requestInit.body as string)
                  : {};
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
        setMessage(
          'This invitation link has expired or is invalid. Please request a new invitation.'
        );
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

  return {
    handleIntermediateLogin,
    checkDeviceAndHandleMagicLink,
    isSameDevice,
    message,
    isLoading,
    status,
    handleMagicLinkConsumption,
    router,
    orgId
  };
}

export default useOrganizationInvite;
