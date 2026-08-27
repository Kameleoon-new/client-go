package events

import (
	"sync/atomic"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
)

// EventManager stores SDK event handlers and fires SDK events.
type EventManager interface {
	// SetEventHandler sets the SDK event handler for the specified event type.
	// Passing nil clears the handler for the specified event type.
	SetEventHandler(eventType EventType, handler EventHandler)

	// FireHttpRequestSucceeded fires an `EventTypeHttpRequest` SDK event for a request
	// completed successfully.
	FireHttpRequestSucceeded(requestType RequestType, httpStatus int, duration time.Duration)

	// FireHttpRequestFailed fires an `EventTypeHttpRequest` SDK event for a failed request.
	FireHttpRequestFailed(requestType RequestType, failure *HttpRequestFailure, duration time.Duration)

	// FireDataFileUpdate fires an `EventTypeDataFileUpdate` SDK event.
	FireDataFileUpdate(event DataFileUpdateEvent)
}

// Handlers are stored in `atomic.Value` to provide lock-free reads on the event firing paths.
// The holder structs are required because `atomic.Value` accepts neither nil values nor values
// of different concrete types.
type httpRequestHandlerHolder struct {
	handler HttpRequestHandler
}

type dataFileUpdateHandlerHolder struct {
	handler DataFileUpdateHandler
}

type EventManagerImpl struct {
	httpRequestHandler    atomic.Value // httpRequestHandlerHolder
	dataFileUpdateHandler atomic.Value // dataFileUpdateHandlerHolder
}

func NewEventManager() *EventManagerImpl {
	return &EventManagerImpl{}
}

func (em *EventManagerImpl) SetEventHandler(eventType EventType, handler EventHandler) {
	switch eventType {
	case EventTypeHttpRequest:
		em.setHttpRequestHandler(eventType, handler)
	case EventTypeDataFileUpdate:
		em.setDataFileUpdateHandler(eventType, handler)
	default:
		logging.Error("Unknown event type '%s'", eventType)
	}
}

func (em *EventManagerImpl) setHttpRequestHandler(eventType EventType, handler EventHandler) {
	var httpRequestHandler HttpRequestHandler
	if handler != nil {
		var ok bool
		if httpRequestHandler, ok = handler.(HttpRequestHandler); !ok {
			logging.Error("Handler for event type '%s' must implement the events.HttpRequestHandler interface",
				eventType)
			return
		}
	}
	em.httpRequestHandler.Store(httpRequestHandlerHolder{handler: httpRequestHandler})
}

func (em *EventManagerImpl) setDataFileUpdateHandler(eventType EventType, handler EventHandler) {
	var dataFileUpdateHandler DataFileUpdateHandler
	if handler != nil {
		var ok bool
		if dataFileUpdateHandler, ok = handler.(DataFileUpdateHandler); !ok {
			logging.Error("Handler for event type '%s' must implement the events.DataFileUpdateHandler interface",
				eventType)
			return
		}
	}
	em.dataFileUpdateHandler.Store(dataFileUpdateHandlerHolder{handler: dataFileUpdateHandler})
}

func (em *EventManagerImpl) FireHttpRequestSucceeded(
	requestType RequestType, httpStatus int, duration time.Duration,
) {
	handler := em.getHttpRequestHandler()
	if handler == nil {
		return
	}
	defer logHandlerPanic("HTTP request")
	handler.OnRequestSucceeded(requestType, httpStatus, duration)
}

func (em *EventManagerImpl) FireHttpRequestFailed(
	requestType RequestType, failure *HttpRequestFailure, duration time.Duration,
) {
	handler := em.getHttpRequestHandler()
	if handler == nil {
		return
	}
	defer logHandlerPanic("HTTP request")
	handler.OnRequestFailed(requestType, failure, duration)
}

func (em *EventManagerImpl) FireDataFileUpdate(event DataFileUpdateEvent) {
	handler := em.getDataFileUpdateHandler()
	if handler == nil {
		return
	}
	defer logHandlerPanic("Data file update")
	handler.OnUpdate(event)
}

func (em *EventManagerImpl) getHttpRequestHandler() HttpRequestHandler {
	holder, ok := em.httpRequestHandler.Load().(httpRequestHandlerHolder)
	if !ok {
		return nil
	}
	return holder.handler
}

func (em *EventManagerImpl) getDataFileUpdateHandler() DataFileUpdateHandler {
	holder, ok := em.dataFileUpdateHandler.Load().(dataFileUpdateHandlerHolder)
	if !ok {
		return nil
	}
	return holder.handler
}

func logHandlerPanic(eventName string) {
	if r := recover(); r != nil {
		logging.Warning("%s event handler failed: %s", eventName, r)
	}
}
