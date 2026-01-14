/**
 * Parse environment variables from text input
 * Supports formats: KEY=VALUE or KEY="VALUE" or KEY='VALUE'
 * Handles escaped characters in double-quoted strings (e.g., \n for newlines)
 */

interface ParsedEnvLine {
  key: string;
  value: string;
}

const unquoteValue = (value: string): string => {
  // Handle single-quoted strings
  if (value.startsWith("'") && value.endsWith("'")) {
    return value.slice(1, -1).replace(/''/g, "'");
  }

  // Handle double-quoted strings - process escape sequences explicitly
  if (value.startsWith('"') && value.endsWith('"')) {
    return value
      .slice(1, -1)
      .replace(/\\n/g, '\n')
      .replace(/\\r/g, '\r')
      .replace(/\\t/g, '\t')
      .replace(/\\\\/g, '\\')
      .replace(/\\"/g, '"');
  }

  return value;
};

export const parseEnvLine = (line: string): ParsedEnvLine | null => {
  const trimmed = line.trim();
  if (!trimmed || trimmed.startsWith('#')) return null;

  const match = trimmed.match(/^([^=]+)=(.*)$/);
  if (!match) return null;

  const key = match[1].trim();
  const rawValue = match[2].trim();
  if (!key) return null;

  return { key, value: unquoteValue(rawValue) };
};

export const parseEnvText = (text: string): Record<string, string> => {
  const result: Record<string, string> = {};
  const lines = text.split('\n');

  for (const line of lines) {
    const parsed = parseEnvLine(line);
    if (parsed) {
      result[parsed.key] = parsed.value;
    }
  }

  return result;
};

export const isMultiLineEnvPaste = (text: string): boolean => {
  return text.includes('\n');
};
