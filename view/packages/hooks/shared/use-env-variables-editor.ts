'use client';

import { useState, useEffect, useCallback, KeyboardEvent, ClipboardEvent } from 'react';
import { parseEnvText, isMultiLineEnvPaste } from '@/packages/utils/parse-env';
import { useTranslation } from '@/packages/hooks/shared/use-translation';

export interface EnvVariable {
  key: string;
  value: string;
  isSecret: boolean;
}

export interface ValidationType {
  isValid: boolean;
  error?: string;
  key?: string;
  value?: string;
}

export interface PastePreviewItem {
  key: string;
  value: string;
  isSecret: boolean;
  selected: boolean;
}

interface UseEnvVariablesEditorProps {
  validator: (value: string) => ValidationType;
  defaultValues?: Record<string, string>;
  form: any;
  formFieldName: string;
}

export function useEnvVariablesEditor({
  validator,
  defaultValues = {},
  form,
  formFieldName
}: UseEnvVariablesEditorProps) {
  const { t } = useTranslation();
  const [variables, setVariables] = useState<EnvVariable[]>([]);
  const [newKey, setNewKey] = useState('');
  const [newValue, setNewValue] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editKey, setEditKey] = useState('');
  const [editValue, setEditValue] = useState('');
  const [revealedIndices, setRevealedIndices] = useState<Set<number>>(new Set());
  const [pastePreviewOpen, setPastePreviewOpen] = useState(false);
  const [pastePreviewItems, setPastePreviewItems] = useState<PastePreviewItem[]>([]);

  useEffect(() => {
    if (defaultValues && Object.keys(defaultValues).length > 0) {
      const vars = Object.entries(defaultValues).map(([key, value]) => ({
        key,
        value,
        isSecret: false
      }));
      setVariables(vars);
    }
  }, [defaultValues]);

  useEffect(() => {
    const record: Record<string, string> = {};
    variables.forEach((v) => {
      record[v.key] = v.value;
    });
    form.setValue(formFieldName, record);
  }, [variables, form, formFieldName]);

  const addVariable = useCallback(() => {
    const input = `${newKey}=${newValue}`;
    const validation = validator(input);

    if (!validation.isValid) {
      setError(validation.error ?? t('selfHost.envEditor.errors.invalidFormat'));
      return;
    }

    const existingIndex = variables.findIndex((v) => v.key === newKey);
    if (existingIndex !== -1) {
      setVariables((prev) =>
        prev.map((v, i) => (i === existingIndex ? { ...v, value: newValue } : v))
      );
    } else {
      setVariables((prev) => [...prev, { key: newKey, value: newValue, isSecret: false }]);
    }

    setNewKey('');
    setNewValue('');
    setError(null);
  }, [newKey, newValue, validator, variables, t]);

  const handleKeyDown = useCallback(
    (e: KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        if (newKey.trim() && newValue.trim()) {
          addVariable();
        }
      }
    },
    [newKey, newValue, addVariable]
  );

  const handlePaste = useCallback((e: ClipboardEvent<HTMLInputElement>) => {
    const pastedText = e.clipboardData.getData('text');
    if (isMultiLineEnvPaste(pastedText)) {
      e.preventDefault();
      const parsed = parseEnvText(pastedText);
      const items: PastePreviewItem[] = Object.entries(parsed).map(([key, value]) => ({
        key,
        value,
        isSecret: false,
        selected: true
      }));
      if (items.length > 0) {
        setPastePreviewItems(items);
        setPastePreviewOpen(true);
      }
    }
  }, []);

  const confirmPasteImport = useCallback(() => {
    const selectedItems = pastePreviewItems.filter((item) => item.selected);
    setVariables((prev) => {
      const newVars = [...prev];
      selectedItems.forEach((item) => {
        const existingIndex = newVars.findIndex((v) => v.key === item.key);
        if (existingIndex !== -1) {
          newVars[existingIndex] = { key: item.key, value: item.value, isSecret: item.isSecret };
        } else {
          newVars.push({ key: item.key, value: item.value, isSecret: item.isSecret });
        }
      });
      return newVars;
    });
    setPastePreviewOpen(false);
    setPastePreviewItems([]);
  }, [pastePreviewItems]);

  const removeVariable = useCallback((index: number) => {
    setVariables((prev) => prev.filter((_, i) => i !== index));
    setRevealedIndices((prev) => {
      const newSet = new Set(prev);
      newSet.delete(index);
      return newSet;
    });
  }, []);

  const startEditing = useCallback(
    (index: number) => {
      const variable = variables[index];
      setEditingIndex(index);
      setEditKey(variable.key);
      setEditValue(variable.value);
    },
    [variables]
  );

  const cancelEdit = useCallback(() => {
    setEditingIndex(null);
    setEditKey('');
    setEditValue('');
    setError(null);
  }, []);

  const saveEdit = useCallback(() => {
    if (editingIndex === null) return;

    const input = `${editKey}=${editValue}`;
    const validation = validator(input);

    if (!validation.isValid) {
      setError(validation.error ?? t('selfHost.envEditor.errors.invalidFormat'));
      return;
    }

    setVariables((prev) =>
      prev.map((v, i) => (i === editingIndex ? { ...v, key: editKey, value: editValue } : v))
    );
    cancelEdit();
  }, [editingIndex, editKey, editValue, validator, t, cancelEdit]);

  const toggleSecret = useCallback((index: number) => {
    setVariables((prev) => prev.map((v, i) => (i === index ? { ...v, isSecret: !v.isSecret } : v)));
  }, []);

  const toggleReveal = useCallback((index: number) => {
    setRevealedIndices((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(index)) {
        newSet.delete(index);
      } else {
        newSet.add(index);
      }
      return newSet;
    });
  }, []);

  const updateNewKey = useCallback(
    (value: string) => {
      setNewKey(value.toUpperCase().replace(/[^A-Z0-9_]/g, ''));
      if (error) setError(null);
    },
    [error]
  );

  const updateNewValue = useCallback(
    (value: string) => {
      setNewValue(value);
      if (error) setError(null);
    },
    [error]
  );

  const togglePastePreviewItemSelection = useCallback((index: number, checked: boolean) => {
    setPastePreviewItems((prev) =>
      prev.map((p, i) => (i === index ? { ...p, selected: checked } : p))
    );
  }, []);

  const togglePastePreviewItemSecret = useCallback((index: number) => {
    setPastePreviewItems((prev) =>
      prev.map((p, i) => (i === index ? { ...p, isSecret: !p.isSecret } : p))
    );
  }, []);

  const updateEditKey = useCallback((value: string) => {
    setEditKey(value);
  }, []);

  const updateEditValue = useCallback((value: string) => {
    setEditValue(value);
  }, []);

  const maskValue = useCallback((value: string) => {
    return '•'.repeat(Math.min(value.length, 20));
  }, []);

  return {
    variables,
    newKey,
    newValue,
    error,
    editingIndex,
    editKey,
    editValue,
    revealedIndices,
    pastePreviewOpen,
    pastePreviewItems,
    updateNewKey,
    updateNewValue,
    updateEditKey,
    updateEditValue,
    handleKeyDown,
    handlePaste,
    addVariable,
    removeVariable,
    startEditing,
    saveEdit,
    cancelEdit,
    toggleSecret,
    toggleReveal,
    togglePastePreviewItemSelection,
    togglePastePreviewItemSecret,
    confirmPasteImport,
    setPastePreviewOpen,
    maskValue
  };
}
