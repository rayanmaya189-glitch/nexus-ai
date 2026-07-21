package nats

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewEvent(eventType, source, subject string, data map[string]interface{}) *Event {
	return &Event{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Type:      eventType,
		Source:    source,
		Subject:   subject,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func (e *Event) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func UnmarshalEvent(data []byte) (*Event, error) {
	event := &Event{}
	if err := json.Unmarshal(data, event); err != nil {
		return nil, err
	}
	return event, nil
}

type EventHandler func(event *Event) error

type Publisher interface {
	Publish(subject string, event *Event) error
	Close()
}

type Subscriber interface {
	Subscribe(subject string, handler EventHandler) error
	Close()
}

type EventBridge interface {
	Publisher
	Subscriber
}

// Event types
const (
	EventUserCreated      = "user.created"
	EventUserUpdated      = "user.updated"
	EventUserDeleted      = "user.deleted"
	EventTenantCreated    = "tenant.created"
	EventTenantUpdated    = "tenant.updated"
	EventAgentCreated     = "agent.created"
	EventAgentUpdated     = "agent.updated"
	EventAgentExecuted    = "agent.executed"
	EventDocumentIngested = "document.ingested"
	EventDocumentDeleted  = "document.deleted"
	EventWorkflowStarted  = "workflow.started"
	EventWorkflowCompleted = "workflow.completed"
	EventAuditLogCreated  = "audit.log.created"
	EventModelPulled      = "model.pulled"
	EventSecurityScan     = "security.scan"
	EventNotificationSent = "notification.sent"
)

// Subject patterns
const (
	SubjectUser      = "nexus.users"
	SubjectTenant    = "nexus.tenants"
	SubjectAgent     = "nexus.agents"
	SubjectDocument  = "nexus.documents"
	SubjectWorkflow  = "nexus.workflows"
	SubjectAudit     = "nexus.audit"
	SubjectModel     = "nexus.models"
	SubjectSecurity  = "nexus.security"
	SubjectNotification = "nexus.notifications"
	SubjectEvent     = "nexus.events"
)
