import { ScrollArea } from '@/components/ui/scroll-area';
import { Container } from '@/redux/services/container/containerApi';

interface DetailsTabProps {
  container: Container;
}

export function DetailsTab({ container }: DetailsTabProps) {
  return (
    <ScrollArea className="h-[600px] rounded-lg border">
      <pre className="text-sm font-mono whitespace-pre-wrap p-4">
        {JSON.stringify(container, null, 2)}
      </pre>
    </ScrollArea>
  );
}