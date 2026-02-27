import { SystemStatsType } from '@/redux/types/monitor';

export interface SystemInfoData {
  hostname: string;
  os: string;
  cpu: string;
  cores: number;
  kernel: string;
  uptime: string;
}

export function useSystemInfoData(systemStats: SystemStatsType): SystemInfoData {
  const { load, os_type, cpu_info, cpu_cores, hostname, architecture, kernel_version } = systemStats;

  // Shorten CPU info - take first part before @ or full if short
  const shortCpu = cpu_info?.split('@')[0]?.trim() || cpu_info || 'N/A';

  return {
    hostname: hostname || 'N/A',
    os: `${os_type || 'N/A'} (${architecture || 'N/A'})`,
    cpu: shortCpu,
    cores: cpu_cores,
    kernel: kernel_version || 'N/A',
    uptime: load.uptime?.replaceAll(/([hms])(\d)/g, '$1 $2') || 'N/A'
  };
}
