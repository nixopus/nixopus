'use client';

import React from 'react';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { Checkbox } from '@/components/ui/checkbox';
import { ScrollArea } from '@/components/ui/scroll-area';
import {
  Eye,
  EyeOff,
  Pencil,
  Trash2,
  Plus,
  Check,
  X,
  Lock,
  FileText,
  AlertCircle
} from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useEnvVariablesEditor,
  type ValidationType,
  type PastePreviewItem,
  type EnvVariable
} from '@/packages/hooks/shared/use-env-variables-editor';
import { cn } from '@/lib/utils';

interface EnvVariablesEditorProps {
  label: string;
  name: string;
  description?: string;
  form: any;
  required?: boolean;
  validator: (value: string) => ValidationType;
  defaultValues?: Record<string, string>;
}

interface EnvVariableRowProps {
  variable: EnvVariable;
  isEditing: boolean;
  editKey: string;
  editValue: string;
  isRevealed: boolean;
  maskValue: (value: string) => string;
  onUpdateEditKey: (value: string) => void;
  onUpdateEditValue: (value: string) => void;
  onSaveEdit: () => void;
  onCancelEdit: () => void;
  onToggleReveal: () => void;
  onToggleSecret: () => void;
  onStartEditing: () => void;
  onRemove: () => void;
}

const EnvVariableRow = ({
  variable,
  isEditing,
  editKey,
  editValue,
  isRevealed,
  maskValue,
  onUpdateEditKey,
  onUpdateEditValue,
  onSaveEdit,
  onCancelEdit,
  onToggleReveal,
  onToggleSecret,
  onStartEditing,
  onRemove
}: EnvVariableRowProps) => {
  const { t } = useTranslation();

  if (isEditing) {
    return (
      <>
        <Input
          value={editKey}
          onChange={(e) => onUpdateEditKey(e.target.value)}
          className="h-8 font-mono text-sm flex-1 max-w-[140px]"
          placeholder={t('selfHost.envEditor.keyPlaceholder')}
        />
        <span className="text-muted-foreground">=</span>
        <Input
          value={editValue}
          onChange={(e) => onUpdateEditValue(e.target.value)}
          className="h-8 font-mono text-sm flex-1"
          placeholder={t('selfHost.envEditor.valuePlaceholder')}
        />
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onSaveEdit}
          className="h-8 w-8 p-0 text-green-600 hover:text-green-700 hover:bg-green-100 dark:hover:bg-green-900/20"
        >
          <Check size={16} />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onCancelEdit}
          className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
        >
          <X size={16} />
        </Button>
      </>
    );
  }

  return (
    <>
      <div className="flex items-center gap-2 min-w-0 flex-1">
        {variable.isSecret && <Lock size={14} className="text-amber-500 flex-shrink-0" />}
        <span className="font-mono text-sm font-medium truncate max-w-[120px] sm:max-w-[160px]">
          {variable.key}
        </span>
        <span className="text-muted-foreground">=</span>
        <span
          className={cn(
            'font-mono text-sm text-muted-foreground truncate flex-1 min-w-0',
            variable.isSecret && !isRevealed && 'select-none'
          )}
          title={
            variable.isSecret && !isRevealed
              ? t('selfHost.envEditor.clickToReveal')
              : variable.value
          }
        >
          {variable.isSecret && !isRevealed ? maskValue(variable.value) : variable.value}
        </span>
      </div>

      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
        {variable.isSecret && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onToggleReveal}
            className="h-8 w-8 p-0"
            title={isRevealed ? t('selfHost.envEditor.hide') : t('selfHost.envEditor.reveal')}
          >
            {isRevealed ? <EyeOff size={14} /> : <Eye size={14} />}
          </Button>
        )}
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onToggleSecret}
          className={cn('h-8 w-8 p-0', variable.isSecret && 'text-amber-500')}
          title={
            variable.isSecret
              ? t('selfHost.envEditor.unmarkSecret')
              : t('selfHost.envEditor.markSecret')
          }
        >
          <Lock size={14} />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onStartEditing}
          className="h-8 w-8 p-0"
          title={t('common.edit')}
        >
          <Pencil size={14} />
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onRemove}
          className="h-8 w-8 p-0 text-muted-foreground hover:text-destructive"
          title={t('common.delete')}
        >
          <Trash2 size={14} />
        </Button>
      </div>
    </>
  );
};

interface PastePreviewDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  items: PastePreviewItem[];
  variables: EnvVariable[];
  onToggleSelection: (index: number, checked: boolean) => void;
  onToggleSecret: (index: number) => void;
  onConfirm: () => void;
}

