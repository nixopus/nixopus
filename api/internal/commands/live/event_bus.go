package live

// EventType identifies the kind of deployment event.
type EventType string

const (
	EventAuth             EventType = "auth"
	EventConfig           EventType = "config"
	EventEnvCheck         EventType = "env_check"
	EventFileDetected     EventType = "file_detected"
	EventConnecting       EventType = "connecting"
	EventConnected        EventType = "connected"
	EventDisconnected     EventType = "disconnected"
	EventReconnecting     EventType = "reconnecting"
	EventSyncStart        EventType = "sync_start"
	EventSyncProgress     EventType = "sync_progress"
	EventSyncDone         EventType = "sync_done"
	EventBuildStarted     EventType = "build_started"
	EventBuildFailed      EventType = "build_failed"
	EventBuildSuccess     EventType = "build_success"
	EventDeploying        EventType = "deploying"
	EventDeployed         EventType = "deployed"
	EventDeployFailed     EventType = "deploy_failed"
	EventError            EventType = "error"
	EventInfo             EventType = "info"
	EventPipelineProgress EventType = "pipeline_progress"
	EventBuildStatus      EventType = "build_status"
	EventBuildLog         EventType = "build_log"
	EventDeploymentStatus EventType = "deployment_status"
)

// EventPayload is the interface for type-safe event data.
// Each event type that carries structured data gets its own struct.
type EventPayload interface{ eventPayload() }

// BuildFailedPayload carries build failure details.
type BuildFailedPayload struct {
	Logs  []string
	Error string
}

func (BuildFailedPayload) eventPayload() {}

// SyncProgressPayload carries file sync counters.
type SyncProgressPayload struct {
	FilesSynced int
	Total       int
}

func (SyncProgressPayload) eventPayload() {}

// EnvCheckPayload carries environment variable check results.
type EnvCheckPayload struct {
	Missing []string
	Found   []string
}

func (EnvCheckPayload) eventPayload() {}

// DeployedPayload carries deployment completion details.
type DeployedPayload struct {
	URL string
}

func (DeployedPayload) eventPayload() {}

// PipelineProgressPayload carries real-time progress from the pipeline agent.
type PipelineProgressPayload struct {
	StageId string
	Message string
}

func (PipelineProgressPayload) eventPayload() {}

// BuildStatusPayload carries build lifecycle events from the server.
type BuildStatusPayload struct {
	Phase   string // "starting", "generating_dockerfile", "dockerfile_ready", "building_container", "error"
	Message string
	Error   string
}

func (BuildStatusPayload) eventPayload() {}

// BuildLogPayload carries a single build log line from the database.
type BuildLogPayload struct {
	Log       string
	Timestamp string
}

func (BuildLogPayload) eventPayload() {}

// DeploymentStatusPayload carries a deployment status change from the database.
type DeploymentStatusPayload struct {
	Status       string
	DeploymentID string
}

func (DeploymentStatusPayload) eventPayload() {}

// Event is a deployment lifecycle event published by various components
// (auth, sync engine, poller, watcher) and consumed by the AgentUI.
type Event struct {
	Type    EventType
	Message string
	Payload EventPayload // nil when no structured data is needed

	// NeedsLLM controls routing: true sends to the agent for an intelligent response,
	// false prints directly to the terminal via the writer.
	NeedsLLM bool
}

// isPriority returns true for events that must never be dropped.
func (e Event) isPriority() bool {
	if e.NeedsLLM ||
		e.Type == EventError ||
		e.Type == EventDeployed ||
		e.Type == EventBuildFailed ||
		e.Type == EventDeployFailed ||
		e.Type == EventBuildSuccess {
		return true
	}
	// Build status errors are priority so the CLI always shows them
	if e.Type == EventBuildStatus {
		if p, ok := e.Payload.(BuildStatusPayload); ok && p.Phase == "error" {
			return true
		}
	}
	return false
}

// EventBus routes events through two internal channels:
//   - priority: unbuffered, blocks the sender until consumed — for events that must not be lost.
//   - normal: buffered, drops on overflow — for high-frequency status ticks.
type EventBus struct {
	priority chan Event
	normal   chan Event
}

// NewEventBus creates an event bus with the given normal-channel buffer size.
func NewEventBus(normalBuf int) *EventBus {
	return &EventBus{
		priority: make(chan Event),
		normal:   make(chan Event, normalBuf),
	}
}

// Send routes the event to the appropriate channel based on priority.
// Priority events block the caller until consumed. Normal events are dropped if the buffer is full.
func (b *EventBus) Send(ev Event) {
	if ev.isPriority() {
		b.priority <- ev
		return
	}
	select {
	case b.normal <- ev:
	default:
	}
}

// Priority returns the channel for high-priority events.
func (b *EventBus) Priority() <-chan Event { return b.priority }

// Normal returns the channel for normal events.
func (b *EventBus) Normal() <-chan Event { return b.normal }

// Close closes both channels. Only the owner (producer side) should call this.
func (b *EventBus) Close() {
	close(b.priority)
	close(b.normal)
}
