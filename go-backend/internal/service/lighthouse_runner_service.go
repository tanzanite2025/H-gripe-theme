package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidSiteQualityRun  = errors.New("invalid Lighthouse runner run")
	ErrSiteQualityJobRequired = errors.New("SiteQuality run requires a leased job")
)

const siteQualityRunnerMaxResponseBytes = 16 << 20

type LighthouseRunnerConfig struct {
	RunnerURL           string
	RunnerToken         string
	StorefrontBaseURL   string
	StorefrontTargetURL string
}

type LighthouseRunnerService struct {
	runs             *repository.SiteQualityRunRepository
	findings         *repository.SiteQualityFindingRepository
	jobs             *repository.SiteQualityJobRepository
	httpClient       *http.Client
	runnerURL        string
	runnerToken      string
	defaultTargetURL string
	runnerTargetURL  string
}

type LighthouseRunnerRunInput struct {
	URL               string `json:"url"`
	Strategy          string `json:"strategy"`
	InitiatedByUserID uint   `json:"-"`
}

// LighthouseRunnerCaptureInput adds Site Quality provenance to one immutable
// provider sample. A capture never changes finding state by itself.
type LighthouseRunnerCaptureInput struct {
	LighthouseRunnerRunInput
	TargetID         *uint  `json:"-"`
	JobID            *uint  `json:"-"`
	LeaseWorkerID    string `json:"-"`
	LeaseGeneration  int64  `json:"-"`
	CanonicalURL     string `json:"-"`
	ReleaseID        string `json:"-"`
	TargetSource     string `json:"-"`
	TargetSourceType string `json:"-"`
	TargetTitle      string `json:"-"`
	TargetLocale     string `json:"-"`
}

type LighthouseRunnerIssue struct {
	ID              string                                                `json:"id"`
	RuleID          string                                                `json:"rule_id"`
	ProviderAuditID string                                                `json:"provider_audit_id,omitempty"`
	Kind            string                                                `json:"kind"`
	RuleVersion     string                                                `json:"rule_version"`
	Title           string                                                `json:"title"`
	Description     string                                                `json:"description,omitempty"`
	Score           *float64                                              `json:"score,omitempty"`
	DisplayValue    string                                                `json:"display_value,omitempty"`
	NumericValue    *float64                                              `json:"numeric_value,omitempty"`
	SavingsMS       *float64                                              `json:"savings_ms,omitempty"`
	SavingsBytes    *int64                                                `json:"savings_bytes,omitempty"`
	Severity        string                                                `json:"severity"`
	Resources       []LighthouseRunnerResource                            `json:"resources,omitempty"`
	Links           []sitequalitydomain.SiteQualityLinkEvidence           `json:"links,omitempty"`
	Headings        []sitequalitydomain.SiteQualityHeadingEvidence        `json:"headings,omitempty"`
	StructuredData  []sitequalitydomain.SiteQualityStructuredDataEvidence `json:"structured_data,omitempty"`
	Remediation     *LighthouseRunnerRemediation                          `json:"remediation,omitempty"`
}

type LighthouseRunnerResource struct {
	URL        string   `json:"url"`
	TotalBytes *int64   `json:"total_bytes,omitempty"`
	WastedMS   *float64 `json:"wasted_ms,omitempty"`
}

type LighthouseRunnerRemediation struct {
	Label string `json:"label"`
	Route string `json:"route"`
}

type LighthouseRunnerRunView struct {
	ID                       uint                    `json:"id"`
	TargetID                 *uint                   `json:"target_id,omitempty"`
	JobID                    *uint                   `json:"job_id,omitempty"`
	TargetURL                string                  `json:"target_url"`
	CanonicalURL             string                  `json:"canonical_url,omitempty"`
	FinalURL                 string                  `json:"final_url,omitempty"`
	Strategy                 string                  `json:"strategy"`
	Status                   string                  `json:"status"`
	InitiatedByUserID        uint                    `json:"initiated_by_user_id"`
	PerformanceScore         *int                    `json:"performance_score,omitempty"`
	AccessibilityScore       *int                    `json:"accessibility_score,omitempty"`
	BestPracticesScore       *int                    `json:"best_practices_score,omitempty"`
	SEOScore                 *int                    `json:"seo_score,omitempty"`
	FirstContentfulPaintMS   *float64                `json:"first_contentful_paint_ms,omitempty"`
	LargestContentfulPaintMS *float64                `json:"largest_contentful_paint_ms,omitempty"`
	InteractionToNextPaintMS *float64                `json:"interaction_to_next_paint_ms,omitempty"`
	CumulativeLayoutShift    *float64                `json:"cumulative_layout_shift,omitempty"`
	TotalBlockingTimeMS      *float64                `json:"total_blocking_time_ms,omitempty"`
	SpeedIndexMS             *float64                `json:"speed_index_ms,omitempty"`
	Issues                   []LighthouseRunnerIssue `json:"issues"`
	ErrorMessage             string                  `json:"error_message,omitempty"`
	CreatedAt                time.Time               `json:"created_at"`
}

