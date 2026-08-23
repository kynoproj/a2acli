package cli

import (
	"context"
	"errors"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// TestStreamEventsCapabilityCheck asserts streamEvents' up-front capability
// gate yields a wrapped a2a.ErrUnsupportedOperation without touching client
// when card advertises no streaming support, so passing a nil
// *a2aclient.Client here is safe.
func TestStreamEventsCapabilityCheck(t *testing.T) {
	card := &a2a.AgentCard{Capabilities: a2a.AgentCapabilities{Streaming: false}}
	_, err := firstEvent(streamEvents(card, nil, context.Background(), &a2a.SendMessageRequest{}))
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, a2a.ErrUnsupportedOperation) {
		t.Errorf("error = %v, want wrapping %v", err, a2a.ErrUnsupportedOperation)
	}
}

// firstEvent drains an iter.Seq2 for its first (event, error) pair.
func firstEvent(events func(func(a2a.Event, error) bool)) (a2a.Event, error) {
	var event a2a.Event
	var err error
	for e, er := range events {
		event, err = e, er
		break
	}
	return event, err
}
