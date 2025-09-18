import * as React from 'react';

import { cn } from '@/lib/utils';
import { Input } from '@/components/ui/input';
import { Eye, EyeOff } from 'lucide-react';

export interface PasswordInputFieldProps extends React.ComponentProps<'input'> {
  containerClassName?: string;
}

const PasswordInputField = React.forwardRef<HTMLInputElement, PasswordInputFieldProps>(
  function PasswordInputField({ className, containerClassName, autoComplete, ...props }, ref) {
    const [showPassword, setShowPassword] = React.useState(false);

    return (
      <div className={cn('relative', containerClassName)}>
        <Input
          ref={ref}
          {...props}
          type={showPassword ? 'text' : 'password'}
          className={cn('pr-10', className)}
          autoComplete={autoComplete ?? 'current-password'}
        />
        <button
          type="button"
          onClick={() => setShowPassword((v) => !v)}
          className="text-muted-foreground hover:text-foreground absolute inset-y-0 right-2 my-auto inline-flex h-6 w-6 items-center justify-center"
          aria-label={showPassword ? 'Hide password' : 'Show password'}
          aria-pressed={showPassword}
          tabIndex={0}
        >
          {showPassword ? (
            <EyeOff className="h-5 w-5" aria-hidden="true" />
          ) : (
            <Eye className="h-5 w-5" aria-hidden="true" />
          )}
        </button>
      </div>
    );
  }
);

export { PasswordInputField };


