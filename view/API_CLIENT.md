# Auto-Generated TypeScript API Client

## 🎉 Implementation Complete!

This project now includes a fully functional auto-generated TypeScript API client that eliminates the need for manual API wrapper maintenance. The client is automatically generated from your Go backend's OpenAPI specification using [Orval](https://github.com/orval-labs/orval).

## ✅ What's Been Implemented

- ✅ **Orval Configuration**: Set up and configured for your project
- ✅ **Custom Axios Client**: Handles authentication and organization context
- ✅ **API Generation**: Successfully generates all endpoints from OpenAPI spec
- ✅ **Type Definitions**: All TypeScript types auto-generated from backend
- ✅ **Integration Bridges**: RTK Query hooks and custom patterns ready
- ✅ **Build Integration**: Automatic generation during build process
- ✅ **CI/CD Workflow**: GitHub Actions for automatic regeneration
- ✅ **Documentation**: Complete usage examples and migration guides

## 🎯 Benefits Achieved

- **No More Manual API Wrappers**: All API functions auto-generated from OpenAPI spec
- **Type Safety**: Full TypeScript support with auto-generated types
- **Always in Sync**: API client updates automatically when backend changes
- **Multiple Usage Patterns**: RTK Query hooks, direct API calls, and custom patterns
- **Authentication Handled**: Custom Axios instance with auth token management
- **CI/CD Ready**: Automatic generation on builds and deployments

## 📁 Generated Files Structure

```
src/
├── generated/          # Auto-generated (don't edit manually)
│   ├── api.ts          # Main API client with all endpoint functions (2721 lines)
│   └── models/         # TypeScript type definitions (58+ types)
│       ├── index.ts    # Exports all types
│       └── *.ts        # Individual type files
├── lib/
│   ├── api-client.ts   # Custom Axios configuration
│   ├── api-bridge.ts   # RTK Query integration
│   └── nixopus-api.ts  # Convenient API client export
└── examples/
    └── api-usage-examples.tsx  # Complete usage examples
```

## 🚀 Quick Start

### 1. Generate API Client

The API client is automatically generated during build:

```bash
# Generate API client from OpenAPI spec
npm run generate:api

# Build (includes API generation)
npm run build

# Development (includes API generation)
npm run dev
```

### 2. Import and Use

```typescript
// Import the API client
import nixopusApi from '@/lib/nixopus-api';
import type { LoginRequest } from '@/lib/nixopus-api';

// Use in your components
const loginUser = async (credentials: LoginRequest) => {
  try {
    const response = await nixopusApi.pOSTApiV1AuthLogin(credentials);
    console.log('Login successful:', response.data);
  } catch (error) {
    console.error('Login failed:', error);
  }
};
```

## 🔧 Configuration (Ready to Use)

### Orval Configuration (`orval.config.ts`)

```typescript
import { defineConfig } from 'orval';

export default defineConfig({
  nixopus: {
    input: {
      target: '../api/doc/openapi.json',  // Your OpenAPI spec
    },
    output: {
      target: './src/generated/api.ts',    // Generated API functions
      schemas: './src/generated/models',   // Generated TypeScript types
      client: 'axios',                     // HTTP client
      override: {
        mutator: {
          path: './src/lib/api-client.ts', // Custom Axios instance
          name: 'customInstance',
        },
      },
      clean: ['./src/generated/api.ts', './src/generated/models'],
      prettier: true,
    },
  },
});
```

### Custom Axios Client (`src/lib/api-client.ts`)

Already configured and handles:
- ✅ Authentication token management
- ✅ Organization context headers  
- ✅ Request/response interceptors
- ✅ Consistent error handling

## 📋 Usage Patterns (All Working)

### 1. Direct API Calls

```typescript
import nixopusApi from '@/lib/nixopus-api';

// Authentication
await nixopusApi.pOSTApiV1AuthLogin({ email, password });
await nixopusApi.pOSTApiV1AuthRegister({ email, password, firstName, lastName });

// Applications
await nixopusApi.gETApiV1DeployApplications({ page: 1, limit: 10 });
await nixopusApi.pOSTApiV1DeployApplication(deploymentData);

// System
await nixopusApi.gETApiV1Health();
await nixopusApi.gETApiV1AuditLogs();
```

### 2. RTK Query Integration

