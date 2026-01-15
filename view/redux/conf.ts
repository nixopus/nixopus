let config: any = null;

async function fetchConfig() {
  if (!config) {
    const response = await fetch('/api/config');
    config = await response.json();
  }
  return config;
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
