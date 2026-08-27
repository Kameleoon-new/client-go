package events

import "time"

// HttpRequestHandler handles `EventTypeHttpRequest` SDK events. The handler is called once per
// each actual HTTP request attempt, including retries.
type HttpRequestHandler interface {
	// OnRequestSucceeded is called when an SDK HTTP request completes successfully.
	//
	// Parameters:
	// - requestType: The type of the request, one of the `RequestType` constants.
	// - httpStatus: The HTTP status code of the response.
	// - duration: The duration of the request.
	OnRequestSucceeded(requestType RequestType, httpStatus int, duration time.Duration)

	// OnRequestFailed is called when an SDK HTTP request fails.
	//
	// Parameters:
	// - requestType: The type of the request, one of the `RequestType` constants.
	// - failure: The failure details.
	// - duration: The duration of the request.
	OnRequestFailed(requestType RequestType, failure *HttpRequestFailure, duration time.Duration)
}