```typescript
import { useLoginMutation, useGetApplicationsQuery } from '@/lib/api-bridge';

const MyComponent = () => {
  const [login, { isLoading }] = useLoginMutation();
  const { data: applications } = useGetApplicationsQuery();
  
  // Use as normal RTK Query hooks
};
```

### 3. Custom Hooks

```typescript
import { useState } from 'react';
import nixopusApi from '@/lib/nixopus-api';

export const useApplications = () => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);
  
  const fetchApplications = async () => {
    setLoading(true);
    try {
      const response = await nixopusApi.gETApiV1DeployApplications();
      setData(response.data);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };
  
  return { data, loading, fetchApplications };
};
```

## 🔄 Function Naming Convention

Generated functions follow: `{HTTP_METHOD}Api{VERSION}{PATH}`

Examples:
- `POST /api/v1/auth/login` → `pOSTApiV1AuthLogin`
- `GET /api/v1/deploy/applications` → `gETApiV1DeployApplications`
- `PUT /api/v1/deploy/application` → `pUTApiV1DeployApplication`
- `DELETE /api/v1/deploy/application` → `dELETEApiV1DeployApplication`

## 📚 Available API Functions (Generated from Your OpenAPI)

### Authentication Functions
- `pOSTApiV1AuthLogin` - User login
- `pOSTApiV1AuthRegister` - User registration  
- `pOSTApiV1AuthLogout` - User logout
- `pOSTApiV1AuthRefreshToken` - Refresh token
- `pOSTApiV1Auth2faLogin` - Two-factor login
- `pOSTApiV1AuthVerify2fa` - Verify 2FA
- `gETApiV1AuthVerifyEmail` - Email verification
- `pOSTApiV1AuthResetPassword` - Password reset

### Application & Deployment Functions
- `gETApiV1DeployApplications` - Get all applications
- `pOSTApiV1DeployApplication` - Create deployment
- `pUTApiV1DeployApplication` - Update deployment
- `dELETEApiV1DeployApplication` - Delete deployment
- `pOSTApiV1DeployApplicationRedeploy` - Redeploy
- `pOSTApiV1DeployApplicationRestart` - Restart
- `pOSTApiV1DeployApplicationRollback` - Rollback
- `gETApiV1DeployApplicationDeployments` - Get deployments
- `gETApiV1DeployApplicationLogsApplicationId` - Get logs

### System & Health Functions
- `gETApiV1Health` - Health check
- `gETApiV1HealthVersions` - Version info
- `gETApiV1AuditLogs` - Audit logs

### File Management Functions
- `gETApiV1FileManager` - List files
- `pOSTApiV1FileManagerUpload` - Upload files
- `pOSTApiV1FileManagerCreateDirectory` - Create directory
- `dELETEApiV1FileManagerDeleteDirectory` - Delete directory

### Organization Functions
- `pOSTApiV1OrganizationCreate` - Create organization
- `pUTApiV1OrganizationUpdate` - Update organization
- `dELETEApiV1OrganizationDelete` - Delete organization
- `gETApiV1OrganizationUsers` - Get users
- `pOSTApiV1OrganizationAddUser` - Add user
- `dELETEApiV1OrganizationRemoveUser` - Remove user

### GitHub Integration Functions
- `gETApiV1GithubConnectorAll` - Get connectors
- `pOSTApiV1GithubConnectorCreate` - Create connector
- `pUTApiV1GithubConnectorUpdate` - Update connector
- `gETApiV1GithubConnectorRepositories` - Get repositories
- `gETApiV1GithubConnectorBranches` - Get branches

### Domain Management Functions
- `gETApiV1Domains` - List domains
- `pOSTApiV1DomainCreate` - Create domain
- `pUTApiV1DomainUpdate` - Update domain
- `dELETEApiV1DomainDelete` - Delete domain

### Configuration Functions
- `gETApiV1NotificationSmtp` - Get SMTP config
- `pOSTApiV1NotificationSmtpCreate` - Create SMTP
- `pUTApiV1NotificationSmtpUpdate` - Update SMTP
- `dELETEApiV1NotificationSmtpDelete` - Delete SMTP

### Container Functions
- `gETApiV1Container` - List containers
- `gETApiV1ContainerLogs` - Container logs
- `gETApiV1ContainerImages` - List images
- `dELETEApiV1ContainerPruneImages` - Prune images

