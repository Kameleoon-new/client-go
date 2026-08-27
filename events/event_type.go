package events

// EventType represents an SDK event type which can be handled with
// `KameleoonClient.SetEventHandler`.
type EventType string

const (
	// EventTypeHttpRequest notifies when an SDK HTTP request completes successfully or fails.
	// Requires a handler implementing the `HttpRequestHandler` interface.
	EventTypeHttpRequest EventType = "http_request"

	// EventTypeDataFileUpdate notifies when the SDK data file (configuration) is updated.
	// Requires a handler implementing the `DataFileUpdateHandler` interface.
	EventTypeDataFileUpdate EventType = "datafile_update"
)
