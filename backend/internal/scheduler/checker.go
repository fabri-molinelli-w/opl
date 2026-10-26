// Package scheduler runs the periodic health checks of CU-03 with a bounded
// worker pool, and the daily retention cleanup.
package scheduler

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/fabri-molinelli-w/opl/backend/internal/domain"
)

// CheckTimeout is the timeout of each probe (CU-03 RN-3).
const CheckTimeout = 5 * time.Second

// Checker probes a service over HTTP(S) or TCP.
type Checker struct {
	http *http.Client
	dial func(ctx context.Context, network, address string) (net.Conn, error)
}

func NewChecker() *Checker {
	dialer := &net.Dialer{Timeout: CheckTimeout}
	return &Checker{
		http: &http.Client{
			Timeout: CheckTimeout,
			// Do not follow redirects: the expected status applies to the target itself.
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
		dial: dialer.DialContext,
	}
}

// Check probes the target and classifies the result (CU-03 RN-1, RN-2).
func (c *Checker) Check(ctx context.Context, cfg domain.CheckConfig) domain.CheckResult {
	ctx, cancel := context.WithTimeout(ctx, CheckTimeout)
	defer cancel()
	switch cfg.CheckType {
	case "http":
		return c.checkHTTP(ctx, cfg)
	case "tcp":
		return c.checkTCP(ctx, cfg)
	default:
		return domain.CheckResult{Status: domain.StatusDown, Error: fmt.Sprintf("tipo de chequeo desconocido: %q", cfg.CheckType)}
	}
}

func (c *Checker) checkHTTP(ctx context.Context, cfg domain.CheckConfig) domain.CheckResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.CheckTarget, nil)
	if err != nil {
		return domain.CheckResult{Status: domain.StatusDown, Error: "URL de chequeo inválida"}
	}
	req.Header.Set("User-Agent", "OPL-healthcheck/1.0")
	start := time.Now()
	resp, err := c.http.Do(req)
	elapsed := msSince(start)
	if err != nil {
		return domain.CheckResult{Status: domain.StatusDown, ResponseMs: &elapsed, Error: shortError(err)}
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64*1024))
	_ = resp.Body.Close()
	code := resp.StatusCode
	return Classify(code == cfg.ExpectedStatus, elapsed, cfg.LatencyThresholdMs, &code,
		fmt.Sprintf("respondió %d, se esperaba %d", code, cfg.ExpectedStatus))
}

func (c *Checker) checkTCP(ctx context.Context, cfg domain.CheckConfig) domain.CheckResult {
	start := time.Now()
	conn, err := c.dial(ctx, "tcp", cfg.CheckTarget)
	elapsed := msSince(start)
	if err != nil {
		return domain.CheckResult{Status: domain.StatusDown, ResponseMs: &elapsed, Error: shortError(err)}
	}
	_ = conn.Close()
	return Classify(true, elapsed, cfg.LatencyThresholdMs, nil, "")
}

// Classify applies RN-1/RN-2: DOWN if the probe failed, DEGRADED if it worked
// but was slower than the threshold, UP otherwise.
func Classify(ok bool, elapsedMs, thresholdMs int, httpStatus *int, failure string) domain.CheckResult {
	result := domain.CheckResult{ResponseMs: &elapsedMs, HTTPStatus: httpStatus}
	switch {
	case !ok:
		result.Status = domain.StatusDown
		result.Error = failure
	case elapsedMs > thresholdMs:
		result.Status = domain.StatusDegraded
	default:
		result.Status = domain.StatusUp
	}
	return result
}

func msSince(start time.Time) int {
	return int(time.Since(start) / time.Millisecond)
}

func shortError(err error) string {
	msg := err.Error()
	if len(msg) > 300 {
		msg = msg[:300]
	}
	return msg
}
