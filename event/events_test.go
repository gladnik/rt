package event

import (
	"testing"
	. "github.com/aandryashin/matchers"
)

func TestFireAndConsume(t *testing.T) {
	eventBus := NewEventBus()
	eventBus.Fire(LaunchStarted, "test-id")
	event := <-eventBus.Events()
	AssertThat(t, event, EqualTo{
		Event{
			Type: LaunchStarted,
			Id:"test-id",
		},
	})
}
