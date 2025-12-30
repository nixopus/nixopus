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
