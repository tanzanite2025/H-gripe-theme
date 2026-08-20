package resilience

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type timeoutRoundTripper struct {
	requests atomic.Int32
}

func (t *timeoutRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	t.requests.Add(1)
	return nil, timeoutTransportError{}
}

type timeoutTransportError struct{}

func (timeoutTransportError) Error() string   { return "upstream timeout" }
func (timeoutTransportError) Timeout() bool   { return true }
func (timeoutTransportError) Temporary() bool { return true }

type probeCircuitController struct{}

func (probeCircuitController) Acquire(context.Context, string) (CircuitPermit, error) {
	return probeCircuitPermit{}, nil
}

type probeCircuitPermit struct{}

func (probeCircuitPermit) IsProbe() bool                 { return true }
func (probeCircuitPermit) RecordSuccess(context.Context) {}
func (probeCircuitPermit) RecordFailure(context.Context) {}
func (probeCircuitPermit) Release(context.Context)       {}

type timedProbeCircuitController struct{}

func (timedProbeCircuitController) Acquire(context.Context, string) (CircuitPermit, error) {
	return timedProbeCircuitPermit{}, nil
}

func (timedProbeCircuitController) ProbeTimeout() time.Duration {
	return 5 * time.Millisecond
}

type timedProbeCircuitPermit struct{}

func (timedProbeCircuitPermit) IsProbe() bool                 { return true }
func (timedProbeCircuitPermit) RecordSuccess(context.Context) {}
func (timedProbeCircuitPermit) RecordFailure(context.Context) {}
func (timedProbeCircuitPermit) Release(context.Context)       {}

type recordingCircuitController struct {
	permit *recordingCircuitPermit
}

func (c recordingCircuitController) Acquire(context.Context, string) (CircuitPermit, error) {
	return c.permit, nil
}

type unavailableCircuitController struct{}

func (unavailableCircuitController) Acquire(context.Context, string) (CircuitPermit, error) {
	return nil, ErrCircuitStoreUnavailable
}

type recordingCircuitPermit struct {
	successes atomic.Int32
	failures  atomic.Int32
	releases  atomic.Int32
}

type firstResponseRoundTripper struct {
	firstResponse chan<- struct{}
}

func (t firstResponseRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	select {
	case t.firstResponse <- struct{}{}:
	default:
	}
	return &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}, nil
}

func (p *recordingCircuitPermit) IsProbe() bool { return false }

func (p *recordingCircuitPermit) RecordSuccess(context.Context) {
	p.successes.Add(1)
}

func (p *recordingCircuitPermit) RecordFailure(context.Context) {
	p.failures.Add(1)
}

func (p *recordingCircuitPermit) Release(context.Context) {
	p.releases.Add(1)
}

func TestExponentialBackoffAddsJitterWithinConfiguredBounds(t *testing.T) {
	originalRandInt63n := randInt63n
	t.Cleanup(func() { randInt63n = originalRandInt63n })
	randInt63n = func(n int64) int64 {
		return n - 1
	}

	policy := BackoffPolicy{
		BaseDelay: 10 * time.Millisecond,
		MaxDelay:  100 * time.Millisecond,
		Jitter:    20 * time.Millisecond,
	}

	require.Equal(t, 30*time.Millisecond-time.Nanosecond, ExponentialBackoff(1, policy))
	require.Equal(t, 40*time.Millisecond-time.Nanosecond, ExponentialBackoff(2, policy))
	require.Equal(t, 60*time.Millisecond-time.Nanosecond, ExponentialBackoff(3, policy))
}

func TestCircuitBreakerOpensAndAllowsOneHalfOpenProbe(t *testing.T) {
	now := time.Date(2026, time.August, 19, 12, 0, 0, 0, time.UTC)
	breaker := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		FailureWindow:    time.Minute,
		OpenDuration:     10 * time.Second,
	})
	breaker.now = func() time.Time { return now }

	require.NoError(t, breaker.Allow("stripe"))
	breaker.RecordFailure("stripe")
	require.NoError(t, breaker.Allow("stripe"))
	breaker.RecordFailure("stripe")
	require.ErrorIs(t, breaker.Allow("stripe"), ErrCircuitOpen)

	now = now.Add(11 * time.Second)
	require.NoError(t, breaker.Allow("stripe"))
	require.ErrorIs(t, breaker.Allow("stripe"), ErrCircuitOpen)
	breaker.RecordSuccess("stripe")
	require.NoError(t, breaker.Allow("stripe"))
}

