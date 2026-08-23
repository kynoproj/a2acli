package cli

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2aevent"
)

// successiveFailureThreshold is the number of consecutive GetTask errors that
// abort a polling loop, matching upstream's fallback behavior.
const successiveFailureThreshold = 3

// streamEvents attempts a real streaming call, yielding a2a.ErrUnsupportedOperation
// up front if card advertises no streaming support (mirroring upstream's
// handleStreaming check) so the caller can fall back to polling. A nil card
// (e.g. the --endpoint path, which bypasses AgentCard resolution) is treated
// as streaming-capable so behavior is unchanged when no card is available.
func streamEvents(card *a2a.AgentCard, client *a2aclient.Client, ctx context.Context, req *a2a.SendMessageRequest) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		if card != nil && !card.Capabilities.Streaming {
			yield(nil, fmt.Errorf("streaming not listed in agent capabilities: %w", a2a.ErrUnsupportedOperation))
			return
		}
		for event, err := range client.SendStreamingMessage(ctx, req) {
			if !yield(event, err) {
				return
			}
		}
	}
}

// handlePolling sends req with ReturnImmediately forced on, then polls
// GetTask every interval, yielding synthetic events recovered from the
// task-state diff until the task reaches a terminal or input-required state.
func handlePolling(ctx context.Context, client *a2aclient.Client, original *a2a.SendMessageRequest, interval time.Duration) iter.Seq2[a2a.Event, error] {
	return func(yield func(a2a.Event, error) bool) {
		req := *original
		if req.Config == nil {
			req.Config = &a2a.SendMessageConfig{ReturnImmediately: true}
		} else if !req.Config.ReturnImmediately {
			config := *original.Config
			config.ReturnImmediately = true
			req.Config = &config
		}

		result, err := client.SendMessage(ctx, &req)
		if err != nil {
			yield(nil, fmt.Errorf("failed to send message: %w", err))
			return
		}
		if !yield(result, nil) {
			return
		}
		if _, ok := result.(*a2a.Message); ok {
			return
		}
		prevState, ok := result.(*a2a.Task)
		if !ok {
			yield(nil, fmt.Errorf("unexpected send result type: %T", result))
			return
		}
		tid := prevState.ID

		successiveFailures := 0
		for !prevState.Status.State.Terminal() && prevState.Status.State != a2a.TaskStateInputRequired {
			select {
			case <-ctx.Done():
				yield(nil, ctx.Err())
				return
			case <-time.After(interval):
			}

			task, err := client.GetTask(ctx, &a2a.GetTaskRequest{ID: tid})
			if err != nil {
				successiveFailures++
				if successiveFailures == successiveFailureThreshold {
					yield(nil, fmt.Errorf("successive polling failure threshold exceeded for task %q: %w", tid, err))
					return
				}
				continue
			}
			successiveFailures = 0

			var events []a2a.Event
			if task.Status.State.Terminal() || task.Status.State == a2a.TaskStateInputRequired {
				events = append(events, task)
			} else {
				events = a2aevent.Recover(prevState, task)
			}

			for _, event := range events {
				if !yield(event, nil) {
					return
				}
			}
			prevState = task
		}
	}
}
