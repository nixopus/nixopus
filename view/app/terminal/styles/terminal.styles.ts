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
      box-shadow: 0 0 8px rgba(34, 211, 238, 0.3);
    }
    50% {
      box-shadow: 0 0 16px rgba(34, 211, 238, 0.5);
    }
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
  
  .terminal-ready-indicator {
    animation: pulseGlow 2s ease-in-out infinite;
  }
`;
