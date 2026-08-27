package events

// RequestType represents the type of an HTTP request performed by the SDK,
// reported to `HttpRequestHandler`.
type RequestType string

const (
	// RequestTypeDataFile is fetching of the SDK data file (configuration).
	RequestTypeDataFile RequestType = "datafile"

	// RequestTypeTracking is sending of tracking data.
	RequestTypeTracking RequestType = "tracking"

	// RequestTypeRemoteVisitorData is fetching of remote visitor data.
	RequestTypeRemoteVisitorData RequestType = "remote_visitor_data"

	// RequestTypeRemoteData is fetching of remote data.
	RequestTypeRemoteData RequestType = "remote_data"

	// RequestTypeAccessToken is fetching of the Kameleoon API access token.
	RequestTypeAccessToken RequestType = "access_token"
)
