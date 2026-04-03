'use client';

interface PluginReduxConfig {
  reducers?: Record<string, any>;
  middleware?: any[];
}

let reduxConfigs: PluginReduxConfig[] = [];
try {
  reduxConfigs = require('./_redux').reduxConfigs;
} catch {}

export function getPluginReducers(): Record<string, any> {
  return reduxConfigs.reduce(
    (acc, cfg) => ({ ...acc, ...(cfg.reducers ?? {}) }),
    {} as Record<string, any>
  );
}

export function getPluginMiddleware(): any[] {
  return reduxConfigs.flatMap((cfg) => cfg.middleware ?? []);
}
