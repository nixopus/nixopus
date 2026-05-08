package service

import (
	"context"
	"fmt"
	"os"

	"github.com/nixopus/nixopus/api/internal/features/agent/service/catalog"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/deploy"
	agentgithub "github.com/nixopus/nixopus/api/internal/features/agent/service/github"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/scheduler"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/usage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/uptrace/bun"
)

// ContextPolicy controls when user-context is injected into the conversation.
type ContextPolicy string

const (
	// ContextPolicyFirstOnly injects full user-context on the first message,
	// instance info on subsequent messages. This is the default.
	ContextPolicyFirstOnly ContextPolicy = "first-only"
	// ContextPolicyAlways re-injects full user-context on every turn.
	// Useful for long conversations where context scrolls out of window.
	ContextPolicyAlways ContextPolicy = "always"
	// ContextPolicyNever skips context injection entirely.
	ContextPolicyNever ContextPolicy = "never"
)

type AgentService struct {
	store         *storage.Store
	ctx           context.Context
	logger        logger.Logger
	provider      llm.Provider
	memory        *memory.PostgresStore
	skills        *llm.SkillStore
	agents        *llm.AgentRegistry
	usage         *usage.Tracker
	preflight     *catalog.Validator
	patterns      *deploy.Store
	github        *agentgithub.Client
	contextPolicy ContextPolicy
	scheduler     *scheduler.Scheduler
	scheduleStore *scheduler.Store
	notifier      types.Notifier
}

func NewAgentService(store *storage.Store, ctx context.Context, l logger.Logger, notifier ...types.Notifier) *AgentService {
	db := store.DB

	provider := buildProvider(l)

	memStore := memory.NewPostgresStore(db)

	skills := llm.NewSkillStore()
	loadSkills(skills, l)

	agents := llm.NewAgentRegistry()

	usageTracker := usage.NewTracker(db, ctx, l)

	patternStore := deploy.NewStore(db, l)

	ghClient := agentgithub.NewClient(db, func(ctx context.Context) string {
		v, _ := ctx.Value(ctxKeyOrgID).(string)
		return v
	})

	specPath := "doc/openapi.json"

	if err := catalog.GenerateAndSet(specPath); err != nil {
		l.Log(logger.Warning, "catalog: auto-generation failed, using fallback", err.Error())
	} else {
		l.Log(logger.Info, "catalog: auto-generated from OpenAPI spec", "")
	}

	ctxPolicy := ContextPolicy(getEnvOrDefault("AGENT_CONTEXT_POLICY", string(ContextPolicyFirstOnly)))
	if ctxPolicy != ContextPolicyFirstOnly && ctxPolicy != ContextPolicyAlways && ctxPolicy != ContextPolicyNever {
		ctxPolicy = ContextPolicyFirstOnly
	}

	var n types.Notifier
	if len(notifier) > 0 && notifier[0] != nil {
		n = notifier[0]
	}

	schedStore := scheduler.NewStore(db, l)

	svc := &AgentService{
		store:         store,
		ctx:           ctx,
		logger:        l,
		provider:      provider,
		memory:        memStore,
		skills:        skills,
		agents:        agents,
		usage:         usageTracker,
		preflight:     catalog.NewValidator(specPath),
		patterns:      patternStore,
		github:        ghClient,
		contextPolicy: ctxPolicy,
		scheduleStore: schedStore,
		notifier:      n,
	}

	svc.registerAgents(db)
	svc.patterns.CreateTables(ctx)
	schedStore.CreateTables(ctx)

	sched := scheduler.New(schedStore, memStore, svc, n, l)
	svc.scheduler = sched
	sched.Start(ctx)

	return svc
}

