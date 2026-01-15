import { formatDistanceToNow, format } from 'date-fns';

export const isRunning = (status: string) => (status || '').toLowerCase() === 'running';

export const formatDate = (date: string) =>
  formatDistanceToNow(new Date(date), { addSuffix: true });

export const formatDateFull = (date: string) => format(new Date(date), 'PPpp');

export const truncateId = (id: string) => id.slice(0, 12);

export const formatImageId = (id: string, length: number = 12) =>
  id.replace('sha256:', '').slice(0, length);

export const bytesToMB = (bytes: number) => (bytes > 0 ? Math.round(bytes / (1024 * 1024)) : 0);

export const getPortsDisplay = (ports: any[], max: number, variant: 'pill' | 'inline' = 'pill') => {
  if (!ports?.length) return null;
  const visible = ports.slice(0, max);
  const remaining = ports.length - max;
  return { visible, remaining };
};
