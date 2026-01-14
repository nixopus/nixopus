import { CPUCore } from '@/redux/types/monitor';

export function useTopCores(perCoreData: CPUCore[], limit: number = 3): CPUCore[] {
  return [...perCoreData].sort((a, b) => b.usage - a.usage).slice(0, limit);
}