type LighthouseRunnerListResult struct {
	RunnerConfigured bool                      `json:"runner_configured"`
	DefaultURL       string                    `json:"default_url,omitempty"`
	Items            []LighthouseRunnerRunView `json:"items"`
	Total            int64                     `json:"total"`
}

type LighthouseRunnerSummary struct {
	RunnerConfigured bool                     `json:"runner_configured"`
	RunCount         int64                    `json:"run_count"`
	LatestRun        *LighthouseRunnerRunView `json:"latest_run,omitempty"`
}

type siteQualityAPIResponse struct {
	LighthouseResult struct {
		FinalURL               string                                  `json:"finalUrl"`
		LighthouseVersion      string                                  `json:"lighthouseVersion"`
		ConfigSettings         json.RawMessage                         `json:"configSettings"`
		Categories             map[string]siteQualityAPICategory       `json:"categories"`
		Audits                 map[string]json.RawMessage              `json:"audits"`
		RenderedHeadings       *siteQualityRenderedHeadingAudit        `json:"renderedHeadings"`
		RenderedStructuredData *siteQualityRenderedStructuredDataAudit `json:"renderedStructuredData"`
	} `json:"lighthouseResult"`
}

type siteQualityRenderedHeadingAudit struct {
	Status   string                                         `json:"status"`
	Source   string                                         `json:"source"`
	FinalURL string                                         `json:"finalUrl"`
	Error    string                                         `json:"error"`
	Headings []sitequalitydomain.SiteQualityHeadingEvidence `json:"headings"`
}

type siteQualityAPICategory struct {
	Score *float64 `json:"score"`
}

type siteQualityAPIAudit struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Score            *float64 `json:"score"`
	DisplayValue     string   `json:"displayValue"`
	NumericValue     *float64 `json:"numericValue"`
	ScoreDisplayMode string   `json:"scoreDisplayMode"`
	Details          struct {
		OverallSavingsMS    *float64          `json:"overallSavingsMs"`
		OverallSavingsBytes *int64            `json:"overallSavingsBytes"`
		Items               []json.RawMessage `json:"items"`
	} `json:"details"`
}

func NewLighthouseRunnerService(
	runs *repository.SiteQualityRunRepository,
	findings *repository.SiteQualityFindingRepository,
	cfg LighthouseRunnerConfig,
) *LighthouseRunnerService {
	defaultTargetURL := strings.TrimSpace(cfg.StorefrontBaseURL)
	if normalized, err := canonicalizeAbsoluteSiteQualityURL(defaultTargetURL); err == nil {
		defaultTargetURL = normalized
	} else {
		defaultTargetURL = strings.TrimRight(defaultTargetURL, "/")
	}
	runnerTargetURL := strings.TrimSpace(cfg.StorefrontTargetURL)
	if runnerTargetURL == "" {
		runnerTargetURL = defaultTargetURL
	}
	if normalized, err := canonicalizeAbsoluteSiteQualityURL(runnerTargetURL); err == nil {
		runnerTargetURL = normalized
	} else {
		runnerTargetURL = strings.TrimRight(runnerTargetURL, "/")
	}
	return &LighthouseRunnerService{
		runs:             runs,
		findings:         findings,
		httpClient:       &http.Client{Timeout: 110 * time.Second},
		runnerURL:        strings.TrimRight(strings.TrimSpace(cfg.RunnerURL), "/"),
		runnerToken:      strings.TrimSpace(cfg.RunnerToken),
		defaultTargetURL: defaultTargetURL,
		runnerTargetURL:  runnerTargetURL,
	}
}

func (s *LighthouseRunnerService) ConfigureJobRepository(jobs *repository.SiteQualityJobRepository) {
	if s == nil {
		return
	}
	s.jobs = jobs
}

func (s *LighthouseRunnerService) ConfigureHTTPClient(client *http.Client, runnerURL string) {
	if s == nil {
		return
	}
	if client != nil {
		s.httpClient = client
	}
	if strings.TrimSpace(runnerURL) != "" {
		s.runnerURL = strings.TrimRight(strings.TrimSpace(runnerURL), "/")
	}
}

func (s *LighthouseRunnerService) Run(
	ctx context.Context,
	input LighthouseRunnerRunInput,
) (*LighthouseRunnerRunView, error) {
	return s.Capture(ctx, LighthouseRunnerCaptureInput{LighthouseRunnerRunInput: input})
}

