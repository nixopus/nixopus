import { Badge } from '@/components/ui/badge';

interface ContainerInfoProps {
  container: any;
}

export const ContainerInfo = ({ container }: ContainerInfoProps) => {
  const containerPorts = container.ports;
  return (
    <div>
      <div className="flex flex-wrap gap-2 mt-1">
        {containerPorts?.map((port: any) => (
          <Badge key={`${port.private_port}-${port.public_port}`} variant="outline">
            {port.public_port} → {port.private_port}
          </Badge>
        ))}
      </div>
    </div>
  );
};
