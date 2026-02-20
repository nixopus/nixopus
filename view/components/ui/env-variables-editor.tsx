'use client';

import React from 'react';
import {
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage
} from '@nixopus/ui';
import { Textarea } from '@nixopus/ui';
import { Button } from '@nixopus/ui';
import AceEditor from '@/components/ui/ace-editor';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@nixopus/ui';
import { Checkbox } from '@nixopus/ui';
import { ScrollArea } from '@nixopus/ui';
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
  AlertCircle,
  Copy,
  CheckCircle2,
  ChevronDown,
  ChevronUp
} from 'lucide-react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import {
  useEnvVariablesEditor,
  type ValidationType,
  type PastePreviewItem,
  type EnvVariable
} from '@/packages/hooks/shared/use-env-variables-editor';
import { isMultiLineEnvPaste } from '@/packages/utils/parse-env';
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
  isExpanded: boolean;
  maskValue: (value: string) => string;
  onUpdateEditKey: (value: string) => void;
  onUpdateEditValue: (value: string) => void;
  onSaveEdit: () => void;
  onCancelEdit: () => void;
  onToggleReveal: () => void;
  onToggleSecret: () => void;
  onStartEditing: () => void;
  onRemove: () => void;
  onCopy: () => void;
  onToggleExpand: () => void;
  copied: boolean;
}

