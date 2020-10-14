package backoff_test

import (
	"context"
	"log"
	"time"

	"effective.ie/cuan/kit/backoff"
	"effective.ie/cuan/kit/errors"
)

func Example() {
	// The number of retries can be limited using a context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// The number of retries can also be limited by the number of operations
	retry := backoff.New(ctx, 10)
	for retry.Ongoing() {
		err := doremotecall()
		if !shouldretry(err) {
			return
		}
		retry.Wait()
	}

	log.Println(retry.Err())
}

func doremotecall() error {
	// here is the remote call operation, potentially failing
	return nil
}

type NetworkError struct {
	ShouldRetry bool
	Underlying  error
}

func shouldretry(err error) bool {
	if err == nil {
		return false
	}

	var ne NetworkError
	if errors.As(err, &ne) {
		return ne.ShouldRetry
	}
	return false
}
