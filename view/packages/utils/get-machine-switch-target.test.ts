import { describe, it, expect } from 'vitest';
import { getMachineSwitchTarget } from './get-machine-switch-target';

describe('getMachineSwitchTarget', () => {
  const m2 = 'machine-2';

  it('sends app / deployment / create routes to the new machine apps list', () => {
    expect(getMachineSwitchTarget(m2, '/apps/application/app-1')).toBe(`/machines/${m2}/apps`);
    expect(getMachineSwitchTarget(m2, '/apps/application/app-1/deployments')).toBe(
      `/machines/${m2}/apps`
    );
    expect(getMachineSwitchTarget(m2, '/apps/application/app-1/deployments/dep-1')).toBe(
      `/machines/${m2}/apps`
    );
    expect(getMachineSwitchTarget(m2, '/apps/create')).toBe(`/machines/${m2}/apps`);
    expect(getMachineSwitchTarget(m2, '/apps/create/repo-1')).toBe(`/machines/${m2}/apps`);
    expect(getMachineSwitchTarget(m2, '/machines/m1/apps/application/x')).toBe(
      `/machines/${m2}/apps`
    );
    expect(getMachineSwitchTarget(m2, '/machines/m1/apps/create/something')).toBe(
      `/machines/${m2}/apps`
    );
  });

  it('keeps machine-scoped index routes and only swaps machine id', () => {
    expect(getMachineSwitchTarget(m2, '/machines/m1/apps')).toBe(`/machines/${m2}/apps`);
    expect(getMachineSwitchTarget(m2, '/machines/m1/charts')).toBe(`/machines/${m2}/charts`);
    expect(getMachineSwitchTarget(m2, '/machines/m1/backups')).toBe(`/machines/${m2}/backups`);
  });

  it('prefixes org-level /apps with the new machine', () => {
    expect(getMachineSwitchTarget(m2, '/apps')).toBe(`/machines/${m2}/apps`);
  });
});