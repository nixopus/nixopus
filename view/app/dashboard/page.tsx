'use client';

import React from 'react';
import use_monitor from './hooks/use_monitor';
import ContainersTable from './components/container_table';
import SystemStats from './components/system_stats';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Package } from 'lucide-react';
import DiskUsageCard from './components/diisk_usage';

function DashboardPage() {
  const { containersData, systemStats } = use_monitor();

  return (
    <div className="p-4 md:p-6 space-y-4 md:space-y-6">
      <div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-2">
        <div>
          <h1 className="text-xl sm:text-2xl font-bold">Dashboard</h1>
          <p className="text-sm sm:text-base text-muted-foreground">Monitor your server</p>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>{systemStats && <SystemStats systemStats={systemStats} />}</div>
        {systemStats && <DiskUsageCard systemStats={systemStats} />}
      </div>
      <Card>
        <CardHeader>
          <CardTitle className="text-xs sm:text-sm font-medium flex items-center">
            <Package className="h-3 w-3 sm:h-4 sm:w-4 mr-1 sm:mr-2" />
            Containers
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ContainersTable containersData={containersData} />
        </CardContent>
      </Card>
    </div>
  );
}

export default DashboardPage;
