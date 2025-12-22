import type { FeatureOption } from "./types";

export const CATEGORIES = {
  network: { title: "Network", description: "Domain & IP configuration" },
  database: { title: "Database", description: "Database configuration" },
  ports: { title: "Ports", description: "Service ports (all optional)" },
  options: { title: "Options", description: "Installation behavior" },
  developer: { title: "Developer", description: "For contributors & forks" }
} as const;

export const COMMANDS = {
  curl: "sudo curl -sSL https://install.nixopus.com | bash",
  nixopus: "nixopus install"
} as const;

export const FLAG_MAPPINGS: Record<string, (value: string) => string> = {
  apiDomain: (v) => `--api-domain ${v}`,
  viewDomain: (v) => `--view-domain ${v}`,
  hostIp: (v) => `--host-ip ${v}`,
  externalDb: (v) => `--external-db-url "${v}"`,
  apiPort: (v) => `--api-port ${v}`,
  viewPort: (v) => `--view-port ${v}`,
  dbPort: (v) => `--db-port ${v}`,
  redisPort: (v) => `--redis-port ${v}`,
  caddyAdminPort: (v) => `--caddy-admin-port ${v}`,
  caddyHttpPort: (v) => `--caddy-http-port ${v}`,
  caddyHttpsPort: (v) => `--caddy-https-port ${v}`,
  supertokensPort: (v) => `--supertokens-port ${v}`,
  timeout: (v) => `--timeout ${v}`,
  configFile: (v) => `--config-file "${v}"`,
  healthCheckTimeout: (v) => `--health-check-timeout ${v}`,
  adminEmail: (v) => `--admin-email "${v}"`,
  adminPassword: (v) => `--admin-password "${v}"`,
  repo: (v) => `--repo ${v}`,
  branch: (v) => `--branch ${v}`
};

export const BOOLEAN_FLAGS: Record<string, string> = {
  verbose: "--verbose",
  dryRun: "--dry-run",
  force: "--force",
  noRollback: "--no-rollback",
  staging: "--staging",
  skipNixopusInstall: "--skip-nixopus-install"
};

const createFeature = (
  id: string,
  label: string,
  description: string,
  category: FeatureOption["category"],
  options: Partial<FeatureOption> = {}
): FeatureOption => ({
  id,
  label,
  description,
  enabled: false,
  value: "",
  category,
  ...options
});

