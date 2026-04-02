'use client';

import React, { useMemo } from 'react';
import { useParams } from 'next/navigation';
import { WorkflowEditor } from '@/packages/components/workflows';
import { useWorkflowDetail } from '@/packages/hooks/workflows';
import { ResourceGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';
import { Skeleton } from '@nixopus/ui';
import { AlertCircle } from 'lucide-react';

export default function WorkflowEditorPage() {
  const params = useParams();
  const applicationId = params.id as string;
  const workflowId = params.workflowId as string;
  const isNew = workflowId === 'new';
  const { dynamicWorkflow, isLoading, error } = useWorkflowDetail({ workflowId, applicationId });

  const { nodes, edges, name, planningMessages, executionMessages, chatThreadId } = useMemo(() => {
    if (dynamicWorkflow) {
      const safeNodes = (dynamicWorkflow.nodes || []).map((n: any, i: number) => ({
        ...n,
        position: n.position ?? { x: 300, y: i * 100 },
        data: { executionStatus: 'idle', ...n.data }
      }));
      return {
        nodes: safeNodes,
        edges: dynamicWorkflow.edges || [],
        name: dynamicWorkflow.name,
        planningMessages: dynamicWorkflow.planningMessages,
        executionMessages: dynamicWorkflow.executionMessages,
        chatThreadId: dynamicWorkflow.chatThreadId
      };
    }

    return {
      nodes: [],
      edges: [],
      name: undefined,
      planningMessages: undefined,
      executionMessages: undefined,
      chatThreadId: undefined
    };
  }, [dynamicWorkflow]);

  if (error && !isNew) {
    return (
      <ResourceGuard
        resource="deploy"
        action="read"
        loadingFallback={<Skeleton className="h-96" />}
      >
        <PageLayout maxWidth="6xl" padding="md" spacing="lg">
          <div className="flex flex-col items-center justify-center py-24 gap-3">
            <AlertCircle className="h-8 w-8 text-destructive" />
            <p className="text-sm text-destructive">{error}</p>
          </div>
        </PageLayout>
      </ResourceGuard>
    );
  }

  if (!isNew && isLoading) {
    return (
      <ResourceGuard
        resource="deploy"
        action="read"
        loadingFallback={<Skeleton className="h-96" />}
      >
        <PageLayout maxWidth="full" padding="none" spacing="none" className="!p-0">
          <div className="h-[calc(100vh-5rem)] flex items-center justify-center">
            <Skeleton className="h-full w-full rounded-none" />
          </div>
        </PageLayout>
      </ResourceGuard>
    );
  }

  return (
    <ResourceGuard resource="deploy" action="read" loadingFallback={<Skeleton className="h-96" />}>
      <PageLayout maxWidth="full" padding="none" spacing="none" className="!p-0">
        <div className="flex flex-col h-[calc(100vh-5rem)]">
          <WorkflowEditor
            applicationId={applicationId}
            workflowId={workflowId}
            workflowName={isNew ? 'New Workflow' : (name ?? workflowId)}
            initialNodes={nodes}
            initialEdges={edges}
            initialPlanningMessages={planningMessages}
            initialExecutionMessages={executionMessages}
            chatThreadId={chatThreadId}
            isDraft={isNew}
            isLoadingMessages={!isNew && isLoading}
          />
        </div>
      </PageLayout>
    </ResourceGuard>
  );
}
