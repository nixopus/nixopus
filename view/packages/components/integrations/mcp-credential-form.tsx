'use client';

import React, { useState } from 'react';
import { Button, Input } from '@nixopus/ui';
import type { MCPProvider, TestMCPServerResult } from '@/redux/types/mcp';
import { useTestMCPServerMutation } from '@/redux/services/settings/mcpApi';

interface Props {
  provider: MCPProvider;
  initialCredentials?: Record<string, string>;
  initialCustomURL?: string;
  initialName?: string;
  onSave: (data: {
    name: string;
    credentials: Record<string, string>;
    custom_url?: string;
    enabled: boolean;
  }) => Promise<void>;
  onBack: () => void;
  isSaving?: boolean;
}

export function MCPCredentialForm({
  provider,
  initialCredentials = {},
  initialCustomURL = '',
  initialName = '',
  onSave,
  onBack,
  isSaving
}: Props) {
  const [name, setName] = useState(initialName || provider.name);
  const [credentials, setCredentials] = useState<Record<string, string>>(initialCredentials);
  const [customURL, setCustomURL] = useState(initialCustomURL);
  const [testResult, setTestResult] = useState<TestMCPServerResult | null>(null);

  const [testServer, { isLoading: isTesting }] = useTestMCPServerMutation();

  const handleTest = async () => {
    setTestResult(null);
    try {
      const result = await testServer({
        provider_id: provider.id,
        credentials,
        custom_url: customURL || undefined
      }).unwrap();
      setTestResult(result);
    } catch {
      setTestResult({ ok: false, error: 'Request failed' });
    }
  };

  const handleSave = async () => {
    await onSave({
      name,
      credentials,
      custom_url: provider.id === 'custom' ? customURL : undefined,
      enabled: true
    });
  };

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-1">
        <label htmlFor="server-name" className="text-sm font-medium">
          Display Name
        </label>
        <Input
          id="server-name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={`e.g. ${provider.name} Prod`}
        />
      </div>

      {provider.id === 'custom' && (
        <div className="flex flex-col gap-1">
          <label htmlFor="custom-url" className="text-sm font-medium">
            Server URL
          </label>
          <Input
            id="custom-url"
            value={customURL}
            onChange={(e) => setCustomURL(e.target.value)}
            placeholder="https://your-mcp-server.com/sse"
          />
        </div>
      )}

      {provider.fields.map((field) => (
        <div key={field.key} className="flex flex-col gap-1">
          <label htmlFor={field.key} className="text-sm font-medium">
            {field.label}
          </label>
          <Input
            id={field.key}
            type={field.sensitive ? 'password' : 'text'}
            value={credentials[field.key] ?? ''}
            onChange={(e) => setCredentials((prev) => ({ ...prev, [field.key]: e.target.value }))}
            placeholder={field.required ? 'Required' : 'Optional'}
          />
        </div>
      ))}

      {testResult && (
        <div
          className={`text-sm p-2 rounded ${
            testResult.ok ? 'text-green-600 bg-green-50' : 'text-destructive bg-destructive/10'
          }`}
        >
          {testResult.ok ? 'Connection successful' : testResult.error}
        </div>
      )}

      <div className="flex gap-2 pt-2">
        <Button variant="outline" onClick={onBack} disabled={isSaving}>
          Cancel
        </Button>
        <Button variant="outline" onClick={handleTest} disabled={isTesting || isSaving}>
          {isTesting ? 'Testing…' : 'Test Connection'}
        </Button>
        <Button onClick={handleSave} disabled={isSaving || !name.trim()}>
          {isSaving ? 'Saving…' : 'Save'}
        </Button>
      </div>
    </div>
  );
}
