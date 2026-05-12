package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/agent/service/deploy"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/usage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
)

type ChatRequest struct {
	ThreadID string `json:"thread_id,omitempty"`
	Input    string `json:"input"`
	Model    string `json:"model,omitempty"`
}

type ChatResponse struct {
	ThreadID     string    `json:"thread_id"`
	Response     string    `json:"response"`
	Usage        llm.Usage `json:"usage,omitempty"`
	BalanceCents *int      `json:"balance_cents,omitempty"`
}

func (s *AgentService) Chat(ctx context.Context, req ChatRequest, authToken, userID, orgID string) (*ChatResponse, error) {
	ctx = context.WithValue(ctx, ctxKeyAuthToken, authToken)
	ctx = context.WithValue(ctx, ctxKeyOrgID, orgID)
	ctx = context.WithValue(ctx, ctxKeySchedulerUserID, userID)

	threadID := req.ThreadID
	if threadID == "" {
		thread, err := s.CreateThread(ctx, userID, orgID)
		if err != nil {
			return nil, fmt.Errorf("create thread: %w", err)
		}
		threadID = thread.ID
	}

	history, err := s.memory.GetMessages(ctx, threadID, 50)
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}

	messages := storedToMessages(history)

	// Preprocess: inject context blocks
	contextPrefix := s.buildContextPrefix(ctx, orgID, history, req.Input)
	augmentedInput := req.Input
	if contextPrefix != "" {
		augmentedInput = contextPrefix + "\n\n" + req.Input
	}

	agent, _ := s.agents.Get("deploy")
	start := time.Now()
	result, err := agent.RunWithOptions(ctx, llm.RunOptions{Model: req.Model}, augmentedInput, messages...)
	if err != nil {
		return nil, fmt.Errorf("agent run: %w", err)
	}
	latencyMs := int(time.Since(start).Milliseconds())

	seq, _ := s.memory.GetMessageCount(ctx, threadID)
	newMessages := extractNewMessages(result.Messages, len(messages)+1, req.Input)
	toStore := memory.MessagesFromLLM(threadID, newMessages, seq)
	if err := s.memory.AppendMessages(ctx, threadID, toStore); err != nil {
		s.logger.Log(logger.Error, "Failed to persist messages", err.Error())
	}

	go s.patterns.RecordDeployOutcome(ctx, orgID, history)

	resp := &ChatResponse{
		ThreadID: threadID,
		Response: result.Content,
		Usage:    result.TotalUsage,
	}

	balance := s.usage.TrackUsage(ctx, usage.TrackingParams{
		OrgID:            orgID,
		ModelID:          usage.ResolveModelID(req.Model),
		PromptTokens:     result.TotalUsage.PromptTokens,
		CompletionTokens: result.TotalUsage.CompletionTokens,
		TotalTokens:      result.TotalUsage.TotalTokens,
		CostUsd:          0,
		RequestType:      "chat",
		SessionID:        threadID,
		LatencyMs:        &latencyMs,
	})
	if balance >= 0 {
		resp.BalanceCents = &balance
	}

	return resp, nil
}

type StreamRequest struct {
	ThreadID string `json:"thread_id,omitempty"`
	Input    string `json:"input"`
	Model    string `json:"model,omitempty"`
}

