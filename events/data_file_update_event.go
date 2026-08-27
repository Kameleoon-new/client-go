package events

import "fmt"

// DataFileUpdateSource represents the source of a data file update.
type DataFileUpdateSource string

const (
	// DataFileUpdateSourcePolling indicates the data file was updated by periodic polling.
	DataFileUpdateSourcePolling DataFileUpdateSource = "polling"

	// DataFileUpdateSourceStreaming indicates the data file was updated by a real-time
	// (streaming) notification.
	DataFileUpdateSourceStreaming DataFileUpdateSource = "streaming"
)

// DataFileUpdateEvent describes an update of the SDK data file (configuration) reported to
// `DataFileUpdateHandler`.
type DataFileUpdateEvent struct {
	// Source is the source of the data file update.
	Source DataFileUpdateSource

	// DateModified is the date of the last modification of the data file,
	// in Unix time milliseconds.
	DateModified int64
}

func (e DataFileUpdateEvent) String() string {
	return fmt.Sprintf("DataFileUpdateEvent{Source:'%s',DateModified:%d}", e.Source, e.DateModified)
}
