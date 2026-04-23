'use client';
import { useRouter, usePathname } from 'next/navigation';
import { ChevronsUpDown, Check, Server, RefreshCw } from 'lucide-react';
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger
} from '@nixopus/ui';
import { useState } from 'react';
import { cn } from '@/lib/utils';
import { useGetServersQuery, useVerifyMachineMutation } from '@/redux/services/servers/serversApi';
import { useMachineContext } from '@/packages/contexts/machine-context';
import { toast } from 'sonner';

export function MachineSwitcher() {
  const router = useRouter();
  const pathname = usePathname();
  const { machineId: contextMachineId } = useMachineContext();
  const { data } = useGetServersQuery({ page: 1, page_size: 100 });
  const [verifyMachine] = useVerifyMachineMutation();
  const [retryingId, setRetryingId] = useState<string | null>(null);

  const servers = data?.servers ?? [];

  if (servers.length < 2) return null;

  const urlMatch = pathname.match(/^\/machines\/([^/]+)/);
  const activeMachineId = urlMatch ? urlMatch[1] : contextMachineId;

  const currentServer = servers.find((s) => s.id === activeMachineId);

  const handleSelect = (server: (typeof servers)[number]) => {
    if (server.id === activeMachineId) return;
    if (!server.is_active) return;

    const target = urlMatch
      ? `/machines/${server.id}${pathname.replace(urlMatch[0], '')}`
      : `/machines/${server.id}${pathname}`;
    router.push(target);
  };

  const handleRetry = async (e: React.MouseEvent, serverId: string) => {
    e.stopPropagation();
    setRetryingId(serverId);
    try {
      await verifyMachine(serverId).unwrap();
      toast.success('Verifying connection, this may take a moment...');
    } catch {
      toast.error('Failed to initiate connection retry');
    } finally {
      setRetryingId(null);
    }
  };

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" className="gap-1.5 px-2 h-7 text-sm font-medium">
          <Server className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          <span className="truncate max-w-[150px]">
            {currentServer?.name?.trim() ||
              currentServer?.provision?.domain?.trim() ||
              currentServer?.host?.trim() ||
              currentServer?.id ||
              'Select Machine'}
          </span>
          <ChevronsUpDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-[250px]" align="start">
        {servers.map((server) => (
          <DropdownMenuItem
            key={server.id}
            onClick={() => handleSelect(server)}
            className={cn(
              'gap-2',
              server.is_active ? 'cursor-pointer' : 'cursor-default opacity-60'
            )}
          >
            <div
              className={cn(
                'h-2 w-2 rounded-full shrink-0',
                server.is_active ? 'bg-green-500' : 'bg-red-500'
              )}
            />
            <span className="truncate text-sm flex-1 min-w-0">
              {server.name?.trim() || server.provision?.domain?.trim() || server.host?.trim() || server.id}
            </span>
            {!server.is_active ? (
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 shrink-0"
                disabled={retryingId === server.id}
                onClick={(e) => handleRetry(e, server.id)}
              >
                <RefreshCw
                  className={cn('h-3.5 w-3.5', retryingId === server.id && 'animate-spin')}
                />
              </Button>
            ) : server.id === activeMachineId ? (
              <Check className="h-4 w-4 shrink-0" />
            ) : null}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}