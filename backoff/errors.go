package backoff

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/lib/pq"
)

// ShouldRetryHTTP can be used to know if a retry is a reasonable strategy to deal with an HTTP
// error (429 or 5xx code).
func ShouldRetryHTTP(rsp *http.Response) bool {
	return rsp.StatusCode == 429 || rsp.StatusCode/100 == 5
}

// RetryAfter reads HTTP headers to see if a duration was provided as an HTTP “retry-after” header
// (RFC 7231, section 7.1.3: Retry-After). It handles both absolute and relative dates.
func RetryAfterHTTP(rsp *http.Response) time.Duration {
	retry := rsp.Header.Get("Retry-After")
	t, err := time.Parse(time.RFC1123, retry)
	if err == nil {
		return time.Until(t)
	}

	d, err := strconv.Atoi(retry)
	if err == nil {
		return time.Duration(d) * time.Second
	}

	return 0
}

// ShouldRetryPostgreSQL can be used to know if a retry is a reasonable strategy to deal with a
// PostgreSQL error.
func ShouldRetryPostgreSQL(err error) bool {
	var pe *pq.Error
	if !errors.As(err, &pe) {
		return false
	}

	// see codes from https://www.postgresql.org/docs/current/errcodes-appendix.html
	switch pe.Code {
	case "08000", "08001", "08004", "08006":
		// network errors
		return true
	case "25000", "25P01", "25P02", "25P03", "2D000":
		// transaction errors
		return true
	default:
		return false
	}
}