// Capture persists one raw SiteQuality provider sample. It deliberately does not
// reconcile findings: transient provider variance must be evaluated by the
// Site Quality job engine before it becomes operational work.
func (s *LighthouseRunnerService) Capture(
	ctx context.Context,
	input LighthouseRunnerCaptureInput,
) (*LighthouseRunnerRunView, error) {
	if s == nil || s.runs == nil || s.findings == nil {
		return nil, errors.New("Lighthouse runner service is unavailable")
	}
	targetURL, strategy, err := s.normalizeRunInput(input.LighthouseRunnerRunInput)
	if err != nil {
		return nil, err
	}
	if input.JobID == nil || *input.JobID == 0 {
		return nil, ErrSiteQualityJobRequired
	}
	if strings.TrimSpace(input.LeaseWorkerID) == "" || input.LeaseGeneration <= 0 {
		return nil, ErrSiteQualityJobRequired
	}

	publicTargetURL := targetURL
	targetURL, err = s.runnerTargetURLFor(publicTargetURL)
	if err != nil {
		return nil, err
	}
	rawResponse, result, requestErr := s.request(ctx, targetURL, strategy, input.ReleaseID)
	canonicalURL := strings.TrimSpace(input.CanonicalURL)
	if canonicalURL == "" {
		canonicalURL = publicTargetURL
	}
	run := sitequalitydomain.SiteQualityRun{
		TargetID:          input.TargetID,
		JobID:             input.JobID,
		TargetURL:         targetURL,
		CanonicalURL:      canonicalURL,
		Strategy:          strategy,
		InitiatedByUserID: input.InitiatedByUserID,
		Provider:          "lighthouse_runner",
		ReleaseID:         strings.TrimSpace(input.ReleaseID),
		Status:            sitequalitydomain.SiteQualityRunStatusSuccess,
		IssuesJSON:        "[]",
		RawResponseJSON:   normalizeSiteQualityRawResponse(rawResponse),
	}
	if requestErr != nil {
		run.Status = sitequalitydomain.SiteQualityRunStatusFailed
		run.ErrorMessage = requestErr.Error()
		if err := s.persistRunWithLease(&run, input); err != nil {
			return nil, err
		}
		view := siteQualityRunView(run)
		return &view, requestErr
	}

	applySiteQualityResult(
		&run,
		result,
		siteQualityStructuredDataPageIntent{
			Source:     input.TargetSource,
			SourceType: input.TargetSourceType,
			Title:      input.TargetTitle,
			Locale:     input.TargetLocale,
		},
	)
	if err := s.persistRunWithLease(&run, input); err != nil {
		return nil, err
	}
	view := siteQualityRunView(run)
	if run.Status != sitequalitydomain.SiteQualityRunStatusSuccess {
		return &view, errors.New(run.ErrorMessage)
	}
	return &view, nil
}

func (s *LighthouseRunnerService) persistRunWithLease(
	run *sitequalitydomain.SiteQualityRun,
	input LighthouseRunnerCaptureInput,
) error {
	if s == nil || s.runs == nil {
		return errors.New("Lighthouse runner service is unavailable")
	}
	if run == nil {
		return errors.New("SiteQuality run is required")
	}
	if input.JobID == nil || *input.JobID == 0 {
		return ErrSiteQualityJobRequired
	}
	if s.jobs == nil {
		return errors.New("SiteQuality job repository is unavailable")
	}
	if strings.TrimSpace(input.LeaseWorkerID) == "" || input.LeaseGeneration <= 0 {
		return errors.New("SiteQuality job lease input is incomplete")
	}

	return s.jobs.Transaction(func(tx *gorm.DB) error {
		if err := s.jobs.WithTx(tx).AssertLease(
			tx,
			*input.JobID,
			input.LeaseWorkerID,
			input.LeaseGeneration,
		); err != nil {
			return err
		}
		return s.runs.WithTx(tx).Create(run)
	})
}

func (s *LighthouseRunnerService) List(
	filter repository.SiteQualityRunListFilter,
) (*LighthouseRunnerListResult, error) {
	if s == nil || s.runs == nil {
		return nil, errors.New("Lighthouse runner service is unavailable")
	}
	runs, total, err := s.runs.List(filter)
	if err != nil {
		return nil, err
	}
	items := make([]LighthouseRunnerRunView, 0, len(runs))
	for _, run := range runs {
		items = append(items, siteQualityRunView(run))
	}
	return &LighthouseRunnerListResult{
		RunnerConfigured: s.RunnerConfigured(),
		DefaultURL:       s.defaultTargetURL,
		Items:            items,
		Total:            total,
	}, nil
}

func (s *LighthouseRunnerService) ListFindings(
	filter repository.SiteQualityFindingListFilter,
) ([]sitequalitydomain.SiteQualityFinding, int64, error) {
	if s == nil || s.findings == nil {
		return nil, 0, errors.New("SiteQuality finding service is unavailable")
	}
	return s.findings.List(filter)
}

func (s *LighthouseRunnerService) FindingStats() (sitequalitydomain.SiteQualityFindingStats, error) {
	if s == nil || s.findings == nil {
		return sitequalitydomain.SiteQualityFindingStats{}, errors.New("SiteQuality finding service is unavailable")
	}
	return s.findings.Stats()
}

func (s *LighthouseRunnerService) GetFinding(
	id uint,
) (*sitequalitydomain.SiteQualityFinding, error) {
	if s == nil || s.findings == nil {
		return nil, errors.New("SiteQuality finding service is unavailable")
	}
	if id == 0 {
		return nil, errors.New("SiteQuality finding ID is required")
	}
	return s.findings.FindByID(id)
}

func (s *LighthouseRunnerService) ListFindingEvents(
	findingID uint,
	page int,
	pageSize int,
) ([]sitequalitydomain.SiteQualityFindingEvent, int64, error) {
	if s == nil || s.findings == nil {
		return nil, 0, errors.New("SiteQuality finding service is unavailable")
	}
	if _, err := s.GetFinding(findingID); err != nil {
		return nil, 0, err
	}
	return s.findings.ListEvents(findingID, page, pageSize)
}