func TestHTTPClientRetriesTransientResponsesAndReturnsSuccess(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) < 3 {
			writer.WriteHeader(http.StatusBadGateway)
			_, _ = writer.Write([]byte("temporary"))
			return
		}
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	policy := DefaultHTTPRetryPolicy()
	policy.Backoff = BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond}
	policy.RetryUnsafeMethods = true
	client := NewHTTPClient(server.Client(), policy, NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 2,
		FailureWindow:    time.Minute,
		OpenDuration:     time.Second,
	}), "test")

	response, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.Equal(t, int32(3), requests.Load())
	require.NoError(t, response.Body.Close())
}

func TestHTTPClientDoesNotRetryNonTransientClientErrors(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = writer.Write([]byte("invalid"))
	}))
	defer server.Close()

	policy := DefaultHTTPRetryPolicy()
	policy.Backoff = BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond}
	policy.RetryUnsafeMethods = true
	client := NewHTTPClient(server.Client(), policy, nil, "test")

	response, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
	require.Equal(t, int32(1), requests.Load())
	require.NoError(t, response.Body.Close())
}

func TestHTTPClientRecordsUnsafeTransientResponseAsCircuitFailureWithoutRetry(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	permit := &recordingCircuitPermit{}
	policy := DefaultHTTPRetryPolicy()
	policy.MaxAttempts = 3
	policy.RetryUnsafeMethods = false
	policy.Backoff = BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond}
	client := NewHTTPClient(server.Client(), policy, recordingCircuitController{permit: permit}, "oauth")

	response, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, server.URL, strings.NewReader("code=authorization-code"))
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusTooManyRequests, response.StatusCode)
	require.Equal(t, int32(1), requests.Load())
	require.Equal(t, int32(0), permit.successes.Load())
	require.Equal(t, int32(1), permit.failures.Load())
	require.NoError(t, response.Body.Close())
}

func TestHTTPClientReturnsCircuitOpenWithoutCallingUpstream(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	breaker := NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: 1,
		FailureWindow:    time.Minute,
		OpenDuration:     time.Minute,
	})
	policy := DefaultHTTPRetryPolicy()
	policy.MaxAttempts = 1
	policy.RetryUnsafeMethods = true
	client := NewHTTPClient(server.Client(), policy, breaker, "test")

	response, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, server.URL, nil)
	})
	require.Error(t, err)
	require.NotNil(t, response)
	require.NoError(t, response.Body.Close())

	_, err = client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, server.URL, nil)
	})
	require.ErrorIs(t, err, ErrCircuitOpen)
	require.Equal(t, int32(1), requests.Load())
}

func TestHTTPClientFailsOpenWhenCircuitStoreIsUnavailable(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(server.Client(), HTTPRetryPolicy{
		MaxAttempts:        1,
		RetryUnsafeMethods: true,
	}, unavailableCircuitController{}, "test")

	response, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, http.StatusNoContent, response.StatusCode)
	require.Equal(t, int32(1), requests.Load())
	require.NoError(t, response.Body.Close())
}

func TestHTTPClientDoesNotTreatCanceledContextAsRetryable(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := NewHTTPClient(&http.Client{}, HTTPRetryPolicy{
		MaxAttempts: 3,
		Backoff:     BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
	}, nil, "")
	_, err := client.Do(ctx, func() (*http.Request, error) {
		return http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:1", nil)
	})
	require.Error(t, err)
	require.True(t, errors.Is(err, context.Canceled))
}

func TestHTTPClientReleasesCircuitPermitWhenRequestDeadlineExpires(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
		writer.WriteHeader(http.StatusGatewayTimeout)
	}))
	defer server.Close()

	permit := &recordingCircuitPermit{}
	client := NewHTTPClient(server.Client(), HTTPRetryPolicy{
		MaxAttempts: 3,
		Backoff:     BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
	}, recordingCircuitController{permit: permit}, "deadline-test")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := client.Do(ctx, func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, server.URL, nil)
	})

	require.Error(t, err)
	require.True(t, errors.Is(err, context.DeadlineExceeded))
	require.Equal(t, int32(0), permit.successes.Load())
	require.Equal(t, int32(0), permit.failures.Load())
	require.Equal(t, int32(1), permit.releases.Load())
}

func TestHTTPClientMarksInFlightUnsafeRequestDeadlineAsUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = writer.Write([]byte(request.Method))
	}))
	defer server.Close()

	client := NewHTTPClient(server.Client(), HTTPRetryPolicy{
		MaxAttempts:        3,
		Backoff:            BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
		RetryUnsafeMethods: true,
	}, nil, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err := client.Do(ctx, func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	})

	require.ErrorIs(t, err, ErrExternalOutcomeUnknown)
}

