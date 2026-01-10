export const config = {
  // Environment
  nodeEnv: process.env.NODE_ENV || 'development',
  isProduction: process.env.NODE_ENV === 'production',
  isDevelopment: process.env.NODE_ENV !== 'production',

  // Server configuration
  port: parseInt(process.env.PORT || '9090', 10),
  host: process.env.HOST || '0.0.0.0',

  // Database/Storage configuration
  // PostgreSQL connection string (e.g., postgresql://user:pass@host:5432/dbname)
  // Required for production, should be set via DATABASE_URL environment variable
  databaseUrl: process.env.DATABASE_URL || '',

  // Logging configuration
  logLevel: (process.env.LOG_LEVEL || 'info') as 'debug' | 'info' | 'warn' | 'error',
  logName: process.env.LOG_NAME || 'Mastra',

  // Observability configuration
  observabilityEnabled: process.env.OBSERVABILITY_ENABLED !== 'false',
  telemetryEnabled: process.env.TELEMETRY_ENABLED === 'true',

  // MCP Server configuration
  mcpServerPath: process.env.MCP_SERVER_PATH || undefined,
  
  // Authentication
  authToken: process.env.AUTH_TOKEN || '',

  // API Configuration (if needed for external APIs)
  apiUrl: process.env.API_URL || '',
  apiKey: process.env.API_KEY || '',

  // Agent/Model configuration
  agentModel: process.env.AGENT_MODEL || 'google/gemini-2.5-pro',
  agentName: process.env.AGENT_NAME || 'Nixopus Agent',

  // CORS configuration
  corsAllowedOrigins: process.env.CORS_ALLOWED_ORIGINS
    ? process.env.CORS_ALLOWED_ORIGINS.split(',').map((origin) => origin.trim())
    : ['http://localhost:3000', 'http://localhost:7443'],
} as const;

