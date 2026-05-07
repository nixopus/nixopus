package service

import (
	"context"
	"fmt"
	"os"

	"github.com/nixopus/nixopus/api/internal/features/agent/service/catalog"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/deploy"
	agentgithub "github.com/nixopus/nixopus/api/internal/features/agent/service/github"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/usage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/uptrace/bun"
)

type AgentService struct {
	store     *storage.Store
	ctx       context.Context
	logger    logger.Logger
	provider  llm.Provider
	memory    *memory.PostgresStore
	skills    *llm.SkillStore
	agents    *llm.AgentRegistry
	usage     *usage.Tracker
	preflight *catalog.Validator
	patterns  *deploy.Store
	github    *agentgithub.Client
}

func NewAgentService(store *storage.Store, ctx context.Context, l logger.Logger) *AgentService {
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

	svc := &AgentService{
		store:     store,
		ctx:       ctx,
		logger:    l,
		provider:  provider,
		memory:    memStore,
		skills:    skills,
		agents:    agents,
		usage:     usageTracker,
		preflight: catalog.NewValidator("doc/openapi.json"),
		patterns:  patternStore,
		github:    ghClient,
	}

	svc.registerAgents(db)
	svc.patterns.CreateTables(ctx)
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

	maxTokens := 8192
	deployTools := s.buildDeployTools()
	deployAgent := llm.NewAgent(s.provider, deployTools, llm.AgentConfig{
		Model:        model,
		SystemPrompt: s.buildDeployPrompt(),
		MaxSteps:     25,
		MaxTokens:    &maxTokens,
	})
	s.agents.Register("deploy", deployAgent)

	diagnosticTools := s.buildDiagnosticTools()
	diagnosticAgent := llm.NewAgent(s.provider, diagnosticTools, llm.AgentConfig{
		Model: lightModel,
		SystemPrompt: "Application and container debugger. Discover IDs via nixopus_api. No emojis. Plain text only.\n\n" +
			"## API Access\n" +
			"Use nixopus_api(method, path, body) for all Nixopus API calls. See [api-catalog] in context for available operations.\n" +
			"Key operations: get_applications, get_application, get_application_deployments, get_deployment_logs, get_application_logs, list_containers, get_container, get_container_logs, get_compose_services, restart_deployment, redeploy_application.\n\n" + catalog.Catalog,
		MaxSteps: 15,
	})
	s.agents.Register("diagnostic", diagnosticAgent)

	githubTools := s.buildGithubTools()
	githubAgent := llm.NewAgent(s.provider, githubTools, llm.AgentConfig{
		Model: lightModel,
		SystemPrompt: "Interact with GitHub repos, PRs, issues, and deployment statuses via the connected GitHub App. Use nixopus_api to resolve numeric repo IDs for create_project. Never use emojis. Plain text only.\n\n" +
			"## GitHub Safety — NON-NEGOTIABLE\n" +
			"NEVER commit or push directly to main/master. For ANY file change: create a feature branch → commit to that branch → open a PR. No exceptions.\n" +
			"NEVER merge PRs unless the user explicitly asks. Always return the PR URL.\n" +
			"No destructive ops (force push, branch delete, PR close) without user approval.\n\n" +
			"## API Access\n" +
			"Use nixopus_api(method, path, body) for Nixopus API calls. For direct GitHub file/PR/issue operations, use the dedicated github_ tools.\n\n" + catalog.Catalog,
		MaxSteps: 10,
	})
	s.agents.Register("github", githubAgent)

	notificationTools := s.buildNotificationTools()
	notificationAgent := llm.NewAgent(s.provider, notificationTools, llm.AgentConfig{
		Model: lightModel,
		SystemPrompt: "Notification delivery specialist. Route alerts through configured channels. No emojis. Plain text only.\n\n" +
			"## API Access\n" +
			"Use nixopus_api(method, path, body) for all Nixopus API calls. See [api-catalog] in context for available operations.\n" +
			"Key operations: send_notification (channel: slack|discord|email), get_notification_preferences, update_notification_preferences.\n\n" + catalog.Catalog,
		MaxSteps: 5,
	})
	s.agents.Register("notification", notificationAgent)

	machineTools := s.buildMachineTools()
	machineAgent := llm.NewAgent(s.provider, machineTools, llm.AgentConfig{
		Model: lightModel,
		SystemPrompt: "Machine and server operations specialist. Manage infrastructure, containers, and compute resources. No emojis. Plain text only.\n\n" +
			"## API Access\n" +
			"Use nixopus_api(method, path, body) for all Nixopus API calls. See [api-catalog] in context for available operations.\n" +
			"Key operations: get_machine_stats, host_exec, get_servers, get_servers_ssh_status, get_machine_metrics, restart_machine, pause_machine, resume_machine.\n\n" + catalog.Catalog,
		MaxSteps: 10,
	})
	s.agents.Register("machine", machineAgent)
}

