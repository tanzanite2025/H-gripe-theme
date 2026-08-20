package resilience

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrCircuitOpen            = errors.New("outbound circuit breaker is open")
	ErrExternalOutcomeUnknown = errors.New("external outcome is unknown")
)

var randInt63n = rand.Int63n

type BackoffPolicy struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
	Jitter    time.Duration
}

func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		BaseDelay: 250 * time.Millisecond,
		MaxDelay:  3 * time.Second,
		Jitter:    250 * time.Millisecond,
	}
}

func ExponentialBackoff(attempt int, policy BackoffPolicy) time.Duration {
	if policy.BaseDelay <= 0 {
		policy.BaseDelay = time.Second
	}
	if attempt <= 0 {
		attempt = 1
	}

	delay := policy.BaseDelay
	for index := 1; index < attempt; index++ {
		if policy.MaxDelay > 0 && delay >= policy.MaxDelay {
			delay = policy.MaxDelay
			break
		}
		if delay > time.Duration(1<<63-1)/2 {
			delay = time.Duration(1<<63 - 1)
			break
		}
		delay *= 2
	}
	if policy.MaxDelay > 0 && delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}
	if policy.Jitter > 0 {
		jitterLimit := policy.Jitter
		if policy.MaxDelay > 0 && delay < policy.MaxDelay && policy.MaxDelay-delay < jitterLimit {
			jitterLimit = policy.MaxDelay - delay
		}
		if jitterLimit > 0 {
			delay += time.Duration(randInt63n(int64(jitterLimit)))
		}
	}
	if policy.MaxDelay > 0 && delay > policy.MaxDelay {
		return policy.MaxDelay
	}
	return delay
}

type CircuitBreakerConfig struct {
	FailureThreshold int
	FailureWindow    time.Duration
	OpenDuration     time.Duration
}

func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		FailureWindow:    60 * time.Second,
		OpenDuration:     30 * time.Second,
	}
}

type circuitState struct {
	failures         int
	windowStarted    time.Time
	openedUntil      time.Time
	halfOpenInFlight bool
}

type CircuitBreaker struct {
	mu     sync.Mutex
	config CircuitBreakerConfig
	now    func() time.Time
	states map[string]*circuitState
}

// CircuitPermit represents one outbound request admitted by a circuit
// controller. A distributed controller can use it to associate the request
// with the single half-open probe lease shared by every application replica.
type CircuitPermit interface {
	IsProbe() bool
	RecordSuccess(context.Context)
	RecordFailure(context.Context)
	Release(context.Context)
}

// CircuitController admits outbound requests and returns a permit used to
// record the terminal request result.
type CircuitController interface {
	Acquire(context.Context, string) (CircuitPermit, error)
}

func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	config = normalizeCircuitBreakerConfig(config)
	return &CircuitBreaker{
		config: config,
		now:    time.Now,
		states: make(map[string]*circuitState),
	}
}

func (b *CircuitBreaker) Acquire(_ context.Context, key string) (CircuitPermit, error) {
	if err := b.Allow(key); err != nil {
		return nil, err
	}
	return localCircuitPermit{breaker: b, key: key}, nil
}

func (b *CircuitBreaker) Allow(key string) error {
	if b == nil {
		return nil
	}
	key = normalizeKey(key)
	now := b.currentTime()

	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.stateForLocked(key)
	if state.openedUntil.After(now) {
		return fmt.Errorf("%w: %s", ErrCircuitOpen, key)
	}
	if !state.openedUntil.IsZero() {
		if state.halfOpenInFlight {
			return fmt.Errorf("%w: %s", ErrCircuitOpen, key)
		}
		state.halfOpenInFlight = true
	}
	if state.windowStarted.IsZero() || now.Sub(state.windowStarted) >= b.config.FailureWindow {
		state.failures = 0
		state.windowStarted = now
	}
	return nil
}

