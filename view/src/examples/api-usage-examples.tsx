/**
 * Usage Examples for Auto-Generated Nixopus API Client
 * 
 * This file demonstrates how to use the auto-generated TypeScript API client
 * in different scenarios within your React/Next.js application.
 */

import React, { useState } from 'react';
import { useLoginMutation } from '../lib/api-bridge';
import nixopusApi from '../lib/nixopus-api';
import type { LoginRequest, RegisterRequest } from '../lib/nixopus-api';

// Example 1: Using RTK Query hooks (recommended for React components)
export const LoginComponent: React.FC = () => {
  const [login, { isLoading, error }] = useLoginMutation();
  const [credentials, setCredentials] = useState<LoginRequest>({
    email: '',
    password: '',
  });

  const handleLogin = async () => {
    try {
      const result = await login(credentials).unwrap();
      console.log('Login successful:', result);
    } catch (err) {
      console.error('Login failed:', err);
    }
  };

  return (
    <div>
      <input
        type="email"
        placeholder="Email"
        value={credentials.email}
        onChange={(e) => setCredentials(prev => ({ ...prev, email: e.target.value }))}
      />
      <input
        type="password"
        placeholder="Password"
        value={credentials.password}
        onChange={(e) => setCredentials(prev => ({ ...prev, password: e.target.value }))}
      />
      <button onClick={handleLogin} disabled={isLoading}>
        {isLoading ? 'Logging in...' : 'Login'}
      </button>
      {error && <div>Error: {JSON.stringify(error)}</div>}
    </div>
  );
};

// Example 2: Using the API client directly (for non-React contexts or manual calls)
export const directApiExample = async () => {
  try {
    // Login example
    const loginData: LoginRequest = {
      email: 'user@example.com',
      password: 'password123',
    };
    
    const loginResponse = await nixopusApi.pOSTApiV1AuthLogin(loginData);
    console.log('Login response:', loginResponse.data);

    // Get applications example
    const appsResponse = await nixopusApi.gETApiV1DeployApplications({
      page: "1",
      page_size: "10",
    });
    console.log('Applications:', appsResponse.data);

    // Health check example
    const healthResponse = await nixopusApi.gETApiV1Health();
    console.log('Health status:', healthResponse.data);

  } catch (error) {
    console.error('API call failed:', error);
  }
};

// Example 3: Custom hook using the API client
export const useApplications = () => {
  const [applications, setApplications] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<any>(null);

  const fetchApplications = async (params: any = {}) => {
    setLoading(true);
    setError(null);
    try {
      const response = await nixopusApi.gETApiV1DeployApplications(params);
      setApplications(response.data);
    } catch (err) {
      setError(err);
    } finally {
      setLoading(false);
    }
  };

  return { applications, loading, error, fetchApplications };
};

// Example 4: TypeScript usage with proper typing
export const typedApiExample = async () => {
  // All request types are automatically generated and imported
  const registerData: RegisterRequest = {
    email: 'newuser@example.com',
    password: 'securepassword',
    username: 'johndoe',
    organization: 'myorg',
  };

  try {
    // TypeScript will provide full IntelliSense and type checking
    const response = await nixopusApi.pOSTApiV1AuthRegister({ 
  email: 'user@example.com', 
  password: 'password123', 
  username: 'user123' 
});
    
    // Response is properly typed based on OpenAPI spec
    console.log('Registration successful:', response.data);
    
    // Error handling is also typed
  } catch (error: any) {
    if (error.response?.status === 400) {
      console.error('Bad request:', error.response.data);
    } else {
      console.error('Unexpected error:', error);
    }
  }
};

// Example 5: Integration with existing Redux store
export const useWithExistingRedux = () => {
  // You can still use the RTK Query API alongside your existing Redux setup
  // The API bridge provides hooks that work with your existing store structure
  
  const [login] = useLoginMutation();
  
  const handleLoginWithReduxIntegration = async (credentials: LoginRequest) => {
    try {
      const result = await login(credentials).unwrap();
      
      // The result is automatically typed and can be used with your existing Redux actions
      // dispatch(setUser(result.user));
      // dispatch(setToken(result.token));
      
      return result;
    } catch (error) {
      // Handle error with existing error handling patterns
      throw error;
    }
  };

  return { handleLoginWithReduxIntegration };
};

// Example 6: Available API Functions
export const apiReference = {
  // All available functions are in nixopusApi object with names like:
  
  // Authentication
  auth: {
    login: 'nixopusApi.pOSTApiV1AuthLogin',
    register: 'nixopusApi.pOSTApiV1AuthRegister',
    logout: 'nixopusApi.pOSTApiV1AuthLogout',
    refreshToken: 'nixopusApi.pOSTApiV1AuthRefreshToken',
    resetPassword: 'nixopusApi.pOSTApiV1AuthResetPassword',
    twoFactorLogin: 'nixopusApi.pOSTApiV1Auth2faLogin',
    verifyEmail: 'nixopusApi.gETApiV1AuthVerifyEmail',
  },
  
  // Applications
  applications: {
    getAll: 'nixopusApi.gETApiV1DeployApplications',
    create: 'nixopusApi.pOSTApiV1DeployApplication',
    update: 'nixopusApi.pUTApiV1DeployApplication',
    delete: 'nixopusApi.dELETEApiV1DeployApplication',
    redeploy: 'nixopusApi.pOSTApiV1DeployApplicationRedeploy',
    restart: 'nixopusApi.pOSTApiV1DeployApplicationRestart',
    rollback: 'nixopusApi.pOSTApiV1DeployApplicationRollback',
  },
  
  // System
  system: {
    health: 'nixopusApi.gETApiV1Health',
    auditLogs: 'nixopusApi.gETApiV1AuditLogs',
  },
  
  // File Management
  files: {
    list: 'nixopusApi.gETApiV1FileManager',
    upload: 'nixopusApi.pOSTApiV1FileManagerUpload',
  },
  
  // And many more... (all endpoints from your OpenAPI spec are available)
};

export default {
  LoginComponent,
  directApiExample,
  useApplications,
  typedApiExample,
  useWithExistingRedux,
  apiReference,
};