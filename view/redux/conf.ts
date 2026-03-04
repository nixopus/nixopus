let config: any = null;
let configPromise: Promise<any> | null = null;

async function fetchConfig() {
  if (config) return config;
  if (!configPromise) {
    configPromise = fetch('/api/config')
      .then((r) => r.json())
      .then((c) => {
        config = c;
        return c;
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

export async function getWebsiteDomain() {
  const { websiteDomain } = await fetchConfig();
  return websiteDomain;
}

export async function getPasswordLoginEnabled() {
  const { passwordLoginEnabled } = await fetchConfig();
  return passwordLoginEnabled;
}

export async function getAgentConfigured() {
  const { agentConfigured } = await fetchConfig();
  return agentConfigured;
}
