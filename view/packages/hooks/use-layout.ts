import React, { useEffect } from 'react';
import { TERMINAL_POSITION } from '../types/layout';
import { useTour } from '@/packages/hooks/shared/useTour';
import { useAppSidebar } from '@/packages/hooks/shared/use-app-sidebar';
import { useTerminalState } from '@/packages/hooks/terminal/use-terminal-state';
import useTeamSwitcher from '@/packages/hooks/shared/use-team-switcher';
import useBreadCrumbs from '@/packages/hooks/shared/use-bread-crumbs';

export const useLayout = () => {
  const {
    user,
    activeOrg,
    hasAnyPermission,
    activeNav,
    refetch,
    t,
    filteredNavItems,
    setActiveNav
  } = useAppSidebar();
  const {
    addTeamModalOpen,
    setAddTeamModalOpen,
    toggleAddTeamModal,
    createTeam,
    teamName,
    teamDescription,
    isLoading,
    handleTeamNameChange,
    handleTeamDescriptionChange
  } = useTeamSwitcher();
  const { getBreadcrumbs } = useBreadCrumbs();
  const breadcrumbs = getBreadcrumbs();
  const { isTerminalOpen, toggleTerminal } = useTerminalState();
  const [TerminalPosition, setTerminalPosition] = React.useState<TERMINAL_POSITION>(() => {
    if (typeof window !== 'undefined') {
      const savedPosition = localStorage.getItem('terminalPosition');
      return (savedPosition as TERMINAL_POSITION) || TERMINAL_POSITION.BOTTOM;
    }
    return TERMINAL_POSITION.BOTTOM;
  });
  const [fitAddonRef, setFitAddonRef] = React.useState<any | null>(null);
  const { startTour } = useTour();

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 't' && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        setTerminalPosition((prevPosition) => {
          const newPosition =
            prevPosition === TERMINAL_POSITION.BOTTOM
              ? TERMINAL_POSITION.RIGHT
              : TERMINAL_POSITION.BOTTOM;
          localStorage.setItem('terminalPosition', newPosition);
          return newPosition;
        });
      }
    };
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return {
    user,
    activeOrg,
    hasAnyPermission,
    activeNav,
    refetch,
    t,
    filteredNavItems,
    setActiveNav,
    addTeamModalOpen,
    setAddTeamModalOpen,
    toggleAddTeamModal,
    createTeam,
    teamName,
    teamDescription,
    isLoading,
    handleTeamNameChange,
    handleTeamDescriptionChange,
    isTerminalOpen,
    toggleTerminal,
    TerminalPosition,
    setTerminalPosition,
    fitAddonRef,
    setFitAddonRef,
    startTour,
    breadcrumbs
  };
};