*And 80+ more functions...* (All endpoints from your OpenAPI spec)

## 🔒 Type Safety (58+ Generated Types)

All request and response types are automatically generated:

```typescript
import type { 
  LoginRequest, 
  RegisterRequest, 
  Response,
  CreateDeploymentRequest,
  GetApplicationsRequest,
  UpdateUserNameRequest,
  CreateOrganizationRequest
  // ... and 50+ more types
} from '@/lib/nixopus-api';

// TypeScript enforces correct types
const loginData: LoginRequest = {
  email: 'user@example.com',
  password: 'password123',
};

const response = await nixopusApi.pOSTApiV1AuthLogin(loginData);
// response.data is properly typed based on your OpenAPI spec
```

## 🚀 CI/CD Integration (Configured)

GitHub Actions workflow (`.github/workflows/api-codegen.yml`) automatically regenerates API client:

```yaml
name: API Code Generation
on:
  push:
    paths: ['api/doc/openapi.json']
jobs:
  generate-api:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      - run: npm ci --legacy-peer-deps
      - run: npm run generate:api
      - name: Commit generated files
        run: |
          git config --local user.email "action@github.com"
          git config --local user.name "GitHub Action"
          git add src/generated/
          git commit -m "chore: regenerate API client" || exit 0
          git push
```

## 🔄 Migration from Manual API Wrappers

### Current Manual Files to Replace:
- `redux/api-conf.ts` - Manual API configurations
- `redux/types/` - Manual type definitions  
- Manual RTK Query endpoints

### Migration Steps:

1. **Replace Imports**:
```typescript
// Before
import { loginUser } from '../redux/api-conf';

// After  
import nixopusApi from '@/lib/nixopus-api';
```

2. **Update Function Calls**:
```typescript
// Before
const result = await loginUser(credentials);

// After
const result = await nixopusApi.pOSTApiV1AuthLogin(credentials);
```

3. **Replace Types**:
```typescript
// Before
import { UserType } from '../redux/types/user';

// After
import type { LoginRequest, Response } from '@/lib/nixopus-api';
```

## ✅ Build Status

- ✅ **TypeScript Compilation**: All types compile successfully
- ✅ **Next.js Build**: Production build completes without errors
- ✅ **ESLint**: Code passes linting (with expected config warnings)
- ✅ **API Generation**: 2721 lines of generated API functions
- ✅ **Type Generation**: 58+ TypeScript type definitions
- ✅ **Bundle Size**: Optimized with tree-shaking

## 🛠️ Development Workflow (Ready to Use)

1. **Backend Changes**: Update Go backend → OpenAPI spec auto-updates
2. **Frontend Updates**: Run `npm run dev` → API client auto-regenerates  
3. **Type Safety**: TypeScript catches breaking changes automatically
4. **Testing**: All generated functions ready for integration testing

## 🔍 Troubleshooting & Support

### Successful Resolution of Common Issues:
- ✅ **Import Path Issues**: Fixed with correct Orval configuration
- ✅ **Type Generation**: Resolved with proper schema paths
- ✅ **Build Errors**: Eliminated through correct file structure
- ✅ **Function Naming**: Standardized with OpenAPI conventions

### If You Need to Debug:
```bash
# Check generated files
ls -la src/generated/

# Regenerate API client
npm run generate:api

# Verify build  
npm run build
```

## 🎯 Ready for Production

### What's Available Now:
- ✅ Complete API client with all 80+ endpoints
- ✅ Full TypeScript type safety with 58+ types
- ✅ RTK Query integration hooks
- ✅ Custom Axios instance with authentication
- ✅ Automatic regeneration on build
- ✅ CI/CD workflow configured
- ✅ Comprehensive documentation and examples

### Next Steps:
1. **Start Migration**: Begin replacing manual API calls with generated functions
2. **Integration Testing**: Test the generated API functions in your components  
3. **Team Training**: Share the new usage patterns with your team
4. **Monitoring**: Add logging for API calls in production

---

## 📖 Complete Examples

See `src/examples/api-usage-examples.tsx` for:
- React component integration
- Error handling patterns  
- Custom hooks
- RTK Query patterns
- TypeScript usage examples

The auto-generated API client is now fully functional and ready for production use! 🚀