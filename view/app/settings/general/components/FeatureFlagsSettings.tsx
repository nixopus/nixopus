import { useAppSelector } from '@/redux/hooks';
import { Switch } from '@/components/ui/switch';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';
import {
  useGetAllFeatureFlagsQuery,
  useUpdateFeatureFlagMutation
} from '@/redux/services/feature-flags/featureFlagsApi';
import { Separator } from '@/components/ui/separator';
import { FeatureFlag, FeatureName, featureGroups } from '@/packages/types/feature-flags';
import { RBACGuard } from '@/components/rbac/RBACGuard';
import { TypographySmall, TypographyMuted, TypographyH3 } from '@/components/ui/typography';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import {
  Server,
  Code,
  BarChart3,
  Bell,
  CheckCircle2,
  XCircle,
  Search,
  Filter,
  Settings
} from 'lucide-react';
import { useState } from 'react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

export default function FeatureFlagsSettings() {
  const { t } = useTranslation();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const { data: featureFlags, isLoading } = useGetAllFeatureFlagsQuery(undefined, {
    skip: !activeOrganization?.id
  });
  const [updateFeatureFlag] = useUpdateFeatureFlagMutation();
  const [searchTerm, setSearchTerm] = useState('');
  const [filterEnabled, setFilterEnabled] = useState<'all' | 'enabled' | 'disabled'>('all');

  const handleToggleFeature = async (featureName: string, isEnabled: boolean) => {
    try {
      await updateFeatureFlag({
        feature_name: featureName,
        is_enabled: isEnabled
      }).unwrap();
      toast.success(t('settings.featureFlags.messages.updated'));
    } catch (error) {
      toast.error(t('settings.featureFlags.messages.updateFailed'));
    }
  };

  const getGroupIcon = (group: string) => {
    const iconMap = {
      infrastructure: Server,
      development: Code,
      monitoring: BarChart3,
      notifications: Bell
    };
    return iconMap[group as keyof typeof iconMap] || Settings;
  };

  const getFilteredFeatures = () => {
    if (!featureFlags) return [];

    return featureFlags.filter((feature) => {
      // Exclude domain and notifications features for now
      // TODO: Add them back later when we have them implemented
      if (feature.feature_name === 'domain' || feature.feature_name === 'notifications') {
        return false;
      }

      const matchesSearch =
        feature.feature_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        t(`settings.featureFlags.features.${feature.feature_name}.title` as any)
          .toLowerCase()
          .includes(searchTerm.toLowerCase());

      const matchesFilter =
        filterEnabled === 'all' ||
        (filterEnabled === 'enabled' && feature.is_enabled) ||
        (filterEnabled === 'disabled' && !feature.is_enabled);

      return matchesSearch && matchesFilter;
    });
  };

  const getGroupedFeatures = () => {
    const filteredFeatures = getFilteredFeatures();
    const grouped = new Map<string, FeatureFlag[]>();

    filteredFeatures.forEach((feature) => {
      for (const [group, features] of Object.entries(featureGroups)) {
        if (features.includes(feature.feature_name as FeatureName)) {
          if (!grouped.has(group)) {
            grouped.set(group, []);
          }
          grouped.get(group)?.push(feature as FeatureFlag);
          return;
        }
      }
    });
    return grouped;
  };

  const groupedFeatures = getGroupedFeatures();
  const totalFeatures = featureFlags?.length || 0;
  const enabledFeatures = featureFlags?.filter((f) => f.is_enabled).length || 0;
  const disabledFeatures = totalFeatures - enabledFeatures;

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div>
          <div className="space-y-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="animate-pulse">
                <div className="h-4 bg-muted rounded w-1/4 mb-2"></div>
                <div className="space-y-2">
                  {[1, 2].map((j) => (
                    <div
                      key={j}
                      className="flex items-center justify-between p-4 border rounded-lg"
                    >
                      <div className="space-y-2">
                        <div className="h-4 bg-muted rounded w-32"></div>
                        <div className="h-3 bg-muted rounded w-48"></div>
                      </div>
                      <div className="h-6 w-11 bg-muted rounded-full"></div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  return (
    <RBACGuard resource="feature-flags" action="read">
      <div className="flex flex-col h-full">
        <div className="flex items-center justify-between mb-6">
          <div>
            <TypographyH3 className="text-lg font-semibold">
              {t('settings.featureFlags.title')}
            </TypographyH3>
            <TypographyMuted className="text-xs mt-1">
              {t('settings.featureFlags.description')}
            </TypographyMuted>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="secondary" className="flex items-center gap-1">
              <CheckCircle2 className="h-3 w-3" />
              {enabledFeatures}
            </Badge>
            <Badge variant="outline" className="flex items-center gap-1">
              <XCircle className="h-3 w-3" />
              {disabledFeatures}
            </Badge>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto space-y-6">
          <div className="flex items-center gap-4">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder={t('settings.featureFlags.searchPlaceholder')}
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10"
              />
            </div>
            <div className="flex items-center gap-2">
              <div className="flex gap-1">
                {(['all', 'enabled', 'disabled'] as const).map((filter) => (
                  <Button
                    key={filter}
                    variant={filterEnabled === filter ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setFilterEnabled(filter)}
                  >
                    {t(`settings.featureFlags.filters.${filter}`)}
                  </Button>
                ))}
              </div>
            </div>
          </div>

          {groupedFeatures.size === 0 ? (
            <Alert>
              <Search className="h-4 w-4" />
              <AlertDescription>
                {searchTerm || filterEnabled !== 'all'
                  ? t('settings.featureFlags.noResults')
                  : t('settings.featureFlags.noFeatures')}
              </AlertDescription>
            </Alert>
          ) : (
            Array.from(groupedFeatures.entries())
              .filter(([group]) => group !== 'notifications')
              .map(([group, features], index) => {
                const GroupIcon = getGroupIcon(group);
                const enabledInGroup = features.filter((f) => f.is_enabled).length;

                return (
                  <div key={group} className="space-y-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <GroupIcon className="h-4 w-4 text-muted-foreground" />
                        <TypographySmall className="font-semibold">
                          {t(`settings.featureFlags.groups.${group}.title` as any)}
                        </TypographySmall>
                        <Badge variant="outline" className="text-xs">
                          {enabledInGroup}/{features.length}
                        </Badge>
                      </div>
                    </div>
                    <div className="space-y-3">
                      {features?.map((feature) => (
                        <div
                          key={feature.feature_name}
                          className="flex items-center justify-between p-4 rounded-md bg-muted/30 transition-colors hover:bg-muted/50"
                        >
                          <div className="space-y-1 flex-1">
                            <div className="flex items-center gap-2">
                              <TypographySmall className="font-medium">
                                {t(
                                  `settings.featureFlags.features.${feature.feature_name}.title` as any
                                )}
                              </TypographySmall>
                            </div>
                            <TypographyMuted className="text-sm">
                              {t(
                                `settings.featureFlags.features.${feature.feature_name}.description` as any
                              )}
                            </TypographyMuted>
                          </div>
                          <RBACGuard resource="feature-flags" action="update">
                            <Switch
                              checked={feature.is_enabled}
                              onCheckedChange={(checked) =>
                                handleToggleFeature(feature.feature_name, checked)
                              }
                            />
                          </RBACGuard>
                        </div>
                      ))}
                    </div>
                  </div>
                );
              })
          )}
        </div>
      </div>
    </RBACGuard>
  );
}
