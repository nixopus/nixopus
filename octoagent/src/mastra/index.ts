
import { Mastra } from '@mastra/core/mastra';
import { PinoLogger } from '@mastra/loggers';
import { PostgresStore } from '@mastra/pg';
import { nixopusAgent } from './agents/nixopus-agent';
import { config } from '../config';

export const postgresStore = new PostgresStore({
  connectionString: config.databaseUrl,
});

export const mastra = new Mastra({
  agents: { nixopusAgent },
  storage: postgresStore,
  logger: new PinoLogger({
    name: config.logName,
    level: config.logLevel,
  }),
  server: {
    port: config.port,
    host: config.host,
    middleware: [
      async (c, next) => {
        const origin = c.req.header('Origin');
        const allowedOrigins = config.corsAllowedOrigins;

        let allowedOrigin: string | null = null;
        
        if (origin && allowedOrigins.includes(origin)) {
          allowedOrigin = origin;
        } else if (allowedOrigins.length > 0) {
          allowedOrigin = allowedOrigins[0];
        }

        if (allowedOrigin) {
          c.header('Access-Control-Allow-Origin', allowedOrigin);
          c.header('Access-Control-Allow-Credentials', 'true');
          c.header(
            'Access-Control-Allow-Methods',
            'GET, POST, PUT, DELETE, OPTIONS, PATCH',
          );
          
          const requestedHeaders = c.req.header('Access-Control-Request-Headers');
          
          const standardHeaders = [
            'Content-Type',
            'Authorization',
          ];
          
          const customHeaders = [
            'X-Session-ID',
            'x-session-id',
            'X-Organization-Id',
            'x-organization-id',
          ];
          
          const allowedHeaders = new Set([...standardHeaders, ...customHeaders]);
          
          if (requestedHeaders) {
            requestedHeaders.split(',').forEach((header) => {
              allowedHeaders.add(header.trim());
            });
          }
          
          c.header(
            'Access-Control-Allow-Headers',
            Array.from(allowedHeaders).join(', '),
          );
          c.header('Access-Control-Expose-Headers', 'Authorization, X-Organization-Id, x-organization-id');
          c.header('Access-Control-Max-Age', '300');
        }

        if (c.req.method === 'OPTIONS') {
          return new Response(null, { status: 204 });
        }

        await next();
      },
    ],
  },
  telemetry: {
    enabled: config.telemetryEnabled, 
  },
  observability: {
    default: { enabled: config.observabilityEnabled }, 
  },
});
