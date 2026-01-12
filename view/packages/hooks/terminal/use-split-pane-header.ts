// Color palette for split panes
const PANE_COLORS = [
  { name: 'blue', active: '#60a5fa', inactive: '#3b82f6' },
  { name: 'emerald', active: '#34d399', inactive: '#10b981' },
  { name: 'amber', active: '#fbbf24', inactive: '#f59e0b' },
  { name: 'purple', active: '#a78bfa', inactive: '#8b5cf6' }
];

type UseSplitPaneHeaderProps = {
  paneIndex: number;
  isActive: boolean;
  totalPanes: number;
};

export const useSplitPaneHeader = ({
  paneIndex,
  isActive,
  totalPanes
}: UseSplitPaneHeaderProps) => {
  const color = PANE_COLORS[paneIndex % PANE_COLORS.length];
  const triangleColor = isActive ? color.active : color.inactive;
  const showTriangle = totalPanes > 1 && isActive;

  return {
    triangleColor,
    showTriangle
  };
};
