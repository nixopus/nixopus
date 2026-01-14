export type TerminalPaneProps = {
  isActive: boolean;
  isTerminalOpen: boolean;
  canCreate: boolean;
  canUpdate: boolean;
  setFitAddonRef: React.Dispatch<React.SetStateAction<any | null>>;
  terminalId: string;
  onFocus: () => void;
  onStatusChange?: (status: SessionStatus) => void;
  exitHandler?: ExitHandler;
};

export type SessionTabProps = {
  session: {
    id: string;
    label: string;
  };
  isActive: boolean;
  status: SessionStatus;
  onSelect: () => void;
  onClose: () => void;
  canClose: boolean;
  index: number;
};

export type SplitPaneHeaderProps = {
  paneIndex: number;
  isActive: boolean;
  canClose: boolean;
  totalPanes: number;
  onFocus: () => void;
  onClose: () => void;
  closeLabel: string;
};

export type TerminalHeaderProps = {
  sessions: Session[];
  activeSessionId: string;
  sessionStatuses: Record<string, SessionStatus>;
  sessionLimit: number;
  maxSplits: number;
  splitPanesCount: number;
  onAddSession: () => void;
  onCloseSession: (id: string) => void;
  onSwitchSession: (id: string) => void;
  onToggleTerminal: () => void;
  onAddSplitPane: () => void;
  terminalPosition: 'bottom' | 'right';
  onTogglePosition: () => void;
  closeLabel: string;
  newTabLabel: string;
};

export type ExitHandler = {
  splitPanesCount: number;
  sessionsCount: number;
  activePaneId: string | null;
  activeSessionId: string | null;
  onCloseSplitPane?: (paneId: string) => void;
  onCloseSession?: (sessionId: string) => void;
  onToggleTerminal?: () => void;
};

export type SessionStatus = 'active' | 'idle' | 'loading';

export type SplitPane = {
  id: string;
  label: string;
  terminalId: string;
};

export type Session = {
  id: string;
  label: string;
  splitPanes: SplitPane[];
};

export type TerminalOutput = {
  data: {
    output_type: string;
    content: string;
  };
  topic: string;
};
