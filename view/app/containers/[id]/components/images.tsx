'use client';

import { useState } from 'react';
import {
  Layers,
  Calendar,
  HardDrive,
  Tag,
  Copy,
  Check,
  ChevronDown,
  Package,
  Clock
} from 'lucide-react';
import { Skeleton } from '@/components/ui/skeleton';
import { useTranslation } from '@/packages/hooks/shared/use-translation';
import { formatBytes } from '@/lib/utils';
import { useGetImagesQuery } from '@/redux/services/container/imagesApi';
import { formatDistanceToNow, format } from 'date-fns';
import { cn } from '@/lib/utils';

interface Image {
  id: string;
  repo_tags: string[];
  repo_digests: string[];
  created: number;
  size: number;
  shared_size: number;
  virtual_size: number;
  labels: Record<string, string>;
}

export function Images({ containerId, imagePrefix }: { containerId: string; imagePrefix: string }) {
  const { data: images = [], isLoading } = useGetImagesQuery({ containerId, imagePrefix });
  const { t } = useTranslation();

  if (isLoading) {
    return <ImagesSectionSkeleton />;
  }

  if (images.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-muted-foreground">
        <Layers className="h-12 w-12 mb-4 opacity-30" />
        <p className="text-sm">{t('containers.images.none')}</p>
      </div>
    );
  }

  const totalSize = images.reduce((acc, img) => acc + img.size, 0);
  const totalLayers = images.length;

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-8">
        <StatItem icon={Layers} value={totalLayers} label="Image Layers" />
        <StatItem icon={HardDrive} value={formatBytes(totalSize)} label="Total Size" />
      </div>

      <div className="space-y-4">
        {images.map((image, index) => (
          <ImageCard key={image.id} image={image} isFirst={index === 0} />
        ))}
      </div>
    </div>
  );
}

function StatItem({
  icon: Icon,
  value,
  label
}: {
  icon: React.ElementType;
  value: string | number;
  label: string;
}) {
  return (
    <div className="flex items-center gap-3">
      <div className="p-2 rounded-lg bg-muted/50">
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div>
        <p className="text-xl font-bold">{value}</p>
        <p className="text-xs text-muted-foreground">{label}</p>
      </div>
    </div>
  );
}

