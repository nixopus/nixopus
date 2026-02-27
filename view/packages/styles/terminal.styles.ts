export const terminalStyles = `
  .xterm-viewport::-webkit-scrollbar {
    width: 8px;
  }
  .xterm-viewport::-webkit-scrollbar-track {
    background: transparent;
  }
  .xterm-viewport::-webkit-scrollbar-thumb {
    background: color-mix(in oklch, var(--foreground) 15%, transparent);
    border-radius: 4px;
  }
  .xterm-viewport::-webkit-scrollbar-thumb:hover {
    background: color-mix(in oklch, var(--foreground) 25%, transparent);
  }
  .dark .xterm-viewport::-webkit-scrollbar-thumb {
    background: color-mix(in oklch, var(--foreground) 10%, transparent);
  }
  .dark .xterm-viewport::-webkit-scrollbar-thumb:hover {
    background: color-mix(in oklch, var(--foreground) 20%, transparent);
  }
  
  /* Ensure terminal respects parent container */
  .terminal-container {
    --terminal-bg: var(--card);
    --terminal-header-bg: color-mix(in oklch, var(--card) 98%, transparent);
    --terminal-border: var(--border);
    --terminal-tab-active: color-mix(in oklch, var(--accent) 10%, transparent);
    --terminal-tab-hover: color-mix(in oklch, var(--foreground) 4%, transparent);
    --terminal-accent: var(--accent);
    --terminal-text: var(--card-foreground);
    --terminal-text-muted: var(--muted-foreground);
    --terminal-glow: 0 0 20px color-mix(in oklch, var(--accent) 15%, transparent);
    --terminal-split-border: var(--border);
    --terminal-split-active: color-mix(in oklch, var(--foreground) 15%, transparent);
    --terminal-status-loading: var(--chart-4);
    --terminal-status-active: var(--chart-1);
    --terminal-status-idle: var(--muted-foreground);
    --terminal-close-hover-bg: color-mix(in oklch, var(--destructive) 10%, transparent);
    --terminal-close-hover-text: var(--destructive);
    width: 100%;
    max-width: 100%;
    box-sizing: border-box;
    contain: inline-size;
  }

  .dark .terminal-container {
    --terminal-bg: var(--card);
    --terminal-header-bg: color-mix(in oklch, var(--card) 95%, transparent);
    --terminal-border: var(--border);
    --terminal-tab-active: color-mix(in oklch, var(--accent) 10%, transparent);
    --terminal-tab-hover: color-mix(in oklch, var(--foreground) 5%, transparent);
    --terminal-accent: var(--accent);
    --terminal-text: var(--card-foreground);
    --terminal-text-muted: var(--muted-foreground);
    --terminal-glow: 0 0 20px color-mix(in oklch, var(--accent) 15%, transparent);
    --terminal-split-border: var(--border);
    --terminal-split-active: color-mix(in oklch, var(--foreground) 20%, transparent);
    --terminal-status-loading: var(--chart-4);
    --terminal-status-active: var(--chart-1);
    --terminal-status-idle: var(--muted-foreground);
    --terminal-close-hover-bg: color-mix(in oklch, var(--destructive) 10%, transparent);
    --terminal-close-hover-text: var(--destructive);
  }
  
  /* Ensure xterm doesn't cause overflow - constrain all xterm elements */
  .terminal-container .xterm,
  .terminal-container .xterm-screen,
  .terminal-container .xterm-viewport,
  .terminal-container .xterm-rows {
    width: 100% !important;
    max-width: 100% !important;
    overflow-x: hidden !important;
  }
  
  .terminal-container .xterm-helper-textarea {
    position: absolute !important;
  }
  
  /* Prevent canvas from expanding */
  .terminal-container canvas {
    max-width: 100% !important;
  }
  
  /* Split pane styles */
  .terminal-split-pane {
    position: relative;
    display: flex;
    flex-direction: column;
    height: 100%;
    width: 100%;
    overflow: hidden;
  }
  
  .terminal-split-pane-header {
    flex-shrink: 0;
    user-select: none;
    height: 24px;
    min-height: 24px;
    backdrop-filter: blur(8px);
  }
  
  .terminal-split-pane-content {
    flex: 1;
    overflow: hidden;
    position: relative;
  }
  
  /* Resizable handle improvements */
  .terminal-container [data-panel-resize-handle-id] {
    transition: background-color 0.2s ease;
    width: 2px;
    background-color: var(--terminal-split-border) !important;
  }

  .terminal-container [data-panel-resize-handle-id]:hover {
    background-color: var(--terminal-split-active) !important;
    width: 3px;
  }

  .terminal-container [data-panel-resize-handle-id]:active {
    background-color: var(--terminal-split-active) !important;
    width: 4px;
  }

  .dark .terminal-container [data-panel-resize-handle-id]:active {
    background-color: var(--terminal-split-active) !important;
  }

  .terminal-container [data-panel-resize-handle-id]:focus,
  .terminal-container [data-panel-resize-handle-id]:focus-visible {
    outline: none !important;
    box-shadow: none !important;
    border: none !important;
  }

  /* Add a visual indicator when dragging */
  .terminal-container [data-panel-resize-handle-id][data-resize-handle-active] {
    background-color: var(--terminal-split-active) !important;
    box-shadow: 0 0 8px color-mix(in oklch, var(--foreground) 20%, transparent);
    width: 4px;
  }

  .dark .terminal-container [data-panel-resize-handle-id][data-resize-handle-active] {
    background-color: var(--terminal-split-active) !important;
    box-shadow: 0 0 8px color-mix(in oklch, var(--foreground) 20%, transparent);
  }
  
  @keyframes terminalFadeIn {
    from {
      opacity: 0;
      transform: translateY(4px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
  
  @keyframes pulseGlow {
    0%, 100% {
      box-shadow: 0 0 6px color-mix(in oklch, var(--terminal-status-active) 60%, transparent), 0 0 12px color-mix(in oklch, var(--terminal-status-active) 30%, transparent);
      transform: scale(1);
    }
    50% {
      box-shadow: 0 0 10px color-mix(in oklch, var(--terminal-status-active) 80%, transparent), 0 0 20px color-mix(in oklch, var(--terminal-status-active) 40%, transparent);
      transform: scale(1.1);
    }
  }
  
  .terminal-ready-indicator {
    animation: pulseGlow 2s ease-in-out infinite;
    filter: brightness(1.1);
  }
  
  .terminal-tab-active::before {
    content: '';
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 60%;
    height: 2px;
    background: linear-gradient(90deg, transparent, var(--terminal-accent), transparent);
    border-radius: 2px;
  }
  
  /* No scrollbar utility class */
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
`;
