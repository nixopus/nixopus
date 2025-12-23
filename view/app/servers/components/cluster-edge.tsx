'use client';

import React from 'react';
import { BaseEdge, type EdgeProps, getBezierPath, EdgeLabelRenderer } from '@xyflow/react';
import { cn } from '@/lib/utils';
import type { ClusterEdgeType } from '../types';

export function ClusterEdge({
  sourceX,
  sourceY,
  targetX,
  targetY,
  sourcePosition,
  targetPosition,
  style = {},
  markerEnd,
  data
}: EdgeProps<ClusterEdgeType>) {
  const [edgePath, labelX, labelY] = getBezierPath({
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition
  });

  const isActive = data?.isActive ?? true;

  return (
    <>
      <BaseEdge
        path={edgePath}
        markerEnd={markerEnd}
        style={{
          ...style,
          stroke: isActive ? 'hsl(var(--primary))' : 'hsl(var(--muted-foreground))',
          strokeWidth: isActive ? 2 : 1,
          strokeDasharray: isActive ? '0' : '5,5',
          opacity: isActive ? 1 : 0.5
        }}
      />
      {data?.label && (
        <EdgeLabelRenderer>
          <div
            style={{
              position: 'absolute',
              transform: `translate(-50%, -50%) translate(${labelX}px,${labelY}px)`,
              pointerEvents: 'all'
            }}
            className={cn(
              'px-2 py-0.5 rounded text-xs font-medium',
              isActive ? 'bg-primary text-primary-foreground' : 'bg-muted text-muted-foreground'
            )}
          >
            {data.label}
          </div>
        </EdgeLabelRenderer>
      )}
    </>
  );
}
