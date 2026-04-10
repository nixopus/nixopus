import { usePathname } from 'next/navigation';

interface BreadcrumbItem {
  href: string;
  label: string;
  external?: boolean;
  isMachineSwitcher?: boolean;
}

function useBreadCrumbs() {
  const pathname = usePathname();
  const hasBasePath = !!process.env.__NEXT_ROUTER_BASEPATH;

  const getBreadcrumbs = (): BreadcrumbItem[] => {
    const machineMatch = pathname.match(/^\/machines\/([^/]+)(\/.*)?$/);
    const isInMachineContext = !!machineMatch;
    const effectivePath = isInMachineContext ? machineMatch![2] || '/' : pathname;

    const segments = effectivePath.split('/').filter((v) => v.length > 0);

    const machinePrefix = isInMachineContext ? `/machines/${machineMatch![1]}` : '';

    const crumblist = segments.map((subpath, idx) => {
      const href = machinePrefix + '/' + segments.slice(0, idx + 1).join('/');
      return { href, label: subpath.charAt(0).toUpperCase() + subpath.slice(1) };
    });

    const dashboardCrumb: BreadcrumbItem = { href: '/', label: 'Machine', external: true };

    const machineCrumb: BreadcrumbItem = {
      href: '/machines',
      label: 'Machines',
      isMachineSwitcher: true
    };

    let result: BreadcrumbItem[];

    if (hasBasePath) {
      result = pathname.startsWith('/chats')
        ? [dashboardCrumb, ...crumblist]
        : [dashboardCrumb, { href: '/chats', label: 'Chats' }, ...crumblist];
    } else {
      result =
        pathname.startsWith('/chats') || (isInMachineContext && effectivePath.startsWith('/chats'))
          ? [...crumblist]
          : [{ href: '/chats', label: 'Chats' }, ...crumblist];
    }

    const machineScopedPaths = ['/charts', '/apps', '/backups'];
    const showSwitcher =
      isInMachineContext || machineScopedPaths.some((p) => pathname.startsWith(p));

    if (showSwitcher) {
      result = [machineCrumb, ...result];
    }

    return result;
  };

  return { getBreadcrumbs };
}

export default useBreadCrumbs;
