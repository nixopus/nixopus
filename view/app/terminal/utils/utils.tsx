export const handleControlKey = (key: string): string => {
  const controlKeys: Record<string, string> = {
    c: '\x03',
    d: '\x04',
    z: '\x1A',
    o: '\x0F',
    m: '\x0D',
    x: '\x18'
  };
  return controlKeys[key] || '';
};

export const specialKeyCodeMappings = {
  38: '\x1b[A',
  40: '\x1b[B',
  39: '\x1b[C',
  37: '\x1b[D',
  36: '\x1bOH',
  35: '\x1bOF',
  46: '\x1b[3~',
  27: '\x1b'
};

export const handleEnterKey = (term: any, currentLine: string, sendJsonMessage: any) => {
  sendJsonMessage({
    action: 'terminal',
    data: currentLine
  });
  return;
};

export const handleTabKey = (currentLine: string, sendJsonMessage: any, term: any) => {
  sendJsonMessage({
    TerminalCommand: {
      command: currentLine
    }
  });
  return currentLine;
};

export const handleBackspaceKey = (term: any, currentLine: string) => {
  if (currentLine.length > 0) {
    term.write('\b \b');
    return currentLine.slice(0, -1);
  }
  return currentLine;
};

export const handlePrintableKey = (term: any, currentLine: string, key: string) => {
  term.write(key);
};
