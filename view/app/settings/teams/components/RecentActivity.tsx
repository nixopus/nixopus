import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';

function RecentActivity() {
  return (
    <Card className="col-span-1 lg:col-span-2">
      <CardHeader>
        <CardTitle>Recent Activities</CardTitle>
        <CardDescription>Recent changes to your organization</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          <div className="flex items-start gap-4">
            <div className="h-2 w-2 mt-2 rounded-full bg-blue-500"></div>
            <div>
              <p className="text-sm font-medium">Jane Smith joined the organization</p>
              <p className="text-xs text-muted-foreground">3 hours ago</p>
            </div>
          </div>
          <div className="flex items-start gap-4">
            <div className="h-2 w-2 mt-2 rounded-full bg-green-500"></div>
            <div>
              <p className="text-sm font-medium">Bob Johnson's role changed to Viewer</p>
              <p className="text-xs text-muted-foreground">Yesterday</p>
            </div>
          </div>
          <div className="flex items-start gap-4">
            <div className="h-2 w-2 mt-2 rounded-full bg-red-500"></div>
            <div>
              <p className="text-sm font-medium">Alex Wilson was removed from the organization</p>
              <p className="text-xs text-muted-foreground">Last week</p>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default RecentActivity;
