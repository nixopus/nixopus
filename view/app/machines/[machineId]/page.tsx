import { redirect } from 'next/navigation';

export default async function MachineIndexPage({
  params
}: {
  params: Promise<{ machineId: string }>;
}) {
  const { machineId } = await params;
  redirect(`/machines/${machineId}/charts`);
}
