// Package backoff implements a convenient mechanism to retry an operation. This is useful when
// talking to a remote system (database, third-party integration) that can fail for any reason
// (e.g. network), and where a retry would usually solve the issue.
package backoff

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// Backoff implements exponential backoff with randomized wait times. It is not safe to share a
// Backoff structure between multiple goroutines.
type Backoff struct {
	MaxRetries int

	ctx    context.Context
	cancel func()

	numRetries int
	nextDelay  time.Duration
}

// New creates a Backoff object that terminates either when the context terminates (built-in
// timeout), or when the maximum number of retries is reached. Passing no maximum number of retries
// means infinite number, in which case the context deadline is used, or a deadline of 2 minutes is
// chosen by default.
func New(ctx context.Context, retries int) *Backoff {
	var cancel func()
	// if no termination is provided, better provide a reasonable default value
	if _, ok := ctx.Deadline(); !ok && retries == 0 {
		ctx, cancel = context.WithTimeout(ctx, 64*time.Second)
	}
	return &Backoff{MaxRetries: retries, ctx: ctx, cancel: cancel}
}

// Ongoing returns true if caller should keep going
func (b *Backoff) Ongoing() bool {
	return b.ctx.Err() == nil && (b.MaxRetries == 0 || b.numRetries < b.MaxRetries)
}

// Err returns the reason for terminating the backoff, or nil if it didn't terminate
func (b *Backoff) Err() error {
	if b.ctx.Err() != nil {
		return b.ctx.Err()
	}
	if b.MaxRetries != 0 && b.numRetries >= b.MaxRetries {
		return fmt.Errorf("terminated after %d retries", b.numRetries)
	}
	return nil
}

// NumRetries returns the number of retries so far
func (b *Backoff) NumRetries() int { return b.numRetries }

// Wait sleeps for the backoff time then increases the retry count and backoff time
// Returns immediately if Context is terminated
func (b *Backoff) Wait() {
	if b.Ongoing() {
		select {
		case <-b.ctx.Done():
			if b.cancel != nil {
				b.cancel()
			}
		case <-time.After(b.NextDelay()):
		}
	}
}

// WaitFor can be used to wait for a specific duration, for example if the duration is provided by
// the remote API. Calling this method does increase the backoff, just like a regular Wait call.
func (b *Backoff) WaitFor(d time.Duration) {
	// as a special case, we handle no duration so users don’t need to worry about it.
	if d == 0 {
		b.Wait()
		return
	}

	if b.Ongoing() {
		b.numRetries++
		select {
		case <-b.ctx.Done():
			if b.cancel != nil {
				b.cancel()
			}
		case <-time.After(d + time.Duration(rand.Intn(maxmilli))*time.Millisecond):
		}
	}
}

// 1000 millisecond seems a reasonable max jitter
// https://cloud.google.com/iot/docs/how-tos/exponential-backoff
const maxmilli = 1000

func (b *Backoff) NextDelay() time.Duration {
	b.numRetries++
	b.nextDelay = (1<<b.numRetries)*time.Second + time.Duration(rand.Intn(maxmilli))*time.Millisecond

	return b.nextDelay
}
