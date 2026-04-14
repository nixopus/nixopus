import type { ChatContext } from './chat-context';

export interface GuidedPrefillPayload {
  promptText: string;
  sampleRepoContext: ChatContext;
}

export function getDefaultGuidedPrompt(t: (key: string) => string): string {
  return t('ai.guidedPrefill.defaultPrompt');
}

export function buildSampleRepoContext(t: (key: string) => string): ChatContext {
  return {
    type: 'Repository',
    id: 'sample-repo-default',
    label: t('ai.guidedPrefill.sampleRepo.label'),
    meta: {
      Language: 'TypeScript',
      Branch: 'main',
      Visibility: 'public'
    }
  };
}

/** True when input and contexts still match the guided starter (same rules as injection). */
export function matchesGuidedPrefillSnapshot(
  inputValue: string,
  contexts: ChatContext[],
  translate: (key: string) => string
): boolean {
  const promptText = getDefaultGuidedPrompt((key) =>
    key === 'ai.guidedPrefill.defaultPrompt' ? translate('ai.guidedPrefill.defaultPrompt') : key
  );
  if (inputValue !== promptText) return false;
  const sampleRepoContext = buildSampleRepoContext((key) =>
    key === 'ai.guidedPrefill.sampleRepo.label'
      ? translate('ai.guidedPrefill.sampleRepo.label')
      : key
  );
  return contexts.some((c) => c.type === sampleRepoContext.type && c.id === sampleRepoContext.id);
}
