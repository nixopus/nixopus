"use client"
import React, { useState } from 'react'
import DashboardPageHeader from '@/components/dashboard-page-header'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import NotificationPreferencesTab from './components/preferenceTab'
import NotificationChannelsTab from './components/channelTab'
import { activityPreferences, securityPreferences, updatePreferences } from './utils/preferences'


export type NotificationChannelConfig = {
  [key: string]: string;
}

const Page: React.FC = () => {
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);

  const handleSaveAll = () => {
    setIsSubmitting(true);
    setTimeout(() => {
      setIsSubmitting(false);
    }, 1000);
  };

  const handleResetDefaults = () => {
    if (confirm('Are you sure you want to reset all settings to defaults?')) {
      console.log('Resetting to defaults');
    }
  };

  return (
    <div className="container mx-auto py-6 space-y-8 max-w-4xl">
      <DashboardPageHeader label="Notifications" description="Manage your notification preferences and channels" />
      
      <Tabs defaultValue="channels" className="space-y-6">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="channels">Notification Channels</TabsTrigger>
          <TabsTrigger value="preferences">Notification Preferences</TabsTrigger>
        </TabsList>
        
        <TabsContent value="channels" className="space-y-6">
          <h3 className="text-lg font-medium">Integration Settings</h3>
          <p className="text-muted-foreground">
            Configure how your application sends notifications through various channels.
          </p>
          
          <NotificationChannelsTab />
        </TabsContent>
        
        <TabsContent value="preferences" className="space-y-6">
          <h3 className="text-lg font-medium">What Notifications You Receive</h3>
          <p className="text-muted-foreground">
            Customize what types of notifications you'd like to receive.
          </p>
          
          <NotificationPreferencesTab
            activityPreferences={activityPreferences}
            securityPreferences={securityPreferences}
            updatePreferences={updatePreferences}
          />
        </TabsContent>
      </Tabs>
      
      <div className="flex justify-end space-x-4 mt-8">
        <Button variant="outline" onClick={handleResetDefaults}>Reset to Defaults</Button>
        <Button disabled={isSubmitting} onClick={handleSaveAll}>
          {isSubmitting ? "Saving..." : "Save All Settings"}
        </Button>
      </div>
    </div>
  )
}

export default Page