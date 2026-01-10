import { config as loadEnv } from 'dotenv';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';
import { spawn } from 'child_process';
import { config } from '../src/config.ts';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const apiEnvPath = resolve(__dirname, '../../api/.env');
loadEnv({ path: apiEnvPath });

console.log(`Loaded environment variables from: ${apiEnvPath}`);

const mastraProcess = spawn('mastra', ['dev'], {
  stdio: 'inherit',
  shell: true,
  env: {
    ...process.env,
    PORT: config.port.toString(),
  },
});

mastraProcess.on('error', (error) => {
  console.error('Failed to start mastra dev:', error);
  process.exit(1);
});

mastraProcess.on('exit', (code) => {
  process.exit(code || 0);
});
