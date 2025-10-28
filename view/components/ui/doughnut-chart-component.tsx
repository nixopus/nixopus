'use client';

import React from 'react';
import { ChartContainer, ChartTooltip, ChartTooltipContent } from '@/components/ui/chart';
import { PieChart, Pie, Cell, Label } from 'recharts';

export interface DoughnutChartDataItem {
  name: string;
  value: number;
  fill: string;
}

interface DoughnutChartComponentProps {
  data: DoughnutChartDataItem[];
  chartConfig: Record<string, { label: string; color: string }>;
  centerLabel?: {
    value: string;
    subLabel?: string;
  };
  innerRadius?: number;
  outerRadius?: number;
  maxHeight?: string;
}

export const DoughnutChartComponent: React.FC<DoughnutChartComponentProps> = ({
  data,
  chartConfig,
  centerLabel,
  innerRadius = 60,
  outerRadius = 80,
  maxHeight = 'max-h-[200px]'
}) => {
  return (
    <ChartContainer config={chartConfig} className={`mx-auto aspect-square ${maxHeight} w-full`}>
      <PieChart width={200} height={200}>
        <ChartTooltip cursor={false} content={<ChartTooltipContent hideLabel />} />
        <Pie
          data={data}
          dataKey="value"
          nameKey="name"
          innerRadius={innerRadius}
          outerRadius={outerRadius}
          strokeWidth={5}
          cx="50%"
          cy="50%"
        >
          {data.map((entry, index) => (
            <Cell key={`cell-${index}`} fill={entry.fill} />
          ))}
          {centerLabel && (
            <Label
              content={({ viewBox }) => {
                if (viewBox && 'cx' in viewBox && 'cy' in viewBox) {
                  return (
                    <text
                      x={viewBox.cx}
                      y={viewBox.cy}
                      textAnchor="middle"
                      dominantBaseline="middle"
                    >
                      <tspan
                        x={viewBox.cx}
                        y={viewBox.cy}
                        className="fill-foreground text-2xl font-bold"
                      >
                        {centerLabel.value}
                      </tspan>
                      {centerLabel.subLabel && (
                        <tspan
                          x={viewBox.cx}
                          y={(viewBox.cy || 0) + 20}
                          className="fill-muted-foreground text-xs"
                        >
                          {centerLabel.subLabel}
                        </tspan>
                      )}
                    </text>
                  );
                }
              }}
            />
          )}
        </Pie>
      </PieChart>
    </ChartContainer>
  );
};