func TestHTTPClientMarksCanceledBackoffAfterUnsafeAttemptAsUnknown(t *testing.T) {
	firstResponse := make(chan struct{}, 1)
	permit := &recordingCircuitPermit{}
	client := NewHTTPClient(
		&http.Client{Transport: firstResponseRoundTripper{firstResponse: firstResponse}},
		HTTPRetryPolicy{
			MaxAttempts:        3,
			Backoff:            BackoffPolicy{BaseDelay: 200 * time.Millisecond, MaxDelay: 200 * time.Millisecond},
			RetryUnsafeMethods: true,
		},
		recordingCircuitController{permit: permit},
		"canceled-backoff",
	)

	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	go func() {
		_, err := client.Do(ctx, func() (*http.Request, error) {
			return http.NewRequest(http.MethodPost, "http://example.test", strings.NewReader("payload"))
		})
		result <- err
	}()

	<-firstResponse
	cancel()

	err := <-result
	require.ErrorIs(t, err, ErrExternalOutcomeUnknown)
	require.Equal(t, int32(1), permit.failures.Load())
	require.Equal(t, int32(0), permit.releases.Load())
}

func TestHTTPClientRetriesTransportTimeouts(t *testing.T) {
	transport := &timeoutRoundTripper{}
	policy := DefaultHTTPRetryPolicy()
	policy.Backoff = BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}
	policy.RetryUnsafeMethods = true
	client := NewHTTPClient(&http.Client{Transport: transport}, policy, nil, "")

	_, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, "http://example.test", strings.NewReader("payload"))
	})
	require.ErrorIs(t, err, ErrExternalOutcomeUnknown)
	require.Equal(t, int32(3), transport.requests.Load())
}

func TestHTTPClientDoesNotMarkGetTransportTimeoutAsUnknown(t *testing.T) {
	transport := &timeoutRoundTripper{}
	policy := DefaultHTTPRetryPolicy()
	policy.Backoff = BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}
	policy.RetryUnsafeMethods = true
	client := NewHTTPClient(&http.Client{Transport: transport}, policy, nil, "")

	_, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, "http://example.test", nil)
	})
	require.Error(t, err)
	require.NotErrorIs(t, err, ErrExternalOutcomeUnknown)
	require.Equal(t, int32(3), transport.requests.Load())
}

func TestHTTPClientMarksUnsafeTransportTimeoutAsUnknownWithoutUnsafeRetries(t *testing.T) {
	transport := &timeoutRoundTripper{}
	policy := DefaultHTTPRetryPolicy()
	policy.MaxAttempts = 3
	policy.Backoff = BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}
	client := NewHTTPClient(&http.Client{Transport: transport}, policy, nil, "")

	_, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, "http://example.test", strings.NewReader("payload"))
	})
	require.ErrorIs(t, err, ErrExternalOutcomeUnknown)
	require.Equal(t, int32(1), transport.requests.Load())
}

func TestHTTPClientDoesNotHideAmbiguousEarlierAttemptBehindLaterClientError(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if requests.Add(1) == 1 {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	policy := DefaultHTTPRetryPolicy()
	policy.MaxAttempts = 2
	policy.Backoff = BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}
	policy.RetryUnsafeMethods = true
	client := NewHTTPClient(server.Client(), policy, nil, "")

	response, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	})
	require.ErrorIs(t, err, ErrExternalOutcomeUnknown)
	require.NotNil(t, response)
	require.Equal(t, int32(2), requests.Load())
	require.NoError(t, response.Body.Close())
}

func TestHTTPClientHalfOpenProbeUsesExactlyOneAttempt(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		writer.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewHTTPClient(server.Client(), HTTPRetryPolicy{
		MaxAttempts:        3,
		Backoff:            BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
		RetryUnsafeMethods: true,
	}, probeCircuitController{}, "test")

	_, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodPost, server.URL, strings.NewReader("payload"))
	})
	require.Error(t, err)
	require.Equal(t, int32(1), requests.Load())
}

func TestHTTPClientBoundsSlowHalfOpenProbeToProbeLease(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests.Add(1)
		<-request.Context().Done()
	}))
	defer server.Close()

	client := NewHTTPClient(server.Client(), HTTPRetryPolicy{
		MaxAttempts:        3,
		Backoff:            BackoffPolicy{BaseDelay: time.Millisecond, MaxDelay: time.Millisecond},
		RetryUnsafeMethods: true,
	}, timedProbeCircuitController{}, "test")

	_, err := client.Do(context.Background(), func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, server.URL, nil)
	})

	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Equal(t, int32(1), requests.Load())
}
