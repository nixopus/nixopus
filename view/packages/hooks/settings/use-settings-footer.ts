'use client';

import { useCheckForUpdatesQuery } from '@/redux/services/users/userApi';

const getClientInfo = () => {
  const userAgent = navigator.userAgent;
  const browser = userAgent.includes('Chrome')
    ? 'Chrome'
    : userAgent.includes('Firefox')
      ? 'Firefox'
      : userAgent.includes('Safari')
        ? 'Safari'
        : userAgent.includes('Edge')
          ? 'Edge'
          : 'Unknown';

  const os = userAgent.includes('Windows')
    ? 'Windows'
    : userAgent.includes('Mac')
      ? 'macOS'
      : userAgent.includes('Linux')
        ? 'Linux'
        : userAgent.includes('Android')
          ? 'Android'
          : userAgent.includes('iOS')
            ? 'iOS'
            : 'Unknown';

  return {
    browser,
    os,
    userAgent,
    screenResolution: `${screen.width}x${screen.height}`,
    language: navigator.language,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone
  };
};

export function useSettingsFooter() {
  const { data: updateInfo, isLoading: isCheckingUpdates } = useCheckForUpdatesQuery();

  const handleSponsor = () => {
    window.open('https://github.com/sponsors/raghavyuva', '_blank');
  };

  const handleReportIssue = () => {
    const clientInfo = getClientInfo();

    const issueBody = `**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Additional context**
- Browser: ${clientInfo.browser}
- Operating System: ${clientInfo.os}
- Screen Resolution: ${clientInfo.screenResolution}
- Language: ${clientInfo.language}
- Timezone: ${clientInfo.timezone}
- User Agent: ${clientInfo.userAgent}

Add any other context about the problem here.`;

    const encodedBody = encodeURIComponent(issueBody);
    const url = `https://github.com/raghavyuva/nixopus/issues/new?template=bug_report.md&body=${encodedBody}`;
    window.open(url, '_blank');
  };

  const handleHelp = () => {
    window.open('https://docs.nixopus.com', '_blank');
  };

  return {
    updateInfo,
    isCheckingUpdates,
    handleSponsor,
    handleReportIssue,
    handleHelp
  };
}