function ImageCard({ image, isFirst }: { image: Image; isFirst: boolean }) {
  const [expanded, setExpanded] = useState(isFirst);
  const [copied, setCopied] = useState<string | null>(null);

  const copyToClipboard = (text: string, key: string) => {
    navigator.clipboard.writeText(text);
    setCopied(key);
    setTimeout(() => setCopied(null), 2000);
  };

  const createdDate = new Date(image.created * 1000);
  const primaryTag = image.repo_tags?.[0] || '<none>';
  const hasLabels = image.labels && Object.keys(image.labels).length > 0;

  return (
    <div className="group">
      <button
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'w-full flex items-center gap-4 p-4 rounded-xl transition-colors text-left',
          'hover:bg-muted/30',
          expanded && 'bg-muted/20'
        )}
      >
        <div
          className={cn(
            'p-3 rounded-xl transition-colors',
            isFirst ? 'bg-emerald-500/10' : 'bg-muted/50'
          )}
        >
          <Package
            className={cn('h-5 w-5', isFirst ? 'text-emerald-500' : 'text-muted-foreground')}
          />
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <span className="font-semibold truncate">{primaryTag}</span>
            {isFirst && (
              <span className="px-2 py-0.5 text-[10px] font-medium uppercase tracking-wide rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
                Current
              </span>
            )}
          </div>
          <p className="text-xs text-muted-foreground mt-0.5 font-mono">
            {image.id.replace('sha256:', '').slice(0, 12)}
          </p>
        </div>

        <div className="hidden sm:flex items-center gap-6 text-sm text-muted-foreground">
          <span className="flex items-center gap-1.5">
            <HardDrive className="h-3.5 w-3.5" />
            {formatBytes(image.size)}
          </span>
          <span className="flex items-center gap-1.5">
            <Clock className="h-3.5 w-3.5" />
            {formatDistanceToNow(createdDate, { addSuffix: true })}
          </span>
        </div>

        <ChevronDown
          className={cn(
            'h-4 w-4 text-muted-foreground transition-transform',
            expanded && 'rotate-180'
          )}
        />
      </button>

      {expanded && (
        <div className="px-4 pb-4 pt-2 space-y-4 ml-[60px]">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <DetailRow
              icon={Tag}
              label="Image ID"
              value={image.id.replace('sha256:', '')}
              displayValue={image.id.replace('sha256:', '').slice(0, 24) + '...'}
              mono
              copyable
              onCopy={() => copyToClipboard(image.id, 'id')}
              copied={copied === 'id'}
            />
            <DetailRow icon={Calendar} label="Created" value={format(createdDate, 'PPpp')} />
            <DetailRow
              icon={HardDrive}
              label="Size"
              value={formatBytes(image.size)}
              sublabel={`Virtual: ${formatBytes(image.virtual_size)}`}
            />
            {image.shared_size > 0 && (
              <DetailRow icon={Layers} label="Shared Size" value={formatBytes(image.shared_size)} />
            )}
          </div>

          {image.repo_tags && image.repo_tags.length > 0 && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                Tags
              </p>
              <div className="flex flex-wrap gap-2">
                {image.repo_tags.map((tag, idx) => (
                  <div
                    key={idx}
                    className="group/tag flex items-center gap-2 px-3 py-1.5 rounded-lg bg-muted/30 text-sm"
                  >
                    <Tag className="h-3 w-3 text-muted-foreground" />
                    <span className="font-mono">{tag}</span>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        copyToClipboard(tag, `tag-${idx}`);
                      }}
                      className="opacity-0 group-hover/tag:opacity-100 transition-opacity text-muted-foreground hover:text-foreground"
                    >
                      {copied === `tag-${idx}` ? (
                        <Check className="h-3 w-3 text-emerald-500" />
                      ) : (
                        <Copy className="h-3 w-3" />
                      )}
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {image.repo_digests && image.repo_digests.length > 0 && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                Digests
              </p>
              <div className="space-y-1">
                {image.repo_digests.map((digest, idx) => (
                  <div
                    key={idx}
                    className="group/digest flex items-center gap-2 p-2 rounded-lg bg-zinc-950 text-zinc-400"
                  >
                    <code className="text-xs font-mono truncate flex-1">{digest}</code>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        copyToClipboard(digest, `digest-${idx}`);
                      }}
                      className="opacity-0 group-hover/digest:opacity-100 transition-opacity text-zinc-500 hover:text-zinc-300 flex-shrink-0"
                    >
                      {copied === `digest-${idx}` ? (
                        <Check className="h-3 w-3 text-emerald-500" />
                      ) : (
                        <Copy className="h-3 w-3" />
                      )}
                    </button>
                  </div>
                ))}
              </div>
            </div>
          )}

          {hasLabels && (
            <div className="space-y-2">
              <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
                Labels
              </p>
              <div className="grid grid-cols-1 gap-1">
                {Object.entries(image.labels).map(([key, value]) => (
                  <div key={key} className="flex items-start gap-2 py-1.5 text-sm">
                    <span className="text-muted-foreground font-mono text-xs">{key}:</span>
                    <span className="font-mono text-xs break-all">{value}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function DetailRow({
  icon: Icon,
  label,
  value,
  displayValue,
  sublabel,
  mono,
  copyable,
  onCopy,
  copied
}: {
  icon: React.ElementType;
  label: string;
  value: string;
  displayValue?: string;
  sublabel?: string;
  mono?: boolean;
  copyable?: boolean;
  onCopy?: () => void;
  copied?: boolean;
}) {
  return (
    <div className="flex items-start gap-2">
      <Icon className="h-4 w-4 mt-0.5 text-muted-foreground flex-shrink-0" />
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <div className="flex items-center gap-2">
          <span className={cn('text-sm truncate', mono && 'font-mono')} title={value}>
            {displayValue || value}
          </span>
          {copyable && onCopy && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                onCopy();
              }}
              className="text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
            >
              {copied ? (
                <Check className="h-3 w-3 text-emerald-500" />
              ) : (
                <Copy className="h-3 w-3" />
              )}
            </button>
          )}
        </div>
        {sublabel && <p className="text-xs text-muted-foreground/60">{sublabel}</p>}
      </div>
    </div>
  );
}

export function ImagesSectionSkeleton() {
  return (
    <div className="space-y-8">
      <div className="flex items-center gap-8">
        {[1, 2].map((i) => (
          <div key={i} className="flex items-center gap-3">
            <Skeleton className="h-10 w-10 rounded-lg" />
            <div className="space-y-1.5">
              <Skeleton className="h-6 w-16" />
              <Skeleton className="h-3 w-20" />
            </div>
          </div>
        ))}
      </div>

      <div className="space-y-4">
        {[1, 2, 3].map((i) => (
          <div key={i} className="flex items-center gap-4 p-4 rounded-xl bg-muted/10">
            <Skeleton className="h-12 w-12 rounded-xl" />
            <div className="flex-1 space-y-2">
              <Skeleton className="h-5 w-48" />
              <Skeleton className="h-3 w-24" />
            </div>
            <Skeleton className="h-4 w-20" />
            <Skeleton className="h-4 w-24" />
          </div>
        ))}
      </div>
    </div>
  );
}
