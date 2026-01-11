export const terminalStyles = `
  .xterm-viewport::-webkit-scrollbar {
    width: 8px;
  }
  .xterm-viewport::-webkit-scrollbar-track {
    background: transparent;
  }
  .xterm-viewport::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
  }
  .xterm-viewport::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.2);
  }
  
  /* Ensure terminal respects parent container */
  .terminal-container {
    --terminal-bg: #0c0c0f;
    --terminal-header-bg: rgba(18, 18, 22, 0.95);
    --terminal-border: rgba(255, 255, 255, 0.06);
    --terminal-tab-active: rgba(34, 211, 238, 0.1);
    --terminal-tab-hover: rgba(255, 255, 255, 0.05);
    --terminal-accent: #22d3ee;
    --terminal-text: #e4e4e7;
    --terminal-text-muted: #71717a;
    --terminal-glow: 0 0 20px rgba(34, 211, 238, 0.15);
    --terminal-split-border: #3a3a3a;
    --terminal-split-active: #4a4a4a;
    width: 100%;
    max-width: 100%;
    box-sizing: border-box;
    contain: inline-size;
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
    background-color: #5a5a5a !important;
    width: 4px;
  }

  .terminal-container [data-panel-resize-handle-id]:focus,
  .terminal-container [data-panel-resize-handle-id]:focus-visible {
    outline: none !important;
    box-shadow: none !important;
    border: none !important;
  }

  /* Add a visual indicator when dragging */
  .terminal-container [data-panel-resize-handle-id][data-resize-handle-active] {
    background-color: #5a5a5a !important;
    box-shadow: 0 0 8px rgba(90, 90, 90, 0.4);
    width: 4px;
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
      box-shadow: 0 0 6px rgba(16, 185, 129, 0.6), 0 0 12px rgba(16, 185, 129, 0.3);
      transform: scale(1);
    }
    50% {
      box-shadow: 0 0 10px rgba(16, 185, 129, 0.8), 0 0 20px rgba(16, 185, 129, 0.4);
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
