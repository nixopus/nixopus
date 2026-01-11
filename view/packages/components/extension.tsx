'use client';

import React from 'react';
import { BadgeGroup, BadgeGroupItem } from '@/components/ui/badge-group';
import { cn } from '@/lib/utils';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { Skeleton } from '@/components/ui/skeleton';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertCircle } from 'lucide-react';
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetDescription,
  SheetFooter
} from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import AceEditor from '@/components/ui/ace-editor';
import { CardWrapper } from '@/components/ui/card-wrapper';
import { TypographyMuted, TypographySmall } from '@/components/ui/typography';
import { DataTable } from '@/components/ui/data-table';
import { ExternalLink, Check, GitFork, Trash2, ChevronDown } from 'lucide-react';
import { CardDescription } from '@/components/ui/card';
import { DeleteDialog } from '@/components/ui/delete-dialog';
import {
  CategoryBadgesProps,
  ExtensionsGridProps,
  ExtensionForkDialogProps,
  ExtensionCardProps,
  ExtensionInputProps
} from '@/packages/types/extension';
import { Sparkles, Globe } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import { ExtensionVariable } from '@/redux/types/extension';
import { DialogWrapper } from '@/components/ui/dialog-wrapper';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { useMemo } from 'react';
import { TableColumn } from '@/components/ui/data-table';
import { Badge } from '@/components/ui/badge';

interface StepsSectionProps {
  tRunLabel: string;
  tValidateLabel: string;
  title: string;
  runSteps: any[];
  validateSteps: any[];
  openRunIndex: number | null;
  setOpenRunIndex: (index: number | null) => void;
  openValidateIndex: number | null;
  setOpenValidateIndex: (index: number | null) => void;
}

function StepsSection({
  tRunLabel,
  tValidateLabel,
  title,
  runSteps,
  validateSteps,
  openRunIndex,
  setOpenRunIndex,
  openValidateIndex,
  setOpenValidateIndex
}: StepsSectionProps) {
  const entryColumns: TableColumn<[string, any]>[] = useMemo(
    () => [
      {
        key: 'key',
        title: 'Key',
        render: ([k]) => k,
        width: '25%',
        className: 'text-muted-foreground'
      },
      {
        key: 'value',
        title: 'Value',
        render: ([, v]) => {
          if (typeof v === 'object' && v !== null) {
            return <pre className="text-xs">{JSON.stringify(v, null, 2)}</pre>;
          }
          return String(v ?? '');
        }
      }
    ],
    []
  );

  const createStepColumns = (
    label: string,
    openIndex: number | null,
    onToggle: (index: number | null) => void
  ): TableColumn<any>[] =>
    useMemo(
      () => [
        {
          key: 'step',
          title: label,
          render: (step, _, index) => {
            const entries = Object.entries(step || {}).filter(
              ([k]) => k !== 'name' && k !== 'type'
            );
            const isOpen = openIndex === index;
            return (
              <Collapsible open={isOpen} onOpenChange={(open) => onToggle(open ? index : null)}>
                <CollapsibleTrigger className="w-full text-left flex items-start gap-2 group">
                  <TypographyMuted className="mt-0.5 shrink-0">{index + 1}.</TypographyMuted>
                  <div className="flex-1 min-w-0">
                    <TypographySmall className="font-medium">
                      {step?.name || step?.type || 'Step'}
                    </TypographySmall>
                    {step?.type && (
                      <Badge variant="outline" className="ml-2">
                        {step.type}
                      </Badge>
                    )}
                  </div>
                  <ChevronDown
                    className={`h-4 w-4 shrink-0 transition-transform duration-200 ${
                      isOpen ? 'rotate-180' : ''
                    }`}
                  />
                </CollapsibleTrigger>
                {entries.length > 0 && entryColumns && (
                  <CollapsibleContent className="mt-3">
                    <DataTable
                      data={entries}
                      columns={entryColumns}
                      showBorder={true}
                      striped={false}
                      containerClassName="rounded-md"
                    />
                  </CollapsibleContent>
                )}
              </Collapsible>
            );
          }
        }
      ],
      [label, openIndex, onToggle, entryColumns]
    );

  const runStepColumns = createStepColumns(tRunLabel, openRunIndex ?? null, setOpenRunIndex);
  const validateStepColumns = createStepColumns(
    tValidateLabel,
    openValidateIndex ?? null,
    setOpenValidateIndex
  );

  const hasRun = runSteps.length > 0;
  const hasValidate = validateSteps.length > 0;

  if (!hasRun && !hasValidate) {
    return null;
  }

  return (
    <div className="flex flex-col gap-4">
      {hasRun && (
        <DataTable
          data={runSteps}
          columns={runStepColumns}
          showBorder={true}
          striped={false}
          rowClassName="border-0"
        />
      )}
      {hasValidate && (
        <DataTable
          data={validateSteps}
          columns={validateStepColumns}
          showBorder={true}
          striped={false}
          rowClassName="border-0"
        />
      )}
    </div>
  );
}

