package live

// EventType identifies the kind of deployment event.
type EventType string

const (
	EventAuth                     EventType = "auth"
	EventConfig                   EventType = "config"
	EventConnecting               EventType = "connecting"
	EventConnected                EventType = "connected"
	EventDisconnected             EventType = "disconnected"
	EventReconnecting             EventType = "reconnecting"
	EventSyncStart                EventType = "sync_start"
	EventSyncProgress             EventType = "sync_progress"
	EventError                    EventType = "error"
	EventInfo                     EventType = "info"
	EventPipelineProgress         EventType = "pipeline_progress"
	EventBuildStatus              EventType = "build_status"
	EventBuildLog                 EventType = "build_log"
	EventDeploymentStatus         EventType = "deployment_status"
	EventApprovalNeeded           EventType = "approval_needed"
	EventDeploymentReasoningChunk EventType = "deployment_reasoning_chunk"
)

// EventPayload is the interface for type-safe event data.
// Each event type that carries structured data gets its own struct.
type EventPayload interface{ eventPayload() }

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

// BuildLogPayload carries a single build log line from the workflow stream (listen-for-build).
type BuildLogPayload struct {
	Step string
	Log  string
}

func (BuildLogPayload) eventPayload() {}

// DeploymentStatusPayload carries a deployment status change from the database.
type DeploymentStatusPayload struct {
	Status       string
	DeploymentID string
}

func (DeploymentStatusPayload) eventPayload() {}

// ApprovalNeededPayload carries a proposal for human approval (Dockerfile, config, or other).
type ApprovalNeededPayload struct {
	Dockerfile      string
	Summary         string
	ValidationScore int
	Suggestions     []string
}

func (ApprovalNeededPayload) eventPayload() {}

// DeploymentReasoningChunkPayload carries a streaming reasoning chunk from the agent.
// Clients append chunks to show LLM reasoning progressively; the final deployment-progress provides the complete message.
type DeploymentReasoningChunkPayload struct {
	Step  string
	Chunk string
}

func (DeploymentReasoningChunkPayload) eventPayload() {}

// Event is a deployment lifecycle event published by various components
// (auth, sync engine, poller, watcher) and consumed by the AgentUI.
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
	if e.NeedsLLM || e.Type == EventError || e.Type == EventApprovalNeeded {
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
