import { usePathname } from 'next/navigation';

interface BreadcrumbItem {
  href: string;
  label: string;
}

function use_bread_crumbs() {
  const pathname = usePathname();

  const getBreadcrumbs = (): BreadcrumbItem[] => {
    const asPathNestedRoutes = pathname.split('/').filter((v) => v.length > 0);

    const crumblist = asPathNestedRoutes.map((subpath, idx) => {
      const href = '/' + asPathNestedRoutes.slice(0, idx + 1).join('/');
      return { href, label: subpath.charAt(0).toUpperCase() + subpath.slice(1) };
    });

    return [{ href: '/', label: 'Dashboard' }, ...crumblist];
  };

  return { getBreadcrumbs };
}

export default use_bread_crumbs;
