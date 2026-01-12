'use client';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue
} from '@/components/ui/select';
import { Application } from '@/redux/types/applications';
import { HealthCheck } from '@/redux/types/healthcheck';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { useHealthCheckDialog } from '@/packages/hooks/applications/use-health-check-dialog';

interface HealthCheckDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  application: Application;
  healthCheck?: HealthCheck;
}

export function HealthCheckDialog({
  open,
  onOpenChange,
  application,
  healthCheck
}: HealthCheckDialogProps) {
  const { t } = useTranslation();
  const {
    endpoint,
    setEndpoint,
    method,
    setMethod,
    intervalSeconds,
    setIntervalSeconds,
    timeoutSeconds,
    setTimeoutSeconds,
    enabled,
    handleSubmit,
    handleDelete,
    handleToggle,
    isLoading
  } = useHealthCheckDialog({
    application,
    healthCheck,
    onSuccess: () => onOpenChange(false)
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {healthCheck
              ? t('selfHost.monitoring.healthCheck.editTitle' as any)
              : t('selfHost.monitoring.healthCheck.createTitle' as any)}
          </DialogTitle>
          <DialogDescription>
            {t('selfHost.monitoring.healthCheck.description' as any)}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <Label>{t('selfHost.monitoring.healthCheck.endpoint' as any)}</Label>
            <Input
              value={endpoint}
              onChange={(e) => setEndpoint(e.target.value)}
              placeholder="/health"
            />
          </div>

          <div className="space-y-2">
            <Label>{t('selfHost.monitoring.healthCheck.method' as any)}</Label>
            <Select value={method} onValueChange={(v) => setMethod(v as 'GET' | 'POST' | 'HEAD')}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GET">GET</SelectItem>
                <SelectItem value="POST">POST</SelectItem>
                <SelectItem value="HEAD">HEAD</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>{t('selfHost.monitoring.healthCheck.intervalSeconds' as any)}</Label>
            <Input
              type="number"
              value={intervalSeconds}
              onChange={(e) => setIntervalSeconds(e.target.value)}
              onBlur={(e) => {
                const inputValue = e.target.value.trim();
                if (inputValue === '') {
                  setIntervalSeconds('60');
                  return;
                }
                const value = parseInt(inputValue, 10);
                if (isNaN(value)) {
                  setIntervalSeconds('60');
                } else if (value < 30) {
                  setIntervalSeconds('30');
                } else if (value > 3600) {
                  setIntervalSeconds('3600');
                }
                // If value is valid and in range, don't update state to avoid resetting user input
              }}
              min={30}
              max={3600}
            />
          </div>

          <div className="space-y-2">
            <Label>{t('selfHost.monitoring.healthCheck.timeoutSeconds' as any)}</Label>
            <Input
              type="number"
              value={timeoutSeconds}
              onChange={(e) => setTimeoutSeconds(e.target.value)}
              onBlur={(e) => {
                const inputValue = e.target.value.trim();
                if (inputValue === '') {
                  setTimeoutSeconds('30');
                  return;
                }
                const value = parseInt(inputValue, 10);
                if (isNaN(value)) {
                  setTimeoutSeconds('30');
                } else if (value < 5) {
                  setTimeoutSeconds('5');
                } else if (value > 120) {
                  setTimeoutSeconds('120');
                }
                // If value is valid and in range, don't update state to avoid resetting user input
              }}
              min={5}
              max={120}
            />
          </div>

          {healthCheck && (
            <div className="flex items-center justify-between">
              <Label>{t('selfHost.monitoring.healthCheck.enabled' as any)}</Label>
              <Switch checked={enabled} onCheckedChange={handleToggle} />
            </div>
          )}
        </div>

        <DialogFooter>
          {healthCheck && (
            <Button variant="destructive" onClick={handleDelete} disabled={isLoading}>
              {t('selfHost.monitoring.healthCheck.delete' as any)}
            </Button>
          )}
          <Button onClick={handleSubmit} disabled={isLoading}>
            {isLoading && (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2" />
            )}
            {healthCheck
              ? t('selfHost.monitoring.healthCheck.update' as any)
              : t('selfHost.monitoring.healthCheck.create' as any)}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
