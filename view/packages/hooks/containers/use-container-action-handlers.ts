import { Action } from '@/packages/types/containers';

export function useContainerActionHandlers(
  containerId: string,
  onAction: (id: string, action: Action) => void
) {
  const handleAction = (action: Action) => (e: React.MouseEvent) => {
    e.stopPropagation();
    onAction(containerId, action);
  };

  return {
    handleAction,
    handleStart: handleAction(Action.START),
    handleStop: handleAction(Action.STOP),
    handleRemove: handleAction(Action.REMOVE)
  };
}
