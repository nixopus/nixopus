import { code } from '@streamdown/code';
import { mermaid } from '@streamdown/mermaid';

export const STREAMDOWN_PLUGINS = { code, mermaid };

export const STREAMDOWN_CONTROLS = {
  code: { copy: true, download: true },
  table: { copy: true, download: true, fullscreen: true },
  mermaid: { fullscreen: true, download: true, copy: true, panZoom: true }
};

export const STREAMDOWN_ANIMATED = {
  animation: 'blurIn' as const,
  duration: 200,
  easing: 'ease-out'
};
