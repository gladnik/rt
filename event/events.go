package event

type EventBus struct {
	events chan Event
}

func NewEventBus() *EventBus {
	return &EventBus{events: make(chan Event)}
}

func (eb *EventBus) Events() <-chan Event {
	return eb.events
}

func (eb *EventBus) Fire(eventType string, id string) {
	go func() {
		eb.events <- Event{Type: eventType, Id: id}
	}()
}

// Events
const (
	LaunchStarted      = "launch_started"
	LaunchFinished     = "launch_finished"
	TestCaseStarted    = "test_case_started"
	TestCaseNotStarted = "test_case_not_started"
	TestCasePassed     = "test_case_finished"
	TestCaseFailed     = "test_case_failed"
	TestCaseRevoked    = "test_case_revoked"
	TestCaseTimedOut   = "test_case_timed_out"
)

type Event struct {
	Type string
	Id   string // Test case ID or launch ID
}
