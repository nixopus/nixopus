'use client';

import { Skeleton } from '@/components/ui/skeleton';
import { Badge } from '@/components/ui/badge';
import { Extension } from '@/redux/types/extension';
import { useTranslation } from '@/hooks/use-translation';
import { useMemo, useState } from 'react';
import YAML from 'yaml';
import { ChevronDown } from 'lucide-react';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';

interface OverviewTabProps {
  extension?: Extension;
  isLoading?: boolean;
}

export default function OverviewTab({ extension, isLoading }: OverviewTabProps) {
  const { t } = useTranslation();
  const [openRunIndex, setOpenRunIndex] = useState<number | null>(null);
  const [openValidateIndex, setOpenValidateIndex] = useState<number | null>(null);
  const parsed = useMemo(() => {
    try {
      if (!extension?.yaml_content) return undefined;
      const y = YAML.parse(extension.yaml_content || '');
      return {
        execution: y?.execution || {}
      } as any;
    } catch {
      return undefined;
    }
  }, [extension?.yaml_content]);

  if (isLoading) {
    return <Skeleton className="h-40 w-full" />;
  }

  return (
    <div className="space-y-6">
      <MetadataSection extension={extension} />
      {extension?.variables && extension.variables.length > 0 && (
        <VariablesTable variables={extension.variables} />
      )}
      {(parsed?.execution?.run?.length || parsed?.execution?.validate?.length) && (
        <StepsSection
          tRunLabel={t('extensions.runSteps') || 'Run steps'}
          tValidateLabel={t('extensions.validateSteps') || 'Validate steps'}
          title={t('extensions.execution') || 'Execution'}
          runSteps={parsed?.execution?.run || []}
          validateSteps={parsed?.execution?.validate || []}
          openRunIndex={openRunIndex}
          setOpenRunIndex={setOpenRunIndex}
          openValidateIndex={openValidateIndex}
          setOpenValidateIndex={setOpenValidateIndex}
        />
      )}
    </div>
  );
}

function MetadataSection({ extension }: { extension?: Extension }) {
  return (
    <div className="space-y-2">
      <div className="text-sm text-muted-foreground">{extension?.description}</div>
      <div className="flex flex-wrap gap-2 pt-1">
        {extension?.category && <Badge variant="secondary">{extension.category}</Badge>}
        {extension?.extension_type && <Badge variant="outline">{extension.extension_type}</Badge>}
        {extension?.version && <Badge>v{extension.version}</Badge>}
        {extension?.is_verified && <Badge>Verified</Badge>}
      </div>
    </div>
  );
}

function VariablesTable({ variables }: { variables: NonNullable<Extension['variables']> }) {
  return (
    <div className="rounded-md border overflow-hidden">
      <div className="grid grid-cols-12 bg-muted/50 px-3 py-2 text-xs font-medium text-muted-foreground">
        <div className="col-span-3">Name</div>
        <div className="col-span-2">Type</div>
        <div className="col-span-2">Required</div>
        <div className="col-span-2">Default</div>
        <div className="col-span-3">Description</div>
      </div>
      <div className="divide-y">
        {variables.map((v) => (
          <div key={v.id} className="grid grid-cols-12 px-3 py-3 text-sm">
            <div className="col-span-3 font-medium">{v.variable_name}</div>
            <div className="col-span-2 text-muted-foreground">{v.variable_type}</div>
            <div className="col-span-2 text-muted-foreground">{v.is_required ? 'Yes' : 'No'}</div>
            <div className="col-span-2 text-muted-foreground truncate">
              {String(v.default_value ?? '')}
            </div>
            <div className="col-span-3 text-muted-foreground">{v.description}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

export function StepsSection({
  title,
  tRunLabel,
  tValidateLabel,
  runSteps,
  validateSteps,
  openRunIndex,
  setOpenRunIndex,
  openValidateIndex,
  setOpenValidateIndex
}: {
  title: string;
  tRunLabel: string;
  tValidateLabel: string;
  runSteps: any[];
  validateSteps: any[];
  openRunIndex: number | null;
  setOpenRunIndex: (i: number | null) => void;
  openValidateIndex: number | null;
  setOpenValidateIndex: (i: number | null) => void;
}) {
  const hasRun = runSteps.length > 0;
  const hasValidate = validateSteps.length > 0;
  return (
    <div className="space-y-3">
      <div className="text-sm font-semibold">{title}</div>
      <div className="flex flex-col gap-4">
        {hasRun && (
          <StepList
            label={tRunLabel}
            steps={runSteps}
            openIndex={openRunIndex}
            onToggle={(i) => setOpenRunIndex(i)}
          />
        )}
        {hasValidate && (
          <StepList
            label={tValidateLabel}
            steps={validateSteps}
            openIndex={openValidateIndex}
            onToggle={(i) => setOpenValidateIndex(i)}
          />
        )}
      </div>
    </div>
  );
}

function StepList({
  label,
  steps,
  openIndex,
  onToggle
}: {
  label: string;
  steps: any[];
  openIndex: number | null;
  onToggle: (i: number | null) => void;
}) {
  return (
    <div className="rounded-md border overflow-hidden">
      <div className="bg-muted/50 px-3 py-2 text-xs font-medium text-muted-foreground">{label}</div>
      <ol className="divide-y">
        {steps.map((s: any, i: number) => {
          const entries = Object.entries(s || {}).filter(([k]) => k !== 'name' && k !== 'type');
          const isOpen = openIndex === i;
          return (
            <li key={i} className="px-3 py-3 text-sm">
              <Collapsible open={isOpen} onOpenChange={(open) => onToggle(open ? i : null)}>
                <CollapsibleTrigger className="w-full text-left flex items-start gap-2 group">
                  <span className="text-muted-foreground mr-1 mt-0.5">{i + 1}.</span>
                  <div className="flex-1">
                    <div className="font-medium">{s?.name || s?.type || 'Step'}</div>
                    {s?.type && (
                      <div className="mt-1">
                        <Badge variant="outline">{s.type}</Badge>
                      </div>
                    )}
                  </div>
                  <ChevronDown
                    className={`ml-2 h-4 w-4 transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
                  />
                </CollapsibleTrigger>
                {entries.length > 0 && (
                  <CollapsibleContent>
                    <div className="mt-3 rounded-md border bg-muted/20">
                      <div className="grid grid-cols-12 gap-2 p-3 text-sm">
                        {entries.map(([k, v]) => (
                          <div key={k} className="contents">
                            <div className="col-span-3 text-muted-foreground">{k}</div>
                            <div className="col-span-9 break-words">
                              {typeof v === 'object' ? (
                                <pre className="whitespace-pre-wrap text-xs">
                                  {JSON.stringify(v, null, 2)}
                                </pre>
                              ) : (
                                String(v)
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  </CollapsibleContent>
                )}
              </Collapsible>
            </li>
          );
        })}
      </ol>
    </div>
  );
}
