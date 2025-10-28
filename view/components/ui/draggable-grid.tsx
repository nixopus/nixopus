'use client';

import React, { useState, useEffect } from 'react';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  DragEndEvent,
  DragOverlay,
  DragStartEvent
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
  useSortable
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical, X } from 'lucide-react';
import { cn } from '@/lib/utils';

export interface DraggableItem {
  id: string;
  component: React.ReactNode;
  className?: string;
}

interface DraggableGridProps {
  items: DraggableItem[];
  onReorder?: (items: DraggableItem[]) => void;
  onDelete?: (itemId: string) => void;
  storageKey?: string;
  className?: string;
  gridCols?: string; // e.g., "grid-cols-1 md:grid-cols-2"
  resetKey?: number; // Change this to trigger a reset
}

function SortableItem({
  item,
  isDragging,
  onDelete
}: {
  item: DraggableItem;
  isDragging?: boolean;
  onDelete?: (id: string) => void;
}) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isItemDragging
  } = useSortable({ id: item.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'relative group h-full',
        item.className,
        isItemDragging && 'opacity-50 z-50',
        'transition-all duration-200'
      )}
    >
      <div
        {...listeners}
        {...attributes}
        className="absolute left-0 top-1/2 -translate-y-1/2 -translate-x-3 opacity-0 group-hover:opacity-100 transition-opacity cursor-grab active:cursor-grabbing z-10 touch-none"
        title="Drag to reorder"
      >
        <div className="bg-primary/10 hover:bg-primary/20 rounded-lg p-2 backdrop-blur-sm border border-primary/20 shadow-sm">
          <GripVertical className="h-4 w-4 text-primary" />
        </div>
      </div>
      {onDelete && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onDelete(item.id);
          }}
          className="absolute right-0 top-1/2 -translate-y-1/2 translate-x-3 opacity-0 group-hover:opacity-100 transition-opacity z-10"
          title="Remove widget"
        >
          <div className="bg-destructive/10 hover:bg-destructive/20 rounded-lg p-2 backdrop-blur-sm border border-destructive/20 shadow-sm">
            <X className="h-4 w-4 text-destructive" />
          </div>
        </button>
      )}
      <div
        className={cn(
          'transition-all duration-200 h-full',
          isItemDragging && 'ring-2 ring-primary/50 rounded-xl'
        )}
      >
        {item.component}
      </div>
    </div>
  );
}

export function DraggableGrid({
  items,
  onReorder,
  onDelete,
  storageKey = 'dashboard-layout',
  className,
  gridCols = 'grid-cols-1',
  resetKey = 0
}: DraggableGridProps) {
  const [orderedItems, setOrderedItems] = useState<DraggableItem[]>(items);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8
      }
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates
    })
  );

  // Load saved order from localStorage only after mount
  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!mounted) return;

    if (typeof window !== 'undefined' && storageKey) {
      const savedOrder = localStorage.getItem(storageKey);
      if (savedOrder) {
        try {
          const orderIds = JSON.parse(savedOrder) as string[];
          const reordered = orderIds
            .map((id) => items.find((item) => item.id === id))
            .filter(Boolean) as DraggableItem[];

          const newItems = items.filter((item) => !orderIds.includes(item.id));

          if (reordered.length > 0) {
            setOrderedItems([...reordered, ...newItems]);
          } else {
            setOrderedItems(items);
          }
        } catch (e) {
          console.error('Failed to load saved order:', e);
          setOrderedItems(items);
        }
      } else {
        // if no saved order, use default
        setOrderedItems(items);
      }
    }
  }, [mounted, storageKey, resetKey, items]);

  // Update items when they change externally (but keep order if saved)
  useEffect(() => {
    if (!mounted) return;

    if (orderedItems.length !== items.length) {
      const savedOrder = localStorage.getItem(storageKey);
      if (!savedOrder) {
        setOrderedItems(items);
      }
    }
  }, [items.length]);

  const handleDragStart = (event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;

    setActiveId(null);

    if (over && active.id !== over.id) {
      const oldIndex = orderedItems.findIndex((item) => item.id === active.id);
      const newIndex = orderedItems.findIndex((item) => item.id === over.id);

      const newItems = arrayMove(orderedItems, oldIndex, newIndex);
      setOrderedItems(newItems);

      if (typeof window !== 'undefined' && storageKey) {
        localStorage.setItem(storageKey, JSON.stringify(newItems.map((item) => item.id)));
      }

      if (onReorder) {
        onReorder(newItems);
      }
    }
  };

  if (!mounted) {
    return (
      <div className={cn('grid gap-4', gridCols, className)}>
        {items.map((item) => (
          <div key={item.id} className={item.className}>
            {item.component}
          </div>
        ))}
      </div>
    );
  }

  const activeItem = orderedItems.find((item) => item.id === activeId);

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCenter}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <SortableContext
        items={orderedItems.map((item) => item.id)}
        strategy={verticalListSortingStrategy}
      >
        <div className={cn('grid gap-4 items-stretch', gridCols, className)}>
          {orderedItems.map((item) => (
            <SortableItem key={item.id} item={item} onDelete={onDelete} />
          ))}
        </div>
      </SortableContext>
      <DragOverlay>
        {activeItem ? (
          <div className="opacity-90 scale-105 shadow-2xl rotate-2">{activeItem.component}</div>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
