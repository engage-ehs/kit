package backoff

import (
	"context"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"effective.ie/cuan/kit/errors"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	t.Run("max retries", func(t *testing.T) {
		t.Parallel()
		retry := New(context.Background(), 10)
		retries := 0
		for retry.Ongoing() {
			t.Logf("next delay: %s", retry.NextDelay())
			retries++
		}

		if retries != 10 {
			t.Errorf("got invalid number of retries %d", retries)
		}

		if retry.Err() == nil || !strings.Contains(retry.Err().Error(), "terminated after 10 retries") {
			t.Error("unterminated retry did not fire a error describing the issue: ", retry.Err())
		}
	})

	t.Run("max wait", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		retry := New(ctx, 0)
		for retry.Ongoing() {
			retry.Wait()
		}

		if retry.Err() == nil || !strings.Contains(retry.Err().Error(), "context deadline exceeded") {
			t.Error("unterminated retry did not fire a error describing the issue: ", retry.Err())
		}
	})

	t.Run("reasonable default", func(t *testing.T) {
		t.Parallel()
		retry := New(context.Background(), 0)
		for retry.Ongoing() {
			retry.Wait()
		}

		if retry.Err() == nil || !strings.Contains(retry.Err().Error(), "context deadline exceeded") {
			t.Error("unterminated retry did not fire a error describing the issue: ", retry.Err())
		}
	})

}

func ExampleBackoff_WaitFor() {
	var apierr error

	retry := New(context.Background(), 10)
	for retry.Ongoing() {
		rsp, err := http.Get("http://example.net/myapi")
		if err != nil {
			apierr = err
			break
		}
		if rsp.StatusCode != http.StatusOK && !ShouldRetryHTTP(rsp) {
			apierr = errors.New("Unexpected status code " + rsp.Status)
			break
		}
		retry.WaitFor(RetryAfterHTTP(rsp))
	}

	log.Println(apierr)
}