const PastePreviewDialog = ({
  open,
  onOpenChange,
  items,
  variables,
  onToggleSelection,
  onToggleSecret,
  onConfirm
}: PastePreviewDialogProps) => {
  const { t } = useTranslation();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('selfHost.envEditor.pastePreview.title')}</DialogTitle>
          <DialogDescription>{t('selfHost.envEditor.pastePreview.description')}</DialogDescription>
        </DialogHeader>

        <ScrollArea className="max-h-[320px] pr-4">
          <div className="space-y-2">
            {items.map((item, index) => (
              <div
                key={`preview-${item.key}-${index}`}
                className={cn(
                  'flex items-center gap-3 p-3 rounded-lg border transition-colors',
                  item.selected ? 'bg-muted/50 border-primary/30' : 'opacity-50'
                )}
              >
                <Checkbox
                  checked={item.selected}
                  onCheckedChange={(checked) => onToggleSelection(index, checked as boolean)}
                />
                <div className="flex-1 min-w-0 space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-sm font-medium">{item.key}</span>
                    {variables.some((v) => v.key === item.key) && (
                      <span className="text-xs text-amber-500 bg-amber-500/10 px-1.5 py-0.5 rounded">
                        {t('selfHost.envEditor.pastePreview.willOverwrite')}
                      </span>
                    )}
                  </div>
                  <p className="font-mono text-xs text-muted-foreground truncate">{item.value}</p>
                </div>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => onToggleSecret(index)}
                  className={cn('h-8 w-8 p-0', item.isSecret && 'text-amber-500')}
                  title={t('selfHost.envEditor.markSecret')}
                >
                  <Lock size={14} />
                </Button>
              </div>
            ))}
          </div>
        </ScrollArea>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onConfirm} disabled={!items.some((item) => item.selected)}>
            {t('selfHost.envEditor.pastePreview.import', {
              count: String(items.filter((item) => item.selected).length)
            })}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

export const EnvVariablesEditor = ({
  label,
  name,
  description,
  form,
  validator,
  required = false,
  defaultValues = {}
}: EnvVariablesEditorProps) => {
  const { t } = useTranslation();
  const {
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
  } = useEnvVariablesEditor({
    validator,
    defaultValues,
    form,
    formFieldName: name
  });

  return (
    <FormField
      control={form.control}
      name={name}
      render={() => (
        <FormItem className="space-y-3">
          <div className="flex gap-2">
            {label && <FormLabel>{label}</FormLabel>}
            <span className="text-destructive w-3 flex-shrink-0 text-right">
              {required ? '*' : ''}
            </span>
          </div>

          <FormControl>
            <div className="space-y-3">
              {variables.length > 0 && (
                <div className="rounded-lg border bg-card">
                  <ScrollArea className={cn(variables.length > 5 && 'h-[280px]')}>
                    <div className="divide-y">
                      {variables.map((variable, index) => (
                        <div
                          key={`${variable.key}-${index}`}
                          className="group flex items-center gap-2 px-3 py-2.5 hover:bg-muted/50 transition-colors"
                        >
                          <EnvVariableRow
                            variable={variable}
                            isEditing={editingIndex === index}
                            editKey={editKey}
                            editValue={editValue}
                            isRevealed={revealedIndices.has(index)}
                            maskValue={maskValue}
                            onUpdateEditKey={updateEditKey}
                            onUpdateEditValue={updateEditValue}
                            onSaveEdit={saveEdit}
                            onCancelEdit={cancelEdit}
                            onToggleReveal={() => toggleReveal(index)}
                            onToggleSecret={() => toggleSecret(index)}
                            onStartEditing={() => startEditing(index)}
                            onRemove={() => removeVariable(index)}
                          />
                        </div>
                      ))}
                    </div>
                  </ScrollArea>
                </div>
              )}

              <div className="flex gap-2">
                <Input
                  value={newKey}
                  onChange={(e) => updateNewKey(e.target.value)}
                  onKeyDown={handleKeyDown}
                  onPaste={handlePaste}
                  placeholder={t('selfHost.envEditor.keyPlaceholder')}
                  className={cn('font-mono flex-1 max-w-[160px]', error && 'border-destructive')}
                />
                <span className="flex items-center text-muted-foreground">=</span>
                <Input
                  value={newValue}
                  onChange={(e) => updateNewValue(e.target.value)}
                  onKeyDown={handleKeyDown}
                  onPaste={handlePaste}
                  placeholder={t('selfHost.envEditor.valuePlaceholder')}
                  className={cn('font-mono flex-1', error && 'border-destructive')}
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={addVariable}
                  disabled={!newKey.trim() || !newValue.trim()}
                  className="flex-shrink-0"
                >
                  <Plus size={16} />
                </Button>
              </div>

              {error && (
                <p className="text-sm font-medium text-destructive flex items-center gap-1.5">
                  <AlertCircle size={14} />
                  {error}
                </p>
              )}

              <p className="text-xs text-muted-foreground flex items-center gap-1.5">
                <FileText size={12} />
                {t('selfHost.envEditor.pasteHint')}
              </p>
            </div>
          </FormControl>

          <FormDescription>{description}</FormDescription>
          <FormMessage />

          <PastePreviewDialog
            open={pastePreviewOpen}
            onOpenChange={setPastePreviewOpen}
            items={pastePreviewItems}
            variables={variables}
            onToggleSelection={togglePastePreviewItemSelection}
            onToggleSecret={togglePastePreviewItemSecret}
            onConfirm={confirmPasteImport}
          />
        </FormItem>
      )}
    />
  );
};