func (s *LighthouseRunnerService) AcknowledgeFinding(
	id uint,
	actorUserID uint,
	input sitequalitydomain.SiteQualityFindingActionInput,
) (*sitequalitydomain.SiteQualityFinding, error) {
	finding, err := s.requireOpenSiteQualityFinding(id)
	if err != nil {
		return nil, err
	}
	if finding.State != sitequalitydomain.SiteQualityFindingStateOpen {
		return nil, errors.New("only open SiteQuality findings can be acknowledged")
	}
	return s.findings.UpdateWithEvent(
		id,
		map[string]interface{}{"state": sitequalitydomain.SiteQualityFindingStateAcknowledged},
		sitequalitydomain.SiteQualityFindingEventAcknowledged,
		actorUserID,
		strings.TrimSpace(input.Note),
		nil,
	)
}

func (s *LighthouseRunnerService) ResolveFinding(
	id uint,
	actorUserID uint,
	input sitequalitydomain.SiteQualityFindingResolutionInput,
) (*sitequalitydomain.SiteQualityFinding, error) {
	finding, err := s.requireOpenSiteQualityFinding(id)
	if err != nil {
		return nil, err
	}
	if finding.State != sitequalitydomain.SiteQualityFindingStateOpen &&
		finding.State != sitequalitydomain.SiteQualityFindingStateAcknowledged {
		return nil, errors.New("only open SiteQuality findings can be resolved")
	}
	note := strings.TrimSpace(input.ResolutionNote)
	if note == "" {
		return nil, errors.New("resolution note is required")
	}
	now := time.Now().UTC()
	return s.findings.UpdateWithEvent(
		id,
		map[string]interface{}{
			"state":           sitequalitydomain.SiteQualityFindingStateResolved,
			"resolution_note": note,
			"resolved_at":     now,
			"verified_at":     nil,
		},
		sitequalitydomain.SiteQualityFindingEventResolutionRecorded,
		actorUserID,
		note,
		nil,
	)
}

func (s *LighthouseRunnerService) Summary() (*LighthouseRunnerSummary, error) {
	result, err := s.List(repository.SiteQualityRunListFilter{Page: 1, PageSize: 1})
	if err != nil {
		return nil, err
	}
	summary := &LighthouseRunnerSummary{
		RunnerConfigured: result.RunnerConfigured,
		RunCount:         result.Total,
	}
	if len(result.Items) > 0 {
		latest := result.Items[0]
		summary.LatestRun = &latest
	}
	return summary, nil
}

func (s *LighthouseRunnerService) LatestSuccessfulAt() (*time.Time, error) {
	if s == nil || s.runs == nil {
		return nil, errors.New("Lighthouse runner service is unavailable")
	}
	return s.runs.LatestSuccessfulAt()
}

func (s *LighthouseRunnerService) RunnerConfigured() bool {
	return s != nil && s.runnerURL != "" && len(s.runnerToken) >= 32
}

