'use client';
import React from 'react';
import { SuperTokensWrapper } from 'supertokens-auth-react';
import SuperTokensReact from 'supertokens-auth-react';
import { AuthRecipeComponentsOverrideContextProvider } from 'supertokens-auth-react/ui';
import { frontendConfig, setRouter } from '@/app/config/frontend';
import { usePathname, useRouter } from 'next/navigation';
import { useIsAdminRegisteredQuery } from '@/redux/services/users/authApi';

if (typeof window !== 'undefined') {
  // we only want to call this init function on the frontend, so we check typeof window !== 'undefined'
  SuperTokensReact.init(frontendConfig());
}

export const SuperTokensProvider: React.FC<React.PropsWithChildren<{}>> = ({
  children,
}) => {
  setRouter(useRouter(), usePathname() || window.location.pathname);
  const { data: isAdminRegistered } = useIsAdminRegisteredQuery();

  return (
    <SuperTokensWrapper>
      <AuthRecipeComponentsOverrideContextProvider
        components={{
          AuthPageComponentList_Override: ({ DefaultComponent, ...props }) => (
            <DefaultComponent
              {...props}
              hasSeparateSignUpView={isAdminRegistered ? false : true}
            />
          ),
        }}
      >
        {children}
      </AuthRecipeComponentsOverrideContextProvider>
    </SuperTokensWrapper>
  );
};
