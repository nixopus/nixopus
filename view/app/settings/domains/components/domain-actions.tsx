import React from 'react';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { Edit, Trash2 } from 'lucide-react';
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
        <Tooltip>
          <TooltipTrigger asChild>
            <Button 
              variant="ghost" 
              size="sm"
              className="h-8 w-8 p-0 text-primary hover:text-primary/80 hover:bg-primary/10" 
              onClick={handleEdit}
            >
              <Edit className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{t('settings.domains.actions.edit')}</p>
          </TooltipContent>
        </Tooltip>
      </ResourceGuard>
      <ResourceGuard resource="domain" action="delete">
        <Tooltip>
          <TooltipTrigger asChild>
            <Button 
              variant="ghost" 
              size="sm"
              className="h-8 w-8 p-0 text-red-500 hover:text-red-600 hover:bg-red-50" 
              onClick={handleDelete}
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>{t('settings.domains.actions.delete')}</p>
          </TooltipContent>
        </Tooltip>
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