func (s *LighthouseRunnerService) request(
	parent context.Context,
	targetURL string,
	strategy string,
	releaseID string,
) (json.RawMessage, *siteQualityAPIResponse, error) {
	if s == nil {
		return nil, nil, errors.New("Lighthouse runner service is unavailable")
	}
	if !s.RunnerConfigured() {
		return nil, nil, errors.New("internal Lighthouse runner is not configured")
	}
	endpoint, err := url.Parse(s.runnerURL + "/v1/lighthouse/run")
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return nil, nil, errors.New("internal Lighthouse runner URL is invalid")
	}
	payload, err := json.Marshal(struct {
		URL       string `json:"url"`
		Strategy  string `json:"strategy"`
		ReleaseID string `json:"release_id,omitempty"`
	}{
		URL:       targetURL,
		Strategy:  strategy,
		ReleaseID: strings.TrimSpace(releaseID),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("encode Lighthouse runner request: %w", err)
	}

	ctx, cancel := context.WithTimeout(parent, 105*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, nil, fmt.Errorf("build Lighthouse runner request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+s.runnerToken)
	request.Header.Set("User-Agent", "tanzanite-site-quality/1.0")
	client := s.httpClient
	if client == nil {
		client = &http.Client{Timeout: 110 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("internal Lighthouse runner request failed: %w", err)
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(response.Body, siteQualityRunnerMaxResponseBytes+1))
	if readErr != nil {
		return nil, nil, fmt.Errorf("read Lighthouse runner response: %w", readErr)
	}
	if len(body) > siteQualityRunnerMaxResponseBytes {
		return nil, nil, errors.New("internal Lighthouse runner response exceeded maximum size")
	}
	raw := json.RawMessage(body)
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return raw, nil, lighthouseRunnerError(response.StatusCode, body)
	}
	var result siteQualityAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return raw, nil, errors.New("internal Lighthouse runner response was not valid JSON")
	}
	if len(result.LighthouseResult.Audits) == 0 {
		return raw, nil, errors.New("internal Lighthouse runner returned no Lighthouse audits")
	}
	if err := s.validateFinalURL(result.LighthouseResult.FinalURL); err != nil {
		return raw, nil, err
	}
	return raw, &result, nil
}

func (s *LighthouseRunnerService) normalizeRunInput(input LighthouseRunnerRunInput) (string, string, error) {
	rawURL := strings.TrimSpace(input.URL)
	if rawURL == "" {
		return "", "", fmt.Errorf("%w: URL is required", ErrInvalidSiteQualityRun)
	}
	if len(rawURL) > 2048 {
		return "", "", fmt.Errorf("%w: URL is too long", ErrInvalidSiteQualityRun)
	}
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed == nil || parsed.Host == "" {
		return "", "", fmt.Errorf("%w: URL must be absolute", ErrInvalidSiteQualityRun)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return "", "", fmt.Errorf("%w: URL must use HTTP or HTTPS", ErrInvalidSiteQualityRun)
	}
	if parsed.User != nil || parsed.Hostname() == "" {
		return "", "", fmt.Errorf("%w: URL cannot include credentials", ErrInvalidSiteQualityRun)
	}
	if s == nil || !sameSiteQualityOrigin(parsed, s.defaultTargetURL) {
		return "", "", fmt.Errorf("%w: URL must belong to the configured storefront origin", ErrInvalidSiteQualityRun)
	}
	parsed.Fragment = ""
	targetURL := parsed.String()

	strategy := strings.ToLower(strings.TrimSpace(input.Strategy))
	if strategy == "" {
		strategy = sitequalitydomain.SiteQualityStrategyMobile
	}
	if strategy != sitequalitydomain.SiteQualityStrategyMobile && strategy != sitequalitydomain.SiteQualityStrategyDesktop {
		return "", "", fmt.Errorf("%w: strategy must be mobile or desktop", ErrInvalidSiteQualityRun)
	}
	return targetURL, strategy, nil
}

func (s *LighthouseRunnerService) runnerTargetURLFor(publicTargetURL string) (string, error) {
	if s == nil {
		return "", errors.New("Lighthouse runner service is unavailable")
	}
	target, err := url.Parse(strings.TrimSpace(publicTargetURL))
	if err != nil || target == nil || target.Scheme == "" || target.Host == "" {
		return "", fmt.Errorf("%w: public target URL is invalid", ErrInvalidSiteQualityRun)
	}
	runnerOrigin, err := url.Parse(strings.TrimSpace(s.runnerTargetURL))
	if err != nil || runnerOrigin == nil || runnerOrigin.Scheme == "" || runnerOrigin.Host == "" {
		return "", errors.New("internal Lighthouse runner storefront target URL is invalid")
	}
	target.Scheme = runnerOrigin.Scheme
	target.Host = runnerOrigin.Host
	target.User = nil
	target.Fragment = ""
	return target.String(), nil
}

func (s *LighthouseRunnerService) validateFinalURL(rawURL string) error {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || !sameSiteQualityOrigin(parsed, s.runnerTargetURL) {
		return errors.New("internal Lighthouse runner reached a URL outside the configured storefront origin")
	}
	return nil
}

func sameSiteQualityOrigin(target *url.URL, rawBaseURL string) bool {
	base, err := url.Parse(strings.TrimSpace(rawBaseURL))
	if err != nil || target == nil || base.Scheme == "" || base.Host == "" {
		return false
	}
	return strings.EqualFold(target.Scheme, base.Scheme) &&
		strings.EqualFold(target.Hostname(), base.Hostname()) &&
		siteQualityURLPort(target) == siteQualityURLPort(base)
}

func siteQualityURLPort(value *url.URL) string {
	if value == nil {
		return ""
	}
	if port := value.Port(); port != "" {
		return port
	}
	if strings.EqualFold(value.Scheme, "https") {
		return "443"
	}
	if strings.EqualFold(value.Scheme, "http") {
		return "80"
	}
	return ""
}

func lighthouseRunnerError(statusCode int, body []byte) error {
	var payload struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &payload) == nil && strings.TrimSpace(payload.Error.Message) != "" {
		return fmt.Errorf("internal Lighthouse runner returned HTTP %d: %s", statusCode, strings.TrimSpace(payload.Error.Message))
	}
	return fmt.Errorf("internal Lighthouse runner returned HTTP %d", statusCode)
}

func normalizeSiteQualityRawResponse(raw json.RawMessage) string {
	if len(raw) == 0 || !json.Valid(raw) {
		return "{}"
	}
	return string(raw)
}

func normalizeSiteQualityEnvironment(raw json.RawMessage) string {
	if len(raw) == 0 || !json.Valid(raw) {
		return "{}"
	}
	return string(raw)
}

