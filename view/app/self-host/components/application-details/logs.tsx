'use client';

import React from 'react';
import './logViewer.css';
import { ApplicationLogs as ApplicationLogsType } from '@/redux/types/applications';
import LogViewer from './log-viewer';

const ApplicationLogs = ({
  logs,
  onRefresh,
  currentPage,
  setCurrentPage
}: {
  logs?: ApplicationLogsType[];
  onRefresh: () => void;
  currentPage: number;
  setCurrentPage: (page: number) => void;
}) => {
  return (
    <LogViewer
      logs={
        logs && logs.length > 0
          ? logs
              .slice()
              .map((logEntry) => {
                if (!logEntry || !logEntry.log) {
                  return '';
                }

                const timestamp = logEntry.created_at
                  ? new Date(logEntry.created_at).toLocaleString()
                  : '';

                const containerInfo = logEntry.application_id || '';

                return `${timestamp} ${containerInfo}: ${logEntry.log}`.trim();
              })
              .join('\n')
          : ''
      }
      onRefresh={onRefresh}
      currentPage={currentPage}
      setCurrentPage={setCurrentPage}
    />
  );
};

export default ApplicationLogs;
