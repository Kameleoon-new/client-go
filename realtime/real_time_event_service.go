package realtime

import (
	"encoding/json"
	"errors"
	"sync/atomic"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
	net "github.com/subchord/go-sse"
)

const dataFileUpdateEvent = "configuration-update-event"

// sseReconnectDelay is the delay before re-establishing a failed SSE connection, so a
// persistent failure does not turn into a hot reconnect loop. A variable rather than a
// constant, so tests can shorten it.
var sseReconnectDelay = 3 * time.Second

type RealTimeEventService struct {
	url        string
	updateChan chan RealTimeEvent
	sse        SseClient
	// closeFlag is accessed atomically (see `isClosed`);
	closeFlag int32
	// connected is accessed atomically (see `IsConnected`);
	connected int32
	closeChan chan bool
}

func NewRealTimeEventService(url string, updateChan chan RealTimeEvent,
	sse SseClient) *RealTimeEventService {
	rtcs := &RealTimeEventService{
		url:        url,
		updateChan: updateChan,
		sse:        sse,
		closeChan:  make(chan bool, 1),
	}
	go rtcs.run()
	return rtcs
}

func (rtcs *RealTimeEventService) Close() {
	if !atomic.CompareAndSwapInt32(&rtcs.closeFlag, 0, 1) {
		return
	}
	rtcs.closeChan <- true
}

func (rtcs *RealTimeEventService) isClosed() bool {
	return atomic.LoadInt32(&rtcs.closeFlag) != 0
}

func (rtcs *RealTimeEventService) IsConnected() bool {
	return atomic.LoadInt32(&rtcs.connected) != 0
}

func (rtcs *RealTimeEventService) run() {
	logging.Info("Real-Time Configuration Service started")
	if rtcs.sse == nil {
		logging.Error("SSE Client is not provided, Real-time Configuration Service is not started")
		// The channel is closed anyway, so its consumer is released rather than leaked.
		close(rtcs.updateChan)
		return
	}
	for !rtcs.isClosed() {
		rtcs.sse.Dispose()
		if err := rtcs.sse.Init(rtcs.url); err != nil {
			logging.Error("Failed to open SSE connection: %s. Next attempt in %s", err, sseReconnectDelay)
			rtcs.waitForReconnect()
			continue
		}
		logging.Info("SSE connection open")
		atomic.StoreInt32(&rtcs.connected, 1)
		rtcs.streamEvents()
		atomic.StoreInt32(&rtcs.connected, 0)
		logging.Info("SSE connection closed")
		if !rtcs.isClosed() {
			rtcs.waitForReconnect()
		}
	}
	close(rtcs.updateChan)
	rtcs.sse.Dispose()
	logging.Info("Real-Time Configuration Service stopped")
}

func (rtcs *RealTimeEventService) streamEvents() {
	for !rtcs.isClosed() {
		select {
		case <-rtcs.closeChan:
			return
		case err, ok := <-rtcs.sse.GetErrorChan():
			if ok {
				logging.Error("Error occurred within SSE client: %s", err)
			} else {
				logging.Error("SSE client closed its error channel")
			}
			return
		case evt, ok := <-rtcs.sse.GetEventChan():
			if !ok {
				// The client closed its event feed; treated as a connection failure.
				logging.Error("SSE client closed its event channel")
				return
			}
			logging.Info("Got %s SSE event", dataFileUpdateEvent)
			if err := rtcs.handleEvent(evt); err != nil {
				logging.Error("Error occurred during SSE event parsing: %s", err)
			}
		}
	}
}

func (rtcs *RealTimeEventService) waitForReconnect() {
	select {
	case <-rtcs.closeChan:
	case <-time.After(sseReconnectDelay):
	}
}

func (rtcs *RealTimeEventService) handleEvent(evt net.Event) error {
	if evt == nil {
		return errors.New("nil SSE event")
	}
	b := []byte(evt.GetData())
	var rtEvent RealTimeEvent
	if err := json.Unmarshal(b, &rtEvent); err != nil {
		return err
	}
	select {
	case rtcs.updateChan <- rtEvent:
	case <-rtcs.closeChan:
	}
	return nil
}
