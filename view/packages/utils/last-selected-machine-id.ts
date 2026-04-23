const key = (orgId: string) => `nixopus.last_machine_id.${orgId}`;

export function getLastSelectedMachineId(orgId: string | null | undefined): string | null {
  if (!orgId || typeof window === 'undefined') return null;
  try {
    return localStorage.getItem(key(orgId));
  } catch (error) {
    // ignore SecurityError, quota errors, etc.
    console.error('Failed to read last selected machine ID:', error);
    return null;
  }
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