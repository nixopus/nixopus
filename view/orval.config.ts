import { defineConfig } from 'orval';

export default defineConfig({
  nixopus: {
    input: {
      target: '../api/doc/openapi.json',
    },
    output: {
      target: './src/generated/api.ts',
      schemas: './src/generated/models',
      client: 'axios',
      override: {
        mutator: {
          path: './src/lib/api-client.ts',
          name: 'customInstance',
        },
      },
      clean: ['./src/generated/api.ts', './src/generated/models'],
      prettier: true,
    },
  },
});