'use client';

import React from 'react';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Cell } from 'recharts';

export interface BarChartDataItem {
  name: string;
  value: number;
  fill: string;
}

interface BarChartComponentProps {
  data: BarChartDataItem[];
  chartConfig: Record<string, { label: string; color: string }>;
  height?: string;
  yAxisLabel?: string;
  xAxisLabel?: string;
  showAxisLabels?: boolean;
}

export const BarChartComponent: React.FC<BarChartComponentProps> = ({
  data,
  chartConfig,
  height = 'h-[180px]',
  yAxisLabel = 'Value',
  xAxisLabel = 'Category',
  showAxisLabels = true
}) => {
  // Calculate max value for Y-axis (round up with 20% padding)
  const maxValue = Math.max(...data.map((item) => item.value));
  const yAxisMax = Math.ceil(maxValue * 1.2);

  return (
    <ChartContainer config={chartConfig} className={`${height} w-full`}>
      <BarChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
        <CartesianGrid strokeDasharray="3 3" vertical={false} />
        <XAxis
          dataKey="name"
          tickLine={false}
          axisLine={false}
          fontSize={12}
          label={
            showAxisLabels
              ? {
                  value: xAxisLabel,
                  position: 'insideBottom',
                  offset: -5,
                  style: { fontSize: '11px', fill: 'hsl(var(--muted-foreground))' }
                }
              : undefined
          }
        />
        <YAxis
          tickLine={false}
          axisLine={false}
          fontSize={12}
          domain={[0, yAxisMax]}
          tickFormatter={(value) => value.toFixed(1)}
          label={
            showAxisLabels
              ? {
                  value: yAxisLabel,
                  angle: -90,
                  position: 'insideLeft',
                  style: { fontSize: '11px', fill: 'hsl(var(--muted-foreground))' }
                }
              : undefined
          }
        />
        <ChartTooltip cursor={{ fill: 'hsl(var(--muted))' }} content={<ChartTooltipContent />} />
        <Bar dataKey="value" radius={[8, 8, 0, 0]} maxBarSize={60}>
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.fill} />
          ))}
        </Bar>
      </BarChart>
    </ChartContainer>
  );
};