func (b *CircuitBreaker) RecordSuccess(key string) {
	if b == nil {
		return
	}
	key = normalizeKey(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.stateForLocked(key)
	state.failures = 0
	state.windowStarted = b.currentTime()
	state.openedUntil = time.Time{}
	state.halfOpenInFlight = false
}

func (b *CircuitBreaker) RecordFailure(key string) {
	if b == nil {
		return
	}
	key = normalizeKey(key)
	now := b.currentTime()
	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.stateForLocked(key)
	if state.halfOpenInFlight {
		state.openedUntil = now.Add(b.config.OpenDuration)
		state.failures = 0
		state.windowStarted = now
		state.halfOpenInFlight = false
		return
	}
	if state.windowStarted.IsZero() || now.Sub(state.windowStarted) >= b.config.FailureWindow {
		state.failures = 0
		state.windowStarted = now
	}
	state.halfOpenInFlight = false
	state.failures++
	if state.failures >= b.config.FailureThreshold {
		state.openedUntil = now.Add(b.config.OpenDuration)
		state.failures = 0
		state.windowStarted = now
	}
}

func (b *CircuitBreaker) Release(key string) {
	if b == nil {
		return
	}
	key = normalizeKey(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	state := b.stateForLocked(key)
	state.halfOpenInFlight = false
}

func (b *CircuitBreaker) stateForLocked(key string) *circuitState {
	state := b.states[key]
	if state == nil {
		state = &circuitState{windowStarted: b.currentTime()}
		b.states[key] = state
	}
	return state
}

func (b *CircuitBreaker) currentTime() time.Time {
	if b != nil && b.now != nil {
		return b.now()
	}
	return time.Now()
}

type localCircuitPermit struct {
	breaker *CircuitBreaker
	key     string
}

func (p localCircuitPermit) RecordSuccess(_ context.Context) {
	p.breaker.RecordSuccess(p.key)
}

func (p localCircuitPermit) IsProbe() bool {
	return false
}

func (p localCircuitPermit) RecordFailure(_ context.Context) {
	p.breaker.RecordFailure(p.key)
}

func (p localCircuitPermit) Release(_ context.Context) {
	p.breaker.Release(p.key)
}

type noopCircuitPermit struct{}

func (noopCircuitPermit) IsProbe() bool                 { return false }
func (noopCircuitPermit) RecordSuccess(context.Context) {}
func (noopCircuitPermit) RecordFailure(context.Context) {}
func (noopCircuitPermit) Release(context.Context)       {}

func normalizeCircuitBreakerConfig(config CircuitBreakerConfig) CircuitBreakerConfig {
	defaults := DefaultCircuitBreakerConfig()
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = defaults.FailureThreshold
	}
	if config.FailureWindow <= 0 {
		config.FailureWindow = defaults.FailureWindow
	}
	if config.OpenDuration <= 0 {
		config.OpenDuration = defaults.OpenDuration
	}
	return config
}

var sharedCircuitBreakers sync.Map

func SharedCircuitBreaker(key string, config CircuitBreakerConfig) *CircuitBreaker {
	key = normalizeKey(key)
	if key == "" {
		return NewCircuitBreaker(config)
	}
	if value, ok := sharedCircuitBreakers.Load(key); ok {
		return value.(*CircuitBreaker)
	}
	breaker := NewCircuitBreaker(config)
	actual, _ := sharedCircuitBreakers.LoadOrStore(key, breaker)
	return actual.(*CircuitBreaker)
}

type HTTPRetryPolicy struct {
	MaxAttempts         int
	Backoff             BackoffPolicy
	RetryUnsafeMethods  bool
	RetryableStatusCode func(int) bool
}

func DefaultHTTPRetryPolicy() HTTPRetryPolicy {
	return HTTPRetryPolicy{
		MaxAttempts:        3,
		Backoff:            DefaultBackoffPolicy(),
		RetryUnsafeMethods: false,
	}
}

type HTTPClient struct {
	Client     *http.Client
	Retry      HTTPRetryPolicy
	Breaker    CircuitController
	BreakerKey string
}

func NewHTTPClient(client *http.Client, retry HTTPRetryPolicy, breaker CircuitController, breakerKey string) *HTTPClient {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	retry = normalizeHTTPRetryPolicy(retry)
	return &HTTPClient{
		Client:     client,
		Retry:      retry,
		Breaker:    breaker,
		BreakerKey: normalizeKey(breakerKey),
	}
}

func (c *HTTPClient) Do(ctx context.Context, requestFactory func() (*http.Request, error)) (*http.Response, error) {
	if c == nil || c.Client == nil {
		return nil, errors.New("outbound HTTP client is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if requestFactory == nil {
		return nil, errors.New("outbound HTTP request factory is required")
	}
	permit, err := c.acquireCircuit(ctx)
	if err != nil {
		if errors.Is(err, ErrCircuitStoreUnavailable) {
			permit = noopCircuitPermit{}
		} else {
			return nil, err
		}
	}

	operationCtx := ctx
	var operationCancel context.CancelFunc
	if permit.IsProbe() {
		if provider, ok := c.Breaker.(interface{ ProbeTimeout() time.Duration }); ok {
			if probeTimeout := provider.ProbeTimeout(); probeTimeout > 0 {
				operationCtx, operationCancel = context.WithTimeout(ctx, probeTimeout)
				defer operationCancel()
			}
		}
	}

	maxAttempts := c.Retry.MaxAttempts
	if permit.IsProbe() {
		maxAttempts = 1
	}
	outcomeMayHaveApplied := false
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := requestFactory()
		if err != nil {
			permit.Release(operationCtx)
			return nil, err
		}
		if req == nil {
			permit.Release(operationCtx)
			return nil, errors.New("outbound HTTP request factory returned nil request")
		}
		req = req.WithContext(operationCtx)

		resp, requestErr := c.Client.Do(req)
		retryable := c.shouldRetry(req, resp, requestErr)
		if !retryable {
			if requestErr != nil {
				if operationCtx.Err() != nil {
					if methodMayHaveExternalSideEffect(req.Method) {
						permit.RecordFailure(operationCtx)
						return resp, fmt.Errorf("%w: %v", ErrExternalOutcomeUnknown, requestErr)
					}
					permit.Release(operationCtx)
					return nil, requestErr
				}
				permit.RecordFailure(operationCtx)
				if methodMayHaveExternalSideEffect(req.Method) {
					return nil, fmt.Errorf("%w: %v", ErrExternalOutcomeUnknown, requestErr)
				}
				return nil, requestErr
			}
			if outcomeMayHaveApplied &&
				methodMayHaveExternalSideEffect(req.Method) &&
				(resp == nil || resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices) {
				permit.RecordFailure(operationCtx)
				statusCode := 0
				if resp != nil {
					statusCode = resp.StatusCode
				}
				return resp, fmt.Errorf("%w: upstream returned %d after an ambiguous attempt", ErrExternalOutcomeUnknown, statusCode)
			}
			if c.isCircuitFailure(resp) {
				permit.RecordFailure(operationCtx)
			} else {
				permit.RecordSuccess(operationCtx)
			}
			return resp, nil
		}

		if attempt >= maxAttempts {
			if operationCtx.Err() != nil {
				if requestErr != nil && methodMayHaveExternalSideEffect(req.Method) {
					permit.RecordFailure(operationCtx)
					return resp, fmt.Errorf("%w: %v", ErrExternalOutcomeUnknown, requestErr)
				}
				permit.Release(operationCtx)
				if requestErr != nil {
					return resp, requestErr
				}
				return resp, operationCtx.Err()
			}
			permit.RecordFailure(operationCtx)
			if requestErr != nil {
				if methodMayHaveExternalSideEffect(req.Method) {
					return resp, fmt.Errorf("%w: %v", ErrExternalOutcomeUnknown, requestErr)
				}
				return resp, requestErr
			}
			if resp != nil && methodMayHaveExternalSideEffect(req.Method) && responseMayHaveApplied(resp.StatusCode) {
				return resp, fmt.Errorf("%w: upstream returned %d", ErrExternalOutcomeUnknown, resp.StatusCode)
			}
			return resp, responseError(resp)
		}

		if resp != nil && resp.Body != nil {
			if methodMayHaveExternalSideEffect(req.Method) && responseMayHaveApplied(resp.StatusCode) {
				outcomeMayHaveApplied = true
			}
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
			_ = resp.Body.Close()
		}
		if requestErr != nil && methodMayHaveExternalSideEffect(req.Method) {
			outcomeMayHaveApplied = true
		}
		delay := ExponentialBackoff(attempt, c.Retry.Backoff)
		if retryAfter := retryAfterDuration(resp); retryAfter > delay {
			delay = retryAfter
			if c.Retry.Backoff.MaxDelay > 0 && delay > c.Retry.Backoff.MaxDelay {
				delay = c.Retry.Backoff.MaxDelay
			}
		}
		if err := sleepContext(operationCtx, delay); err != nil {
			if outcomeMayHaveApplied && methodMayHaveExternalSideEffect(req.Method) {
				permit.RecordFailure(operationCtx)
				return nil, fmt.Errorf("%w: %v", ErrExternalOutcomeUnknown, err)
			}
			permit.Release(operationCtx)
			return nil, err
		}
	}

	permit.RecordFailure(operationCtx)
	return nil, errors.New("outbound HTTP retry loop exhausted")
}

func (c *HTTPClient) acquireCircuit(ctx context.Context) (CircuitPermit, error) {
	if c == nil || c.Breaker == nil {
		return noopCircuitPermit{}, nil
	}
	return c.Breaker.Acquire(ctx, c.BreakerKey)
}

func (c *HTTPClient) shouldRetry(req *http.Request, resp *http.Response, err error) bool {
	if req == nil || !methodMayRetry(req.Method, c.Retry.RetryUnsafeMethods) {
		return false
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return false
		}
		if errors.Is(err, context.DeadlineExceeded) && req.Context().Err() != nil {
			return false
		}
		var networkError interface{ Timeout() bool }
		if errors.As(err, &networkError) && networkError.Timeout() {
			return true
		}
		return true
	}
	if resp == nil {
		return true
	}
	if c.Retry.RetryableStatusCode != nil {
		return c.Retry.RetryableStatusCode(resp.StatusCode)
	}
	switch resp.StatusCode {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (c *HTTPClient) isCircuitFailure(resp *http.Response) bool {
	if resp == nil {
		return true
	}
	if c != nil && c.Retry.RetryableStatusCode != nil {
		return c.Retry.RetryableStatusCode(resp.StatusCode)
	}
	switch resp.StatusCode {
	case http.StatusRequestTimeout, http.StatusTooEarly, http.StatusTooManyRequests,
		http.StatusInternalServerError, http.StatusBadGateway,
		http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func retryAfterDuration(resp *http.Response) time.Duration {
	if resp == nil {
		return 0
	}
	value := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	if timestamp, err := http.ParseTime(value); err == nil {
		delay := time.Until(timestamp)
		if delay > 0 {
			return delay
		}
	}
	return 0
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func normalizeHTTPRetryPolicy(policy HTTPRetryPolicy) HTTPRetryPolicy {
	defaults := DefaultHTTPRetryPolicy()
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = defaults.MaxAttempts
	}
	if policy.Backoff.BaseDelay <= 0 {
		policy.Backoff.BaseDelay = defaults.Backoff.BaseDelay
	}
	if policy.Backoff.MaxDelay <= 0 {
		policy.Backoff.MaxDelay = defaults.Backoff.MaxDelay
	}
	if policy.Backoff.Jitter < 0 {
		policy.Backoff.Jitter = defaults.Backoff.Jitter
	}
	return policy
}

func responseError(resp *http.Response) error {
	if resp == nil {
		return errors.New("unexpected outbound HTTP response")
	}
	return fmt.Errorf("unexpected outbound HTTP status %d", resp.StatusCode)
}

func responseMayHaveApplied(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout,
		http.StatusTooEarly,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func methodMayRetry(method string, retryUnsafe bool) bool {
	if retryUnsafe {
		return true
	}
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

func methodMayHaveExternalSideEffect(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func normalizeKey(value string) string {
	return strings.TrimSpace(value)
}
