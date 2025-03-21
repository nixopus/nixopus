import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';

export const initializeXtermTerminal = (
    container: HTMLDivElement,
    resolvedTheme: string,
    themeColors: any,
    terminalContent: string[],
    sendJsonMessage: (message: any) => void,
) => {
    const term = new Terminal({
        cursorBlink: true,
        fontFamily: '"Menlo", "DejaVu Sans Mono", "Consolas", monospace',
        fontSize: 14,
        theme: {
            foreground: `hsl(${themeColors[resolvedTheme ?? 'light'].foreground})`,
            background: `hsl(${themeColors[resolvedTheme ?? 'light'].background})`,
            cursor: `hsl(${themeColors[resolvedTheme ?? 'light'].foreground})`,
        },
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);

    term.open(container);
    fitAddon.fit();
    container.style.padding = '10px';

    if (terminalContent.length > 0) {
        terminalContent.forEach((content) => formatAndWriteOutput(term, content));
    }

    let currentLine = '';

    term.onKey(({ key, domEvent }) => {
        const printable = !domEvent.altKey && !domEvent.ctrlKey && !domEvent.metaKey;

        if (domEvent.keyCode === 13) {
            sendJsonMessage({
                TerminalCommand: {
                    command: currentLine,
                },
            });
            term.write('\r\n');
            currentLine = '';
        } else if (domEvent.keyCode === 8) {
            if (currentLine.length > 0) {
                currentLine = currentLine.slice(0, -1);
                term.write('\b \b');
            }
        } else if (printable) {
            currentLine += key;
            term.write(key);
        } else {
            let command = '';
            if (domEvent.ctrlKey) {
                switch (domEvent.key.toLowerCase()) {
                    case 'c':
                        command = '\x03';
                        break;
                    case 'd':
                        command = '\x04';
                        break;
                    case 'z':
                        command = '\x1A';
                        break;
                    case 'o':
                        command = '\x0F';
                        break;
                    case 'm':
                        command = '\x0D';
                        break;
                    case 'x':
                        command = '\x18';
                        break;
                }
            } else {
                switch (domEvent.key) {
                    case 'ArrowUp':
                        command = '\x1b[A';
                        break;
                    case 'ArrowDown':
                        command = '\x1b[B';
                        break;
                    case 'ArrowRight':
                        command = '\x1b[C';
                        break;
                    case 'ArrowLeft':
                        command = '\x1b[D';
                        break;
                    case 'Home':
                        command = '\x1bOH';
                        break;
                    case 'End':
                        command = '\x1bOF';
                        break;
                    case 'Delete':
                        command = '\x1b[3~';
                        break;
                    case 'Tab':
                        command = '\t';
                        break;
                    case 'Escape':
                        command = '\x1b';
                        break;
                }
            }
            if (command) {
                sendJsonMessage({
                    TerminalCommand: {
                        command: command,
                    },
                });
            }
        }
    });

    return {
        ...term,
        writeOutput: (output: string) => formatAndWriteOutput(term, output),
        setTheme: (resolvedTheme: string, themeColors: any) => {
            const newColors = {
                foreground: `hsl(${themeColors[resolvedTheme ?? 'light']?.foreground})`,
                background: `hsl(${themeColors[resolvedTheme ?? 'light']?.background})`,
                cursor: `hsl(${themeColors[resolvedTheme ?? 'light']?.foreground})`,
            };
            term.options.theme = { ...newColors };
        },
    };
};

const formatAndWriteOutput = (term: Terminal, output: string) => {
    const lines = output.split('\n');
    lines.forEach((line, index) => {
        const lineEnding = index < lines.length - 1 ? '\r\n' : '';
        term.write(line + lineEnding);
    });
};
