package store

import (
	"context"
	"encoding/json"
	"time"
)

// Event is a WebSocket-ready notification emitted when agent_results change.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Watcher polls agent_results every 2s and emits Events on new rows.
type Watcher struct {
	store    *Store
	out      chan<- Event
	interval time.Duration
}

func NewWatcher(s *Store, out chan<- Event) *Watcher {
	return &Watcher{store: s, out: out, interval: 2 * time.Second}
}

func (w *Watcher) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	var lastSeen int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			results, maxID, err := w.store.GetAgentResultsSince(lastSeen)
			if err != nil || len(results) == 0 {
				continue
			}
			for _, r := range results {
				payload, _ := json.Marshal(r)
				select {
				case w.out <- Event{Type: "agent_result", Payload: payload}:
				default:
				}
			}
			lastSeen = maxID
		}
	}
}
