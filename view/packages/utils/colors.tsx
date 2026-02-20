import React from 'react';
import { CheckCircle2, AlertCircle, Loader2, Clock } from 'lucide-react';

export interface ThemeColors {
  [key: string]: {
    background: string;
    foreground: string;
    card: string;
    'card-foreground': string;
    popover: string;
    'popover-foreground': string;
    primary: string;
    'primary-foreground': string;
    secondary: string;
    'secondary-foreground': string;
    muted: string;
    'muted-foreground': string;
    accent: string;
    'accent-foreground': string;
    destructive: string;
    'destructive-foreground': string;
    border: string;
    input: string;
    ring: string;
    radius?: string;
    'chart-1'?: string;
    'chart-2'?: string;
    'chart-3'?: string;
    'chart-4'?: string;
    'chart-5'?: string;
  };
}

export const palette = [
  'light',
  'dark'
];

export const themeColors: ThemeColors = {
  light: {
    background: '0 0% 100%',
    foreground: '240 10% 3.9%',
    card: '0 0% 100%',
    'card-foreground': '240 10% 3.9%',
    popover: '0 0% 100%',
    'popover-foreground': '240 10% 3.9%',
    primary: '240 5.9% 10%',
    'primary-foreground': '0 0% 98%',
    secondary: '240 4.8% 95.9%',
    'secondary-foreground': '240 5.9% 10%',
    muted: '240 4.8% 95.9%',
    'muted-foreground': '240 3.8% 46.1%',
    accent: '240 4.8% 95.9%',
    'accent-foreground': '240 5.9% 10%',
    destructive: '0 84.2% 60.2%',
    'destructive-foreground': '0 0% 98%',
    border: '240 5.9% 90%',
    input: '240 5.9% 90%',
    ring: '240 5.9% 10%',
    radius: '0.5rem',
    'chart-1': '12 76% 61%',
    'chart-2': '173 58% 39%',
    'chart-3': '197 37% 24%',
    'chart-4': '43 74% 66%',
    'chart-5': '27 87% 67%'
  },
  dark: {
    background: '20 14.3% 4.1%',
    foreground: '0 0% 95%',
    card: '24 9.8% 10%',
    'card-foreground': '0 0% 95%',
    popover: '0 0% 9%',
    'popover-foreground': '0 0% 95%',
    primary: '142.1 70.6% 45.3%',
    'primary-foreground': '144.9 80.4% 10%',
    secondary: '240 3.7% 15.9%',
    'secondary-foreground': '0 0% 98%',
    muted: '0 0% 15%',
    'muted-foreground': '240 5% 64.9%',
    accent: '12 6.5% 15.1%',
    'accent-foreground': '0 0% 98%',
    destructive: '0 62.8% 30.6%',
    'destructive-foreground': '0 85.7% 97.3%',
    border: '240 3.7% 15.9%',
    input: '240 3.7% 15.9%',
    ring: '142.4 71.8% 29.2%',
    'chart-1': '220 70% 50%',
    'chart-2': '160 60% 45%',
    'chart-3': '30 80% 55%',
    'chart-4': '280 65% 60%',
    'chart-5': '340 75% 55%'
  }
};

export function getDeploymentStatusIcon(status?: string): React.ReactElement {
  const statusLower = String(status || '').toLowerCase();
  
  switch (statusLower) {
    case 'deployed':
      return <CheckCircle2 className="h-4 w-4 text-primary" />;
    case 'failed':
      return <AlertCircle className="h-4 w-4 text-destructive" />;
    case 'in_progress':
    case 'building':
    case 'deploying':
      return <Loader2 className="h-4 w-4 text-primary animate-spin" />;
    default:
      return <Clock className="h-4 w-4 text-muted-foreground" />;
  }
}

export function getDeploymentStatusColor(status?: string): string {
  const statusLower = String(status || '').toLowerCase();
  
  switch (statusLower) {
    case 'deployed':
      return 'bg-primary';
    case 'failed':
      return 'bg-destructive';
    case 'in_progress':
    case 'building':
    case 'deploying':
      return 'bg-primary';
    default:
      return 'bg-muted';
  }
}

export function getDeploymentStatusBadgeClasses(status?: string): string {
  const statusLower = String(status || '').toLowerCase();
  
  switch (statusLower) {
    case 'deployed':
      return 'bg-primary/10 text-primary';
    case 'failed':
      return 'bg-destructive/10 text-destructive';
    case 'in_progress':
    case 'building':
    case 'deploying':
      return 'bg-primary/10 text-primary';
    default:
      return 'bg-muted text-muted-foreground';
  }
}

