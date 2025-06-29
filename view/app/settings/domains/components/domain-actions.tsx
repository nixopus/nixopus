import React from 'react';
import { Button } from '@/components/ui/button';
import UpdateDomainDialog from './update-domain';
import { Domain } from '@/redux/types/domain';
import DeleteDomain from './delete-domain';
import { useTranslation } from '@/hooks/use-translation';
import { ResourceGuard } from '@/components/rbac/PermissionGuard';

interface DomainActionsProps {
  domain: Domain;
}

export function DomainActions({ domain }: DomainActionsProps) {
  const { t } = useTranslation();
  const [isEditModalOpen, setIsEditModalOpen] = React.useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = React.useState(false);

  const handleEdit = () => {
    setIsEditModalOpen(true);
  };

  const handleDelete = () => {
    setIsDeleteModalOpen(true);
  };

  return (
    <div className="flex justify-end gap-2">
      <ResourceGuard resource="domain" action="update">
        <Button variant="ghost" className="text-primary p-0 m-0" onClick={handleEdit}>
          {t('settings.domains.actions.edit')}
        </Button>
      </ResourceGuard>
      <ResourceGuard resource="domain" action="delete">
        <Button variant="ghost" className="text-red-500 p-0 m-0" onClick={handleDelete}>
          {t('settings.domains.actions.delete')}
        </Button>
      </ResourceGuard>
      {isEditModalOpen && (
        <UpdateDomainDialog
          open={isEditModalOpen}
          setOpen={setIsEditModalOpen}
          id={domain.id}
          data={domain}
        />
      )}
      {isDeleteModalOpen && (
        <DeleteDomain open={isDeleteModalOpen} setOpen={setIsDeleteModalOpen} id={domain.id} />
      )}
    </div>
  );
}
