import { MachineProvider } from '@/packages/contexts/machine-context';

export default async function MachineLayout({
  children,
  params
}: {
  children: React.ReactNode;
  params: Promise<{ machineId: string }>;
}) {
  const { machineId } = await params;
  return <MachineProvider machineId={machineId}>{children}</MachineProvider>;
}
