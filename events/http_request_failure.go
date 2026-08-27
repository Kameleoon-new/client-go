package events

import (
	"fmt"
	"strconv"
)

// HttpRequestFailureReason represents the reason of an SDK HTTP request failure.
type HttpRequestFailureReason string

const (
	// FailureReasonHttpStatus indicates the request completed with an unexpected HTTP status code.
	FailureReasonHttpStatus HttpRequestFailureReason = "http_status"

	// FailureReasonError indicates the request failed with an error (e.g. a network error).
	FailureReasonError HttpRequestFailureReason = "error"

	// FailureReasonCancelled indicates the request was cancelled (for example, due to a timeout).
	FailureReasonCancelled HttpRequestFailureReason = "cancelled"
)

// HttpRequestFailure describes the failure of an SDK HTTP request reported to
// `HttpRequestHandler`.
type HttpRequestFailure struct {
	reason     HttpRequestFailureReason
	httpStatus *int
	cause      error
}

// NewHttpRequestFailureFromHttpStatus creates an `HttpRequestFailure` for a request completed
// with an unexpected HTTP status code.
func NewHttpRequestFailureFromHttpStatus(httpStatus int) *HttpRequestFailure {
	return &HttpRequestFailure{reason: FailureReasonHttpStatus, httpStatus: &httpStatus}
}

// NewHttpRequestFailureFromError creates an `HttpRequestFailure` for a request failed
// with an error.
func NewHttpRequestFailureFromError(cause error) *HttpRequestFailure {
	return &HttpRequestFailure{reason: FailureReasonError, cause: cause}
}

// NewHttpRequestFailureOfCancellation creates an `HttpRequestFailure` for a cancelled request.
func NewHttpRequestFailureOfCancellation() *HttpRequestFailure {
	return &HttpRequestFailure{reason: FailureReasonCancelled}
}

// Reason returns the reason of the failure.
func (f *HttpRequestFailure) Reason() HttpRequestFailureReason {
	return f.reason
}

// HttpStatus returns the HTTP status code of the response. Not nil only if `Reason()` is
// `FailureReasonHttpStatus`.
func (f *HttpRequestFailure) HttpStatus() *int {
	return f.httpStatus
}

// Cause returns the error caused the failure. Not nil only if `Reason()` is
// `FailureReasonError`.
func (f *HttpRequestFailure) Cause() error {
	return f.cause
}

func (f *HttpRequestFailure) String() string {
	httpStatus := "<nil>"
	if f.httpStatus != nil {
		httpStatus = strconv.Itoa(*f.httpStatus)
	}
	return fmt.Sprintf("HttpRequestFailure{Reason:'%s',HttpStatus:%s,Cause:'%v'}",
		f.reason, httpStatus, f.cause)
}
