# New User Guided Chat Prefill Design

Date: 2026-04-14
Status: Approved for implementation planning
Owner: Chat onboarding flow

## Goal

Guide first-time users through what Nixopus can do by preloading a starter prompt and attaching a default sample repository context in the chat welcome screen.

The prefill should feel assistive, not forced:

- Never auto-send.
- Keep the prompt editable.
- Preserve user control of when to send.

## Product Decision

Chosen behavior:

1. Prefill guided starter prompt + attach sample repository context.
2. User must click Send manually.
3. First-time behavior is global to the account experience.
4. If GitHub is not connected, keep showing the guided prefill on each visit.
5. Once GitHub is connected and guided prefill has been shown once, stop showing it.

## UX Contract

When the chat page is in welcome state (no current messages):

- If `githubConnected === false`: show guided prefill every visit.
- If `githubConnected === true` and `guidedPrefillSeen === false`: show guided prefill once and mark as seen.
- If `githubConnected === true` and `guidedPrefillSeen === true`: show default empty welcome state.

Interaction rules:

- Guided prompt text is prefilled in the input box.
- Default sample repository is attached as a chat context automatically.
- Input remains editable.
- Prompt is not auto-submitted.
- If user already typed text, do not overwrite it.
- If user clears/edits prefill, do not re-inject in the same session view.
- Show helper copy near input: "Starter prompt loaded - edit before sending."

## Architecture and Component Boundaries

Primary integration point:

- `view/packages/hooks/ai/use-chat-page.ts`

Reason:

- Existing prefill behavior (`pendingDeployPrompt`) already lives here.
- It already orchestrates welcome state, thread setup, and input updates.

UI rendering point:

- `view/packages/components/ai-chat.tsx`

Reason:

- Renders `ChatWelcomeView` and welcome input.
- Can render helper text and consume guided-prefill state exposed by the hook.

## State Model

Local persisted flag:

- `chat_guided_prefill_seen` (localStorage boolean)

Session-only guard:

- in-memory ref to prevent repeated reinjection after user interaction in a single mounted view.

Derived decision input:

- `githubConnected` (from existing GitHub connection source)
- `isThreadsInitialized`
- welcome state (`messages.length === 0` + existing welcome criteria)
- current input emptiness (`inputValue.trim().length === 0`)

Derived output:

- `shouldInjectGuidedPrefill: boolean`
- `guidedPrefillPayload` with:
  - `promptText`
  - `sampleRepoContext` (`ChatContext` shape)

## Data Flow

1. Chat page reaches welcome-eligible state.
2. Hook computes eligibility for guided prefill:
   - `!githubConnected` OR (`githubConnected && !guidedPrefillSeen`)
3. If eligible and input is empty, hook:
   - sets input value to guided prompt text.
   - adds sample repository context (deduped).
4. If GitHub is connected, mark `chat_guided_prefill_seen = true` after successful injection.
5. User edits or sends manually via existing submit flow.
6. No auto-submit path is introduced.

## Precedence Rules

To avoid conflicting prefills:

- Deploy deep-link prefill (`pendingDeployPrompt`) has higher priority.
- Guided prefill runs only when deploy prefill is not active.
- Guided prefill runs only in welcome state.

## Error Handling and Fallbacks

- If sample repository context data cannot be built, still inject prompt text.
- If GitHub connection status is unknown or fails to load, treat as not connected for safe onboarding behavior.
- If localStorage is unavailable, fallback to session-only behavior (no crash).
- If sample repo context already exists in selected contexts, do not duplicate.

## Telemetry (Recommended)

Track lightweight events:

- `guided_prefill_shown`
- `guided_prefill_sent`
- `guided_prefill_edited_before_send`
- `guided_prefill_cleared`

Purpose:

- Validate whether onboarding guidance helps activation without increasing friction.

## Verification Strategy (Manual Only)

Per request, no automated tests are required for this change.

Manual verification checklist:

1. New user with GitHub disconnected:
   - guided prefill appears on welcome.
   - sample repo context is attached.
   - no auto-send.
   - revisiting chat shows prefill again.
2. User connects GitHub:
   - guided prefill appears once.
   - after reload/revisit, prefill no longer appears.
3. Existing user with seen flag and GitHub connected:
   - normal empty welcome input.
4. User starts typing before injection condition resolves:
   - typed input is preserved (not overwritten).
5. Deploy deep-link case:
   - deploy prefill still appears and has precedence.

## Out of Scope

- Server-synced onboarding flags across devices.
- New backend storage for onboarding state.
- Full onboarding tour orchestration beyond chat prefill + sample repo context.

## Risks and Mitigations

Risk: Users may feel forced if behavior repeats too often while disconnected.
Mitigation: Keep prompt editable, never auto-send, and use short helper copy.

Risk: Prefill collisions with existing flows.
Mitigation: Explicit precedence rules and single hook integration point.

Risk: Local-only seen flag not shared across devices.
Mitigation: Accept for now; leave migration path open for server-backed flags later.