func applySiteQualityResult(
	run *sitequalitydomain.SiteQualityRun,
	result *siteQualityAPIResponse,
	intents ...siteQualityStructuredDataPageIntent,
) {
	if run == nil || result == nil {
		return
	}
	audits, err := decodeSiteQualityAudits(result.LighthouseResult.Audits)
	if err != nil {
		run.Status = sitequalitydomain.SiteQualityRunStatusFailed
		run.ErrorMessage = fmt.Sprintf("decode Lighthouse runner audits: %v", err)
		return
	}
	run.FinalURL = strings.TrimSpace(result.LighthouseResult.FinalURL)
	run.LighthouseVersion = strings.TrimSpace(result.LighthouseResult.LighthouseVersion)
	run.EnvironmentJSON = normalizeSiteQualityEnvironment(result.LighthouseResult.ConfigSettings)
	run.PerformanceScore = siteQualityCategoryScore(result.LighthouseResult.Categories["performance"])
	run.AccessibilityScore = siteQualityCategoryScore(result.LighthouseResult.Categories["accessibility"])
	run.BestPracticesScore = siteQualityCategoryScore(result.LighthouseResult.Categories["best-practices"])
	run.SEOScore = siteQualityCategoryScore(result.LighthouseResult.Categories["seo"])
	run.FirstContentfulPaintMS = siteQualityAuditNumericValue(audits, "first-contentful-paint")
	run.LargestContentfulPaintMS = siteQualityAuditNumericValue(audits, "largest-contentful-paint")
	run.InteractionToNextPaintMS = siteQualityAuditNumericValue(audits, "interaction-to-next-paint")
	run.CumulativeLayoutShift = siteQualityAuditNumericValue(audits, "cumulative-layout-shift")
	run.TotalBlockingTimeMS = siteQualityAuditNumericValue(audits, "total-blocking-time")
	run.SpeedIndexMS = siteQualityAuditNumericValue(audits, "speed-index")

	issues, err := normalizeSiteQualityIssues(audits)
	if err != nil {
		run.Status = sitequalitydomain.SiteQualityRunStatusFailed
		run.ErrorMessage = fmt.Sprintf("normalize Lighthouse runner response: %v", err)
		return
	}
	issues = removeSiteQualityRenderedHeadingManagedIssues(issues)
	issues = removeSiteQualityRenderedStructuredDataManagedIssues(issues)
	headingIssues := siteQualityRenderedHeadingAuditIssues(
		run.TargetURL,
		result.LighthouseResult.FinalURL,
		result.LighthouseResult.RenderedHeadings,
	)
	issues = append(issues, headingIssues...)
	structuredDataIssues := siteQualityRenderedStructuredDataAuditIssues(
		run.TargetURL,
		result.LighthouseResult.FinalURL,
		result.LighthouseResult.RenderedStructuredData,
		intents...,
	)
	issues = append(issues, structuredDataIssues...)
	decorateSiteQualityIssueIDs(issues)
	sortSiteQualityIssues(issues)
	encoded, err := json.Marshal(issues)
	if err != nil {
		run.Status = sitequalitydomain.SiteQualityRunStatusFailed
		run.ErrorMessage = fmt.Sprintf("encode Lighthouse runner findings: %v", err)
		return
	}
	run.IssuesJSON = string(encoded)
}

func decodeSiteQualityAudits(rawAudits map[string]json.RawMessage) (map[string]siteQualityAPIAudit, error) {
	audits := make(map[string]siteQualityAPIAudit)
	for auditID, rawAudit := range rawAudits {
		if !siteQualityAuditRequired(auditID) {
			continue
		}
		var audit siteQualityAPIAudit
		if err := json.Unmarshal(rawAudit, &audit); err != nil {
			return nil, fmt.Errorf("decode audit %q: %w", auditID, err)
		}
		if strings.TrimSpace(audit.ID) == "" {
			audit.ID = auditID
		}
		audits[auditID] = audit
	}
	return audits, nil
}

func siteQualityAuditRequired(auditID string) bool {
	switch auditID {
	case "first-contentful-paint",
		"largest-contentful-paint",
		"interaction-to-next-paint",
		"cumulative-layout-shift",
		"total-blocking-time",
		"speed-index":
		return true
	}
	_, actionable := siteQualityLookupAuditRule(auditID)
	return actionable
}

func siteQualityCategoryScore(category siteQualityAPICategory) *int {
	if category.Score == nil {
		return nil
	}
	score := int(math.Round(*category.Score * 100))
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return &score
}

func siteQualityAuditNumericValue(audits map[string]siteQualityAPIAudit, id string) *float64 {
	audit, found := audits[id]
	if !found || audit.NumericValue == nil {
		return nil
	}
	value := *audit.NumericValue
	return &value
}

func normalizeSiteQualityIssues(audits map[string]siteQualityAPIAudit) ([]LighthouseRunnerIssue, error) {
	issues := make([]LighthouseRunnerIssue, 0)
	for fallbackID, audit := range audits {
		id := strings.TrimSpace(audit.ID)
		if id == "" {
			id = fallbackID
		}
		if id == "" {
			continue
		}
		rule, actionable := siteQualityLookupAuditRule(id)
		if !actionable || !siteQualityAuditMeetsRule(rule, audit) {
			continue
		}
		issue := LighthouseRunnerIssue{
			ID:              id,
			RuleID:          siteQualityRuleIDForAudit(id),
			ProviderAuditID: siteQualityProviderAuditIDForAudit(id),
			Kind:            rule.Kind,
			RuleVersion:     siteQualityAuditRuleVersion,
			Title:           strings.TrimSpace(audit.Title),
			Description:     strings.TrimSpace(audit.Description),
			Score:           copyFloat64(audit.Score),
			DisplayValue:    strings.TrimSpace(audit.DisplayValue),
			SavingsMS:       copyFloat64(audit.Details.OverallSavingsMS),
			SavingsBytes:    copyInt64(audit.Details.OverallSavingsBytes),
			Severity:        siteQualityAuditSeverity(audit),
			Resources:       siteQualityAuditResources(audit),
			Links:           siteQualityAuditLinkEvidence(id, audit),
			Headings:        siteQualityAuditHeadingEvidence(id, audit),
			Remediation:     siteQualityIssueRemediation(id),
		}
		if rule.DefaultSeverity != "" {
			issue.Severity = rule.DefaultSeverity
		}
		if audit.NumericValue != nil {
			issue.NumericValue = copyFloat64(audit.NumericValue)
		}
		if issue.Title == "" {
			issue.Title = id
		}
		issues = append(issues, issue)
	}
	sortSiteQualityIssues(issues)
	return issues, nil
}

