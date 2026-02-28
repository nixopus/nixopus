'use client';

import React, { useMemo } from 'react';
import { useParams } from 'next/navigation';
import { WorkflowEditor } from '@/packages/components/workflows';
import { useWorkflowDetail } from '@/packages/hooks/workflows';
import { ResourceGuard } from '@/packages/components/rbac';
import PageLayout from '@/packages/layouts/page-layout';
import { Skeleton } from '@nixopus/ui';
import { AlertCircle, Sparkles } from 'lucide-react';
import { isAgentConfigured } from '@/packages/lib/agent-client';

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

  if (!isAgentConfigured()) {
    return (
      <ResourceGuard
        resource="deploy"
        action="read"
        loadingFallback={<Skeleton className="h-96" />}
      >
        <PageLayout maxWidth="full" padding="md" spacing="lg">
          <div className="flex h-full w-full items-center justify-center py-24">
            <div className="text-center max-w-md space-y-4 px-4">
              <div className="flex items-center justify-center size-16 rounded-2xl bg-muted mx-auto">
                <Sparkles className="size-8 text-muted-foreground" />
              </div>
              <h3 className="text-lg font-semibold">AI Agent Not Configured</h3>
              <p className="text-sm text-muted-foreground">
                The AI-powered deployment assistant is not enabled on this instance. To get access,
                reach out to us and we&apos;ll help you get set up.
              </p>
              <a
                href="mailto:support@nixopus.com"
                className="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 transition-colors"
              >
                Contact support@nixopus.com
              </a>
            </div>
          </div>
        </PageLayout>
      </ResourceGuard>
    );
  }

  if (error && !isNew) {
    return (
      <ResourceGuard
        resource="deploy"
        action="read"
        loadingFallback={<Skeleton className="h-96" />}
      >
        <PageLayout maxWidth="full" padding="md" spacing="lg">
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
