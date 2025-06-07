---
outline: deep
---

# Terminal

Experience the power of a secure, web-based terminal that brings the convenience of cloud shells to your fingertips. Inspired by [VS Code's terminal](https://code.visualstudio.com/docs/terminal/basics) and similar to [Google Cloud Shell](https://cloud.google.com/shell) and [AWS Cloud Shell](https://aws.amazon.com/cloudshell/), our terminal ensures safe command execution while providing a seamless interface.

## Architecture

We have built nixopus terminal using the following architecture:

<div style="margin: 2rem 0; padding: 1rem; border-radius: 8px;">
<svg viewBox="0 0 800 500" xmlns="http://www.w3.org/2000/svg">
  <rect x="0" y="0" width="800" height="500" fill="var(--vp-c-bg)" />
  
  <rect x="50" y="50" width="300" height="170" rx="10" ry="10" fill="var(--vp-c-brand-soft)" stroke="var(--vp-c-brand)" stroke-width="2" />
  <text x="200" y="80" font-family="Arial" font-size="18" text-anchor="middle" font-weight="bold" fill="var(--vp-c-text-1)">nixopus-view</text>
  
  <rect x="450" y="50" width="300" height="170" rx="10" ry="10" fill="var(--vp-c-success-soft)" stroke="var(--vp-c-success)" stroke-width="2" />
  <text x="600" y="80" font-family="Arial" font-size="18" text-anchor="middle" font-weight="bold" fill="var(--vp-c-text-1)">nixopus-api</text>
  
  <rect x="320" y="280" width="160" height="80" rx="10" ry="10" fill="var(--vp-c-danger-soft)" stroke="var(--vp-c-danger)" stroke-width="2" />
  <text x="400" y="325" font-family="Arial" font-size="16" text-anchor="middle" fill="var(--vp-c-text-1)">nixopus-realtime</text>
  
  <rect x="620" y="120" width="120" height="60" rx="5" ry="5" fill="var(--vp-c-success-soft)" stroke="var(--vp-c-success)" stroke-width="2" />
  <text x="680" y="155" font-family="Arial" font-size="14" text-anchor="middle" fill="var(--vp-c-text-1)">goph ssh</text>
  
  <rect x="80" y="120" width="120" height="60" rx="5" ry="5" fill="var(--vp-c-brand-soft)" stroke="var(--vp-c-brand)" stroke-width="2" />
  <text x="140" y="155" font-family="Arial" font-size="14" text-anchor="middle" fill="var(--vp-c-text-1)">XTerm.js</text>
  
  <rect x="220" y="120" width="120" height="60" rx="5" ry="5" fill="var(--vp-c-brand-soft)" stroke="var(--vp-c-brand)" stroke-width="2" />
  <text x="280" y="155" font-family="Arial" font-size="14" text-anchor="middle" fill="var(--vp-c-text-1)">React-DOM</text>
  
  <rect x="530" y="400" width="220" height="60" rx="10" ry="10" fill="var(--vp-c-warning-soft)" stroke="var(--vp-c-warning)" stroke-width="2" />
  <text x="640" y="435" font-family="Arial" font-size="16" text-anchor="middle" fill="var(--vp-c-text-1)">VPS</text>
  
  <path d="M200 150 L220 150" stroke="var(--vp-c-text-2)" stroke-width="2" fill="none" marker-end="url(#arrowhead)" />
  <path d="M280 180 L280 230 L320 230 L320 280" stroke="var(--vp-c-text-2)" stroke-width="2" fill="none" marker-end="url(#arrowhead)" />
  <path d="M320 320 L300 320 L300 230 L280 230 L280 180" stroke="var(--vp-c-text-2)" stroke-width="2" fill="none" marker-end="url(#arrowhead)" stroke-dasharray="5,5" />
  <path d="M480 280 L480 180 L620 180" stroke="var(--vp-c-text-2)" stroke-width="2" fill="none" marker-end="url(#arrowhead)" />
  <path d="M620 160 L480 160 L480 280" stroke="var(--vp-c-text-2)" stroke-width="2" fill="none" marker-end="url(#arrowhead)" stroke-dasharray="5,5" />
  <path d="M680 180 L680 300 L640 300 L640 400" stroke="var(--vp-c-text-2)" stroke-width="2" fill="none" marker-end="url(#arrowhead)" />
  <path d="M600 400 L600 300 L660 300 L660 180" stroke="var(--vp-c-text-2)" stroke-width="2" fill="none" marker-end="url(#arrowhead)" stroke-dasharray="5,5" />
  
  <defs>
    <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
      <polygon points="0 0, 10 3.5, 0 7" fill="var(--vp-c-text-2)" />
    </marker>
  </defs>
  
  <text x="600" y="30" font-family="Arial" font-size="20" text-anchor="middle" font-weight="bold" fill="var(--vp-c-text-1)">Nixopus Terminal Architecture</text>
  <text x="200" y="135" font-family="Arial" font-size="12" text-anchor="middle" fill="var(--vp-c-text-2)">User Input</text>
  <text x="200" y="165" font-family="Arial" font-size="12" text-anchor="middle" fill="var(--vp-c-text-2)">Terminal Output</text>
  <text x="350" y="250" font-family="Arial" font-size="12" text-anchor="middle" fill="var(--vp-c-text-2)">WebSocket</text>
  <text x="595" y="130" font-family="Arial" font-size="12" text-anchor="middle" fill="var(--vp-c-text-2)">SSH Protocol</text>
  <text x="595" y="170" font-family="Arial" font-size="12" text-anchor="middle" fill="var(--vp-c-text-2)">Data Stream</text>
  <text x="700" y="300" font-family="Arial" font-size="12" text-anchor="middle" fill="var(--vp-c-text-2)">Secure Connection</text>
  
  <rect x="50" y="470" width="20" height="10" fill="var(--vp-c-brand-soft)" stroke="var(--vp-c-brand)" stroke-width="1" />
  <text x="80" y="480" font-family="Arial" font-size="12" text-anchor="start" fill="var(--vp-c-text-2)">nixopus-view</text>
  
  <rect x="200" y="470" width="20" height="10" fill="var(--vp-c-success-soft)" stroke="var(--vp-c-success)" stroke-width="1" />
  <text x="230" y="480" font-family="Arial" font-size="12" text-anchor="start" fill="var(--vp-c-text-2)">nixopus-api</text>
  
  <rect x="350" y="470" width="20" height="10" fill="var(--vp-c-danger-soft)" stroke="var(--vp-c-danger)" stroke-width="1" />
  <text x="380" y="480" font-family="Arial" font-size="12" text-anchor="start" fill="var(--vp-c-text-2)">nixopus-realtime</text>
  
  <line x1="500" y1="475" x2="530" y2="475" stroke="var(--vp-c-text-2)" stroke-width="2" marker-end="url(#arrowhead)" />
  <text x="560" y="480" font-family="Arial" font-size="12" text-anchor="start" fill="var(--vp-c-text-2)">Request</text>
  
  <line x1="650" y1="475" x2="680" y2="475" stroke="var(--vp-c-text-2)" stroke-width="2" stroke-dasharray="5,5" marker-end="url(#arrowhead)" />
  <text x="690" y="480" font-family="Arial" font-size="12" text-anchor="start" fill="var(--vp-c-text-2)">Response</text>
</svg>
</div>

Let's break down how our terminal works. Think of it as a team of three main players working together: the frontend (`nixopus-view`), the backend (`nixopus-api`), and our real-time communication hub (`nixopus-realtime`). The frontend is what you see and interact with - it uses XTerm.js to create that familiar terminal look and feel, while React-DOM helps us build a smooth, responsive interface. When you type commands, they travel through our backend, which uses goph-ssh to securely connect to your VPS. The real-time server keeps everything in sync, making sure your commands and their responses flow smoothly back and forth. We've built this whole system with security in mind, so your data stays safe as it moves between these components.

## Features

### Key Bindings

* `CTRL + J`: Toggle terminal visibility
* `CTRL + T`: Switch terminal position (bottom/right)

### Resize

Drag the terminal's top edge to resize, similar to VS Code's terminal behavior.

### Themes and Fonts

The terminal inherits your application's theme and font settings. Customize them in [Theme and Fonts](/preferences/index.md).

### Terminal Focus

When the terminal is focused, application shortcuts are disabled to prevent conflicts with terminal commands. For example, `CTRL + C` sends SIGINT instead of copying.

## What's Coming Next

* Command sanitization and System protection against harmful commands
* Privacy safeguards and following best practices
* Prompt customization
* Multi-editor support
* Command auto-completion
* AI-powered command suggestions
* Multi-shell support (bash/zsh/fish)
