import React from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { ExternalLink, Github, Star, GitFork, AlertCircle, Lock, Unlock } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { GithubRepository } from '@/redux/types/github';
import { Skeleton } from '@/components/ui/skeleton';

const GithubRepositories = ({
  name,
  html_url: url,
  description,
  private: isPrivate,
  stargazers_count,
  forks_count,
  open_issues_count,
  license,
  topics,
  setSelectedRepository
}: GithubRepository & { setSelectedRepository: (repo: string) => void }) => {
  return (
    <Card
      className="group relative  w-full max-w-md cursor-pointer overflow-hidden transition-all duration-300 hover:bg-muted hover:shadow-lg"
      onClick={() => setSelectedRepository(name)}
    >
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center gap-2 text-lg font-bold">
          <Github className="text-primary" size={24} />
          {name || 'Unnamed Repository'}
          {url && (
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              className="ml-auto text-muted-foreground transition-colors duration-200 hover:text-primary"
              title="View on GitHub"
            >
              <ExternalLink size={18} />
            </a>
          )}
        </CardTitle>
        {description && (
          <CardDescription className="line-clamp-2 text-sm text-muted-foreground">
            {description}
          </CardDescription>
        )}
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={isPrivate ? 'secondary' : 'outline'} className="text-xs font-medium">
            {isPrivate ? (
              <Lock size={12} className="mr-1" />
            ) : (
              <Unlock size={12} className="mr-1" />
            )}
            {isPrivate ? 'Private' : 'Public'}
          </Badge>
          {license && license.spdx_id && (
            <Badge variant="outline" className="text-xs font-medium">
              {license.spdx_id}
            </Badge>
          )}
        </div>
        <div className="flex items-center gap-4 text-sm text-muted-foreground">
          <div className="flex items-center gap-1">
            <Star size={16} />
            <span>{stargazers_count?.toLocaleString() || '0'}</span>
          </div>
          <div className="flex items-center gap-1">
            <GitFork size={16} />
            <span>{forks_count?.toLocaleString() || '0'}</span>
          </div>
          <div className="flex items-center gap-1">
            <AlertCircle size={16} />
            <span>{open_issues_count?.toLocaleString() || '0'}</span>
          </div>
        </div>
        {topics && topics.length > 0 && (
          <div className="flex flex-wrap items-center gap-2">
            {topics.slice(0, 2).map((topic) => (
              <Badge key={topic} variant="secondary" className="text-xs font-medium">
                {topic}
              </Badge>
            ))}
            {topics.length > 2 && (
              <Badge variant="secondary" className="text-xs font-medium">
                +{topics.length - 2} more
              </Badge>
            )}
          </div>
        )}
      </CardContent>
    </Card>
  );
};

export default GithubRepositories;

export const GithubRepositoriesSkeletonLoader: React.FC = () => {
  return (
    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-3">
      {[...Array(6)].map((_, index) => (
        <Card
          key={index}
          className="group relative w-full max-w-md overflow-hidden transition-all duration-300 hover:bg-muted hover:shadow-lg"
        >
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center gap-2 text-lg font-bold">
              <Skeleton className="h-6 w-6 rounded-full" />
              <Skeleton className="h-6 w-40" />
              <Skeleton className="ml-auto h-6 w-6 rounded-full" />
            </CardTitle>
            <CardDescription>
              <Skeleton className="mt-2 h-4 w-full" />
              <Skeleton className="mt-1 h-4 w-3/4" />
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <Skeleton className="h-5 w-16 rounded-full" />
              <Skeleton className="h-5 w-20 rounded-full" />
            </div>
            <div className="flex items-center gap-4 text-sm">
              <Skeleton className="h-4 w-12" />
              <Skeleton className="h-4 w-12" />
              <Skeleton className="h-4 w-12" />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Skeleton className="h-5 w-16 rounded-full" />
              <Skeleton className="h-5 w-16 rounded-full" />
              <Skeleton className="h-5 w-16 rounded-full" />
              <Skeleton className="h-5 w-16 rounded-full" />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
};
