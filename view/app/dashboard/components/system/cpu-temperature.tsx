'use client';

import React from 'react';
import { Thermometer } from 'lucide-react';
import { SystemStatsType } from '@/redux/types/monitor';
import { TypographySmall, TypographyMuted } from '@/components/ui/typography';
import { SystemMetricCard } from './system-metric-card';
import { useSystemMetric } from '../../hooks/use-system-metric';
import { DEFAULT_METRICS } from '../utils/constants';
import { CPUTemperatureCardSkeletonContent } from './skeletons/cpu-temperature';

interface CPUTemperatureCardProps {
  systemStats: SystemStatsType | null;
}

interface CPUTemperatureDisplayProps {
  temperature: number;
  label: string;
}

interface TemperatureGaugeProps {
  temperature: number;
}

const getTemperatureColor = (temp: number): string => {
  if (temp < 50) return 'text-blue-500';
  if (temp < 70) return 'text-green-500';
  if (temp < 85) return 'text-orange-500';
  return 'text-red-500';
};

const getTemperatureStatus = (temp: number): string => {
  if (temp < 50) return 'Cool';
  if (temp < 70) return 'Normal';
  if (temp < 85) return 'Warm';
  return 'Hot';
};

const CPUTemperatureDisplay: React.FC<CPUTemperatureDisplayProps> = ({ temperature, label }) => {
  const tempColor = getTemperatureColor(temperature);
  const status = getTemperatureStatus(temperature);

  return (
    <div className="text-center">
      <TypographyMuted className="text-xs">{label}</TypographyMuted>
      <div className={`text-4xl font-bold ${tempColor} mt-1`}>
        {temperature > 0 ? `${temperature.toFixed(1)}°C` : 'N/A'}
      </div>
      {temperature > 0 && (
        <TypographyMuted className="text-xs mt-1">{status}</TypographyMuted>
      )}
    </div>
  );
};

const TemperatureGauge: React.FC<TemperatureGaugeProps> = ({ temperature }) => {
  // Calculate percentage based on 0-100°C range
  const percentage = Math.min((temperature / 100) * 100, 100);
  const gaugeColor = getTemperatureColor(temperature);

  return (
    <div className="space-y-2">
      <div className="flex justify-between text-xs text-muted-foreground">
        <span>0°C</span>
        <span>50°C</span>
        <span>100°C</span>
      </div>
      <div className="w-full bg-secondary rounded-full h-4 overflow-hidden">
        <div
          className={`h-full ${gaugeColor} transition-all duration-300 rounded-full`}
          style={{
            width: `${percentage}%`,
            backgroundColor: 'currentColor'
          }}
        />
      </div>
      <div className="flex justify-between text-xs">
        <TypographyMuted>Min: 0°C</TypographyMuted>
        <TypographyMuted>Max: {temperature > 0 ? `${temperature.toFixed(1)}°C` : 'N/A'}</TypographyMuted>
      </div>
    </div>
  );
};

const CPUTemperatureCard: React.FC<CPUTemperatureCardProps> = ({ systemStats }) => {
  const {
    data: cpu,
    isLoading,
    t
  } = useSystemMetric({
    systemStats,
    extractData: (stats) => stats.cpu,
    defaultData: DEFAULT_METRICS.cpu
  });

  const temperature = cpu.temperature || 0;

  return (
    <SystemMetricCard
      title={t('dashboard.cpu.temperature')}
      icon={Thermometer}
      isLoading={isLoading}
      skeletonContent={<CPUTemperatureCardSkeletonContent />}
    >
      <div className="space-y-6">
        <CPUTemperatureDisplay temperature={temperature} label={t('dashboard.cpu.current_temp')} />
        <TemperatureGauge temperature={temperature} />
        {temperature === 0 && (
          <div className="text-center">
            <TypographySmall className="text-muted-foreground">
              Temperature sensors not available on this system
            </TypographySmall>
          </div>
        )}
      </div>
    </SystemMetricCard>
  );
};

export default CPUTemperatureCard;

