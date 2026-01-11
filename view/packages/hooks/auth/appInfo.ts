import { getBaseUrl, getWebsiteDomain } from '@/redux/conf';

export const getAppInfo = async () => {
  const baseUrl = await getBaseUrl();
  const apiDomain = baseUrl.replace('://', 'TEMP').replace('/api', '').replace('TEMP', '://');
  const websiteDomain = await getWebsiteDomain();

  return {
    // learn more about this on https://supertokens.com/docs/thirdpartyemailpassword/appinfo
    appName: 'Nixopus',
    apiDomain,
    websiteDomain,
    apiBasePath: '/auth',
    websiteBasePath: '/auth'
  };
};
