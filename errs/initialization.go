package errs

import "fmt"

// Initialization indicates that the SDK could not be initialized: its configuration failed
// to load or no initialization result was available within the requested timeout.
// `KameleoonClient.WaitInit()` returns this error. The failure which prevented the
// initialization is available as the wrapped error (see `Unwrap`).
type Initialization struct {
	KameleoonError
	cause error
}

// NewInitialization creates an Initialization error. The `cause` parameter is the failure
// which prevented the SDK from loading its configuration; it is reported both in the
// message and as the wrapped error (see `Unwrap`).
func NewInitialization(siteCode string, environment string, cause error) *Initialization {
	return &Initialization{
		KameleoonError: NewKameleoonError(fmt.Sprintf(
			"SDK is not ready for siteCode: '%s', environment: '%s'. Reason: %s",
			siteCode, environment, describeCause(cause))),
		cause: cause,
	}
}

// Unwrap returns the failure which prevented the SDK from loading its configuration.
func (e *Initialization) Unwrap() error {
	return e.cause
}

// describeCause describes the cause by its message. It is reported in the message so a
// caller which only logs the message still learns why the SDK is not ready.
func describeCause(cause error) string {
	if cause == nil {
		return "unknown"
	}
	return cause.Error()
}
