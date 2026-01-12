import EmailPasswordReact from 'supertokens-auth-react/recipe/emailpassword';
import PasswordlessReact from 'supertokens-auth-react/recipe/passwordless';
import SessionReact from 'supertokens-auth-react/recipe/session';
import { getAppInfo } from './appInfo';
import { useRouter } from 'next/navigation';
import { SuperTokensConfig } from 'supertokens-auth-react/lib/build/types';

const routerInfo: { router?: ReturnType<typeof useRouter>; pathName?: string } = {};

export function setRouter(router: ReturnType<typeof useRouter>, pathName: string) {
  routerInfo.router = router;
  routerInfo.pathName = pathName;
}

export const frontendConfig = async (): Promise<SuperTokensConfig> => {
  const appInfo = await getAppInfo();

  return {
    appInfo,
    recipeList: [
      EmailPasswordReact.init({
        resetPasswordUsingTokenFeature: {
          disableDefaultUI: true
        }
      }),
      PasswordlessReact.init({
        contactMethod: 'EMAIL'
      }),
      SessionReact.init()
    ],
    windowHandler: (original: any) => ({
      ...original,
      location: {
        ...original.location,
        getPathName: () => routerInfo.pathName!,
        assign: (url: any) => routerInfo.router!.push(url.toString()),
        setHref: (url: any) => routerInfo.router!.push(url.toString())
      }
    })
  };
};
