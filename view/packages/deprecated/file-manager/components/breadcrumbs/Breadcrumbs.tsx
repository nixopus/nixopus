import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator
} from '@/components/ui/breadcrumb';
import { HomeIcon } from 'lucide-react';
import React from 'react';

export interface BreadcrumbsProps {
  breadcrumbs: string[];
  fileClicked: (filePath: string) => void;
}

export const Breadcrumbs = ({ breadcrumbs, fileClicked }: BreadcrumbsProps) => {
  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem className="cursor-pointer" onClick={() => fileClicked('/')}>
          <BreadcrumbPage>
            <HomeIcon className="h-5 w-5" />
          </BreadcrumbPage>
        </BreadcrumbItem>
        {breadcrumbs.map((crumb: string, index: number) => (
          <React.Fragment key={index}>
            <BreadcrumbItem
              onClick={() => fileClicked('/' + breadcrumbs.slice(0, index + 1).join('/'))}
              className="cursor-pointer"
            >
              {index === breadcrumbs.length - 1 ? (
                <BreadcrumbPage>{crumb}</BreadcrumbPage>
              ) : (
                <BreadcrumbLink asChild>
                  <h1>{crumb}</h1>
                </BreadcrumbLink>
              )}
            </BreadcrumbItem>
            {index < breadcrumbs.length - 1 && <BreadcrumbSeparator />}
          </React.Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  );
};
