import { useAppSelector } from '@/redux/hooks';
import { Switch } from '@/components/ui/switch';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useTranslation } from '@/hooks/use-translation';
import { toast } from 'sonner';
import { TabsContent } from '@/components/ui/tabs';
import {
  useGetAllFeatureFlagsQuery,
  useUpdateFeatureFlagMutation
} from '@/redux/services/feature-flags/featureFlagsApi';
import { Separator } from '@/components/ui/separator';
import { FeatureFlag, FeatureName, featureGroups } from '@/types/feature-flags';
import { RBACGuard } from '@/components/rbac/RBACGuard';

export default function FeatureFlagsSettings() {
  const { t } = useTranslation();
  const activeOrganization = useAppSelector((state) => state.user.activeOrganization);
  const { data: featureFlags, isLoading } = useGetAllFeatureFlagsQuery(undefined, {
    skip: !activeOrganization?.id
  });
  const [updateFeatureFlag] = useUpdateFeatureFlagMutation();

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

  if (isLoading) {
    return <div>{t('common.loading')}</div>;
  }

  const getGroupedFeatures = () => {
    const grouped = new Map<string, FeatureFlag[]>();
    featureFlags?.forEach((feature) => {
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

  return (
    <RBACGuard resource="feature-flags" action="read">
      <TabsContent value="feature-flags" className="space-y-6 mt-4">
        <Card>
          <CardHeader>
            <CardTitle>{t('settings.featureFlags.title')}</CardTitle>
            <CardDescription>{t('settings.featureFlags.description')}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {Array.from(groupedFeatures.entries()).map(([group, features], index) => (
              <div key={group} className="space-y-4">
                <div className="space-y-2">
                  <h3 className="text-lg font-semibold">
                    {t(`settings.featureFlags.groups.${group}.title`)}
                  </h3>
                </div>
                <div className="space-y-4">
                  {features?.map((feature) => (
                    <div
                      key={feature.feature_name}
                      className="flex items-center justify-between p-2 rounded-lg"
                    >
                      <div className="space-y-1">
                        <h4 className="text-sm font-medium">
                          {t(`settings.featureFlags.features.${feature.feature_name}.title`)}
                        </h4>
                        <p className="text-sm text-muted-foreground">
                          {t(`settings.featureFlags.features.${feature.feature_name}.description`)}
                        </p>
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
                {index !== groupedFeatures.size - 1 && <Separator />}
              </div>
            ))}
          </CardContent>
        </Card>
      </TabsContent>
    </RBACGuard>
  );
}
