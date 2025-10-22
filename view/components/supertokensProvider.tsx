'use client';
import React, { useEffect, useState } from 'react';
import { SuperTokensWrapper } from 'supertokens-auth-react';
import SuperTokensReact from 'supertokens-auth-react';
import { AuthRecipeComponentsOverrideContextProvider } from 'supertokens-auth-react/ui';
import { frontendConfig, setRouter } from '@/app/config/frontend';
import { usePathname, useRouter } from 'next/navigation';
import { useIsAdminRegisteredQuery } from '@/redux/services/users/authApi';

let isInitialized = false;

export const SuperTokensProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
  const [isReady, setIsReady] = useState(isInitialized);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    const initializeSuperTokens = async () => {
      if (!isInitialized) {
        try {
          const config = await frontendConfig();
          SuperTokensReact.init(config);
          isInitialized = true;
          setIsReady(true);
        } catch (error) {
          console.error('Failed to initialize SuperTokens:', error);
          setIsReady(true);
        }
      } else {
        setIsReady(true);
      }
    };

    initializeSuperTokens();
  }, []);

  useEffect(() => {
    setRouter(router, pathname || window.location.pathname);
  }, [router, pathname]);

  const { data: isAdminRegistered } = useIsAdminRegisteredQuery();

  if (!isReady) {
    return <div>Loading...</div>;
  }

  return (
    <SuperTokensWrapper>
      <AuthRecipeComponentsOverrideContextProvider
        components={{
          AuthPageComponentList_Override: ({ DefaultComponent, ...props }) => (
            <DefaultComponent {...props} hasSeparateSignUpView={isAdminRegistered ? false : true} />
          )
        }}
      >
        {children}
      </AuthRecipeComponentsOverrideContextProvider>
    </SuperTokensWrapper>
  );
};
