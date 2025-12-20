/**
 * Parse environment variables from text input
 * Supports formats: KEY=VALUE or KEY="VALUE" or KEY='VALUE'
 */

interface ParsedEnvLine {
  key: string;
  value: string;
}

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

const unquoteValue = (value: string): string => {
  const singleQuoteMatch = value.match(/^'(.*)'$/);
  if (singleQuoteMatch) return singleQuoteMatch[1];

  const doubleQuoteMatch = value.match(/^"(.*)"$/);
  if (doubleQuoteMatch) return doubleQuoteMatch[1];

  return value;
};

export const parseEnvText = (text: string): Record<string, string> => {
  const lines = text.split('\n');
  const result: Record<string, string> = {};

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
