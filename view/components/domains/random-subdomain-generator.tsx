import React from 'react';
import { useGenerateRandomSubdomainQuery } from '@/redux/services/settings/domainsApi';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

interface RandomSubdomainGeneratorProps {
  onSubdomainGenerated?: (subdomain: string) => void;
}

export const RandomSubdomainGenerator: React.FC<RandomSubdomainGeneratorProps> = ({
  onSubdomainGenerated
}) => {
  const {
    data: randomSubdomainData,
    isLoading,
    error,
    refetch
  } = useGenerateRandomSubdomainQuery();

  const handleGenerateSubdomain = async () => {
    try {
      const result = await refetch();
      if (result.data) {
        toast.success(`Generated subdomain: ${result.data.subdomain}`);
        onSubdomainGenerated?.(result.data.subdomain);
      }
    } catch (err) {
      toast.error('Failed to generate random subdomain');
    }
  };

  return (
    <div className="flex flex-col gap-2">
      <Button onClick={handleGenerateSubdomain} disabled={isLoading} variant="outline">
        {isLoading ? 'Generating...' : 'Generate Random Subdomain'}
      </Button>

      {randomSubdomainData && (
        <div className="text-sm text-muted-foreground">
          <p>
            <strong>Subdomain:</strong> {randomSubdomainData.subdomain}
          </p>
          <p>
            <strong>Domain:</strong> {randomSubdomainData.domain}
          </p>
        </div>
      )}

      {error && (
        <p className="text-sm text-destructive">Failed to generate subdomain. Please try again.</p>
      )}
    </div>
  );
};