export const DEFAULT_FEATURES: FeatureOption[] = [
  // Network
  createFeature(
    "apiDomain",
    "API Domain",
    "HTTPS endpoint for API",
    "network",
    {
      inputType: "text",
      placeholder: "api.example.com",
      tooltip:
        "Domain for API server. Requires App Domain for auto HTTPS with Let's Encrypt."
    }
  ),
  createFeature(
    "viewDomain",
    "App Domain",
    "HTTPS endpoint for dashboard",
    "network",
    {
      inputType: "text",
      placeholder: "app.example.com",
      tooltip:
        "Domain for dashboard. Requires API Domain for auto HTTPS with Let's Encrypt."
    }
  ),
  createFeature(
    "hostIp",
    "Host IP",
    "Server IP (auto-detected if empty)",
    "network",
    {
      inputType: "text",
      placeholder: "192.168.1.100",
      tooltip: "Custom IP for HTTP-only mode. Mutually exclusive with domains."
    }
  ),

  // Database
  createFeature(
    "externalDb",
    "External Database",
    "If disabled, use local PostgreSQL",
    "database",
    {
      inputType: "text",
      placeholder: "postgresql://user:pass@host:5432/db",
      tooltip:
        "Use existing PostgreSQL instead of local DB. Ignores DB port settings."
    }
  ),

  // Ports
  createFeature("apiPort", "API Port", "Default: 8443", "ports", {
    inputType: "number",
    placeholder: "8443",
    tooltip: "Custom API port. Ports < 1024 need root/sudo privileges."
  }),
  createFeature("viewPort", "Dashboard Port", "Default: 7443", "ports", {
    inputType: "number",
    placeholder: "7443",
    tooltip: "Custom dashboard port. Must differ from API port."
  }),
  createFeature("dbPort", "Database Port", "Default: 5432", "ports", {
    inputType: "number",
    placeholder: "5432",
    tooltip: "Local PostgreSQL port. Ignored when using external database."
  }),
  createFeature("redisPort", "Redis Port", "Default: 6379", "ports", {
    inputType: "number",
    placeholder: "6379",
    tooltip:
      "Port for local Redis service. Redis runs independently from the database."
  }),
  createFeature(
    "caddyAdminPort",
    "Caddy Admin Port",
    "Default: 2019",
    "ports",
    {
      inputType: "number",
      placeholder: "2019",
      tooltip:
        "Port for Caddy admin API. Used for proxy configuration management."
    }
  ),
  createFeature("caddyHttpPort", "Caddy HTTP Port", "Default: 80", "ports", {
    inputType: "number",
    placeholder: "80",
    tooltip:
      "Port for Caddy HTTP traffic. Ports < 1024 need root/sudo privileges."
  }),
  createFeature("caddyHttpsPort", "Caddy HTTPS Port", "Default: 443", "ports", {
    inputType: "number",
    placeholder: "443",
    tooltip:
      "Port for Caddy HTTPS traffic. Ports < 1024 need root/sudo privileges."
  }),
  createFeature(
    "supertokensPort",
    "SuperTokens Port",
    "Default: 3567",
    "ports",
    {
      inputType: "number",
      placeholder: "3567",
      tooltip: "Port for SuperTokens authentication service."
    }
  ),

  // Options
  createFeature("verbose", "Verbose", "Show detailed logs", "options", {
    inputType: "toggle",
    tooltip: "Show detailed installation logs for debugging."
  }),
  createFeature("dryRun", "Dry Run", "Preview without changes", "options", {
    inputType: "toggle",
    tooltip: "Preview installation without making actual changes."
  }),
  createFeature("force", "Force", "Overwrite existing files", "options", {
    inputType: "toggle",
    tooltip: "Overwrite existing config files. Use with caution."
  }),
  createFeature("timeout", "Timeout", "Default: 300 seconds", "options", {
    inputType: "number",
    placeholder: "300",
    tooltip: "Max wait time in seconds. Default: 300s (5 min)."
  }),
  createFeature(
    "configFile",
    "Config File",
    "Custom config file path",
    "options",
    {
      inputType: "text",
      placeholder: "/path/to/config.yaml",
      tooltip:
        "Path to custom config file. Defaults to built-in config if not provided."
    }
  ),
  createFeature(
    "noRollback",
    "No Rollback",
    "Disable automatic rollback",
    "options",
    {
      inputType: "toggle",
      tooltip:
        "Disable automatic rollback on installation failure. Not recommended for production."
    }
  ),
  createFeature(
    "healthCheckTimeout",
    "Health Check Timeout",
    "Default: 120 seconds",
    "options",
    {
      inputType: "number",
      placeholder: "120",
      tooltip:
        "Maximum time to wait for services to become healthy (in seconds)."
    }
  ),
  createFeature(
    "adminEmail",
    "Admin Email",
    "Email for admin user",
    "options",
    {
      inputType: "text",
      placeholder: "admin@example.com",
      tooltip:
        "Email for admin user registration. Automatically registered after installation."
    }
  ),
  createFeature(
    "adminPassword",
    "Admin Password",
    "Password for admin user",
    "options",
    {
      inputType: "text",
      placeholder: "secure-password",
      tooltip:
        "Password for admin user registration. Automatically registered after installation."
    }
  ),
  createFeature(
    "skipNixopusInstall",
    "Skip Nixopus Install",
    "Install CLI only",
    "options",
    {
      inputType: "toggle",
      tooltip:
        "Install only the Nixopus CLI tool without running the installation process. Useful if you want to run 'nixopus install' manually later."
    }
  ),

  // Developer
  createFeature("repo", "Custom Repo", "Fork repository URL", "developer", {
    inputType: "text",
    placeholder: "https://github.com/user/fork",
    tooltip: "Install from custom fork instead of official repo."
  }),
  createFeature("branch", "Branch", "Git branch name", "developer", {
    inputType: "text",
    placeholder: "main",
    tooltip: 'Install from specific Git branch. Defaults to "main".'
  }),
  createFeature("staging", "Staging", "Use staging compose file", "developer", {
    inputType: "toggle",
    tooltip: "Use staging config for development/testing only."
  })
];
