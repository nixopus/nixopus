import React from 'react';
import { Button } from '@/components/ui/button';
import UpdateDomainDialog from './update-domain';
import { Domain } from '@/redux/types/domain';
import DeleteDomain from './delete-domain';

interface DomainActionsProps {
  domain: Domain;
}

export function DomainActions({ domain }: DomainActionsProps) {
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
      <Button variant="ghost" className="text-primary p-0 m-0" onClick={handleEdit}>
        Edit
      </Button>
      <Button variant="ghost" className="text-red-500 p-0 m-0" onClick={handleDelete}>
        Delete
      </Button>
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
