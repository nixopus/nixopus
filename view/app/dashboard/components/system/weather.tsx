'use client';

import React from 'react';
import { Cloud, CloudRain, CloudSnow, Sun, CloudDrizzle, Wind } from 'lucide-react';
import { SystemMetricCard } from './system-metric-card';
import { WeatherCardSkeletonContent } from './skeletons/weather';
import useWeather from '../../hooks/use-weather';

const getWeatherIcon = (condition: string) => {
  const lowerCondition = condition.toLowerCase();
  if (lowerCondition.includes('rain')) return CloudRain;
  if (lowerCondition.includes('snow')) return CloudSnow;
  if (lowerCondition.includes('drizzle')) return CloudDrizzle;
  if (lowerCondition.includes('wind')) return Wind;
  if (lowerCondition.includes('clear') || lowerCondition.includes('sun')) return Sun;
  return Cloud;
};

interface WeatherWidgetProps {
  latitude?: number;
  longitude?: number;
  location?: string;
}

const WeatherWidget: React.FC<WeatherWidgetProps> = ({
  latitude,
  longitude,
  location = 'Default Location'
}) => {
  const { weather, isLoading, error } = useWeather({ latitude, longitude, location });

  if (isLoading) {
    return (
      <SystemMetricCard
        title="Weather"
        icon={Cloud}
        isLoading={true}
        skeletonContent={<WeatherCardSkeletonContent />}
      >
        <div className="flex flex-col items-center justify-center h-full space-y-3">
          <div className="text-5xl font-bold text-primary tabular-nums">
            --
          </div>
          <div className="text-sm text-muted-foreground">
            Loading...
          </div>
        </div>
      </SystemMetricCard>
    );
  }

  if (error && !weather) {
    return (
      <SystemMetricCard
        title="Weather"
        icon={Cloud}
        isLoading={false}
        skeletonContent={<WeatherCardSkeletonContent />}
      >
        <div className="flex flex-col items-center justify-center h-full space-y-2 text-center">
          <div className="text-sm text-muted-foreground">
            {error}
          </div>
        </div>
      </SystemMetricCard>
    );
  }

  const Icon = weather ? getWeatherIcon(weather.condition) : Cloud;

  return (
    <SystemMetricCard
      title="Weather"
      icon={Cloud}
      isLoading={false}
      skeletonContent={<WeatherCardSkeletonContent />}
    >
      {weather && (
        <div className="flex flex-col items-center justify-center h-full space-y-3">
          <Icon className="h-12 w-12 text-blue-500" />
          <div className="text-5xl font-bold text-primary tabular-nums">
            {weather.temperature}°F
          </div>
          <div className="text-sm text-center text-muted-foreground capitalize">
            {weather.description}
          </div>
          <div className="flex gap-4 text-xs text-muted-foreground">
            <span>H: {weather.humidity}%</span>
            <span>W: {weather.windSpeed} mph</span>
          </div>
        </div>
      )}
    </SystemMetricCard>
  );
};

export default WeatherWidget;

