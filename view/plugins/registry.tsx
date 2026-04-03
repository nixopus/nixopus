import type { ComponentType, ReactElement, ReactNode } from 'react';

export interface PluginNavItem {
  title: string;
  url: string;
  icon: ComponentType<{ className?: string }>;
  resource: string;
}

export interface PluginManifest {
  name: string;
  navigation?: PluginNavItem[];
  middleware?: {
    privatePaths?: string[];
    publicPaths?: string[];
  };
  providers?: ComponentType<{ children: ReactNode }>;
  banners?: ComponentType[];
}

// Plugin manifests are loaded from a generated file that the route
// generation script produces at predev/prebuild time. When no
// plugins are installed the file exports an empty array.
let plugins: PluginManifest[] = [];
try {
  plugins = require('./_manifests').manifests;
} catch {}

export function getPluginNavItems(): PluginNavItem[] {
  return plugins.flatMap((p) => p.navigation ?? []);
}

export function getPluginPrivatePaths(): string[] {
  return plugins.flatMap((p) => p.middleware?.privatePaths ?? []);
}

export function getPluginPublicPaths(): string[] {
  return plugins.flatMap((p) => p.middleware?.publicPaths ?? []);
}

export function getPluginBanners(): ComponentType[] {
  return plugins.flatMap((p) => p.banners ?? []);
}

export function getPluginProviders(): ComponentType<{ children: ReactNode }> | null {
  const providerComponents = plugins
    .map((p) => p.providers)
    .filter((p): p is ComponentType<{ children: ReactNode }> => !!p);

  if (providerComponents.length === 0) return null;

  return function ComposedPluginProviders({ children }: { children: ReactNode }) {
    return providerComponents.reduceRight(
      (acc, Provider) => <Provider>{acc}</Provider>,
      children
    ) as ReactElement;
  };
}
