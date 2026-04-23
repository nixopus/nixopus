/**
 * Where to go when changing the active machine from the switcher.
 * App / deployment / create paths embed IDs that are scoped to a machine, so
 * we must not keep those segments when switching — use the new machine’s apps
 * list instead. Other machine-prefixed paths (charts, backups, /apps list) keep
 * the same suffix with the new machine id.
 */
export function getMachineSwitchTarget(newMachineId: string, pathname: string): string {
  if (isNonPortableAppPath(pathname)) {
    return `/machines/${newMachineId}/apps`;
  }

  const urlMatch = pathname.match(/^\/machines\/[^/]+/);
  if (urlMatch) {
    return `/machines/${newMachineId}${pathname.slice(urlMatch[0].length)}`;
  }

  return `/machines/${newMachineId}${pathname}`;
}

/** Routes under /apps (or .../apps/) whose segments after /apps/ are not portable across machines. */
function isNonPortableAppPath(pathname: string): boolean {
  if (/^\/apps\/(application|create)(\/|$)/.test(pathname)) {
    return true;
  }
  if (/^\/machines\/[^/]+\/apps\/(application|create)(\/|$)/.test(pathname)) {
    return true;
  }
  return false;
}