func (s *AgentService) StreamChat(parentCtx context.Context, w http.ResponseWriter, req StreamRequest, authToken, userID, orgID string) error {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	ctx = context.WithValue(ctx, ctxKeyAuthToken, authToken)
	ctx = context.WithValue(ctx, ctxKeyOrgID, orgID)
	ctx = context.WithValue(ctx, ctxKeySchedulerUserID, userID)

	threadID := req.ThreadID
	if threadID == "" {
		thread, err := s.CreateThread(ctx, userID, orgID)
		if err != nil {
			return fmt.Errorf("create thread: %w", err)
		}
		threadID = thread.ID
	} else {
		if _, err := s.memory.GetThread(ctx, threadID); err != nil {
			if _, createErr := s.CreateThreadWithOpts(ctx, userID, orgID, threadID, ""); createErr != nil {
				return fmt.Errorf("ensure thread: %w", createErr)
			}
		}
	}

	RegisterStream(threadID, cancel)
	defer UnregisterStream(threadID)

	history, err := s.memory.GetMessages(ctx, threadID, 50)
	if err != nil {
		return fmt.Errorf("get history: %w", err)
	}

	messages := storedToMessages(history)

	// Preprocess: inject context blocks
	contextPrefix := s.buildContextPrefix(ctx, orgID, history, req.Input)
	augmentedInput := req.Input
	if contextPrefix != "" {
		augmentedInput = contextPrefix + "\n\n" + req.Input
	}

	agent, _ := s.agents.Get("deploy")
	start := time.Now()

	handler := llm.NewStreamHandler(agent)
	handler.ThreadID = threadID
	handler.Model = req.Model
	handler.OnDone = func(result *llm.StreamRunResult) {
		bgCtx := context.Background()

		seq, _ := s.memory.GetMessageCount(bgCtx, threadID)
		newMessages := extractNewMessages(result.Messages, len(messages)+1, req.Input)
		toStore := memory.MessagesFromLLM(threadID, newMessages, seq)
		_ = s.memory.AppendMessages(bgCtx, threadID, toStore)

		latencyMs := int(time.Since(start).Milliseconds())
		balance := s.usage.TrackUsage(bgCtx, usage.TrackingParams{
			OrgID:            orgID,
			ModelID:          usage.ResolveModelID(req.Model),
			PromptTokens:     result.TotalUsage.PromptTokens,
			CompletionTokens: result.TotalUsage.CompletionTokens,
			TotalTokens:      result.TotalUsage.TotalTokens,
			CostUsd:          0,
			RequestType:      "stream",
			SessionID:        threadID,
			LatencyMs:        &latencyMs,
		})
		if balance >= 0 {
			w.Header().Set("X-Credits-Remaining", fmt.Sprintf("%d", balance))
		}

		go s.patterns.RecordDeployOutcome(bgCtx, orgID, history)
	}

	body, _ := json.Marshal(struct {
		Input   string        `json:"input"`
		History []llm.Message `json:"history,omitempty"`
	}{
		Input:   augmentedInput,
		History: messages,
	})

	r, _ := http.NewRequest("POST", "/", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(ctx)

	handler.ServeHTTP(w, r)
	return nil
}

func (s *AgentService) CreateThread(ctx context.Context, userID, orgID string) (*memory.Thread, error) {
	return s.CreateThreadWithOpts(ctx, userID, orgID, "", "")
}

func (s *AgentService) CreateThreadWithOpts(ctx context.Context, userID, orgID, id, title string) (*memory.Thread, error) {
	thread := &memory.Thread{
		ID:       id,
		UserID:   userID,
		Title:    title,
		Metadata: memory.JSONMap{"org_id": orgID},
	}
	if err := s.memory.CreateThread(ctx, thread); err != nil {
		return nil, err
	}
	return thread, nil
}

func (s *AgentService) GetThreadMessages(ctx context.Context, threadID, userID string, limit int) ([]memory.StoredMessage, error) {
	if limit <= 0 {
		limit = 50
	}
	if _, err := s.memory.GetThreadForUser(ctx, threadID, userID); err != nil {
		return nil, fmt.Errorf("thread not found or access denied: %w", err)
	}
	return s.memory.GetMessages(ctx, threadID, limit)
}

func (s *AgentService) ListThreads(ctx context.Context, userID string) ([]memory.Thread, error) {
	return s.memory.ListThreads(ctx, userID, 100, 0)
}

func (s *AgentService) UpdateThread(ctx context.Context, threadID, userID, title string) error {
	return s.memory.UpdateThread(ctx, threadID, userID, title)
}

func (s *AgentService) DeleteThread(ctx context.Context, threadID, userID string) error {
	return s.memory.DeleteThread(ctx, threadID, userID)
}

func storedToMessages(stored []memory.StoredMessage) []llm.Message {
	messages := make([]llm.Message, 0, len(stored))
	for _, m := range stored {
		msg := llm.Message{
			Role:       llm.Role(m.Role),
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if len(m.ToolCalls) > 0 {
			msg.ToolCalls = m.ToolCalls
		}
		messages = append(messages, msg)
	}
	return messages
}

// extractNewMessages pulls out the new messages from a completed agent run.
// buildMessages produces [system, ...history, user(augmented)], so everything
// from index `offset` onward is new. The first new message is the user prompt
// with the original (un-augmented) input text.
func extractNewMessages(all []llm.Message, offset int, originalInput string) []llm.Message {
	if offset >= len(all) {
		return []llm.Message{
			{Role: llm.RoleUser, Content: originalInput},
		}
	}

	newMsgs := make([]llm.Message, 0, len(all)-offset+1)
	newMsgs = append(newMsgs, llm.Message{Role: llm.RoleUser, Content: originalInput})

	for _, m := range all[offset:] {
		newMsgs = append(newMsgs, m)
	}
	return newMsgs
}

// buildContextPrefix assembles the preprocessed context blocks to prepend to user input.
// Context injection behaviour is controlled by s.contextPolicy:
//   - "first-only" (default): full user-context on first message, instance info on subsequent
//   - "always": full user-context re-injected every turn
//   - "never": skip user-context entirely (instance info still included)
func (s *AgentService) buildContextPrefix(ctx context.Context, orgID string, history []memory.StoredMessage, currentInput string) string {
	var blocks []string

	switch s.contextPolicy {
	case ContextPolicyNever:
		blocks = append(blocks, s.fetchInstanceContext())
	case ContextPolicyAlways:
		if uc := s.injectUserContext(ctx, orgID); uc != "" {
			blocks = append(blocks, uc)
		} else {
			blocks = append(blocks, s.fetchInstanceContext())
		}
	default: // ContextPolicyFirstOnly
		if len(history) == 0 {
			if uc := s.injectUserContext(ctx, orgID); uc != "" {
				blocks = append(blocks, uc)
			}
		} else {
			blocks = append(blocks, s.fetchInstanceContext())
		}
	}

	// 2. Deploy flow: auto-inject on deploy intent so the LLM has full instructions
	if hasDeployIntent(currentInput) {
		if flow := s.loadDeployFlowSkill(); flow != "" {
			blocks = append(blocks, "[deploy-flow]\n"+flow+"\n[/deploy-flow]")
		}
	}

	// 3. Onboarding: auto-inject when frontend sends __ONBOARD__ signal
	if strings.TrimSpace(currentInput) == "__ONBOARD__" {
		if ob := s.loadOnboardingSkill(); ob != "" {
			blocks = append(blocks, "[onboarding]\n"+ob+"\n[/onboarding]")
		}
		if gh := s.loadSkillContent("github-onboarding"); gh != "" {
			blocks = append(blocks, "[github-onboarding]\n"+gh+"\n[/github-onboarding]")
		}
	}

	// 4. Deploy state: always include
	deployStateBlock := deploy.ExtractDeployState(history)
	if deployStateBlock != "" {
		blocks = append(blocks, deployStateBlock)
	}

	// 5. Deploy patterns: include when ecosystem is detected
	ecosystem := deploy.DetectEcosystem(history, currentInput)
	if ecosystem != "" {
		patterns := s.patterns.GetPatterns(ctx, ecosystem)
		if patternsBlock := deploy.FormatPatterns(ecosystem, patterns); patternsBlock != "" {
			blocks = append(blocks, patternsBlock)
		}
	}

	if len(blocks) == 0 {
		return ""
	}

	result := strings.Join(blocks, "\n\n")

	// Token budget: cap total injected context at ~15K chars (~3.7K tokens)
	// to prevent context explosion in long conversations with many apps/patterns.
	const maxContextChars = 15_000
	if len(result) > maxContextChars {
		result = result[:maxContextChars] + "\n... [context truncated]"
	}

	return result
}

var deployIntentRe = regexp.MustCompile(`(?i)\b(deploy\w*|launch|ship|go\s+live|push\s+to\s+prod|put\s+(?:it\s+)?(?:on|up)|host|set\s*up)\b`)

func hasDeployIntent(input string) bool {
	return deployIntentRe.MatchString(input)
}

func (s *AgentService) loadDeployFlowSkill() string {
	skill, ok := s.skills.Get("deploy-flow")
	if !ok || skill.Content == "" {
		return ""
	}
	return skill.Content
}

func (s *AgentService) loadOnboardingSkill() string {
	skill, ok := s.skills.Get("onboarding")
	if !ok || skill.Content == "" {
		return ""
	}
	return skill.Content
}

func (s *AgentService) loadSkillContent(name string) string {
	skill, ok := s.skills.Get(name)
	if !ok || skill.Content == "" {
		return ""
	}
	return skill.Content
}
