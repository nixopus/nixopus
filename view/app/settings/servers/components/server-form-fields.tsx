import React from 'react';
import { Label } from '@/components/ui/label';
import { AuthenticationType } from '@/redux/types/server';
import FormInputField from '@/components/ui/form-input-field';
import FormTextareaField from '@/components/ui/form-textarea-field';

interface ServerFormFieldsProps {
  form: any;
  t: (key: any) => string;
  authType: AuthenticationType;
  setAuthType: (type: AuthenticationType) => void;
  isEditMode: boolean;
}

export function ServerFormFields({ form, t, authType, setAuthType, isEditMode }: ServerFormFieldsProps) {
  const inputFields = [
    {
      name: 'name',
      label: t('servers.create.dialog.fields.name.label'),
      placeholder: t('servers.create.dialog.fields.name.placeholder')
    },
    {
      name: 'username',
      label: t('servers.create.dialog.fields.username.label'),
      placeholder: t('servers.create.dialog.fields.username.placeholder')
    }
  ];

  const gridFields = [
    {
      name: 'host',
      label: t('servers.create.dialog.fields.host.label'),
      placeholder: t('servers.create.dialog.fields.host.placeholder')
    },
    {
      name: 'port',
      label: t('servers.create.dialog.fields.port.label'),
      placeholder: t('servers.create.dialog.fields.port.placeholder'),
      type: 'number' as const,
      onValueChange: (v: string) => parseInt(v) || 22
    }
  ];

  const descriptionField = {
    name: 'description',
    label: t('servers.create.dialog.fields.description.label'),
    placeholder: t('servers.create.dialog.fields.description.placeholder')
  };

  const authField =
    authType === AuthenticationType.PASSWORD
      ? {
          name: 'ssh_password',
          label: t('servers.create.dialog.fields.sshPassword.label'),
          placeholder: isEditMode
            ? 'Leave empty to keep current password'
            : t('servers.create.dialog.fields.sshPassword.placeholder'),
          type: 'password' as const,
          required: false
        }
      : {
          name: 'ssh_private_key_path',
          label: t('servers.create.dialog.fields.privateKeyPath.label'),
          placeholder: isEditMode
            ? 'Leave empty to keep current private key path'
            : t('servers.create.dialog.fields.privateKeyPath.placeholder'),
          required: false
        };

  return (
    <>
      {inputFields.map((cfg) => (
        <FormInputField
          key={cfg.name}
          form={form}
          name={cfg.name}
          label={cfg.label}
          placeholder={cfg.placeholder}
        />
      ))}

      <FormTextareaField
        form={form}
        name={descriptionField.name}
        label={descriptionField.label}
        placeholder={descriptionField.placeholder}
        rows={4}
      />

      <div className="grid grid-cols-2 gap-4">
        {gridFields.map((cfg) => (
          <FormInputField
            key={cfg.name}
            form={form}
            name={cfg.name}
            label={cfg.label}
            placeholder={cfg.placeholder}
            type={(cfg as any).type}
            onValueChange={(cfg as any).onValueChange}
          />
        ))}
      </div>

      <div className="space-y-4">
        <Label className="text-sm font-medium">{t('servers.create.dialog.fields.authMethod.label')}</Label>
        <div className="flex space-x-6">
          <div className="flex items-center space-x-2">
            <input
              type="radio"
              id="password"
              name="authType"
              value={AuthenticationType.PASSWORD}
              checked={authType === AuthenticationType.PASSWORD}
              onChange={(e) => setAuthType(e.target.value as AuthenticationType)}
              className="h-4 w-4 text-primary focus:ring-primary border-input"
            />
            <Label htmlFor="password">{t('servers.create.dialog.fields.authMethod.password')}</Label>
          </div>
          <div className="flex items-center space-x-2">
            <input
              type="radio"
              id="private-key"
              name="authType"
              value={AuthenticationType.PRIVATE_KEY}
              checked={authType === AuthenticationType.PRIVATE_KEY}
              onChange={(e) => setAuthType(e.target.value as AuthenticationType)}
              className="h-4 w-4 text-primary focus:ring-primary border-input"
            />
            <Label htmlFor="private-key">{t('servers.create.dialog.fields.authMethod.privateKey')}</Label>
          </div>
        </div>

        <FormInputField
          form={form}
          name={authField.name}
          label={authField.label}
          placeholder={authField.placeholder}
          type={(authField as any).type}
          required={authField.required as boolean}
        />
      </div>
    </>
  );
}

export default ServerFormFields;