// buildProvider creates the LLM provider based on environment configuration.
// Uses OpenAI-compatible API format which works with OpenRouter, OpenAI, Together,
// Groq, Ollama, and any provider exposing /v1/chat/completions.
func buildProvider(l logger.Logger) llm.Provider {
	baseURL := getEnvOrDefault("LLM_BASE_URL", "https://openrouter.ai/api/v1")
	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	headers := map[string]string{}
	if os.Getenv("LLM_REFERER") != "" {
		headers["HTTP-Referer"] = os.Getenv("LLM_REFERER")
	} else {
		headers["HTTP-Referer"] = "https://nixopus.com"
	}
	headers["X-Title"] = "Nixopus Agent"

	l.Log(logger.Info, "LLM provider configured", fmt.Sprintf("base_url=%s", baseURL))

	return llm.NewOpenAIProvider(llm.OpenAIConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Headers: headers,
	})
}

func (s *AgentService) registerAgents(db *bun.DB) {
	model := getEnvOrDefault("AGENT_MODEL", "anthropic/claude-sonnet-4")
	lightModel := getEnvOrDefault("AGENT_LIGHT_MODEL", "openai/gpt-4o-mini")

	profiles := s.buildToolProfiles()

	deployAgent := llm.NewAgent(s.provider, profiles.Build(llm.ProfileDeploy), llm.AgentConfig{
		Model:        model,
		SystemPrompt: s.buildDeployPrompt(),
		MaxSteps:     100,
	})
	s.agents.Register("deploy", deployAgent)

	diagnosticPrompt := llm.NewPromptBuilder()
	diagnosticPrompt.Add(llm.PromptSection{Name: "identity", Priority: 0, Content: "Application and container debugger. Discover IDs via nixopus_api. No emojis. Plain text only."})
	diagnosticPrompt.Add(llm.PromptSection{Name: "api-access", Priority: 1, Content: "## API Access\nUse nixopus_api(method, path, body) for all Nixopus API calls. See [api-catalog] below for available operations with required fields and types.\nKey operations: get_applications, get_application, get_application_deployments, get_deployment_logs, get_application_logs, list_containers, get_container, get_container_logs, get_compose_services, restart_deployment, redeploy_application."})
	diagnosticPrompt.Add(llm.PromptSection{Name: "catalog", Priority: 2, Content: catalog.Catalog, MaxChars: 30_000})
	diagnosticAgent := llm.NewAgent(s.provider, profiles.Build(llm.ProfileDiagnostic), llm.AgentConfig{
		Model:        lightModel,
		SystemPrompt: diagnosticPrompt.Build(),
		MaxSteps:     15,
	})
	s.agents.Register("diagnostic", diagnosticAgent)

	githubPrompt := llm.NewPromptBuilder()
	githubPrompt.Add(llm.PromptSection{Name: "identity", Priority: 0, Content: "Interact with GitHub repos, PRs, issues, and deployment statuses via the connected GitHub App. Use nixopus_api to resolve numeric repo IDs for create_project. Never use emojis. Plain text only."})
	githubPrompt.Add(llm.PromptSection{Name: "safety", Priority: 1, Content: "## GitHub Safety — NON-NEGOTIABLE\nNEVER commit or push directly to main/master. For ANY file change: create a feature branch → commit to that branch → open a PR. No exceptions.\nNEVER merge PRs unless the user explicitly asks. Always return the PR URL.\nNo destructive ops (force push, branch delete, PR close) without user approval."})
	githubPrompt.Add(llm.PromptSection{Name: "api-access", Priority: 2, Content: "## API Access\nUse nixopus_api(method, path, body) for Nixopus API calls. For direct GitHub file/PR/issue operations, use the dedicated github_ tools."})
	githubPrompt.Add(llm.PromptSection{Name: "catalog", Priority: 3, Content: catalog.Catalog, MaxChars: 30_000})
	githubAgent := llm.NewAgent(s.provider, profiles.Build(llm.ProfileGitHub), llm.AgentConfig{
		Model:        lightModel,
		SystemPrompt: githubPrompt.Build(),
		MaxSteps:     10,
	})
	s.agents.Register("github", githubAgent)

	notificationPrompt := llm.NewPromptBuilder()
	notificationPrompt.Add(llm.PromptSection{Name: "identity", Priority: 0, Content: "Notification delivery specialist. Route alerts through configured channels. No emojis. Plain text only."})
	notificationPrompt.Add(llm.PromptSection{Name: "api-access", Priority: 1, Content: "## API Access\nUse nixopus_api(method, path, body) for all Nixopus API calls. See [api-catalog] below for available operations with required fields and types.\nKey operations: send_notification (channel: slack|discord|email), get_notification_preferences, update_notification_preferences."})
	notificationPrompt.Add(llm.PromptSection{Name: "catalog", Priority: 2, Content: catalog.Catalog, MaxChars: 30_000})
	notificationAgent := llm.NewAgent(s.provider, profiles.Build(llm.ProfileNotify), llm.AgentConfig{
		Model:        lightModel,
		SystemPrompt: notificationPrompt.Build(),
		MaxSteps:     5,
	})
	s.agents.Register("notification", notificationAgent)

	machinePrompt := llm.NewPromptBuilder()
	machinePrompt.Add(llm.PromptSection{Name: "identity", Priority: 0, Content: "Machine and server operations specialist. Manage infrastructure, containers, and compute resources. No emojis. Plain text only."})
	machinePrompt.Add(llm.PromptSection{Name: "api-access", Priority: 1, Content: "## API Access\nUse nixopus_api(method, path, body) for all Nixopus API calls. See [api-catalog] below for available operations with required fields and types.\nKey operations: get_machine_stats, host_exec, get_servers, get_servers_ssh_status, get_machine_metrics, restart_machine, pause_machine, resume_machine."})
	machinePrompt.Add(llm.PromptSection{Name: "catalog", Priority: 2, Content: catalog.Catalog, MaxChars: 30_000})
	machineAgent := llm.NewAgent(s.provider, profiles.Build(llm.ProfileMachine), llm.AgentConfig{
		Model:        lightModel,
		SystemPrompt: machinePrompt.Build(),
		MaxSteps:     10,
	})
	s.agents.Register("machine", machineAgent)
}

