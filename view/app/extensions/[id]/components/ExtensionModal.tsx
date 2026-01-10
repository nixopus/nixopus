'use client';

import { DialogAction } from '@/components/ui/dialog-wrapper';
import { ExtensionInput } from '@/packages/components/extension';
import { useExtensionInput } from '@/packages/hooks/extensions/use-extension-input';
import type { Extension } from '@/redux/types/extension';
import { useTranslation } from '@/hooks/use-translation';

interface ExtensionModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  extension?: Extension;
  onSubmit: (values: Record<string, unknown>) => Promise<void>;
}

export function ExtensionModal({ open, onOpenChange, extension, onSubmit }: ExtensionModalProps) {
  const { t } = useTranslation();
  const { values, errors, handleChange, handleSubmit, requiredFields } = useExtensionInput({
    extension,
    open,
    onSubmit,
    onClose: () => onOpenChange(false)
  });
  const actions: DialogAction[] = [
    {
      label: t('common.cancel'),
      onClick: () => onOpenChange(false),
      variant: 'ghost'
    }
  ];
  const isOnlyProxyDomain =
    requiredFields.length === 1 &&
    (requiredFields[0].variable_name.toLowerCase() === 'proxy_domain' ||
      requiredFields[0].variable_name.toLowerCase() === 'domain');
  const noFieldsToShow = requiredFields.length === 0;
  return (
    <ExtensionInput
      open={open}
      onOpenChange={onOpenChange}
      extension={extension}
      onSubmit={onSubmit}
      t={t}
      actions={actions}
      isOnlyProxyDomain={isOnlyProxyDomain}
      noFieldsToShow={noFieldsToShow}
      values={values}
      errors={errors}
      handleChange={handleChange}
      handleSubmit={handleSubmit}
      requiredFields={requiredFields}
    />
  );
}