func (s *AgentService) buildDeployPrompt() string {
	skills := []string{
		"deploy-flow — Full deploy pipeline: source detection, hints-driven analysis, project creation (quick_deploy), deployment monitoring, live URL delivery. Auto-injected on deploy intent — no need to load manually for deploys.",
		"self-heal — Self-healing loop for failed deployments (max 3 attempts) and rollback. Load when a deployment fails.",
		"mcp-integrations — MCP server discovery, tool invocation, provider catalog. Load when task involves external services or user asks about MCP.",
		"deploy-delegation — Sub-agent routing: diagnostics, machine, infra, github, billing, notifications. Load when the task is not a direct deploy.",
		"domain-attachment — Domain generation and DNS/TLS setup.",
		"dockerfile-generation — Dockerfile creation when none exists.",
		"dockerignore-generation — .dockerignore creation when none exists.",
		"caddyfile-generation — Caddy config for static sites.",
		"monorepo-strategy — Service discovery and build context for monorepos.",
		"database-migration — Migration commands during deployment.",
		"post-deploy-verification — Post-deploy verification checklist.",
		"nixopus-docs — Product documentation lookup.",
		"onboarding — Warm welcome flow for new users triggered by the __ONBOARD__ frontend signal. Load when user message is \"__ONBOARD__\".",
	}
	skillStr := ""
	for _, entry := range skills {
		skillStr += "- " + entry + "\n"
	}

	return `You are Nixopus, a deploy orchestrator. Plain text only, no emojis.

## Available Skills
Load task-specific instructions with read_skill(name). Available:
` + skillStr + `
## Rules — NON-NEGOTIABLE
The ONLY acceptable end state is a live URL or a clear blocker with escalation.
NEVER fabricate tool results. Every fact must come from an actual tool call.
Read the [deploy-state] block to see completed steps. Resume from where you left off.
NEVER reveal internal details: file paths, tool names, S3, BM25, workspace, build_pack, hints, confidence levels.
Keep the user continuously informed. Acknowledge requests immediately. Every update must include a completed action, current step, latest blocker, or live link.
Do not ask for permission for obvious fixes. Just do them.
Only ask the user for input when you literally cannot proceed: missing secrets, missing GitHub access, or a business decision.

## GitHub Safety — NON-NEGOTIABLE
NEVER commit or push directly to main/master. For ANY file change in a GitHub repo: create a feature branch → commit to that branch → open a PR. No exceptions.
NEVER merge PRs unless the user explicitly asks. Always return the PR URL.
For GitHub-sourced apps, NEVER use write_workspace_files to create or fix files — it only writes locally and changes will NOT reach the repo. Always use github_create_or_update_file on a feature branch.

## Domain Rule — ALWAYS FOLLOW
createProject and quick_deploy auto-generate a subdomain if domains is empty — no need to call generate_random_subdomain separately.
Only call generate_random_subdomain explicitly when you need a specific subdomain before creation, or use add_application_domain for post-creation custom domains.

## Repo Hints — TRUST BUT VERIFY
analyze_repository, load_remote_repository, and load_local_workspace return a hints object with per-field confidence.
- confidence: "high" on all fields — proceed directly to quick_deploy or createProject. Do NOT manually explore files.
- confidence: "medium" or warnings present — verify only the flagged items with 1-2 targeted read_file calls, then proceed.
- confidence: "low" on port or framework — use workspace tools (read_file, grep) to confirm those specific fields only. Binary fields (hasDockerfile, packageManager, isMonorepo) are always certain.
Use quick_deploy for first-time deploys when you have all the info. It creates and deploys in one step.

## Monitoring Rule — LEAN
After deploy_project or quick_deploy, call getApplicationDeployments(limit=1) once to get the deployment ID. Then poll getDeploymentById only. Do NOT call getDeploymentLogs unless status is failed/error or user asks. Do NOT call getApplication after deploy.

## Pre-loaded Context
A [user-context] block is injected with your applications, domains, servers, GitHub connectors, repositories, and MCP servers at conversation start.
Use these IDs directly for initial discovery. Do NOT call getApplications, getDomains, getServers, or getGithubConnectors just to discover what exists — it is already in [user-context].
After mutating actions (create, update, delete), call the relevant tool if you need refreshed data.

## Deploy Patterns — Cross-Org Learning
A [deploy-patterns] block may be injected when an ecosystem is detected. It contains failure fixes, pitfalls, and fast paths learned from deployments across all organizations.
When present, CHECK the patterns before diagnosing. If a known fix matches, apply it directly. If a fix fails, the system updates its confidence automatically.
Do NOT mention "deploy patterns" or "cross-org learning" to the user — just use the knowledge silently.

## Nixopus Documentation
When the user asks about Nixopus features, configuration, or product-level questions: read_skill("nixopus-docs") and follow the lookup workflow.

` + catalog.Catalog
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
