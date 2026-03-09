const CONFIG_CACHE_KEY = 'nixopus_config';
const CONFIG_CACHE_TTL_MS = 5 * 60 * 1000;

let config: any = null;
let configPromise: Promise<any> | null = null;

function isValidConfig(c: unknown): c is { baseUrl: string } {
  return !!c && typeof c === 'object' && typeof (c as any).baseUrl === 'string';
}

function getCachedConfig(): any {
  if (typeof window === 'undefined') return null;
  try {
    const raw = sessionStorage.getItem(CONFIG_CACHE_KEY);
    if (!raw) return null;
    const { data, ts } = JSON.parse(raw);
    if (Date.now() - ts > CONFIG_CACHE_TTL_MS) return null;
    if (!isValidConfig(data)) return null;
    return data;
  } catch {
    return null;
  }
}

function setCachedConfig(data: any) {
  if (typeof window === 'undefined') return;
  try {
    sessionStorage.setItem(CONFIG_CACHE_KEY, JSON.stringify({ data, ts: Date.now() }));
  } catch {}
}

async function fetchConfig() {
  if (config) return config;
  const cached = getCachedConfig();
  if (cached && isValidConfig(cached)) {
    config = cached;
    fetch('/api/config')
      .then((r) => r.json())
      .then((c) => {
        if (isValidConfig(c)) {
          config = c;
          setCachedConfig(c);
        }
      })
      .catch(() => {});
    return config;
  }
  if (!configPromise) {
    configPromise = fetch('/api/config')
      .then((r) => r.json())
      .then((c) => {
        if (isValidConfig(c)) {
          config = c;
          setCachedConfig(c);
          return c;
        }
        configPromise = null;
        throw new Error('Invalid config: missing baseUrl');
      });
  }
  return configPromise;
}

export async function getBaseUrl() {
  const { baseUrl } = await fetchConfig();
  return baseUrl;
}

export async function getWebsocketUrl() {
  const { websocketUrl } = await fetchConfig();
  return websocketUrl;
}

export async function getWebhookUrl() {
  const { webhookUrl } = await fetchConfig();
  return webhookUrl;
}

export async function getPasswordLoginEnabled() {
  const { passwordLoginEnabled } = await fetchConfig();
  return passwordLoginEnabled;
}

export async function getAgentConfigured() {
  const { agentConfigured } = await fetchConfig();
  return agentConfigured;
}

export async function getGithubAppSlug() {
  const { githubAppSlug } = await fetchConfig();
  return githubAppSlug as string;
}

export async function getSelfHosted() {
  const { selfHosted } = await fetchConfig();
  return selfHosted as boolean;
}

export async function getPostHogKey() {
  const c = await fetchConfig();
  return (c?.posthogKey || '') as string;
}

export async function getPostHogHost() {
  const c = await fetchConfig();
  return (c?.posthogHost || '') as string;
}