func (s *AgentService) buildDeployPrompt() string {
	b := llm.NewPromptBuilder()

	b.Add(llm.PromptSection{
		Name:     "identity",
		Priority: 0,
		Content:  "You are Nixopus, a deploy orchestrator. Plain text only, no emojis.",
	})

	b.Add(llm.PromptSection{
		Name:     "rules",
		Priority: 1,
		Content: `## Rules — NON-NEGOTIABLE
The ONLY acceptable end state is a live URL or a clear blocker with escalation.
NEVER fabricate tool results. Every fact must come from an actual tool call.
Read the [deploy-state] block to see completed steps. Resume from where you left off.
NEVER reveal internal details: file paths, tool names, S3, BM25, workspace, build_pack, hints, confidence levels.
Keep the user continuously informed. Acknowledge requests immediately. Every update must include a completed action, current step, latest blocker, or live link.
Do not ask for permission for obvious fixes. Just do them.
Only ask the user for input when you literally cannot proceed: missing secrets, missing GitHub access, or a business decision.`,
	})

	b.Add(llm.PromptSection{
		Name:     "github-safety",
		Priority: 2,
		Content: `## GitHub Safety — NON-NEGOTIABLE
NEVER commit or push directly to main/master. For ANY file change in a GitHub repo: create a feature branch → commit to that branch → open a PR. No exceptions.
NEVER merge PRs unless the user explicitly asks. Always return the PR URL.
For GitHub-sourced apps, NEVER use write_workspace_files to create or fix files — it only writes locally and changes will NOT reach the repo. Always use github_create_or_update_file on a feature branch.`,
	})

	b.Add(llm.PromptSection{
		Name:     "api-catalog",
		Priority: 3,
		Content:  catalog.Catalog,
		MaxChars: 30_000,
	})

	b.Add(llm.PromptSection{
		Name:     "skills",
		Priority: 4,
		Content:  "## Available Skills\nLoad task-specific instructions with read_skill(name). Available:\n" + s.skills.FormattedCatalog(),
	})

	b.Add(llm.PromptSection{
		Name:     "deploy-guidance",
		Priority: 5,
		Content: `## Domain Rule — ALWAYS FOLLOW
createProject and quick_deploy auto-generate a subdomain if domains is empty — no need to call generate_random_subdomain separately.
Only call generate_random_subdomain explicitly when you need a specific subdomain before creation, or use add_application_domain for post-creation custom domains.

## Repo Hints — TRUST BUT VERIFY
analyze_repository, load_remote_repository, and load_local_workspace return a hints object with per-field confidence.
- confidence: "high" on all fields — proceed directly to quick_deploy or createProject. Do NOT manually explore files.
- confidence: "medium" or warnings present — verify only the flagged items with 1-2 targeted read_file calls, then proceed.
- confidence: "low" on port or framework — use workspace tools (read_file, grep) to confirm those specific fields only. Binary fields (hasDockerfile, packageManager, isMonorepo) are always certain.
Use quick_deploy for first-time deploys when you have all the info. It creates and deploys in one step.

## Monitoring Rule — LEAN
After deploy_project or quick_deploy, call getApplicationDeployments(limit=1) once to get the deployment ID. Then poll getDeploymentById only. Do NOT call getDeploymentLogs unless status is failed/error or user asks. Do NOT call getApplication after deploy.`,
	})

	b.Add(llm.PromptSection{
		Name:     "context-info",
		Priority: 6,
		Content: `## Pre-loaded Context
A [user-context] block is injected with your applications, domains, servers, GitHub connectors, repositories, and MCP servers at conversation start.
Use these IDs directly for initial discovery. Do NOT call getApplications, getDomains, getServers, or getGithubConnectors just to discover what exists — it is already in [user-context].
After mutating actions (create, update, delete), call the relevant tool if you need refreshed data.

## Deploy Patterns — Cross-Org Learning
A [deploy-patterns] block may be injected when an ecosystem is detected. It contains failure fixes, pitfalls, and fast paths learned from deployments across all organizations.
When present, CHECK the patterns before diagnosing. If a known fix matches, apply it directly. If a fix fails, the system updates its confidence automatically.
Do NOT mention "deploy patterns" or "cross-org learning" to the user — just use the knowledge silently.

## Nixopus Documentation
When the user asks about Nixopus features, configuration, or product-level questions: read_skill("nixopus-docs") and follow the lookup workflow.

## Scheduled Tasks
You can create recurring scheduled tasks for users. When a user asks for periodic monitoring, alerts, or recurring actions:
1. Parse their intent into a cron expression and a clear task prompt.
2. Call create_schedule with the name, prompt, cron expression, and delivery channel.
3. The prompt should be self-contained: specify which tools to call, what data to check, and what to report.
4. Use list_schedules to show active schedules, pause_schedule/resume_schedule to toggle, delete_schedule to remove.
Cron examples: "*/30 * * * *" (every 30 min), "0 9 * * *" (daily 9 AM), "0 */2 * * *" (every 2 hours), "@hourly", "@daily".`,
	})

	return b.Build()
}

func loadSkills(skills *llm.SkillStore, l logger.Logger) {
	skillPaths := []string{"./skills", "../skills", "/app/skills"}
	for _, p := range skillPaths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			if err := skills.LoadFromFS(os.DirFS(p)); err != nil {
				l.Log(logger.Error, "Failed to load skills from "+p, err.Error())
			} else {
				l.Log(logger.Info, "Loaded skills", fmt.Sprintf("path=%s count=%d", p, skills.Count()))
			}
			return
		}
	}
	l.Log(logger.Warning, "No skills directory found", "")
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
