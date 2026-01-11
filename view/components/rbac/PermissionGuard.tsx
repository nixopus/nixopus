import React from 'react';
import { RBACGuard } from './RBACGuard';
import { Action, Permission, Resource } from '@/lib/rbac';

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
