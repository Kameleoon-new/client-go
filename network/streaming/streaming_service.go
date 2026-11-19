package streaming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
)

const (
	// The server pushes `configuration-update-event` when a configuration is published and
	// `connection-open-event` on every (re)connection; both carry the timestamp of the current
	// configuration.
	configurationUpdateEvent = "configuration-update-event"
	connectionOpenEvent      = "connection-open-event"
	// maxRetries is the number of reconnections attempted, without a received event in
	// between, before the service gives up.
	maxRetries = 1
	// sseReconnectDelay prevents a persistent failure from becoming a hot reconnect loop.
	sseReconnectDelay = 10 * time.Second
)

// StreamingService keeps an SSE connection to `url` and reports the configuration events it
// receives. Every end of a connection is followed by a delayed reconnection, whether it is a
// routine close by the server (which recycles connections periodically) or a failure (error
// response, network error, stalled stream). Every received event resets the count of
// disconnects, and the server sends the connection open event on every connection, so a
// working stream keeps the count low. Once more than `maxRetries` connections in a row end
// without an event, the service stops itself and reports it with `onUnavailable`, which
// receives the service so a late report of a replaced service can be told from the current
// one. `onDataFileUpdate` is called from the service goroutine, so the stream is not read
// while it runs.
type StreamingService struct {
	url              string
	sse              SseClient
	reconnectDelay   time.Duration
	onDataFileUpdate func(DataFileUpdateMessage)
	onUnavailable    func(*StreamingService)
	ctx              context.Context
	cancel           context.CancelFunc
	// consecutiveDisconnects is only accessed by the service goroutine
	consecutiveDisconnects int
}

func NewStreamingService(url string, sse SseClient, onDataFileUpdate func(DataFileUpdateMessage),
	onUnavailable func(*StreamingService)) *StreamingService {
	return newStreamingService(url, sse, onDataFileUpdate, onUnavailable, sseReconnectDelay)
}

func newStreamingService(url string, sse SseClient, onDataFileUpdate func(DataFileUpdateMessage),
	onUnavailable func(*StreamingService), reconnectDelay time.Duration) *StreamingService {
	ctx, cancel := context.WithCancel(context.Background())
	s := &StreamingService{
		url:              url,
		sse:              sse,
		reconnectDelay:   reconnectDelay,
		onDataFileUpdate: onDataFileUpdate,
		onUnavailable:    onUnavailable,
		ctx:              ctx,
		cancel:           cancel,
	}
	go s.run()
	return s
}

// Close stops the service. It does not wait for the service goroutine, which may be running
// `onDataFileUpdate` at that moment.
func (s *StreamingService) Close() {
	s.cancel()
}

func (s *StreamingService) isClosed() bool {
	return s.ctx.Err() != nil
}

func (s *StreamingService) run() {
	if s.sse == nil {
		logging.Error("SSE Client is not provided, configuration streaming service is not started")
		s.reportUnavailable()
		return
	}
	logging.Info("Starting configuration streaming service")
	for !s.isClosed() {
		stream, err := s.sse.Connect(s.ctx, s.url)
		connected := err == nil
		if connected {
			logging.Info("Established a new SSE connection with: %s", s.url)
			err = s.streamEvents(stream)
			stream.Close()
		}
		if s.isClosed() {
			break
		}
		s.consecutiveDisconnects++
		givingUp := s.consecutiveDisconnects > maxRetries
		s.logDisconnect(err, connected && isClosedByServer(err), givingUp)
		if givingUp {
			s.reportUnavailable()
			return
		}
		s.waitForReconnect()
	}
	logging.Info("SSE connection: %s was closed by SDK", s.url)
}

// logDisconnect logs why a connection ended and what follows. A close by the server is routine
// (it recycles connections periodically), so it is an error only when the service gives up.
func (s *StreamingService) logDisconnect(err error, closedByServer bool, givingUp bool) {
	level := logging.ERROR
	if closedByServer && !givingUp {
		level = logging.INFO
	}
	logging.Log(level, func() string {
		reason := "was closed by server"
		if !closedByServer {
			reason = fmt.Sprintf("failed: %s", err)
		}
		outcome := fmt.Sprintf("reconnecting in %d ms", s.reconnectDelay.Milliseconds())
		if givingUp {
			outcome = fmt.Sprintf("streaming is unavailable after %d consecutive disconnects, switching to polling",
				s.consecutiveDisconnects)
		}
		return fmt.Sprintf("SSE connection: '%s' %s; %s", s.url, reason, outcome)
	})
}

// reportUnavailable stops the service and reports it. No connection is open at that point,
// so stopping only marks the service closed.
func (s *StreamingService) reportUnavailable() {
	s.cancel()
	if s.onUnavailable != nil {
		s.onUnavailable(s)
	}
}

// streamEvents reports the configuration events of `stream` until it ends, and returns the
// error which ended it.
func (s *StreamingService) streamEvents(stream SseStream) error {
	for {
		evt, err := stream.Next()
		if err != nil {
			return err
		}
		// Closing does not discard the events which the stream has already buffered
		if s.isClosed() {
			return s.ctx.Err()
		}
		// Only a delivered event proves the stream works, whatever its type: a connection may be
		// established (e.g. by a proxy) and never deliver anything.
		s.consecutiveDisconnects = 0
		if evt.Name != configurationUpdateEvent && evt.Name != connectionOpenEvent {
			continue
		}
		logging.Info("Received %s SSE event: %s", evt.Name, evt.Data)
		var message DataFileUpdateMessage
		if err := json.Unmarshal([]byte(evt.Data), &message); err != nil {
			logging.Error("Received event from SSE server with wrong format: %s; %s", evt.Data, err)
			continue
		}
		s.onDataFileUpdate(message)
	}
}

// isClosedByServer tells a routine end of an established stream from a failure. The
// streaming server recycles connections periodically, which the SDK sees as an EOF.
func isClosedByServer(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}

// waitForReconnect waits for the reconnect delay, or until the service is closed.
func (s *StreamingService) waitForReconnect() {
	timer := time.NewTimer(s.reconnectDelay)
	defer timer.Stop()
	select {
	case <-s.ctx.Done():
	case <-timer.C:
	}
}
