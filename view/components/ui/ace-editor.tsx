import dynamic from 'next/dynamic';
import React from 'react';

const Ace = dynamic(
  async () => {
    const ace = await import('react-ace');
    await import('ace-builds/src-noconflict/mode-yaml');
    await import('ace-builds/src-noconflict/mode-sh');
    await import('ace-builds/src-noconflict/theme-terminal');
    return ace;
  },
  { ssr: false }
);

export interface AceEditorProps {
  mode: 'yaml' | 'sh';
  value: string;
  onChange: (value: string) => void;
  name: string;
  theme?: string;
  readOnly?: boolean;
  onLoad?: (editor: any) => void;
  height?: string;
  markers?: any[];
}

const AceEditor = React.forwardRef<any, AceEditorProps>(
  (
    { mode, value, onChange, name, readOnly = false, onLoad, height = '600px', markers = [] },
    ref
  ) => (
    <Ace
      mode={mode}
      theme="terminal"
      onChange={onChange}
      value={value}
      name={name}
      editorProps={{ $blockScrolling: true }}
      fontSize={14}
      lineHeight={19}
      showPrintMargin={true}
      onLoad={onLoad}
      showGutter={true}
      highlightActiveLine={true}
      readOnly={readOnly}
      markers={markers}
      setOptions={{
        enableBasicAutocompletion: true,
        enableLiveAutocompletion: true,
        enableSnippets: true,
        showLineNumbers: true,
        tabSize: 2,
        useWorker: true,
        wrap: true,
        autoScrollEditorIntoView: true,
        animatedScroll: true,
        highlightGutterLine: true,
        highlightSelectedWord: true
      }}
      style={{ width: '100%', height, overflow: 'hidden' }}
    />
  )
);

AceEditor.displayName = 'AceEditor';

export default AceEditor;
