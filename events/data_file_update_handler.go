package events

// DataFileUpdateHandler handles `EventTypeDataFileUpdate` SDK events.
type DataFileUpdateHandler interface {
	// OnUpdate is called when the SDK data file (configuration) is updated.
	//
	// Parameters:
	// - event: The data file update details.
	OnUpdate(event DataFileUpdateEvent)
}
