import React from 'react';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { AlertTriangle } from 'lucide-react';
import { Action, Permission, Resource } from '@/packages/utils/rbac';
import { useRBAC } from '@/packages/utils/rbac';
import { AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';

interface AccessDeniedProps {
  title?: string;
  description?: string;
  className?: string;
}

export const AccessDenied: React.FC<AccessDeniedProps> = ({
  title,
  description,
  className = ''
}) => {
  const { t } = useTranslation();

  return (
    <div className={`flex h-full items-center justify-center ${className}`}>
      <div className="text-center">
        <AlertTriangle className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
        <h2 className="text-2xl font-bold">{title || t('common.accessDenied')}</h2>
        <p className="text-muted-foreground">{description || t('common.noPermissionView')}</p>
      </div>
    </div>
  );
};

interface BasePermissionGuardProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
  loadingFallback?: React.ReactNode;
  errorFallback?: React.ReactNode;
}

interface SinglePermissionGuardProps extends BasePermissionGuardProps {
  permission: Permission;
}

interface MultiplePermissionGuardProps extends BasePermissionGuardProps {
  permissions: Permission[];
}

interface ResourceGuardProps extends BasePermissionGuardProps {
  resource: Resource;
  action: Action;
}

export const PermissionGuard: React.FC<SinglePermissionGuardProps> = ({
  children,
  permission,
  fallback,
  loadingFallback,
  errorFallback
}) => (
  <RBACGuard
    permission={permission}
    fallback={fallback}
    loadingFallback={loadingFallback}
    errorFallback={errorFallback}
  >
    {children}
  </RBACGuard>
);

export const AnyPermissionGuard: React.FC<MultiplePermissionGuardProps> = ({
  children,
  permissions,
  fallback,
  loadingFallback,
  errorFallback
}) => (
  <RBACGuard
    permissions={permissions}
    requireAll={false}
    fallback={fallback}
    loadingFallback={loadingFallback}
    errorFallback={errorFallback}
  >
    {children}
  </RBACGuard>
);

export const AllPermissionsGuard: React.FC<MultiplePermissionGuardProps> = ({
  children,
  permissions,
  fallback,
  loadingFallback,
  errorFallback
}) => (
  <RBACGuard
    permissions={permissions}
    requireAll={true}
    fallback={fallback}
    loadingFallback={loadingFallback}
    errorFallback={errorFallback}
  >
    {children}
  </RBACGuard>
);

export const ResourceGuard: React.FC<ResourceGuardProps> = ({
  children,
  resource,
  action,
  fallback,
  loadingFallback,
  errorFallback
}) => (
  <RBACGuard
    resource={resource}
    action={action}
    fallback={fallback}
    loadingFallback={loadingFallback}
    errorFallback={errorFallback}
  >
    {children}
  </RBACGuard>
);

interface RBACGuardProps {
  children: React.ReactNode;
  resource?: Resource;
  action?: Action;
  permission?: Permission;
  permissions?: Permission[];
  requireAll?: boolean;
  fallback?: React.ReactNode;
  loadingFallback?: React.ReactNode;
  errorFallback?: React.ReactNode;
}

export const RBACGuard: React.FC<RBACGuardProps> = ({
  children,
  resource,
  action,
  permission,
  permissions,
  requireAll = false,
  fallback = null,
  loadingFallback = null,
  errorFallback = null
}) => {
  const { canAccessResource, hasPermission, hasAnyPermission, hasAllPermissions, isLoading } =
    useRBAC();

  if (isLoading) {
    return <>{loadingFallback}</>;
  }

  let shouldRender = false;

  try {
    if (resource && action) {
      shouldRender = canAccessResource(resource, action);
    } else if (permission) {
      shouldRender = hasPermission(permission);
    } else if (permissions && permissions.length > 0) {
      shouldRender = requireAll ? hasAllPermissions(permissions) : hasAnyPermission(permissions);
    } else {
      shouldRender = true;
    }
  } catch (error) {
    console.error('Permission check failed:', error);
    return <>{errorFallback}</>;
  }

  return shouldRender ? <>{children}</> : <>{fallback}</>;
};

function DisabledFeature() {
  const { t } = useTranslation();
  return (
    <div className="flex h-[calc(100vh-200px)] items-center justify-center p-4">
      <Card className="w-full max-w-md p-6 text-center border-none">
        <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-muted">
          <AlertCircle className="h-6 w-6 text-muted-foreground" />
        </div>
        <h2 className="mb-2 text-2xl font-semibold tracking-tight">
          {t('common.featureDisabled')}
        </h2>
        <p className="mb-6 text-sm text-muted-foreground">{t('common.featureNotAvailable')}</p>
        <div className="space-y-3">
          <Button variant="outline" className="w-full" onClick={() => window.history.back()}>
            {t('common.goBack')}
          </Button>
          <Button variant="ghost" className="w-full" onClick={() => window.location.reload()}>
            {t('common.refreshPage')}
          </Button>
        </div>
      </Card>
    </div>
  );
}

export default DisabledFeature;
