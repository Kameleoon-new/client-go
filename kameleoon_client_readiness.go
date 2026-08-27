package kameleoon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Kameleoon/client-go/v3/errs"
	"github.com/Kameleoon/client-go/v3/logging"
)

// readinessState is a single settlement of the readiness state: `err` is written at most
// once, before `done` is closed, and never changes afterwards, so a waiter which obtained
// the state observes exactly the outcome it was registered for.
type readinessState struct {
	done chan struct{}
	err  error
}

func newReadinessState() *readinessState {
	return &readinessState{done: make(chan struct{})}
}

func (s *readinessState) isSettled() bool {
	select {
	case <-s.done:
		return true
	default:
		return false
	}
}

// kameleoonClientReadiness tracks whether the SDK has successfully loaded its configuration
// and exposes that state to `KameleoonClient.WaitInit()` and `KameleoonClient.IsReady()`:
//   - a successful fetch settles the state successfully;
//   - a failed fetch settles it with an `errs.Initialization` error.
//
// Readiness is monotonic: once the SDK is ready it stays ready, and a later failed fetch
// can never revert it. A failure reported before the first success does not prevent a
// subsequent successful retry from marking the SDK ready.
type kameleoonClientReadiness struct {
	mx          sync.Mutex
	state       *readinessState
	siteCode    string
	environment string
}

func newKameleoonClientReadiness(siteCode string, environment string) *kameleoonClientReadiness {
	logging.Debug("CALL/RETURN: newKameleoonClientReadiness(siteCode: %s, environment: %s)",
		siteCode, environment)
	return &kameleoonClientReadiness{
		state:       newReadinessState(),
		siteCode:    siteCode,
		environment: environment,
	}
}

// MarkReady marks the SDK ready and releases all waiters.
func (r *kameleoonClientReadiness) MarkReady() {
	logging.Debug("CALL: kameleoonClientReadiness.MarkReady(siteCode: %s, environment: %s)",
		r.siteCode, r.environment)
	r.mx.Lock()
	if r.state.isSettled() {
		if r.state.err == nil {
			r.mx.Unlock()
			logging.Debug("RETURN: kameleoonClientReadiness.MarkReady(siteCode: %s, environment: %s)",
				r.siteCode, r.environment)
			return // already ready - readiness is monotonic
		}
		// Recovery after a failure: the failed state is left settled for its waiters,
		// and a fresh, ready state is published for the new ones.
		state := newReadinessState()
		close(state.done)
		r.state = state
	} else {
		close(r.state.done)
	}
	r.mx.Unlock()
	logging.Debug("RETURN: kameleoonClientReadiness.MarkReady(siteCode: %s, environment: %s)",
		r.siteCode, r.environment)
}

// MarkNotReady reports a failed fetch with an `errs.Initialization` error, but only while the
// SDK is not yet ready. A failure never reverts an already-ready client and is reported at
// most once, so the first reported cause is the one waiters observe.
func (r *kameleoonClientReadiness) MarkNotReady(cause error) {
	logging.Debug("CALL: kameleoonClientReadiness.MarkNotReady(siteCode: %s, environment: %s, cause: %s)",
		r.siteCode, r.environment, cause)
	r.mx.Lock()
	if !r.state.isSettled() {
		r.state.err = errs.NewInitialization(r.siteCode, r.environment, cause)
		close(r.state.done)
	} // else: already ready, or the failure was already reported
	r.mx.Unlock()
	logging.Debug("RETURN: kameleoonClientReadiness.MarkNotReady(siteCode: %s, environment: %s, cause: %s)",
		r.siteCode, r.environment, cause)
}

// IsReady returns `true` if the SDK has been successfully initialized, `false` otherwise
// (including while the initialization is still pending or has failed). It never blocks.
func (r *kameleoonClientReadiness) IsReady() bool {
	state := r.currentState()
	return state.isSettled() && (state.err == nil)
}

// Wait blocks until the readiness state is settled. It returns nil once the SDK is ready,
// or an `errs.Initialization` error if the configuration could not be loaded. Once the SDK
// becomes ready, a subsequent call returns nil.
func (r *kameleoonClientReadiness) Wait() error {
	state := r.currentState()
	<-state.done
	return state.err
}

// WaitWithTimeout blocks until the readiness state is settled, but no longer than
// `timeout`. It returns the result of the configuration fetch: nil once the SDK is ready,
// or the `errs.Initialization` error the fetch failure was reported with (fail-fast — it does
// not wait for background retries). If no fetch result is available within the timeout, it
// returns an `errs.Initialization` error caused by (wrapping) `context.DeadlineExceeded`.
// A non-positive timeout expires immediately unless a result is already available.
func (r *kameleoonClientReadiness) WaitWithTimeout(timeout time.Duration) error {
	state := r.currentState()

	select {
	case <-state.done:
		return state.err
	default:
	}

	if timeout <= 0 {
		return r.newTimeoutFailure(timeout)
	}

	timer := time.NewTimer(timeout)
	defer stopTimer(timer)
	select {
	case <-state.done:
		return state.err
	case <-timer.C:
		return r.newTimeoutFailure(timeout)
	}
}

func (r *kameleoonClientReadiness) newTimeoutFailure(timeout time.Duration) error {
	cause := fmt.Errorf("initialization did not complete within %s: %w",
		timeout, context.DeadlineExceeded)
	return errs.NewInitialization(r.siteCode, r.environment, cause)
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}

func (r *kameleoonClientReadiness) currentState() *readinessState {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.state
}