export default function CategoryBadges({
  categories,
  selected = null,
  onChange,
  className,
  showAll = true
}: CategoryBadgesProps) {
  const handleSelect = (value: string | null) => {
    onChange?.(value);
  };

  return (
    <BadgeGroup className={cn('w-full', className)} selectable>
      {showAll && (
        <BadgeGroupItem
          selected={selected === null}
          clickable
          onClick={() => handleSelect(null)}
          className="px-3 py-1"
        >
          All
        </BadgeGroupItem>
      )}
      {categories.map((cat) => (
        <BadgeGroupItem
          key={cat}
          selected={selected === cat}
          clickable
          onClick={() => handleSelect(cat === selected ? null : cat)}
          className="px-3 py-1"
        >
          {cat}
        </BadgeGroupItem>
      ))}
    </BadgeGroup>
  );
}

export function ExtensionGrid({
  extensions = [],
  isLoading = false,
  error,
  onInstall,
  onViewDetails,
  onForkClick,
  setConfirmOpen,
  expanded,
  setExpanded,
  forkOpen,
  setForkOpen,
  confirmOpen,
  forkYaml,
  setForkYaml,
  preview,
  variableColumns,
  doFork,
  selectedExtension
}: ExtensionsGridProps) {
  const { t } = useTranslation();

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircle className="h-4 w-4" />
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    );
  }

  if (isLoading) {
    return <ExtensionGridSkeleton />;
  }

  if (extensions.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="mx-auto max-w-md">
          <div className="text-6xl mb-4">🔍</div>
          <h3 className="text-lg font-semibold mb-2">{t('extensions.noExtensions')}</h3>
          <p className="text-muted-foreground">
            Try adjusting your search or filters to find more extensions.
          </p>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        {extensions.map((extension) => (
          <ExtensionCard
            key={extension.id}
            extension={extension}
            onInstall={onInstall}
            onViewDetails={onViewDetails}
            onForkClick={onForkClick}
            setConfirmOpen={setConfirmOpen}
            expanded={expanded}
            setExpanded={setExpanded}
            t={t}
            confirmOpen={confirmOpen}
          />
        ))}
      </div>
      {selectedExtension && (
        <ExtensionForkDialog
          open={forkOpen}
          onOpenChange={setForkOpen}
          extension={selectedExtension}
          t={t}
          forkYaml={forkYaml}
          setForkYaml={setForkYaml}
          preview={preview}
          variableColumns={variableColumns}
          doFork={doFork}
          isLoading={false}
        />
      )}
    </>
  );
}

export function ExtensionGridSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <div
          key={i}
          className="group h-full transition-all duration-200 bg-card border-border p-6 rounded-xl"
        >
          <div className="space-y-4">
            <div className="flex items-start gap-4">
              <Skeleton className="h-12 w-12 rounded-full flex-shrink-0" />
              <div className="flex-1 min-w-0">
                <Skeleton className="h-6 w-48 mb-2" />
                <Skeleton className="h-4 w-32" />
              </div>
            </div>
            <Skeleton className="h-4 w-full" />
            <Skeleton className="h-4 w-3/4" />
            <div className="flex gap-2 pt-6">
              <Skeleton className="h-10 flex-1" />
              <Skeleton className="h-10 w-10" />
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}

