import { Info, Cpu, Network, Terminal } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { formatDistanceToNow } from 'date-fns';
import { Container } from '@/redux/services/container/containerApi';
import { useTranslation } from '@/hooks/use-translation';

interface OverviewTabProps {
  container: Container;
}

export function OverviewTab({ container }: OverviewTabProps) {
  const { t } = useTranslation();

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            {t('containers.basic_info')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('containers.status')}</span>
              <Badge variant={container.status === 'running' ? 'default' : 'secondary'}>
                {container.status}
              </Badge>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('containers.created')}</span>
              <span className="text-sm">
                {formatDistanceToNow(new Date(container.created), { addSuffix: true })}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('containers.ip_address')}</span>
              <span className="text-sm font-mono">{container.ip_address}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Cpu className="h-5 w-5" />
            {t('containers.resources')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('containers.cpu_shares')}</span>
              <span className="text-sm">{container.host_config.cpu_shares} shares</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">{t('containers.memory')}</span>
              <span className="text-sm">
                {(container.host_config.memory / (1024 * 1024)).toFixed(2)} MB
              </span>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Network className="h-5 w-5" />
            {t('containers.ports')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-wrap gap-2">
            {container?.ports?.map((port, index) => (
              <Badge key={`${port.private_port}-${port.public_port}-${index}`} variant="outline">
                {port.public_port} → {port.private_port} ({port.type})
              </Badge>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Terminal className="h-5 w-5" />
            {t('containers.command')}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="bg-muted p-4 rounded-lg">
            <p className="text-sm font-mono">{container.command}</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
