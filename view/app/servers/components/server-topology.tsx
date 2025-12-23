'use client';

import React from 'react';
import {
  ReactFlow,
  Controls,
  Background,
  BackgroundVariant,
  ConnectionMode,
  type NodeTypes,
  type EdgeTypes
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { ServerNode } from './server-node';
import { ClusterEdge } from './cluster-edge';
import { useServerTopology } from '../hooks/use-server-topology';
import type { ServerTopologyProps } from '../types';

const nodeTypes: NodeTypes = {
  serverNode: ServerNode as unknown as NodeTypes[string]
};

const edgeTypes: EdgeTypes = {
  clusterEdge: ClusterEdge as unknown as EdgeTypes[string]
};

export function ServerTopology({ servers, selectedServer, onServerSelect }: ServerTopologyProps) {
  const { nodes, edges, onNodesChange, onEdgesChange, handleNodeClick, handlePaneClick } =
    useServerTopology({ servers, selectedServer, onServerSelect });

  return (
    <>
      <style
        dangerouslySetInnerHTML={{
          __html: `
        .react-flow__controls-button svg {
          fill: hsl(var(--primary)) !important;
          stroke: hsl(var(--primary)) !important;
        }
        .react-flow__controls-button:hover svg {
          fill: hsl(var(--primary)) !important;
          stroke: hsl(var(--primary)) !important;
        }
      `
        }}
      />
      <div className="h-full w-full rounded-xl border bg-background/50 backdrop-blur-sm overflow-hidden">
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onNodeClick={handleNodeClick}
          onPaneClick={handlePaneClick}
          nodeTypes={nodeTypes}
          edgeTypes={edgeTypes}
          connectionMode={ConnectionMode.Loose}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          minZoom={0.3}
          maxZoom={1.5}
          defaultViewport={{ x: 0, y: 0, zoom: 0.8 }}
          proOptions={{ hideAttribution: true }}
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={20}
            size={1}
            className="!bg-background"
            color="hsl(var(--muted-foreground) / 0.2)"
          />
          <Controls showInteractive={false} className="!bg-card !border-border !shadow-lg" />
        </ReactFlow>
      </div>
    </>
  );
}
