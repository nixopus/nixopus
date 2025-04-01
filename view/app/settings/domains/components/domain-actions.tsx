import React from 'react';
import { Button } from '@/components/ui/button';
import UpdateDomainDialog from './update-domain';
import { Domain } from '@/redux/types/domain';
import DeleteDomain from './delete-domain';
import { useAppSelector } from '@/redux/hooks';
import { useResourcePermissions } from '@/lib/permission';
import { useTranslation } from '@/hooks/use-translation';

interface DomainActionsProps {
  domain: Domain;
}

export function DomainActions({ domain }: DomainActionsProps) {
  const { t } = useTranslation();
  const [isEditModalOpen, setIsEditModalOpen] = React.useState(false);
  const [isDeleteModalOpen, setIsDeleteModalOpen] = React.useState(false);
  const user = useAppSelector((state) => state.auth.user);
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const { canUpdate, canDelete } = useResourcePermissions(
    user,
    'organization',
    activeOrganization?.id
  );

  const handleEdit = () => {
    setIsEditModalOpen(true);
  };

  const handleDelete = () => {
    setIsDeleteModalOpen(true);
  };

  return (
    <div className="flex justify-end gap-2">
      {canUpdate && (
        <Button variant="ghost" className="text-primary p-0 m-0" onClick={handleEdit}>
          {t('settings.domains.actions.edit')}
        </Button>
      )}
      {canDelete && (
        <Button variant="ghost" className="text-red-500 p-0 m-0" onClick={handleDelete}>
          {t('settings.domains.actions.delete')}
        </Button>
      )}
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