func decorateSiteQualityIssueIDs(issues []LighthouseRunnerIssue) {
	for index := range issues {
		issue := &issues[index]
		auditID := strings.TrimSpace(issue.ID)
		if auditID == "" {
			auditID = strings.TrimSpace(issue.RuleID)
		}
		issue.RuleID, issue.ProviderAuditID = sitequalitydomain.NormalizeRuleIdentity(
			issue.RuleID,
			auditID,
			issue.ProviderAuditID,
		)
	}
}

func siteQualityAuditResources(audit siteQualityAPIAudit) []LighthouseRunnerResource {
	resources := make([]LighthouseRunnerResource, 0, len(audit.Details.Items))
	for _, rawItem := range audit.Details.Items {
		var item struct {
			URL        string   `json:"url"`
			TotalBytes *int64   `json:"totalBytes"`
			WastedMS   *float64 `json:"wastedMs"`
		}
		if err := json.Unmarshal(rawItem, &item); err != nil {
			continue
		}
		resourceURL := strings.TrimSpace(item.URL)
		if resourceURL == "" {
			continue
		}
		resources = append(resources, LighthouseRunnerResource{
			URL:        resourceURL,
			TotalBytes: copyInt64(item.TotalBytes),
			WastedMS:   copyFloat64(item.WastedMS),
		})
	}
	sort.Slice(resources, func(i, j int) bool {
		iWastedMS, jWastedMS := siteQualityResourceWastedMS(resources[i]), siteQualityResourceWastedMS(resources[j])
		if iWastedMS != jWastedMS {
			return iWastedMS > jWastedMS
		}
		iBytes, jBytes := siteQualityResourceTotalBytes(resources[i]), siteQualityResourceTotalBytes(resources[j])
		if iBytes != jBytes {
			return iBytes > jBytes
		}
		return resources[i].URL < resources[j].URL
	})
	return resources
}

func siteQualityAuditLinkEvidence(
	auditID string,
	audit siteQualityAPIAudit,
) []sitequalitydomain.SiteQualityLinkEvidence {
	if auditID != siteQualityLinkTextAuditID {
		return nil
	}
	links := make([]sitequalitydomain.SiteQualityLinkEvidence, 0, len(audit.Details.Items))
	seen := make(map[string]struct{}, len(audit.Details.Items))
	for _, rawItem := range audit.Details.Items {
		var item struct {
			Href     string `json:"href"`
			Text     string `json:"text"`
			TextLang string `json:"textLang"`
		}
		if err := json.Unmarshal(rawItem, &item); err != nil {
			continue
		}
		link := sitequalitydomain.SiteQualityLinkEvidence{
			Href:     strings.TrimSpace(item.Href),
			Text:     strings.TrimSpace(item.Text),
			TextLang: strings.TrimSpace(item.TextLang),
		}
		if link.Href == "" && link.Text == "" {
			continue
		}
		key := strings.Join([]string{link.Href, link.Text, link.TextLang}, "\x00")
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		links = append(links, link)
	}
	sort.SliceStable(links, func(i, j int) bool {
		if links[i].Href != links[j].Href {
			return links[i].Href < links[j].Href
		}
		if links[i].Text != links[j].Text {
			return links[i].Text < links[j].Text
		}
		return links[i].TextLang < links[j].TextLang
	})
	return links
}

func siteQualityAuditHeadingEvidence(
	auditID string,
	audit siteQualityAPIAudit,
) []sitequalitydomain.SiteQualityHeadingEvidence {
	if auditID != "heading-order" {
		return nil
	}
	headings := make([]sitequalitydomain.SiteQualityHeadingEvidence, 0, len(audit.Details.Items))
	seen := make(map[string]struct{})
	for _, rawItem := range audit.Details.Items {
		var item struct {
			Node struct {
				Selector    string `json:"selector"`
				Snippet     string `json:"snippet"`
				NodeLabel   string `json:"nodeLabel"`
				Explanation string `json:"explanation"`
			} `json:"node"`
		}
		if err := json.Unmarshal(rawItem, &item); err != nil {
			continue
		}
		heading := sitequalitydomain.SiteQualityHeadingEvidence{
			Level:       siteQualityHeadingLevel(item.Node.Snippet),
			Text:        strings.TrimSpace(item.Node.NodeLabel),
			Snippet:     strings.TrimSpace(item.Node.Snippet),
			Selector:    strings.TrimSpace(item.Node.Selector),
			Explanation: strings.TrimSpace(item.Node.Explanation),
		}
		if heading.Text == "" && heading.Snippet == "" && heading.Selector == "" {
			continue
		}
		key := strings.Join([]string{heading.Selector, heading.Snippet, heading.Text}, "\x00")
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		headings = append(headings, heading)
	}
	return headings
}

