'use client';

import React from 'react';
import { ExternalLink, Globe } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';
import { ApplicationDomain } from '@/redux/types/applications';
import { cn } from '@/lib/utils';
import { Controller } from 'react-hook-form';
import {
  FormItem,
  FormLabel,
  FormControl,
  FormDescription,
  FormMessage
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { X, Plus } from 'lucide-react';
import { useDomainDropdown } from '@/packages/hooks/applications/use_multiple_domains';
import {
  useMultipleDomains,
  defaultValidator
} from '@/packages/hooks/applications/use_multiple_domains';

interface DropdownWrapperProps {
  trigger: React.ReactNode;
  items: Array<{ value: string; onClick: (e?: React.MouseEvent) => void }>;
  align?: 'start' | 'end' | 'center';
  onContentClick?: (e: React.MouseEvent) => void;
}

function DropdownWrapper({ trigger, items, align = 'end', onContentClick }: DropdownWrapperProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>{trigger}</DropdownMenuTrigger>
      <DropdownMenuContent align={align} onClick={onContentClick}>
        {items.map((item, index) => (
          <DropdownMenuItem
            key={index}
            onClick={item.onClick}
            className="flex items-center gap-2 cursor-pointer"
          >
            <ExternalLink className="h-4 w-4" />
            <span className="font-mono text-sm">{item.value}</span>
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

interface DomainInputRowProps {
  domain: string;
  placeholder: string;
  actualIndex: number;
  isLastVisible: boolean;
  isLastInAll: boolean;
  canRemove: boolean;
  canAdd: boolean;
  hasError: boolean;
  onUpdate: (index: number, value: string) => void;
  onRemove: (index: number) => void;
  onAdd: () => void;
  onBlur: () => void;
}

const DomainInputRow = React.memo<DomainInputRowProps>(function DomainInputRow({
  domain,
  placeholder,
  actualIndex,
  isLastVisible,
  isLastInAll,
  canRemove,
  canAdd,
  hasError,
  onUpdate,
  onRemove,
  onAdd,
  onBlur
}) {
  return (
    <div className="flex gap-2">
      <FormControl>
        <Input
          placeholder={placeholder}
          value={domain}
          onChange={(e) => onUpdate(actualIndex, e.target.value)}
          onBlur={onBlur}
          className={cn('flex-1', hasError && 'border-destructive')}
        />
      </FormControl>
      {canRemove && (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={() => onRemove(actualIndex)}
          className="flex-shrink-0"
          aria-label="Remove domain"
        >
          <X className="h-4 w-4" />
        </Button>
      )}
      {isLastVisible && isLastInAll && canAdd && (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          onClick={onAdd}
          className="flex-shrink-0"
          aria-label="Add domain"
        >
          <Plus className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
});

interface DomainDropdownProps {
  domains?: ApplicationDomain[];
  variant?: 'icon' | 'button';
  className?: string;
  onDomainClick?: (domain: string) => void;
}

export function DomainDropdown({
  domains,
  variant = 'icon',
  className,
  onDomainClick
}: DomainDropdownProps) {
  const { allDomains, handleDomainClick } = useDomainDropdown({ domains, onDomainClick });

  if (allDomains.length === 0) {
    return null;
  }

  const trigger = (
    <Button
      variant={variant === 'icon' ? 'ghost' : 'outline'}
      size={variant === 'icon' ? 'icon' : 'sm'}
      className={cn(
        variant === 'icon' ? 'h-8 w-8 text-muted-foreground hover:text-foreground' : 'text-sm',
        className
      )}
      onClick={(e) => e.stopPropagation()}
    >
      {variant === 'icon' ? (
        <ExternalLink className="h-4 w-4" />
      ) : (
        <span className="flex items-center gap-1">
          <Globe className="h-3 w-3" />
          {allDomains.length > 1 ? `${allDomains.length} domains` : allDomains[0]}
        </span>
      )}
    </Button>
  );

  const dropdownItems = allDomains.map((domainValue) => ({
    value: domainValue,
    onClick: (e?: React.MouseEvent) => handleDomainClick(domainValue, e)
  }));

  return (
    <DropdownWrapper
      trigger={trigger}
      items={dropdownItems}
      align="end"
      onContentClick={(e) => e.stopPropagation()}
    />
  );
}

interface MultipleDomainInputProps {
  form: any;
  label: string;
  name: string;
  description?: string | React.ReactNode;
  placeholder?: string;
  required?: boolean;
  maxDomains?: number;
  validator?: (value: string) => boolean;
}

export function MultipleDomainInput({
  form,
  label,
  name,
  description,
  placeholder = 'example.com',
  required = false,
  maxDomains = 5,
  validator = defaultValidator
}: MultipleDomainInputProps) {
  const {
    domains,
    showAll,
    setShowAll,
    addDomain,
    removeDomain,
    updateDomain,
    visibleDomains,
    remainingCount,
    startIndex,
    validationRules,
    maxVisible
  } = useMultipleDomains({
    form,
    name,
    maxDomains,
    validator
  });

  return (
    <Controller
      control={form.control}
      name={name as any}
      rules={validationRules as any}
      render={({ field, fieldState }: { field: any; fieldState: any }) => (
        <FormItem>
          <div className="flex gap-2">
            {label && <FormLabel>{label}</FormLabel>}
            <span className="text-destructive w-3 flex-shrink-0 text-right">
              {required ? '*' : ''}
            </span>
          </div>
          <div className="space-y-2">
            {visibleDomains.map((domain: string, visibleIndex: number) => {
              const actualIndex = startIndex + visibleIndex;
              const isLastVisible = visibleIndex === visibleDomains.length - 1;
              const isLastInAll = actualIndex === domains.length - 1;

              return (
                <DomainInputRow
                  key={actualIndex}
                  domain={domain}
                  placeholder={placeholder}
                  actualIndex={actualIndex}
                  isLastVisible={isLastVisible}
                  isLastInAll={isLastInAll}
                  canRemove={domains.length > 1}
                  canAdd={domains.length < maxDomains}
                  hasError={!!fieldState.error}
                  onUpdate={updateDomain}
                  onRemove={removeDomain}
                  onAdd={addDomain}
                  onBlur={field.onBlur}
                />
              );
            })}
            {!showAll && remainingCount > 0 && (
              <div className="flex items-center gap-2 text-sm text-muted-foreground py-1">
                <span>and {remainingCount} more</span>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowAll(true)}
                  className="h-auto p-0 text-primary hover:text-primary/80 underline"
                >
                  Show all
                </Button>
              </div>
            )}
            {showAll && domains.length > maxVisible && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setShowAll(false)}
                className="text-sm text-muted-foreground hover:text-foreground"
              >
                Show less
              </Button>
            )}
          </div>
          {description && <FormDescription>{description}</FormDescription>}
          <FormMessage />
        </FormItem>
      )}
    />
  );
}
