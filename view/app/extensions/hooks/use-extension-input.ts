import { useEffect, useMemo, useState } from 'react';
import { Extension, ExtensionVariable } from '@/redux/types/extension';

type SubmitFn = (values: Record<string, unknown>) => void;

const normalizeDefault = (value: unknown, type: string): unknown => {
  if (value === undefined || value === null) {
    if (type === 'boolean') return false;
    if (type === 'array') return [];
    return '';
  }
  if (type === 'boolean') {
    return typeof value === 'boolean' ? value : String(value).toLowerCase() === 'true';
  }
  if (type === 'integer') {
    if (typeof value === 'number') return value;
    const num = Number(value);
    return Number.isNaN(num) ? '' : num;
  }
  if (type === 'array') {
    if (Array.isArray(value)) return value;
    const s = typeof value === 'string' ? value : String(value);
    return s
      .split(/[\,\n]/)
      .map((x) => x.trim())
      .filter((x) => x.length > 0);
  }
  if (typeof value === 'string') {
    let s = value.trim();
    for (let i = 0; i < 3; i++) {
      try {
        const parsed = JSON.parse(s);
        if (typeof parsed === 'string') {
          s = String(parsed).trim();
          continue;
        }
      } catch {}
      break;
    }
    if (s === '""' || s === "''") return '';
    const isQuote = (c: string) => c === '"' || c === "'";
    while (s.length >= 2 && isQuote(s[0]) && isQuote(s[s.length - 1])) {
      s = s.slice(1, -1).trim();
    }
    return s;
  }
  return String(value);
};

const buildInitialValues = (variables: ExtensionVariable[]) => {
  const acc: Record<string, unknown> = {};
  variables.forEach((v) => {
    acc[v.variable_name] = normalizeDefault(v.default_value, v.variable_type);
  });
  return acc;
};

// Check if a variable has an empty/missing default value
const hasEmptyDefault = (v: ExtensionVariable): boolean => {
  const defaultVal = v.default_value;
  if (defaultVal === undefined || defaultVal === null) return true;
  if (v.variable_type === 'string') {
    return typeof defaultVal === 'string' && defaultVal.trim() === '';
  }
  if (v.variable_type === 'array') {
    if (Array.isArray(defaultVal)) return defaultVal.length === 0;
    if (typeof defaultVal === 'string') return defaultVal.trim() === '';
    return true;
  }
  if (v.variable_type === 'integer') {
    return defaultVal === '' || defaultVal === undefined;
  }
  return false;
};

export function useExtensionInput(args: {
  extension?: Extension | null;
  open: boolean;
  onSubmit?: SubmitFn;
  onClose?: () => void;
}) {
  const { extension, open, onSubmit, onClose } = args;

  const variables = useMemo(() => extension?.variables || [], [extension]);
  const initialValues = useMemo(() => buildInitialValues(variables), [variables]);

  const [values, setValues] = useState<Record<string, unknown>>(initialValues);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    setValues(initialValues);
    setErrors({});
  }, [initialValues, open]);

  // Only show fields that actually need user input:
  // - Required fields without defaults
  // - proxy_domain/domain (commonly customized)
  const requiredFields = useMemo(
    () =>
      variables.filter((v) => {
        const isProxyDomain =
          v.variable_name.toLowerCase() === 'proxy_domain' ||
          v.variable_name.toLowerCase() === 'domain';
        // Always show proxy_domain/domain
        if (isProxyDomain) return true;
        // Show required fields only if they don't have a default
        return v.is_required && hasEmptyDefault(v);
      }),
    [variables]
  );

  const handleChange = (name: string, value: unknown) => {
    setValues((prev) => ({ ...prev, [name]: value }));
  };

  const handleSubmit = () => {
    const nextErrors: Record<string, string> = {};
    variables.forEach((v) => {
      if (v.is_required) {
        const val = values[v.variable_name];
        if (v.variable_type === 'string' || v.variable_type === 'array') {
          if (v.variable_type === 'string') {
            if (!val || (typeof val === 'string' && val.trim() === ''))
              nextErrors[v.variable_name] = 'Required';
          } else if (v.variable_type === 'array') {
            if (!Array.isArray(val) || val.length === 0) nextErrors[v.variable_name] = 'Required';
          }
        }
        if (v.variable_type === 'integer') {
          if (val === '' || val === undefined || val === null || Number.isNaN(val))
            nextErrors[v.variable_name] = 'Required';
        }
      }
    });
    setErrors(nextErrors);
    if (Object.keys(nextErrors).length > 0) return;
    onSubmit?.(values);
    onClose?.();
  };

  return {
    values,
    errors,
    handleChange,
    handleSubmit,
    requiredFields
  };
}
