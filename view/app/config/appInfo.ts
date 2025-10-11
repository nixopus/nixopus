import { getBaseUrl } from '@/redux/conf';

export const getAppInfo = async () => {
  const baseUrl = await getBaseUrl();
  const apiDomain = baseUrl.replace('://', 'TEMP').replace('/api', '').replace('TEMP', '://');

  return {
    // learn more about this on https://supertokens.com/docs/thirdpartyemailpassword/appinfo
    appName: 'Nixopus',
    apiDomain,
    websiteDomain: process.env.NEXT_PUBLIC_WEBSITE_DOMAIN || 'http://localhost:3000',
    apiBasePath: '/auth',
    websiteBasePath: '/auth'
  };
};