export function ExtensionForkDialog({
  open,
  onOpenChange,
  extension,
  t,
  forkYaml,
  setForkYaml,
  preview,
  variableColumns,
  doFork,
  isLoading
}: ExtensionForkDialogProps) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="sm:max-w-3xl md:max-w-4xl w-full p-0">
        <SheetHeader className="px-4 pt-4 pb-0">
          <SheetTitle>{t('extensions.fork') || 'Fork Extension'}</SheetTitle>
          <SheetDescription>{extension.description}</SheetDescription>
        </SheetHeader>
        <div className="px-4 flex-1 overflow-hidden flex flex-col min-h-0">
          <Tabs defaultValue="edit" className="w-full flex-1 flex flex-col overflow-hidden min-h-0">
            <TabsList>
              <TabsTrigger value="edit">{t('common.edit') || 'Edit'}</TabsTrigger>
              <TabsTrigger value="preview">{t('common.preview') || 'Preview'}</TabsTrigger>
            </TabsList>
            <TabsContent value="edit" className="overflow-hidden flex-1 flex flex-col min-h-0">
              <div className="flex flex-col gap-2 h-full">
                <Label>{t('extensions.forkYaml') || 'YAML (optional)'}</Label>
                <div className="flex-1">
                  <AceEditor
                    mode="yaml"
                    name="fork-yaml-editor"
                    value={forkYaml}
                    onChange={setForkYaml}
                    height="100%"
                  />
                </div>
              </div>
            </TabsContent>
            <TabsContent value="preview" className="flex-1 overflow-y-auto min-h-0">
              <div className="space-y-4">
                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <CardWrapper compact>
                    <TypographyMuted className="text-xs mb-1">
                      {t('extensions.type') || 'Type'}
                    </TypographyMuted>
                    <TypographySmall>
                      {preview?.metadata?.type || extension.extension_type}
                    </TypographySmall>
                  </CardWrapper>
                  <CardWrapper compact>
                    <TypographyMuted className="text-xs mb-1">
                      {t('extensions.category') || 'Category'}
                    </TypographyMuted>
                    <TypographySmall>
                      {preview?.metadata?.category || extension.category}
                    </TypographySmall>
                  </CardWrapper>
                </div>
                {preview?.variables && preview.variables.length > 0 && (
                  <DataTable
                    data={preview.variables}
                    columns={variableColumns}
                    showBorder={true}
                    striped={false}
                    rowClassName="border-0"
                  />
                )}
                {(preview?.execution?.run?.length || preview?.execution?.validate?.length) && (
                  <StepsSection
                    tRunLabel={t('extensions.runSteps') || 'Run steps'}
                    tValidateLabel={t('extensions.validateSteps') || 'Validate steps'}
                    title={t('extensions.execution') || 'Execution'}
                    runSteps={preview?.execution?.run || []}
                    validateSteps={preview?.execution?.validate || []}
                    openRunIndex={null}
                    setOpenRunIndex={() => {}}
                    openValidateIndex={null}
                    setOpenValidateIndex={() => {}}
                  />
                )}
              </div>
            </TabsContent>
          </Tabs>
        </div>
        <SheetFooter className="flex-row justify-end border-t gap-2">
          <Button variant="ghost" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button onClick={doFork} disabled={isLoading}>
            {t('extensions.fork') || 'Fork'}
          </Button>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  );
}

const MAX_DESCRIPTION_CHARS = 90;

