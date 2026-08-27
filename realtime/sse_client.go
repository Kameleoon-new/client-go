package realtime

import (
	"sync"

	"github.com/Kameleoon/client-go/v3/logging"

	net "github.com/subchord/go-sse"
)

type SseClient interface {
	Init(url string) error
	Dispose()
	GetErrorChan() <-chan error
	GetEventChan() <-chan net.Event
}

// NetSseClient adapts `go-sse` to the `SseClient` interface. A single instance is shared
// by consecutive `RealTimeEventService` instances, and two services can briefly overlap
// during a mode switch (the old one disposing while the new one initializes); the mutex
// keeps such an overlap free of races and panics. At worst it closes the fresh connection,
// which the owning service re-establishes on its next reconnect attempt.
type NetSseClient struct {
	mx   sync.Mutex
	feed *net.SSEFeed
	sub  *net.Subscription
}

func (sse *NetSseClient) Init(url string) error {
	headers := map[string][]string{
		"Accept":        {"text/event-stream"},
		"Cache-Control": {"no-cache"},
		"Connection":    {"Keep-Alive"},
	}
	feed, err := net.ConnectWithSSEFeed(url, headers)
	if err != nil {
		return err
	}
	sub, err := feed.Subscribe(dataFileUpdateEvent)
	if err != nil {
		feed.Close() // the connection is not leaked on a partially failed init
		return err
	}
	sse.mx.Lock()
	sse.feed = feed
	sse.sub = sub
	sse.mx.Unlock()
	return nil
}

func (sse *NetSseClient) Dispose() {
	sse.mx.Lock()
	feed, sub := sse.feed, sse.sub
	sse.feed = nil
	sse.sub = nil
	sse.mx.Unlock()
	if feed == nil {
		return
	}
	disposed := make(chan struct{})
	if sub != nil {
		go func() {
			for {
				select {
				case _, ok := <-sub.Feed():
					if !ok {
						return
					}
				case <-disposed:
					return
				}
			}
		}()
	}
	defer close(disposed)
	defer func() {
		if err := recover(); err != nil {
			logging.Warning("Panic occurred during SSE dispose process: %s", err)
		}
	}()
	feed.Close()
}

// GetErrorChan returns the error channel of the current subscription, or nil (a channel
// which never delivers) when the client is not initialized.
func (sse *NetSseClient) GetErrorChan() <-chan error {
	sse.mx.Lock()
	defer sse.mx.Unlock()
	if sse.sub == nil {
		return nil
	}
	return sse.sub.ErrFeed()
}

// GetEventChan returns the event channel of the current subscription, or nil (a channel
// which never delivers) when the client is not initialized.
func (sse *NetSseClient) GetEventChan() <-chan net.Event {
	sse.mx.Lock()
	defer sse.mx.Unlock()
	if sse.sub == nil {
		return nil
	}
	return sse.sub.Feed()
}
