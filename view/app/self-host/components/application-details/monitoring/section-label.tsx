'use client';

interface SectionLabelProps {
  children: React.ReactNode;
}

export function SectionLabel({ children }: SectionLabelProps) {
  return (
    <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
      {children}
    </h3>
  );
}