const EnvVariableRow = ({
  variable,
  isEditing,
  editKey,
  editValue,
  isRevealed,
  isExpanded,
  maskValue,
  onUpdateEditKey,
  onUpdateEditValue,
  onSaveEdit,
  onCancelEdit,
  onToggleReveal,
  onToggleSecret,
  onStartEditing,
  onRemove,
  onCopy,
  onToggleExpand,
  copied
}: EnvVariableRowProps) => {
  const { t } = useTranslation();
  const isLongValue = variable.value.length > 60;
  const displayValue =
    variable.isSecret && !isRevealed ? maskValue(variable.value) : variable.value;
  const shouldShowExpand = isLongValue && !isExpanded;

  if (isEditing) {
    return (
      <div className="flex items-start gap-2 w-full py-1">
        <Textarea
          value={editKey}
          onChange={(e) => onUpdateEditKey(e.target.value)}
          className="font-mono text-sm min-h-[36px] max-w-[200px] resize-none"
          placeholder={t('selfHost.envEditor.keyPlaceholder')}
          rows={1}
        />
        <span className="text-muted-foreground flex-shrink-0 mt-2">=</span>
        <Textarea
          value={editValue}
          onChange={(e) => onUpdateEditValue(e.target.value)}
          className="font-mono text-sm min-h-[36px] flex-1 resize-none"
          placeholder={t('selfHost.envEditor.valuePlaceholder')}
          rows={editValue.length > 100 ? 3 : 1}
        />
        <div className="flex items-start gap-1 flex-shrink-0 pt-1">
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
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="flex items-start gap-2 min-w-0 flex-1 max-w-full">
        {variable.isSecret && <Lock size={14} className="text-amber-500 flex-shrink-0 mt-1" />}
        <div className="flex-1 min-w-0 max-w-full overflow-hidden">
          <div className="flex items-center gap-2 mb-1 min-w-0">
            <span className="font-mono text-sm font-medium break-all min-w-0">{variable.key}</span>
            <span className="text-muted-foreground flex-shrink-0">=</span>
          </div>
          <div className="flex items-start gap-2 min-w-0">
            <div className="flex-1 min-w-0 max-w-full overflow-hidden">
              <div
                className={cn(
                  'font-mono text-sm text-muted-foreground',
                  variable.isSecret && !isRevealed && 'select-none',
                  !isExpanded && isLongValue && 'line-clamp-2'
                )}
                style={{
                  wordBreak: 'break-all',
                  overflowWrap: 'anywhere',
                  wordWrap: 'break-word',
                  maxWidth: '100%'
                }}
              >
                {displayValue}
              </div>
              {shouldShowExpand && (
                <button
                  onClick={onToggleExpand}
                  className="text-xs text-primary hover:underline mt-1 flex items-center gap-1"
                >
                  {t('selfHost.envEditor.showMore')}
                  <ChevronDown size={12} />
                </button>
              )}
              {isExpanded && isLongValue && (
                <button
                  onClick={onToggleExpand}
                  className="text-xs text-primary hover:underline mt-1 flex items-center gap-1"
                >
                  {t('selfHost.envEditor.showLess')}
                  <ChevronUp size={12} />
                </button>
              )}
            </div>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onCopy}
          className={cn('h-8 w-8 p-0', copied && 'text-green-600')}
          title={copied ? t('selfHost.envEditor.copied') : t('selfHost.envEditor.copy')}
        >
          {copied ? <CheckCircle2 size={14} /> : <Copy size={14} />}
        </Button>
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

interface ImportVariablesDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  importText: string;
  onImportTextChange: (text: string) => void;
  onSubmit: () => void;
}

const ImportVariablesDialog = ({
  open,
  onOpenChange,
  importText,
  onImportTextChange,
  onSubmit
}: ImportVariablesDialogProps) => {
  const { t } = useTranslation();

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-2xl max-h-[90vh] flex flex-col">
        <DialogHeader>
          <DialogTitle>{t('selfHost.envEditor.pastePreview.title')}</DialogTitle>
          <DialogDescription>{t('selfHost.envEditor.pastePreview.description')}</DialogDescription>
        </DialogHeader>

        <div className="space-y-3 flex-1 min-h-0 flex flex-col">
          <div className="flex-1 min-h-0 border rounded-md overflow-hidden">
            <AceEditor
              mode="text"
              value={importText}
              onChange={onImportTextChange}
              name="import-variables-editor"
              height="400px"
            />
          </div>
          <p className="text-xs text-muted-foreground flex items-center gap-1.5">
            <FileText size={12} />
            {t('selfHost.envEditor.pasteHint')}
          </p>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={onSubmit} disabled={!importText.trim()}>
            {t('selfHost.envEditor.importPasted')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
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
  const [importModalOpen, setImportModalOpen] = React.useState(false);
  const [importText, setImportText] = React.useState('');

  const {
    variables,
    newKey,
    newValue,
    error,
    editingIndex,
    editKey,
    editValue,
    revealedIndices,
    expandedIndices,
    copiedIndex,
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
    maskValue,
    copyVariable,
    copyAllVariables,
    toggleExpand,
    importVariablesDirectly
  } = useEnvVariablesEditor({
    validator,
    defaultValues,
    form,
    formFieldName: name
  });

  const handleImportSubmit = React.useCallback(() => {
    if (importText.trim()) {
      importVariablesDirectly(importText);
      setImportText('');
      setImportModalOpen(false);
    }
  }, [importText, importVariablesDirectly]);

  const handleKeyPaste = React.useCallback(
    (e: React.ClipboardEvent<HTMLTextAreaElement>) => {
      const pastedText = e.clipboardData.getData('text');
      if (isMultiLineEnvPaste(pastedText)) {
        e.preventDefault();
        setImportText(pastedText);
        setImportModalOpen(true);
      } else {
        // Allow normal paste for single line
        handlePaste(e);
      }
    },
    [handlePaste]
  );

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
              {/* Add New Variable Input - Middle for primary action */}
              <div className="flex flex-col gap-2">
                <div className="flex items-center gap-2">
                  <Textarea
                    value={newKey}
                    onChange={(e) => updateNewKey(e.target.value)}
                    onKeyDown={handleKeyDown}
                    onPaste={handleKeyPaste}
                    placeholder={t('selfHost.envEditor.keyPlaceholder')}
                    className={cn(
                      'font-mono text-sm min-h-[36px] max-w-[200px] resize-none',
                      error && 'border-destructive'
                    )}
                    rows={1}
                  />
                  <span className="flex items-center text-muted-foreground flex-shrink-0">=</span>
                  <Textarea
                    value={newValue}
                    onChange={(e) => updateNewValue(e.target.value)}
                    onKeyDown={handleKeyDown}
                    onPaste={handlePaste}
                    placeholder={t('selfHost.envEditor.valuePlaceholder')}
                    className={cn(
                      'font-mono text-sm min-h-[36px] flex-1 resize-none',
                      error && 'border-destructive'
                    )}
                    rows={newValue.length > 100 ? 3 : 1}
                  />
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={addVariable}
                    disabled={!newKey.trim() || !newValue.trim()}
                    className="flex-shrink-0 h-[36px] w-[36px]"
                  >
                    <Plus size={16} />
                  </Button>
                </div>
              </div>

              {/* Variables List - Bottom to show all added variables */}
              {variables.length > 0 && (
                <div className="rounded-lg border bg-card">
                  <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
                    <span className="text-sm font-medium">
                      {variables.length}{' '}
                      {variables.length === 1
                        ? t('selfHost.envEditor.variable')
                        : t('selfHost.envEditor.variables')}
                    </span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={copyAllVariables}
                      className="h-7 text-xs gap-1.5"
                      title={t('selfHost.envEditor.copyAll')}
                    >
                      <Copy size={12} />
                      {t('selfHost.envEditor.copyAll')}
                    </Button>
                  </div>
                  <ScrollArea className={cn(variables.length > 5 && 'h-[400px]')}>
                    <div className="divide-y">
                      {variables.map((variable, index) => (
                        <div
                          key={`${variable.key}-${index}`}
                          className="group flex items-start gap-2 px-3 py-3 hover:bg-muted/50 transition-colors min-w-0 max-w-full overflow-hidden"
                        >
                          <EnvVariableRow
                            variable={variable}
                            isEditing={editingIndex === index}
                            editKey={editKey}
                            editValue={editValue}
                            isRevealed={revealedIndices.has(index)}
                            isExpanded={expandedIndices.has(index)}
                            maskValue={maskValue}
                            onUpdateEditKey={updateEditKey}
                            onUpdateEditValue={updateEditValue}
                            onSaveEdit={saveEdit}
                            onCancelEdit={cancelEdit}
                            onToggleReveal={() => toggleReveal(index)}
                            onToggleSecret={() => toggleSecret(index)}
                            onStartEditing={() => startEditing(index)}
                            onRemove={() => removeVariable(index)}
                            onCopy={() => copyVariable(index)}
                            onToggleExpand={() => toggleExpand(index)}
                            copied={copiedIndex === index}
                          />
                        </div>
                      ))}
                    </div>
                  </ScrollArea>
                </div>
              )}

              {error && (
                <p className="text-sm font-medium text-destructive flex items-center gap-1.5">
                  <AlertCircle size={14} />
                  {error}
                </p>
              )}
            </div>
          </FormControl>

          <FormDescription>{description}</FormDescription>
          <FormMessage />

          <ImportVariablesDialog
            open={importModalOpen}
            onOpenChange={setImportModalOpen}
            importText={importText}
            onImportTextChange={setImportText}
            onSubmit={handleImportSubmit}
          />
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