export function ExtensionCard({
  extension,
  onInstall,
  onViewDetails,
  onFork,
  onForkClick,
  onRemove,
  setConfirmOpen,
  expanded,
  setExpanded,
  t,
  confirmOpen
}: ExtensionCardProps) {
  const cardActions = (
    <div className="flex items-center gap-1 shrink-0">
      {!extension.parent_extension_id && (
        <Button
          variant="ghost"
          size="icon"
          aria-label={t('extensions.fork') || 'Fork'}
          onClick={() => onForkClick?.(extension)}
        >
          <GitFork className="h-4 w-4" />
        </Button>
      )}
      {extension.parent_extension_id && (
        <Button
          variant="ghost"
          size="icon"
          aria-label={t('extensions.remove') || 'Remove'}
          onClick={() => setConfirmOpen(true)}
          className="text-destructive hover:text-destructive"
        >
          <Trash2 className="h-4 w-4" />
        </Button>
      )}
    </div>
  );

  const customHeader = (
    <div className="flex items-start gap-4 w-full">
      <div className="flex size-12 items-center justify-center rounded-full bg-muted shrink-0">
        <span className="text-lg font-bold text-muted-foreground">{extension.icon}</span>
      </div>
      <div className="flex-1 min-w-0">
        <h3 className="text-lg font-bold mb-1">{extension.name}</h3>
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">
            {t('extensions.madeBy')} {extension.author}
          </span>
          {extension.is_verified && (
            <div className="flex size-4 items-center justify-center rounded-full bg-primary">
              <Check className="size-2.5 text-primary-foreground" />
            </div>
          )}
        </div>
      </div>
      {cardActions}
    </div>
  );

  return (
    <>
      <CardWrapper
        className="h-full transition-shadow hover:shadow-lg"
        header={customHeader}
        contentClassName="space-y-4"
      >
        <CardDescription>
          {expanded || extension.description.length <= MAX_DESCRIPTION_CHARS
            ? extension.description
            : `${extension.description.slice(0, MAX_DESCRIPTION_CHARS)}…`}
          {extension.description.length > MAX_DESCRIPTION_CHARS && (
            <Button
              variant="link"
              className="ml-2 text-primary hover:underline"
              size="sm"
              onClick={() => setExpanded(!expanded)}
            >
              {expanded ? t('common.readLess') || 'Read less' : t('common.readMore') || 'Read more'}{' '}
              <ChevronDown className="size-4" />
            </Button>
          )}
        </CardDescription>
        <div className="flex gap-2">
          <Button onClick={() => onInstall?.(extension)} className="min-w-[100px]">
            {extension.extension_type === 'install' ? t('extensions.install') : t('extensions.run')}
          </Button>
          <Button
            variant="ghost"
            onClick={() => onViewDetails?.(extension)}
            className="min-w-[100px]"
          >
            {t('extensions.viewDetails')}
            <ExternalLink className="ml-2 h-4 w-4" />
          </Button>
        </div>
      </CardWrapper>
      <DeleteDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title={t('extensions.confirmDeleteTitle') || 'Remove fork?'}
        description={
          t('extensions.confirmDeleteMessage') ||
          'This will remove your forked extension. This action cannot be undone.'
        }
        confirmText={t('common.delete') || 'Delete'}
        cancelText={t('common.cancel') || 'Cancel'}
        variant="destructive"
        onConfirm={async () => {
          await onRemove?.(extension);
        }}
      />
    </>
  );
}

export function ExtensionInput({
  open,
  onOpenChange,
  extension,
  onSubmit,
  t,
  actions,
  isOnlyProxyDomain,
  noFieldsToShow,
  values,
  errors,
  handleChange,
  handleSubmit,
  requiredFields
}: ExtensionInputProps) {
  if (noFieldsToShow) {
    return (
      <DialogWrapper
        open={open}
        onOpenChange={onOpenChange}
        title={
          <div className="flex items-center gap-2">
            <Sparkles className="h-5 w-5 text-primary" />
            <span>{extension?.name || t('extensions.run')}</span>
          </div>
        }
        description={extension?.description}
        actions={actions}
        size="md"
      >
        <div className="py-2">
          <p className="text-sm text-muted-foreground">No additional fields required.</p>
        </div>
      </DialogWrapper>
    );
  }

  return (
    <DialogWrapper
      open={open}
      onOpenChange={onOpenChange}
      title={
        <div className="flex items-center gap-2">
          <Sparkles className="h-5 w-5 text-primary" />
          <span>{extension?.name || t('extensions.run')}</span>
        </div>
      }
      description={extension?.description}
      actions={actions}
      size={isOnlyProxyDomain ? 'md' : 'lg'}
    >
      <div className="py-2 space-y-4">
        {requiredFields.map((variable) => {
          const isProxyDomain =
            variable.variable_name.toLowerCase() === 'proxy_domain' ||
            variable.variable_name.toLowerCase() === 'domain';

          if (isProxyDomain) {
            return (
              <ProxyDomainInput
                key={variable.id}
                variable={variable}
                value={values[variable.variable_name]}
                error={errors[variable.variable_name]}
                onChange={handleChange}
              />
            );
          }

          return (
            <VariableInput
              key={variable.id}
              variable={variable}
              value={values[variable.variable_name]}
              error={errors[variable.variable_name]}
              onChange={handleChange}
            />
          );
        })}
      </div>
    </DialogWrapper>
  );
}

