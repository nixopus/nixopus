import { Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useTranslation } from '@/hooks/use-translation';
import { Container } from '@/redux/services/container/containerApi';

interface LogsTabProps {
  container: Container;
  logs: string;
  onLoadMore: () => void;
}

export function LogsTab({ container, logs, onLoadMore }: LogsTabProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-4">
      <ScrollArea className="h-[600px] rounded-lg border">
        <pre className="text-sm font-mono whitespace-pre-wrap p-4">
          {logs || t('containers.no_logs')}
        </pre>
      </ScrollArea>
      <div className="flex justify-center">
        <Button
          variant="outline"
          onClick={onLoadMore}
          className="w-full max-w-[200px]"
        >
          <Loader2 className="mr-2 h-4 w-4" />
          {t('containers.load_more_logs')}
        </Button>
      </div>
    </div>
  );
} 