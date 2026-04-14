import type { ChatContext, ContextProviderData } from './chat-context';

export interface MentionMatch {
  start: number;
  end: number;
  query: string;
}

export interface MentionItem {
  key: string;
  label: string;
  type: string;
  ctx: ChatContext;
}

const OPEN_BRACKETS = new Set(['(', '[', '{', '<']);

function isMentionBoundaryBefore(input: string, atIndex: number): boolean {
  if (atIndex === 0) return true;
  const prev = input[atIndex - 1]!;
  if (/\s/.test(prev)) return true;
  if (OPEN_BRACKETS.has(prev)) return true;
  return false;
}

/**
 * Nearest @-mention before `cursor`: @ has a valid left boundary, and there is
 * no whitespace in the query segment between @ and the cursor.
 */
export function getActiveMentionToken(input: string, cursor: number): MentionMatch | null {
  const c = Math.max(0, Math.min(cursor, input.length));

  let at = -1;
  for (let i = c - 1; i >= 0; i--) {
    if (input.charCodeAt(i) !== 64 /* @ */) continue;
    if (isMentionBoundaryBefore(input, i)) {
      at = i;
      break;
    }
  }
  if (at === -1) return null;

  const after = input.slice(at + 1, c);
  if (/\s/.test(after)) return null;

  return {
    start: at,
    end: c,
    query: after.toLowerCase()
  };
}

export function flattenMentionItems(providers: ContextProviderData[]): MentionItem[] {
  const out: MentionItem[] = [];
  for (const p of providers) {
    for (const ctx of p.items) {
      out.push({
        key: `${ctx.type}:${ctx.id}`,
        label: ctx.label,
        type: ctx.type,
        ctx
      });
    }
  }
  return out;
}

type Rank = 0 | 1;

function mentionRank(labelLower: string, q: string): Rank | null {
  if (q.length === 0) return 0;
  if (labelLower.startsWith(q)) return 0;
  if (labelLower.includes(q)) return 1;
  return null;
}

export function filterMentionItems(items: MentionItem[], query: string, limit = 8): MentionItem[] {
  const q = query.toLowerCase();
  const scored: { item: MentionItem; rank: Rank; labelLower: string }[] = [];

  for (const item of items) {
    const labelLower = item.label.toLowerCase();
    const rank = mentionRank(labelLower, q);
    if (rank === null) continue;
    scored.push({ item, rank, labelLower });
  }

  scored.sort((a, b) => {
    if (a.rank !== b.rank) return a.rank - b.rank;
    const lbl = a.labelLower.localeCompare(b.labelLower);
    if (lbl !== 0) return lbl;
    const typ = a.item.type.localeCompare(b.item.type);
    if (typ !== 0) return typ;
    return a.item.key.localeCompare(b.item.key);
  });

  return scored.slice(0, limit).map((s) => s.item);
}

/** Remove the matched @token span and collapse spaces at the join. */
export function replaceMentionToken(input: string, match: MentionMatch): string {
  const left = input.slice(0, match.start);
  const right = input.slice(match.end);
  const leftTrim = left.replace(/\s+$/, '');
  const rightTrim = right.replace(/^\s+/, '');
  if (leftTrim.length > 0 && rightTrim.length > 0) {
    return `${leftTrim} ${rightTrim}`;
  }
  return `${leftTrim}${rightTrim}`;
}
