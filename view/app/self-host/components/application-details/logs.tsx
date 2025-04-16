'use client';

import React from 'react';
import './logViewer.css';
import LogViewer from './log-viewer';

const ApplicationLogs = ({
  id,
  currentPage,
  setCurrentPage
}: {
  id: string;
  currentPage: number;
  setCurrentPage: (page: number) => void;
}) => {
  return <LogViewer id={id} currentPage={currentPage} setCurrentPage={setCurrentPage} />;
};

export default ApplicationLogs;
