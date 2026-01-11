import * as React from 'react';
import { LucideIcon } from 'lucide-react';

import { cn } from '@/lib/utils';
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle
} from '@/components/ui/card';
import { TypographySmall } from '@/components/ui/typography';

export interface CardWrapperProps extends React.ComponentProps<typeof Card> {
  /**
   * The title of the card
   */
  title?: string;
  /**
   * Optional description text displayed below the title
   */
  description?: string;
  /**
   * Optional icon to display next to the title
   */
  icon?: LucideIcon;
  /**
   * Optional actions/buttons to display in the header (right side)
   */
  actions?: React.ReactNode;
  /**
   * Optional footer content
   */
  footer?: React.ReactNode;
  /**
   * The main content of the card
   */
  children: React.ReactNode;
  /**
   * Custom header content (overrides title/icon/actions if provided)
   */
  header?: React.ReactNode;
  /**
   * Custom className for the header
   */
  headerClassName?: string;
  /**
   * Custom className for the content
   */
  contentClassName?: string;
  /**
   * Custom className for the footer
   */
  footerClassName?: string;
  /**
   * Whether to show the title with smaller text (for compact cards)
   * @default false
   */
  compact?: boolean;
}

function CardWrapper({
  className,
  title,
  description,
  icon: Icon,
  actions,
  footer,
  children,
  header,
  headerClassName,
  contentClassName,
  footerClassName,
  compact = false,
  ...props
}: CardWrapperProps) {
  const hasHeader = title || Icon || actions || header;

  return (
    <Card className={cn('flex flex-col', className)} {...props}>
      {hasHeader && (
        <CardHeader
          data-slot="card-wrapper-header"
          className={cn(
            'flex flex-row items-center justify-between gap-2',
            compact && 'pb-2',
            headerClassName
          )}
        >
          {header ? (
            header
          ) : (
            <>
              {(title || Icon) && (
                <CardTitle
                  data-slot="card-wrapper-title"
                  className={cn(
                    'flex items-center gap-1 sm:gap-2',
                    compact ? 'text-xs sm:text-sm' : 'text-sm',
                    compact ? 'font-medium' : 'font-bold'
                  )}
                >
                  {Icon && (
                    <Icon
                      className={cn(
                        'text-muted-foreground shrink-0',
                        compact ? 'h-3 w-3 sm:h-4 sm:w-4' : 'h-4 w-4'
                      )}
                    />
                  )}
                  {title && (
                    <TypographySmall className={compact ? 'text-xs sm:text-sm' : ''}>
                      {title}
                    </TypographySmall>
                  )}
                </CardTitle>
              )}
              {description && (
                <CardDescription data-slot="card-wrapper-description" className="mt-1">
                  {description}
                </CardDescription>
              )}
              {actions && (
                <div
                  data-slot="card-wrapper-actions"
                  className="flex items-center gap-2 shrink-0 ml-auto"
                >
                  {actions}
                </div>
              )}
            </>
          )}
        </CardHeader>
      )}
      <CardContent data-slot="card-wrapper-content" className={cn('flex-1', contentClassName)}>
        {children}
      </CardContent>
      {footer && (
        <CardFooter data-slot="card-wrapper-footer" className={cn(footerClassName)}>
          {footer}
        </CardFooter>
      )}
    </Card>
  );
}

export { CardWrapper };
