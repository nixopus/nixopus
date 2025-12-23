'use client';

import { useCallback, useMemo, useEffect } from 'react';
import { useNodesState, useEdgesState, type Node, type Edge } from '@xyflow/react';
import type { Server } from '@/redux/types/servers';
import type { ServerNodeData, ClusterEdgeData, UseServerTopologyProps } from '../types';
import { NODE_DIMENSIONS } from '../constants';

function calculateHierarchicalLayout(servers: Server[]): {
  nodes: Node<ServerNodeData>[];
  edges: Edge<ClusterEdgeData>[];
} {
  const nodes: Node<ServerNodeData>[] = [];
  const edges: Edge<ClusterEdgeData>[] = [];

  const { WIDTH, HEIGHT, HORIZONTAL_GAP, VERTICAL_GAP } = NODE_DIMENSIONS;

  // Group servers by cluster
  const clusters = new Map<string | null, { managers: Server[]; workers: Server[] }>();
  const standalones: Server[] = [];

  servers.forEach((server) => {
    if (server.role === 'standalone') {
      standalones.push(server);
      return;
    }

    const clusterId = server.cluster_id || 'unassigned';
    if (!clusters.has(clusterId)) {
      clusters.set(clusterId, { managers: [], workers: [] });
    }

    const cluster = clusters.get(clusterId)!;
    if (server.role === 'manager') {
      cluster.managers.push(server);
    } else if (server.role === 'worker') {
      cluster.workers.push(server);
    }
  });

  // Calculate positions for clusters
  let clusterX = 0;
  const clusterPositions = new Map<string | null, { x: number; width: number }>();

  clusters.forEach((cluster, clusterId) => {
    const clusterManagers = cluster.managers;
    const clusterWorkers = cluster.workers;

    // Calculate cluster width based on max nodes in a row
    const maxNodesInRow = Math.max(clusterManagers.length, clusterWorkers.length);
    const clusterWidth =
      maxNodesInRow > 0 ? maxNodesInRow * WIDTH + (maxNodesInRow - 1) * HORIZONTAL_GAP : WIDTH;

    clusterPositions.set(clusterId, { x: clusterX, width: clusterWidth });
    clusterX += clusterWidth + HORIZONTAL_GAP * 2;
  });

  // Position nodes within clusters
  clusters.forEach((cluster, clusterId) => {
    const { x: clusterX, width: clusterWidth } = clusterPositions.get(clusterId)!;
    const clusterManagers = cluster.managers;
    const clusterWorkers = cluster.workers;

    // Position managers at the top
    clusterManagers.forEach((manager, index) => {
      const totalWidth =
        clusterManagers.length * WIDTH + (clusterManagers.length - 1) * HORIZONTAL_GAP;
      const startX = clusterX + (clusterWidth - totalWidth) / 2;
      const x = startX + index * (WIDTH + HORIZONTAL_GAP);

      nodes.push({
        id: manager.id,
        type: 'serverNode',
        position: { x, y: 0 },
        data: { server: manager, isSelected: false, onSelect: () => {} }
      });
    });

    // Position workers below managers
    clusterWorkers.forEach((worker, index) => {
      const totalWidth =
        clusterWorkers.length * WIDTH + (clusterWorkers.length - 1) * HORIZONTAL_GAP;
      const startX = clusterX + (clusterWidth - totalWidth) / 2;
      const x = startX + index * (WIDTH + HORIZONTAL_GAP);

      nodes.push({
        id: worker.id,
        type: 'serverNode',
        position: { x, y: HEIGHT + VERTICAL_GAP },
        data: { server: worker, isSelected: false, onSelect: () => {} }
      });
    });

    // Create edges between managers and workers in the same cluster
    clusterManagers.forEach((manager) => {
      clusterWorkers.forEach((worker) => {
        edges.push({
          id: `${manager.id}-${worker.id}`,
          source: manager.id,
          target: worker.id,
          type: 'clusterEdge',
          data: {
            isActive: manager.status === 'online' && worker.status === 'online'
          }
        });
      });
    });
  });

  // Position standalone servers on the right
  const standaloneStartX = clusterX + HORIZONTAL_GAP;
  standalones.forEach((server, index) => {
    nodes.push({
      id: server.id,
      type: 'serverNode',
      position: {
        x: standaloneStartX,
        y: index * (HEIGHT + VERTICAL_GAP / 2)
      },
      data: { server, isSelected: false, onSelect: () => {} }
    });
  });

  // Center the entire layout
  if (nodes.length > 0) {
    const minX = Math.min(...nodes.map((n) => n.position.x));
    const maxX = Math.max(...nodes.map((n) => n.position.x + WIDTH));
    const centerX = (minX + maxX) / 2;
    const offsetX = -centerX;

    nodes.forEach((node) => {
      node.position.x += offsetX;
    });
  }

  return { nodes, edges };
}

export function useServerTopology({
  servers,
  selectedServer,
  onServerSelect
}: UseServerTopologyProps) {
  const { nodes: initialNodes, edges: initialEdges } = useMemo(
    () => calculateHierarchicalLayout(servers),
    [servers]
  );

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  useEffect(() => {
    setNodes((nds) =>
      nds.map((node) => {
        const server = servers.find((s) => s.id === node.id);
        if (!server) return node;

        return {
          ...node,
          data: {
            ...node.data,
            server,
            isSelected: selectedServer?.id === node.id,
            onSelect: onServerSelect
          }
        };
      })
    );
  }, [servers, selectedServer, onServerSelect, setNodes]);

  useEffect(() => {
    const { nodes: newNodes, edges: newEdges } = calculateHierarchicalLayout(servers);

    const updatedNodes = newNodes.map((node) => ({
      ...node,
      data: {
        ...node.data,
        isSelected: selectedServer?.id === node.id,
        onSelect: onServerSelect
      }
    }));

    setNodes(updatedNodes);
    setEdges(newEdges);
  }, [servers.length, setNodes, setEdges, selectedServer, onServerSelect, servers]);

  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: Node) => {
      const server = servers.find((s) => s.id === node.id);
      onServerSelect(server || null);
    },
    [servers, onServerSelect]
  );

  const handlePaneClick = useCallback(() => {
    onServerSelect(null);
  }, [onServerSelect]);

  const getNodeColor = useCallback(
    (node: Node) => {
      const server = servers.find((s) => s.id === node.id);
      if (!server) return 'hsl(var(--muted))';
      if (server.status === 'online') return 'hsl(var(--primary))';
      if (server.status === 'error') return 'hsl(var(--destructive))';
      return 'hsl(var(--muted-foreground))';
    },
    [servers]
  );

  return {
    nodes,
    edges,
    onNodesChange,
    onEdgesChange,
    handleNodeClick,
    handlePaneClick,
    getNodeColor
  };
}
