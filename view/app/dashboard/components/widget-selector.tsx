'use client';

import React from 'react';
import { Plus } from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu';
import { Button } from '@/components/ui/button';

interface AvailableWidget {
  id: string;
  label: string;
}

interface WidgetSelectorProps {
  availableWidgets: AvailableWidget[];
  onAddWidget: (widgetId: string) => void;
}

export function WidgetSelector({ availableWidgets, onAddWidget }: WidgetSelectorProps) {
  if (availableWidgets.length === 0) {
    return null;
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="outline" size="sm" className="gap-2">
          <Plus className="h-4 w-4" />
          Add Widget
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-48">
        {availableWidgets.map((widget) => (
          <DropdownMenuItem
            key={widget.id}
            onClick={() => onAddWidget(widget.id)}
            className="cursor-pointer"
          >
            {widget.label}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
