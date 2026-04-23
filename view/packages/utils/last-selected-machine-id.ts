const key = (orgId: string) => `nixopus.last_machine_id.${orgId}`;

export function getLastSelectedMachineId(orgId: string | null | undefined): string | null {
  if (!orgId || typeof window === 'undefined') return null;
  return localStorage.getItem(key(orgId));
}

export function setLastSelectedMachineId(
  orgId: string | null | undefined,
  machineId: string
): void {
  if (!orgId || typeof window === 'undefined') return;
  try {
    localStorage.setItem(key(orgId), machineId);
  } catch {
    // ignore quota / private mode
  }
}