func siteQualityHeadingLevel(snippet string) int {
	normalized := strings.ToLower(strings.TrimSpace(snippet))
	for level := 1; level <= 6; level++ {
		if strings.Contains(normalized, fmt.Sprintf("<h%d", level)) {
			return level
		}
	}
	return 0
}

func siteQualityResourceWastedMS(resource LighthouseRunnerResource) float64 {
	if resource.WastedMS == nil {
		return 0
	}
	return *resource.WastedMS
}

func siteQualityResourceTotalBytes(resource LighthouseRunnerResource) int64 {
	if resource.TotalBytes == nil {
		return 0
	}
	return *resource.TotalBytes
}

func siteQualityAuditSeverity(audit siteQualityAPIAudit) string {
	if audit.Score != nil {
		return siteQualityIssueSeverity(*audit.Score)
	}

	savingsMS := 0.0
	if audit.Details.OverallSavingsMS != nil {
		savingsMS = *audit.Details.OverallSavingsMS
	}
	switch {
	case savingsMS >= 1000:
		return "high"
	case savingsMS >= 250:
		return "medium"
	default:
		return "low"
	}
}

func siteQualityIssueSeverity(score float64) string {
	switch {
	case score < 0.25:
		return "critical"
	case score < 0.5:
		return "high"
	case score < 0.75:
		return "medium"
	default:
		return "low"
	}
}

func siteQualityIssueRank(issue LighthouseRunnerIssue) float64 {
	score := 1.0
	if issue.Score != nil {
		score = *issue.Score
	}
	savings := 0.0
	if issue.SavingsMS != nil {
		savings += *issue.SavingsMS / 1000
	}
	if issue.SavingsBytes != nil {
		savings += float64(*issue.SavingsBytes) / (1024 * 1024)
	}
	return siteQualitySeverityRank(issue.Severity) + (1-score)*100 + savings
}

func siteQualitySeverityRank(severity string) float64 {
	switch severity {
	case "critical":
		return 400
	case "high":
		return 300
	case "medium":
		return 200
	case "low":
		return 100
	default:
		return 0
	}
}

func sortSiteQualityIssues(issues []LighthouseRunnerIssue) {
	sort.SliceStable(issues, func(i, j int) bool {
		iRank, jRank := siteQualityIssueRank(issues[i]), siteQualityIssueRank(issues[j])
		if iRank != jRank {
			return iRank > jRank
		}
		return issues[i].ID < issues[j].ID
	})
}

func siteQualityIssueRemediation(id string) *LighthouseRunnerRemediation {
	switch id {
	case "uses-long-cache-ttl":
		return &LighthouseRunnerRemediation{
			Label: "检查 Cloudflare 缓存策略",
			Route: "/services/cloudflare?tab=cache",
		}
	default:
		return nil
	}
}

func (s *LighthouseRunnerService) requireOpenSiteQualityFinding(
	id uint,
) (*sitequalitydomain.SiteQualityFinding, error) {
	if id == 0 {
		return nil, errors.New("SiteQuality finding ID is required")
	}
	return s.GetFinding(id)
}

func siteQualityRunView(run sitequalitydomain.SiteQualityRun) LighthouseRunnerRunView {
	issues := []LighthouseRunnerIssue{}
	if strings.TrimSpace(run.IssuesJSON) != "" {
		_ = json.Unmarshal([]byte(run.IssuesJSON), &issues)
	}
	decorateSiteQualityIssueIDs(issues)
	return LighthouseRunnerRunView{
		ID:                       run.ID,
		TargetID:                 run.TargetID,
		JobID:                    run.JobID,
		TargetURL:                run.TargetURL,
		CanonicalURL:             run.CanonicalURL,
		FinalURL:                 run.FinalURL,
		Strategy:                 run.Strategy,
		Status:                   run.Status,
		InitiatedByUserID:        run.InitiatedByUserID,
		PerformanceScore:         run.PerformanceScore,
		AccessibilityScore:       run.AccessibilityScore,
		BestPracticesScore:       run.BestPracticesScore,
		SEOScore:                 run.SEOScore,
		FirstContentfulPaintMS:   run.FirstContentfulPaintMS,
		LargestContentfulPaintMS: run.LargestContentfulPaintMS,
		InteractionToNextPaintMS: run.InteractionToNextPaintMS,
		CumulativeLayoutShift:    run.CumulativeLayoutShift,
		TotalBlockingTimeMS:      run.TotalBlockingTimeMS,
		SpeedIndexMS:             run.SpeedIndexMS,
		Issues:                   issues,
		ErrorMessage:             run.ErrorMessage,
		CreatedAt:                run.CreatedAt,
	}
}

func copyFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func copyInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}
