export type ContainerData = {
  Id: string;
  Names: string[];
  Image: string;
  ImageID: string;
  Command: string;
  Created: number;
  Ports: Port[];
  SizeRw?: number;
  SizeRootFs?: number;
  Labels: { [key: string]: string };
  State: string;
  Status: string;
  HostConfig: {
    NetworkMode?: string;
    Annotations?: { [key: string]: string };
  };
};

export interface Port {
  IP?: string;
  PrivatePort: number;
  PublicPort?: number;
  Type: string;
}

export interface MemoryStats {
  used: number;
  total: number;
  percentage: number;
  rawInfo: string;
}

export interface LoadStats {
  oneMin: number;
  fiveMin: number;
  fifteenMin: number;
  uptime: string;
}

export interface DiskMount {
  filesystem: string;
  size: string;
  used: string;
  avail: string;
  capacity: string;
  mountPoint: string;
}

export interface DiskStats {
  total: number;
  used: number;
  available: number;
  percentage: number;
  mountPoint: string;
  allMounts: DiskMount[];
}

export interface CPUCore {
  core_id: number;
  usage: number;
}

export interface CPUStats {
  overall: number;
  per_core: CPUCore[];
}

export interface NetworkInterface {
  name: string;
  bytesSent: number;
  bytesRecv: number;
  packetsSent: number;
  packetsRecv: number;
  errorIn: number;
  errorOut: number;
  dropIn: number;
  dropOut: number;
}

export interface NetworkStats {
  totalBytesSent: number;
  totalBytesRecv: number;
  totalPacketsSent: number;
  totalPacketsRecv: number;
  interfaces: NetworkInterface[];
  uploadSpeed: number;
  downloadSpeed: number;
}

export interface SystemStatsType {
  os_type: string;
  hostname: string;
  cpu_info: string;
  cpu_cores: number;
  cpu: CPUStats;
  memory: MemoryStats;
  load: LoadStats;
  disk: DiskStats;
  network: NetworkStats;
  kernel_version: string;
  architecture: string;
  timestamp: number;
}