function ProxyDomainInput({
  variable,
  value,
  error,
  onChange
}: {
  variable: ExtensionVariable;
  value: unknown;
  error?: string;
  onChange: (name: string, value: unknown) => void;
}) {
  const id = `var-${variable.variable_name}`;
  return (
    <div className="space-y-2">
      <Label htmlFor={id} className="text-sm font-medium flex items-center gap-2">
        <Globe className="h-4 w-4 text-muted-foreground" />
        {variable.variable_name}
        {variable.is_required && <span className="text-destructive">*</span>}
      </Label>
      <Input
        id={id}
        type="text"
        value={(value as string) ?? ''}
        onChange={(e) => onChange(variable.variable_name, e.target.value)}
        placeholder={variable.description || 'app.example.com'}
        className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
        autoFocus
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
      {variable.description && (
        <p className="text-xs text-muted-foreground">{variable.description}</p>
      )}
    </div>
  );
}

function VariableInput({
  variable,
  value,
  error,
  onChange
}: {
  variable: ExtensionVariable;
  value: unknown;
  error?: string;
  onChange: (name: string, value: unknown) => void;
}) {
  const id = `var-${variable.variable_name}`;

  const renderInput = () => {
    switch (variable.variable_type) {
      case 'integer':
        return (
          <Input
            id={id}
            type="number"
            value={(value as number) ?? ''}
            onChange={(e) => {
              const num = e.target.value === '' ? '' : Number(e.target.value);
              onChange(variable.variable_name, num);
            }}
            placeholder={variable.description || String(variable.default_value ?? '')}
            className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
          />
        );
      case 'array':
        const arrayValue = Array.isArray(value) ? value.join(', ') : ((value as string) ?? '');
        return (
          <Input
            id={id}
            type="text"
            value={arrayValue}
            onChange={(e) => {
              const arr = e.target.value
                .split(',')
                .map((x) => x.trim())
                .filter((x) => x.length > 0);
              onChange(variable.variable_name, arr);
            }}
            placeholder={variable.description || 'comma-separated values'}
            className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
          />
        );
      case 'boolean':
        const checked = value === true || value === 'true' || value === 1;
        return (
          <div className="flex items-center gap-2">
            <Checkbox
              id={id}
              checked={checked}
              onCheckedChange={(checked) => onChange(variable.variable_name, checked)}
            />
            <Label htmlFor={id} className="text-sm font-normal cursor-pointer">
              {variable.description || 'Enable'}
            </Label>
          </div>
        );
      default:
        // string or other types
        const isLongText = variable.description && variable.description.length > 100;
        if (isLongText) {
          return (
            <Textarea
              id={id}
              value={(value as string) ?? ''}
              onChange={(e) => onChange(variable.variable_name, e.target.value)}
              placeholder={variable.description || ''}
              className={cn(
                'min-h-[100px]',
                error && 'border-destructive focus-visible:ring-destructive'
              )}
            />
          );
        }
        return (
          <Input
            id={id}
            type="text"
            value={(value as string) ?? ''}
            onChange={(e) => onChange(variable.variable_name, e.target.value)}
            placeholder={variable.description || String(variable.default_value ?? '')}
            className={cn('h-10', error && 'border-destructive focus-visible:ring-destructive')}
          />
        );
    }
  };

  return (
    <div className="space-y-2">
      <Label htmlFor={id} className="text-sm font-medium">
        {variable.variable_name}
        {variable.is_required && <span className="text-destructive ml-1">*</span>}
        {variable.variable_type && (
          <span className="text-xs text-muted-foreground ml-2 font-normal">
            ({variable.variable_type})
          </span>
        )}
      </Label>
      {renderInput()}
      {error && <p className="text-xs text-destructive">{error}</p>}
      {variable.description && variable.variable_type !== 'boolean' && (
        <p className="text-xs text-muted-foreground">{variable.description}</p>
      )}
    </div>
  );
}
