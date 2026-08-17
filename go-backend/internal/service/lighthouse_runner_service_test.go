package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const testLighthouseRunnerToken = "01234567890123456789012345678901"

func TestLighthouseRunnerServiceCaptureNormalizesAndPersistsLeasedResult(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, "Bearer "+testLighthouseRunnerToken, request.Header.Get("Authorization"))
		var payload struct {
			URL      string `json:"url"`
			Strategy string `json:"strategy"`
		}
		require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
		require.Equal(t, "https://example.com/products/wheel", payload.URL)
		require.Equal(t, sitequalitydomain.SiteQualityStrategyMobile, payload.Strategy)
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(siteQualityRunnerTestResponseWithCleanHeadings(t, `{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {
					"performance": {"score": 0.47},
					"accessibility": {"score": 0.91},
					"best-practices": {"score": 0.87},
					"seo": {"score": 1}
				},
				"audits": {
					"first-contentful-paint": {"id": "first-contentful-paint", "score": 0.5, "scoreDisplayMode": "metric", "numericValue": 1600},
					"largest-contentful-paint": {"id": "largest-contentful-paint", "score": 0.2, "scoreDisplayMode": "metric", "numericValue": 4200},
					"interaction-to-next-paint": {"id": "interaction-to-next-paint", "score": 0.7, "scoreDisplayMode": "metric", "numericValue": 280},
					"cumulative-layout-shift": {"id": "cumulative-layout-shift", "score": 0.8, "scoreDisplayMode": "metric", "numericValue": 0.18},
					"total-blocking-time": {"id": "total-blocking-time", "score": 0.5, "scoreDisplayMode": "metric", "numericValue": 630},
					"speed-index": {"id": "speed-index", "score": 0.4, "scoreDisplayMode": "metric", "numericValue": 3700},
					"uses-long-cache-ttl": {
						"id": "uses-long-cache-ttl",
						"title": "Serve static assets with an efficient cache policy",
						"description": "Cache policy description",
						"score": 0,
						"displayValue": "Potential savings of 1,200 KiB",
						"scoreDisplayMode": "numeric",
						"details": {"overallSavingsMs": 950, "overallSavingsBytes": 1228800}
					},
					"uses-text-compression": {
						"id": "uses-text-compression",
						"title": "Enable text compression",
						"score": 0.5,
						"scoreDisplayMode": "numeric"
					},
					"render-blocking-resources": {
						"id": "render-blocking-resources",
						"title": "Render blocking requests",
						"description": "Network requests delay the initial render and can affect LCP.",
						"displayValue": "Potential savings of 280 ms",
						"scoreDisplayMode": "informative",
						"details": {
							"overallSavingsMs": 280,
							"items": [
								{"url": "https://example.com/_nuxt/app.css", "totalBytes": 42300, "wastedMs": 190},
								{"url": "https://example.com/_nuxt/vendor.css", "totalBytes": 19800, "wastedMs": 90}
							]
						}
					},
					"unsized-images": {
						"id": "unsized-images",
						"title": "Image elements do not have explicit width and height",
						"description": "Set an explicit width and height on image elements to reduce layout shifts.",
						"score": 0,
						"scoreDisplayMode": "metricSavings",
						"details": {
							"items": [
								{"url": "https://example.com/icons/payment/visa.svg"},
								{"url": "https://example.com/uploads/site/logo.webp"}
							]
						}
					},
					"passed-audit": {"id": "passed-audit", "score": 1, "scoreDisplayMode": "numeric"}
					,
					"unknown-provider-audit": {
						"id": "unknown-provider-audit",
						"score": 1,
						"details": {
							"items": [
								{"url": {"unsupported": "object"}, "totalBytes": {"unsupported": true}}
							]
						}
					}
				}
			}
		}`)))
	}))
	t.Cleanup(server.Close)

	service := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         server.URL,
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)
	service.ConfigureHTTPClient(server.Client(), server.URL)
	job := createLeasedSiteQualityJob(
		t,
		db,
		"https://example.com/products/wheel",
		sitequalitydomain.SiteQualityStrategyMobile,
		"runner-normalize-job",
	)
	service.ConfigureJobRepository(repository.NewSiteQualityJobRepository(db))

	result, err := service.Capture(context.Background(), LighthouseRunnerCaptureInput{
		LighthouseRunnerRunInput: LighthouseRunnerRunInput{
			URL:               "https://example.com/products/wheel",
			Strategy:          sitequalitydomain.SiteQualityStrategyMobile,
			InitiatedByUserID: 42,
		},
		TargetID:        &job.TargetID,
		JobID:           &job.ID,
		LeaseWorkerID:   job.LockedBy,
		LeaseGeneration: job.LeaseGeneration,
		CanonicalURL:    "https://example.com/products/wheel",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotZero(t, result.ID)
	require.NotNil(t, result.JobID)
	require.Equal(t, job.ID, *result.JobID)
	require.Equal(t, sitequalitydomain.SiteQualityRunStatusSuccess, result.Status)
	require.Equal(t, 47, *result.PerformanceScore)
	require.Equal(t, 1600.0, *result.FirstContentfulPaintMS)
	require.Equal(t, 4200.0, *result.LargestContentfulPaintMS)
	require.Len(t, result.Issues, 2)
	require.Equal(t, "uses-long-cache-ttl", result.Issues[0].ID)
	require.NotNil(t, result.Issues[0].Remediation)
	require.Equal(t, "/services/cloudflare?tab=cache", result.Issues[0].Remediation.Route)
	var renderBlockingIssue *LighthouseRunnerIssue
	var unsizedImagesIssue *LighthouseRunnerIssue
	for index := range result.Issues {
		if result.Issues[index].ID == "render-blocking-resources" {
			renderBlockingIssue = &result.Issues[index]
		}
		if result.Issues[index].ID == "unsized-images" {
			unsizedImagesIssue = &result.Issues[index]
		}
	}
	require.NotNil(t, renderBlockingIssue)
	require.Nil(t, renderBlockingIssue.Score)
	require.Equal(t, "Potential savings of 280 ms", renderBlockingIssue.DisplayValue)
	require.Equal(t, "medium", renderBlockingIssue.Severity)
	require.NotNil(t, renderBlockingIssue.SavingsMS)
	require.Equal(t, 280.0, *renderBlockingIssue.SavingsMS)
	require.Len(t, renderBlockingIssue.Resources, 2)
	require.Equal(t, "https://example.com/_nuxt/app.css", renderBlockingIssue.Resources[0].URL)
	require.NotNil(t, renderBlockingIssue.Resources[0].WastedMS)
	require.Equal(t, 190.0, *renderBlockingIssue.Resources[0].WastedMS)
	require.Nil(t, unsizedImagesIssue)

	list, err := service.List(repository.SiteQualityRunListFilter{Page: 1, PageSize: 20})
	require.NoError(t, err)
	require.EqualValues(t, 1, list.Total)
	require.Len(t, list.Items, 1)
	require.Equal(t, result.ID, list.Items[0].ID)

	findings, findingTotal, err := service.ListFindings(repository.SiteQualityFindingListFilter{
		Page:     1,
		PageSize: 20,
		State:    "all",
	})
	require.NoError(t, err)
	require.EqualValues(t, 0, findingTotal)
	require.Empty(t, findings)
}

func TestLighthouseRunnerCaptureAppliesTargetSourceTypeToSchemaChecks(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(siteQualityRunnerTestResponseWithCleanHeadings(t, `{
			"lighthouseResult": {
				"finalUrl": "https://example.com/p/legacy-wheel",
				"categories": {"performance": {"score": 1}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`)))
	}))
	t.Cleanup(server.Close)

	service := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         server.URL,
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)
	service.ConfigureHTTPClient(server.Client(), server.URL)
	job := createLeasedSiteQualityJob(
		t,
		db,
		"https://example.com/p/legacy-wheel",
		sitequalitydomain.SiteQualityStrategyDesktop,
		"runner-product-intent-job",
	)
	service.ConfigureJobRepository(repository.NewSiteQualityJobRepository(db))

	result, err := service.Capture(context.Background(), LighthouseRunnerCaptureInput{
		LighthouseRunnerRunInput: LighthouseRunnerRunInput{
			URL:      "https://example.com/p/legacy-wheel",
			Strategy: sitequalitydomain.SiteQualityStrategyDesktop,
		},
		TargetID:         &job.TargetID,
		JobID:            &job.ID,
		LeaseWorkerID:    job.LockedBy,
		LeaseGeneration:  job.LeaseGeneration,
		CanonicalURL:     "https://example.com/p/legacy-wheel",
		TargetSource:     sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
		TargetSourceType: seodomain.RouteSourceProduct,
	})

	require.NoError(t, err)
	require.Len(t, result.Issues, 1)
	require.Equal(t, siteQualityStructuredDataMissingRequiredTypeAuditID, result.Issues[0].ID)
}

func TestSiteQualityEngineConfirmsAndVerifiesFindingLifecycle(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	responses := []string{
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.91}},
				"audits": {
					"render-blocking-resources": {
						"id": "render-blocking-resources",
						"title": "Render blocking requests",
						"scoreDisplayMode": "informative",
						"details": {
							"overallSavingsMs": 280,
							"items": [
								{"url": "https://example.com/_nuxt/app.css", "totalBytes": 42300, "wastedMs": 280}
							]
						}
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.91}},
				"audits": {
					"render-blocking-resources": {
						"id": "render-blocking-resources",
						"title": "Render blocking requests",
						"scoreDisplayMode": "informative",
						"details": {
							"overallSavingsMs": 280,
							"items": [
								{"url": "https://example.com/_nuxt/app.css", "totalBytes": 42300, "wastedMs": 280}
							]
						}
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.91}},
				"audits": {
					"render-blocking-resources": {
						"id": "render-blocking-resources",
						"title": "Render blocking requests",
						"scoreDisplayMode": "informative",
						"details": {"overallSavingsMs": 280}
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.91}},
				"audits": {
					"render-blocking-resources": {
						"id": "render-blocking-resources",
						"title": "Render blocking requests",
						"scoreDisplayMode": "informative",
						"details": {"overallSavingsMs": 280}
					}
				}
			}
		}`,
		`{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 0.98}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`,
	}
	responseIndex := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(siteQualityRunnerTestResponseWithCleanHeadings(t, responses[responseIndex])))
		responseIndex++
	}))
	t.Cleanup(server.Close)

	service := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         server.URL,
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)
	service.ConfigureHTTPClient(server.Client(), server.URL)

	engine := NewSiteQualityEngineService(
		repository.NewSiteQualityTargetRepository(db),
		repository.NewSiteQualityJobRepository(db),
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		nil,
		service,
		SiteQualityEngineConfig{
			BaseURL:                  "https://example.com",
			SampleCount:              3,
			RequiredConfirmations:    2,
			RequiredCleanEvaluations: 2,
			WorkerBatchLimit:         1,
			ProviderRequestInterval:  time.Nanosecond,
		},
	)
	_, err := engine.EnqueueManualTarget(context.Background(),
		"https://example.com/products/wheel",
		sitequalitydomain.SiteQualityStrategyMobile,
		42,
		sitequalitydomain.SiteQualityJobKindManual,
	)
	require.NoError(t, err)
	result, err := engine.ProcessReady(context.Background(), time.Now().UTC(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Succeeded)

	findings, total, err := service.ListFindings(repository.SiteQualityFindingListFilter{
		Page:     1,
		PageSize: 20,
		State:    "all",
	})
	require.NoError(t, err)
	require.EqualValues(t, 1, total)
	require.Len(t, findings, 1)
	require.Equal(t, sitequalitydomain.SiteQualityFindingStateOpen, findings[0].State)
	require.Equal(t, 1, findings[0].ResourceCount)
	require.NotEmpty(t, findings[0].LatestEvidence)
	require.Equal(t, 3, findings[0].SampleCount)
	require.Equal(t, 2, findings[0].Confirmations)
	require.InDelta(t, 2.0/3.0, findings[0].Confidence, 0.001)

	resolved, err := service.ResolveFinding(findings[0].ID, 42, sitequalitydomain.SiteQualityFindingResolutionInput{
		ResolutionNote: "Moved non-critical stylesheet loading behind the initial render.",
	})
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityFindingStateResolved, resolved.State)

	_, err = engine.EnqueueRecheckFinding(resolved, 42)
	require.NoError(t, err)
	result, err = engine.ProcessReady(context.Background(), time.Now().UTC(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Succeeded)
	cleanOnce, err := service.GetFinding(findings[0].ID)
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityFindingStateResolved, cleanOnce.State)
	require.Equal(t, 1, cleanOnce.ConsecutiveClean)

	_, err = engine.EnqueueRecheckFinding(cleanOnce, 42)
	require.NoError(t, err)
	result, err = engine.ProcessReady(context.Background(), time.Now().UTC(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Succeeded)
	verified, err := service.GetFinding(findings[0].ID)
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityFindingStateVerified, verified.State)
	require.NotNil(t, verified.VerifiedAt)
	require.Equal(t, resolved.ResolutionNote, verified.ResolutionNote)

	_, err = engine.EnqueueRecheckFinding(verified, 42)
	require.NoError(t, err)
	result, err = engine.ProcessReady(context.Background(), time.Now().UTC(), 1)
	require.NoError(t, err)
	require.Equal(t, 1, result.Succeeded)
	reopened, err := service.GetFinding(findings[0].ID)
	require.NoError(t, err)
	require.Equal(t, sitequalitydomain.SiteQualityFindingStateOpen, reopened.State)
	require.Empty(t, reopened.ResolutionNote)
	require.Nil(t, reopened.ResolvedAt)
	require.Nil(t, reopened.VerifiedAt)

	events, eventTotal, err := service.ListFindingEvents(findings[0].ID, 1, 20)
	require.NoError(t, err)
	require.EqualValues(t, 4, eventTotal)
	require.Len(t, events, 4)
	require.ElementsMatch(t, []string{
		sitequalitydomain.SiteQualityFindingEventDetected,
		sitequalitydomain.SiteQualityFindingEventResolutionRecorded,
		sitequalitydomain.SiteQualityFindingEventVerificationPassed,
		sitequalitydomain.SiteQualityFindingEventReopened,
	}, []string{
		events[0].EventType,
		events[1].EventType,
		events[2].EventType,
		events[3].EventType,
	})
}

func TestSiteQualityEvaluationCleanRequiresAuditAbsentFromAllSamples(t *testing.T) {
	savingsMS := 280.0
	decision, _ := evaluateSiteQualityRuns(
		sitequalitydomain.SiteQualityTarget{
			ID:           1,
			CanonicalURL: "https://example.com/products/wheel",
		},
		sitequalitydomain.SiteQualityJob{
			TargetID:              1,
			Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
			SampleCount:           3,
			RequiredConfirmations: 2,
		},
		[]LighthouseRunnerRunView{
			{
				ID: 1,
				Issues: []LighthouseRunnerIssue{
					{
						ID:          "render-blocking-resources",
						Title:       "Render blocking requests",
						Severity:    "medium",
						SavingsMS:   &savingsMS,
						RuleVersion: siteQualityAuditRuleVersion,
					},
				},
			},
			{ID: 2},
			{ID: 3},
		},
	)

	require.Empty(t, decision.Confirmed)
	require.Contains(t, decision.Observed, "render-blocking-resources")
	require.NotContains(t, decision.Clean, "render-blocking-resources")
}

func TestSiteQualityAuditRegistryExcludesImageDimensionRules(t *testing.T) {
	for _, auditID := range []string{
		"uses-responsive-images",
		"unsized-images",
		"uses-responsive-images-snapshot",
	} {
		_, ok := siteQualityLookupAuditRule(auditID)
		require.False(t, ok, auditID)
	}

	for _, auditID := range []string{
		"uses-optimized-images",
		"modern-image-formats",
		"efficient-animated-content",
		"heading-order",
		siteQualityHeadingMissingH1AuditID,
		siteQualityHeadingMultipleH1AuditID,
		siteQualityHeadingScanFailedAuditID,
		siteQualityStructuredDataScanFailedAuditID,
		siteQualityStructuredDataInvalidJSONLDAuditID,
		siteQualityStructuredDataMissingStructuredDataAuditID,
		siteQualityStructuredDataMissingRequiredTypeAuditID,
		siteQualityStructuredDataDuplicatePrimaryTypeAuditID,
		siteQualityStructuredDataURLMismatchAuditID,
		siteQualityStructuredDataBreadcrumbInvalidAuditID,
		siteQualityStructuredDataProductInvalidAuditID,
		siteQualityStructuredDataFAQInvalidAuditID,
		siteQualityStructuredDataFAQContentMismatchAuditID,
		siteQualityStructuredDataArticleInvalidAuditID,
		siteQualityStructuredDataOrganizationInvalidAuditID,
		siteQualityStructuredDataWebPageInvalidAuditID,
	} {
		_, ok := siteQualityLookupAuditRule(auditID)
		require.True(t, ok, auditID)
	}
}

func TestNormalizeSiteQualityIssuesCapturesHeadingOrderEvidence(t *testing.T) {
	var audit siteQualityAPIAudit
	require.NoError(t, json.Unmarshal([]byte(`{
		"id": "heading-order",
		"title": "Heading elements are not in a sequentially-descending order",
		"description": "Properly ordered headings make pages easier to browse with assistive technology.",
		"score": 0,
		"scoreDisplayMode": "numeric",
		"details": {
			"items": [
				{
					"node": {
						"selector": ".home-card > h3",
						"snippet": "<h3 class=\"text-xl\">Secure Payment</h3>",
						"nodeLabel": "Secure Payment",
						"explanation": "Fix any skipped heading levels."
					}
				}
			]
		}
	}`), &audit))

	issues, err := normalizeSiteQualityIssues(map[string]siteQualityAPIAudit{
		"heading-order": audit,
	})

	require.NoError(t, err)
	require.Len(t, issues, 1)
	require.Equal(t, "heading-order", issues[0].ID)
	require.Equal(t, "headings", issues[0].Kind)
	require.Equal(t, "high", issues[0].Severity)
	require.Len(t, issues[0].Headings, 1)
	require.Equal(t, 3, issues[0].Headings[0].Level)
	require.Equal(t, "Secure Payment", issues[0].Headings[0].Text)
	require.Equal(t, ".home-card > h3", issues[0].Headings[0].Selector)
}

func TestSiteQualityHeadingOutlineIssuesDetectMissingAndMultipleH1(t *testing.T) {
	missingIssues := siteQualityHeadingOutlineIssues([]siteQualityHeadingNode{
		{Level: 2, Text: "Shop with Confidence", Selector: "main > h2"},
		{Level: 3, Text: "Secure Payment", Selector: "main > section > h3"},
	})
	require.Len(t, missingIssues, 2)
	require.Equal(t, siteQualityHeadingMissingH1AuditID, missingIssues[0].ID)
	require.Equal(t, "critical", missingIssues[0].Severity)
	require.Len(t, missingIssues[0].Headings, 2)
	require.Equal(t, "heading-order", missingIssues[1].ID)
	require.Equal(t, "high", missingIssues[1].Severity)

	multipleIssues := siteQualityHeadingOutlineIssues([]siteQualityHeadingNode{
		{Level: 1, Text: "Carbon Rims", Selector: "main > h1:nth-of-type(1)"},
		{Level: 1, Text: "Wheelsets", Selector: "main > h1:nth-of-type(2)"},
		{Level: 2, Text: "Shop", Selector: "main > h2"},
	})
	require.Len(t, multipleIssues, 1)
	require.Equal(t, siteQualityHeadingMultipleH1AuditID, multipleIssues[0].ID)
	require.Equal(t, "medium", multipleIssues[0].Severity)
	require.Len(t, multipleIssues[0].Headings, 2)
}

func TestRenderedHeadingAuditIssuesUseFinalDOMSnapshot(t *testing.T) {
	issues := siteQualityRenderedHeadingAuditIssues(
		"https://example.com/products/wheel",
		"https://example.com/products/wheel",
		&siteQualityRenderedHeadingAudit{
			Status:   "complete",
			FinalURL: "https://example.com/products/wheel",
			Headings: []sitequalitydomain.SiteQualityHeadingEvidence{
				{Level: 1, Text: "Carbon Rims", Selector: "main > h1"},
				{Level: 3, Text: "Secure Payment", Selector: "main > section > h3"},
			},
		},
	)

	require.Len(t, issues, 1)
	require.Equal(t, "heading-order", issues[0].ID)
	require.Equal(t, "headings", issues[0].Kind)
	require.Len(t, issues[0].Headings, 2)
	require.Contains(t, issues[0].Headings[1].Explanation, "H3 follows H1")
}

func TestRenderedHeadingAuditFailureBlocksCleanHeadingResult(t *testing.T) {
	issues := siteQualityRenderedHeadingAuditIssues(
		"https://example.com/products/wheel",
		"https://example.com/products/wheel",
		&siteQualityRenderedHeadingAudit{
			Status:   "failed",
			FinalURL: "https://example.com/products/wheel",
			Error:    "navigation timeout",
		},
	)

	require.Len(t, issues, 1)
	require.Equal(t, siteQualityHeadingScanFailedAuditID, issues[0].ID)
	require.Equal(t, "critical", issues[0].Severity)
	require.Contains(t, issues[0].Description, "navigation timeout")
}

func TestRenderedStructuredDataAuditIssuesValidateProductSchema(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/shop/carbon-rim",
		"https://example.com/shop/carbon-rim",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/shop/carbon-rim",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/shop/carbon-rim",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Index:    0,
					Selector: "html > head > script:nth-of-type(1)",
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"Product"},
							Type:      "Product",
							Name:      "Carbon Rim",
							URL:       "https://example.com/shop/carbon-rim",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "Product",
								"name": "Carbon Rim",
								"url": "https://example.com/shop/carbon-rim",
								"image": ["https://example.com/uploads/rim.webp"],
								"offers": {
									"@type": "Offer",
									"price": 899,
									"priceCurrency": "USD",
									"availability": "https://schema.org/InStock"
								}
							}`),
						},
					},
				},
			},
		},
	)

	require.Empty(t, issues)
}

func TestRenderedStructuredDataAuditIssuesDetectProductDefects(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/shop/carbon-rim",
		"https://example.com/shop/carbon-rim",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/shop/carbon-rim",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/shop/carbon-rim",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Index:    0,
					Selector: "html > head > script:nth-of-type(1)",
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"Product"},
							Type:      "Product",
							Name:      "Carbon Rim",
							URL:       "https://example.com/shop/other-rim",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "Product",
								"name": "Carbon Rim",
								"url": "https://example.com/shop/other-rim",
								"image": ["https://example.com/uploads/rim.webp"]
							}`),
						},
					},
				},
			},
		},
	)

	require.Len(t, issues, 2)
	require.Equal(t, siteQualityStructuredDataProductInvalidAuditID, issues[0].ID)
	require.Equal(t, "schema", issues[0].Kind)
	require.Equal(t, "critical", issues[0].Severity)
	require.Len(t, issues[0].StructuredData, 1)
	require.Equal(t, "offers", issues[0].StructuredData[0].Property)
	require.Equal(t, siteQualityStructuredDataURLMismatchAuditID, issues[1].ID)
	require.Equal(t, "url", issues[1].StructuredData[0].Property)
}

func TestRenderedStructuredDataAuditIssuesDetectInvalidJSONLD(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/",
		"https://example.com/",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Index:      0,
					Selector:   "html > head > script:nth-of-type(1)",
					Raw:        `{"@context": "https://schema.org",`,
					ParseError: "Unexpected end of JSON input",
				},
			},
		},
	)

	require.NotEmpty(t, issues)
	require.Equal(t, siteQualityStructuredDataInvalidJSONLDAuditID, issues[0].ID)
	require.Equal(t, "critical", issues[0].Severity)
	require.Contains(t, issues[0].StructuredData[0].Explanation, "Unexpected end")
}

func TestRenderedStructuredDataAuditUsesTargetSourceTypeForProductIntent(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/p/legacy-wheel",
		"https://example.com/p/legacy-wheel",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/p/legacy-wheel",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/p/legacy-wheel",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"Organization"},
							Type:      "Organization",
							Name:      "TANZANITE",
							URL:       "https://example.com/",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "Organization",
								"name": "TANZANITE",
								"url": "https://example.com/"
							}`),
						},
					},
				},
			},
		},
		siteQualityStructuredDataPageIntent{
			Source:     sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
			SourceType: seodomain.RouteSourceProduct,
		},
	)

	require.Len(t, issues, 1)
	require.Equal(t, siteQualityStructuredDataMissingRequiredTypeAuditID, issues[0].ID)
	require.Contains(t, issues[0].Description, "Product or ProductGroup")
}

func TestRenderedStructuredDataAuditUsesTargetSourceTypeForArticleIntent(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/guides/road-bike",
		"https://example.com/guides/road-bike",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/guides/road-bike",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/guides/road-bike",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"WebPage"},
							Type:      "WebPage",
							Name:      "Road Bike Guide",
							URL:       "https://example.com/guides/road-bike",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "WebPage",
								"name": "Road Bike Guide",
								"url": "https://example.com/guides/road-bike"
							}`),
						},
					},
				},
			},
		},
		siteQualityStructuredDataPageIntent{
			Source:     sitequalitydomain.SiteQualityTargetSourceRouteCatalog,
			SourceType: seodomain.RouteSourceBlog,
		},
	)

	require.Len(t, issues, 1)
	require.Equal(t, siteQualityStructuredDataMissingRequiredTypeAuditID, issues[0].ID)
	require.Contains(t, issues[0].Description, "Article, BlogPosting, or NewsArticle")
}

func TestRenderedStructuredDataAuditAcceptsCompleteArticleSchema(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/blog/rim-testing",
		"https://example.com/blog/rim-testing",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/blog/rim-testing",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/blog/rim-testing",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"BlogPosting"},
							Type:      "BlogPosting",
							Name:      "Rim Testing",
							URL:       "https://example.com/blog/rim-testing",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "BlogPosting",
								"headline": "Rim Testing",
								"image": ["https://example.com/uploads/rim-testing.webp"],
								"datePublished": "2026-08-10",
								"author": {"@type": "Person", "name": "TANZANITE Lab"},
								"publisher": {"@type": "Organization", "name": "TANZANITE"},
								"url": "https://example.com/blog/rim-testing"
							}`),
						},
					},
				},
			},
		},
		siteQualityStructuredDataPageIntent{
			SourceType: seodomain.RouteSourceBlog,
		},
	)

	require.Empty(t, issues)
}

func TestRenderedStructuredDataAuditDetectsIncompleteArticleSchema(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/blog/rim-testing",
		"https://example.com/blog/rim-testing",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/blog/rim-testing",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/blog/rim-testing",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"Article"},
							Type:      "Article",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "Article",
								"headline": "Rim Testing"
							}`),
						},
					},
				},
			},
		},
		siteQualityStructuredDataPageIntent{
			SourceType: seodomain.RouteSourceBlog,
		},
	)

	require.Len(t, issues, 1)
	require.Equal(t, siteQualityStructuredDataArticleInvalidAuditID, issues[0].ID)
	require.ElementsMatch(t, []string{"image", "datePublished", "author", "url"}, structuredDataIssueProperties(issues[0]))
}

func TestRenderedStructuredDataAuditRequiresOrganizationLogoOnHome(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/",
		"https://example.com/",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"Organization"},
							Type:      "Organization",
							Name:      "TANZANITE",
							URL:       "https://example.com/",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "Organization",
								"name": "TANZANITE",
								"url": "https://example.com/"
							}`),
						},
					},
				},
			},
		},
	)

	require.Len(t, issues, 1)
	require.Equal(t, siteQualityStructuredDataOrganizationInvalidAuditID, issues[0].ID)
	require.Equal(t, "logo", issues[0].StructuredData[0].Property)
}

func TestRenderedStructuredDataAuditDetectsFAQContentMismatch(t *testing.T) {
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/support/faqs",
		"https://example.com/support/faqs",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/support/faqs",
			Page: siteQualityStructuredDataPage{
				CanonicalURL:     "https://example.com/support/faqs",
				FAQQuestionCount: 4,
			},
			JSONLD: []siteQualityStructuredDataScript{
				{
					Nodes: []siteQualityStructuredDataNode{
						{
							Types:     []string{"FAQPage"},
							Type:      "FAQPage",
							GraphPath: "script[0]",
							Data: json.RawMessage(`{
								"@context": "https://schema.org",
								"@type": "FAQPage",
								"mainEntity": [
									{"@type": "Question", "name": "Question one", "acceptedAnswer": {"@type": "Answer", "text": "Answer one"}},
									{"@type": "Question", "name": "Question two", "acceptedAnswer": {"@type": "Answer", "text": "Answer two"}}
								]
							}`),
						},
					},
				},
			},
		},
	)

	require.Len(t, issues, 1)
	require.Equal(t, siteQualityStructuredDataFAQContentMismatchAuditID, issues[0].ID)
	require.Contains(t, issues[0].StructuredData[0].Explanation, "Rendered FAQ-like questions: 4")
}

func TestRenderedStructuredDataAuditDoesNotFlagRepeatedSameProductEntity(t *testing.T) {
	productData := `{
		"@context": "https://schema.org",
		"@type": "Product",
		"@id": "https://example.com/shop/carbon-rim#product",
		"name": "Carbon Rim",
		"url": "https://example.com/shop/carbon-rim",
		"image": ["https://example.com/uploads/rim.webp"],
		"offers": {"@type": "Offer", "price": 899, "priceCurrency": "USD", "availability": "https://schema.org/InStock"}
	}`
	issues := siteQualityRenderedStructuredDataAuditIssues(
		"https://example.com/shop/carbon-rim",
		"https://example.com/shop/carbon-rim",
		&siteQualityRenderedStructuredDataAudit{
			Status:   "complete",
			Source:   "chrome-rendered-dom",
			FinalURL: "https://example.com/shop/carbon-rim",
			Page: siteQualityStructuredDataPage{
				CanonicalURL: "https://example.com/shop/carbon-rim",
			},
			JSONLD: []siteQualityStructuredDataScript{
				{Nodes: []siteQualityStructuredDataNode{
					{Types: []string{"Product"}, Type: "Product", ID: "https://example.com/shop/carbon-rim#product", URL: "https://example.com/shop/carbon-rim", GraphPath: "script[0]", Data: json.RawMessage(productData)},
				}},
				{Nodes: []siteQualityStructuredDataNode{
					{Types: []string{"Product"}, Type: "Product", ID: "https://example.com/shop/carbon-rim#product", URL: "https://example.com/shop/carbon-rim", GraphPath: "script[1]", Data: json.RawMessage(productData)},
				}},
			},
		},
	)

	require.Empty(t, issues)
}

func structuredDataIssueProperties(issue LighthouseRunnerIssue) []string {
	properties := make([]string, 0, len(issue.StructuredData))
	for _, evidence := range issue.StructuredData {
		properties = append(properties, evidence.Property)
	}
	return properties
}

func TestApplySiteQualityResultReplacesLighthouseHeadingOrderWithRenderedDOM(t *testing.T) {
	var result siteQualityAPIResponse
	require.NoError(t, json.Unmarshal([]byte(`{
		"lighthouseResult": {
			"finalUrl": "https://example.com/products/wheel",
			"renderedHeadings": {
				"status": "complete",
				"source": "chrome-rendered-dom",
				"finalUrl": "https://example.com/products/wheel",
				"headings": [
					{"level": 1, "text": "Carbon Rims", "selector": "main > h1"}
				]
			},
			"renderedStructuredData": {
				"status": "complete",
				"source": "chrome-rendered-dom",
				"finalUrl": "https://example.com/products/wheel",
				"page": {"title": "Carbon Rims", "canonicalUrl": "https://example.com/products/wheel"},
				"jsonLd": [
					{
						"index": 0,
						"selector": "html > head > script:nth-of-type(1)",
						"raw": "{\"@context\":\"https://schema.org\",\"@type\":\"Organization\",\"name\":\"TANZANITE\",\"url\":\"https://example.com/\"}",
						"nodes": [
							{
								"types": ["Organization"],
								"type": "Organization",
								"name": "TANZANITE",
								"url": "https://example.com/",
								"graphPath": "script[0]",
								"data": {"@context": "https://schema.org", "@type": "Organization", "name": "TANZANITE", "url": "https://example.com/"}
							}
						]
					}
				],
				"microdata": [],
				"rdfa": []
			},
			"categories": {"performance": {"score": 1}},
			"audits": {
				"heading-order": {
					"id": "heading-order",
					"title": "Heading elements are not in a sequentially-descending order",
					"description": "Provider stale evidence",
					"score": 0,
					"scoreDisplayMode": "numeric",
					"details": {
						"items": [
							{
								"node": {
									"selector": ".stale > h3",
									"snippet": "<h3>Stale</h3>",
									"nodeLabel": "Stale"
								}
							}
						]
					}
				}
			}
		}
	}`), &result))

	run := sitequalitydomain.SiteQualityRun{
		TargetURL: "https://example.com/products/wheel",
		Strategy:  sitequalitydomain.SiteQualityStrategyDesktop,
		Status:    sitequalitydomain.SiteQualityRunStatusSuccess,
	}
	applySiteQualityResult(&run, &result)

	require.Equal(t, sitequalitydomain.SiteQualityRunStatusSuccess, run.Status)
	var issues []LighthouseRunnerIssue
	require.NoError(t, json.Unmarshal([]byte(run.IssuesJSON), &issues))
	require.Empty(t, issues)
}

func TestFetchSiteQualityHeadingOutlineParsesHTMLHeadings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = writer.Write([]byte(`<!doctype html>
			<html>
				<body>
					<main id="content">
						<h1>Carbon Rims</h1>
						<section><h2>Shop with Confidence</h2><h3 class="card-title">Secure Payment</h3></section>
						<h2 hidden>Hidden title</h2>
					</main>
				</body>
			</html>`))
	}))
	t.Cleanup(server.Close)

	headings, err := fetchSiteQualityHeadingOutline(context.Background(), server.URL)

	require.NoError(t, err)
	require.Len(t, headings, 3)
	require.Equal(t, []int{1, 2, 3}, []int{headings[0].Level, headings[1].Level, headings[2].Level})
	require.Equal(t, "Secure Payment", headings[2].Text)
	require.Contains(t, headings[2].Snippet, `<h3 class="card-title">`)
	require.Contains(t, headings[2].Selector, "main#content")
}

func TestSiteQualityEngineSummaryDegradesOnLatestFailureAndKeepsLatestSuccess(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	successAt := time.Date(2026, time.August, 17, 9, 0, 0, 0, time.UTC)
	failureAt := successAt.Add(time.Minute)
	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityRun{
		TargetURL: "https://example.com/products/wheel",
		Strategy:  sitequalitydomain.SiteQualityStrategyMobile,
		Status:    sitequalitydomain.SiteQualityRunStatusSuccess,
		CreatedAt: successAt,
		UpdatedAt: successAt,
	}).Error)
	require.NoError(t, db.Create(&sitequalitydomain.SiteQualityRun{
		TargetURL:    "https://example.com/products/wheel",
		Strategy:     sitequalitydomain.SiteQualityStrategyMobile,
		Status:       sitequalitydomain.SiteQualityRunStatusFailed,
		ErrorMessage: "runner timed out",
		CreatedAt:    failureAt,
		UpdatedAt:    failureAt,
	}).Error)

	runner := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         "http://runner.internal",
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)
	engine := NewSiteQualityEngineService(
		repository.NewSiteQualityTargetRepository(db),
		repository.NewSiteQualityJobRepository(db),
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		nil,
		runner,
		SiteQualityEngineConfig{BaseURL: "https://example.com"},
	)

	summary, err := engine.Summary(failureAt.Add(time.Minute))
	require.NoError(t, err)
	require.True(t, summary.RunnerConfigured)
	require.Equal(t, "degraded", summary.Status)
	require.NotNil(t, summary.LatestRun)
	require.Equal(t, sitequalitydomain.SiteQualityRunStatusFailed, summary.LatestRun.Status)
	require.NotNil(t, summary.LatestSuccessAt)
	require.Equal(t, successAt, *summary.LatestSuccessAt)
	require.Contains(t, summary.Warnings, "latest Lighthouse sample failed")
}

func TestLighthouseRunnerServiceCapturePersistsRunnerFailureWithLease(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusTooManyRequests)
		_, _ = writer.Write([]byte(`{"error":{"message":"quota exceeded"}}`))
	}))
	t.Cleanup(server.Close)

	service := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         server.URL,
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)
	service.ConfigureHTTPClient(server.Client(), server.URL)
	job := createLeasedSiteQualityJob(
		t,
		db,
		"https://example.com/",
		sitequalitydomain.SiteQualityStrategyDesktop,
		"runner-failure-job",
	)
	service.ConfigureJobRepository(repository.NewSiteQualityJobRepository(db))

	result, err := service.Capture(context.Background(), LighthouseRunnerCaptureInput{
		LighthouseRunnerRunInput: LighthouseRunnerRunInput{
			URL:      "https://example.com/",
			Strategy: sitequalitydomain.SiteQualityStrategyDesktop,
		},
		TargetID:        &job.TargetID,
		JobID:           &job.ID,
		LeaseWorkerID:   job.LockedBy,
		LeaseGeneration: job.LeaseGeneration,
		CanonicalURL:    "https://example.com/",
	})
	require.Error(t, err)
	require.NotNil(t, result)
	require.NotZero(t, result.ID)
	require.Equal(t, sitequalitydomain.SiteQualityRunStatusFailed, result.Status)
	require.Contains(t, result.ErrorMessage, "HTTP 429")

	list, listErr := service.List(repository.SiteQualityRunListFilter{Page: 1, PageSize: 20})
	require.NoError(t, listErr)
	require.EqualValues(t, 1, list.Total)
	require.Equal(t, sitequalitydomain.SiteQualityRunStatusFailed, list.Items[0].Status)
}

func TestLighthouseRunnerServiceRunRequiresJobLease(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requests++
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	service := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         server.URL,
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)
	service.ConfigureHTTPClient(server.Client(), server.URL)

	result, err := service.Run(context.Background(), LighthouseRunnerRunInput{
		URL:      "https://example.com/",
		Strategy: sitequalitydomain.SiteQualityStrategyMobile,
	})
	require.ErrorIs(t, err, ErrSiteQualityJobRequired)
	require.Nil(t, result)
	require.Zero(t, requests)

	list, listErr := service.List(repository.SiteQualityRunListFilter{Page: 1, PageSize: 20})
	require.NoError(t, listErr)
	require.EqualValues(t, 0, list.Total)
}

func TestLighthouseRunnerServiceCaptureRejectsExpiredLeaseBeforePersistingRun(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	now := time.Now().UTC()
	expiredAt := now.Add(-time.Minute)
	job := sitequalitydomain.SiteQualityJob{
		TargetID:              1,
		Strategy:              sitequalitydomain.SiteQualityStrategyMobile,
		Kind:                  sitequalitydomain.SiteQualityJobKindManual,
		Status:                sitequalitydomain.SiteQualityJobStatusProcessing,
		IdempotencyKey:        "expired-lease-job",
		SampleCount:           1,
		RequiredConfirmations: 1,
		MaxAttempts:           4,
		Attempts:              1,
		AvailableAt:           expiredAt,
		LockedAt:              &expiredAt,
		LockedBy:              "worker-1",
		LeaseGeneration:       1,
		LeaseExpiresAt:        &expiredAt,
		HeartbeatAt:           &expiredAt,
	}
	require.NoError(t, db.Create(&job).Error)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(siteQualityRunnerTestResponseWithCleanHeadings(t, `{
			"lighthouseResult": {
				"finalUrl": "https://example.com/products/wheel",
				"categories": {"performance": {"score": 1}},
				"audits": {
					"first-contentful-paint": {
						"id": "first-contentful-paint",
						"score": 1,
						"scoreDisplayMode": "metric",
						"numericValue": 900
					}
				}
			}
		}`)))
	}))
	t.Cleanup(server.Close)

	runner := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         server.URL,
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)
	runner.ConfigureHTTPClient(server.Client(), server.URL)
	runner.ConfigureJobRepository(repository.NewSiteQualityJobRepository(db))

	targetID := uint(1)
	jobID := job.ID
	result, err := runner.Capture(context.Background(), LighthouseRunnerCaptureInput{
		LighthouseRunnerRunInput: LighthouseRunnerRunInput{
			URL:      "https://example.com/products/wheel",
			Strategy: sitequalitydomain.SiteQualityStrategyMobile,
		},
		TargetID:        &targetID,
		JobID:           &jobID,
		LeaseWorkerID:   "worker-1",
		LeaseGeneration: 1,
		CanonicalURL:    "https://example.com/products/wheel",
	})
	require.ErrorIs(t, err, repository.ErrSiteQualityLeaseLost)
	require.Nil(t, result)

	list, listErr := runner.List(repository.SiteQualityRunListFilter{Page: 1, PageSize: 20})
	require.NoError(t, listErr)
	require.EqualValues(t, 0, list.Total)
}

func TestLighthouseRunnerServiceRejectsOffStorefrontTarget(t *testing.T) {
	db := newLighthouseRunnerTestDB(t)
	service := NewLighthouseRunnerService(
		repository.NewSiteQualityRunRepository(db),
		repository.NewSiteQualityFindingRepository(db),
		LighthouseRunnerConfig{
			RunnerURL:         "http://runner.internal",
			RunnerToken:       testLighthouseRunnerToken,
			StorefrontBaseURL: "https://example.com",
		},
	)

	result, err := service.Run(context.Background(), LighthouseRunnerRunInput{
		URL: "http://127.0.0.1:3000/",
	})
	require.ErrorIs(t, err, ErrInvalidSiteQualityRun)
	require.Nil(t, result)

	list, listErr := service.List(repository.SiteQualityRunListFilter{Page: 1, PageSize: 20})
	require.NoError(t, listErr)
	require.EqualValues(t, 0, list.Total)
}

func newLighthouseRunnerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&sitequalitydomain.SiteQualityTarget{},
		&sitequalitydomain.SiteQualityJob{},
		&sitequalitydomain.SiteQualityProviderSlot{},
		&sitequalitydomain.SiteQualityEvaluation{},
		&sitequalitydomain.SiteQualityRun{},
		&sitequalitydomain.SiteQualityFinding{},
		&sitequalitydomain.SiteQualityFindingEvent{},
	))
	return db
}

func createLeasedSiteQualityJob(
	t *testing.T,
	db *gorm.DB,
	targetURL string,
	strategy string,
	idempotencyKey string,
) *sitequalitydomain.SiteQualityJob {
	t.Helper()
	now := time.Now().UTC()
	target, err := repository.NewSiteQualityTargetRepository(db).Upsert(sitequalitydomain.SiteQualityTargetInput{
		CanonicalURL:            targetURL,
		Source:                  sitequalitydomain.SiteQualityTargetSourceOperator,
		SamplingTier:            sitequalitydomain.SiteQualityTargetTierStandard,
		SamplingIntervalSeconds: 604800,
		Enabled:                 true,
	}, now)
	require.NoError(t, err)

	leaseExpiresAt := now.Add(10 * time.Minute)
	job := sitequalitydomain.SiteQualityJob{
		TargetID:              target.ID,
		Strategy:              strategy,
		Kind:                  sitequalitydomain.SiteQualityJobKindManual,
		Status:                sitequalitydomain.SiteQualityJobStatusProcessing,
		IdempotencyKey:        idempotencyKey,
		SampleCount:           1,
		RequiredConfirmations: 1,
		Attempts:              1,
		MaxAttempts:           4,
		AvailableAt:           now,
		LockedAt:              &now,
		LockedBy:              "worker-lease-test",
		LeaseGeneration:       1,
		LeaseExpiresAt:        &leaseExpiresAt,
		HeartbeatAt:           &now,
	}
	require.NoError(t, db.Create(&job).Error)
	return &job
}

func siteQualityRunnerTestResponseWithCleanHeadings(t *testing.T, raw string) string {
	t.Helper()
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(raw), &payload))
	lighthouseResult, ok := payload["lighthouseResult"].(map[string]interface{})
	require.True(t, ok, "test response must include lighthouseResult")
	if _, exists := lighthouseResult["renderedHeadings"]; !exists {
		finalURL, _ := lighthouseResult["finalUrl"].(string)
		lighthouseResult["renderedHeadings"] = map[string]interface{}{
			"status":   "complete",
			"source":   "chrome-rendered-dom",
			"finalUrl": finalURL,
			"headings": []map[string]interface{}{
				{
					"level":    1,
					"text":     "Carbon Rims",
					"snippet":  "<h1>Carbon Rims</h1>",
					"selector": "main > h1",
				},
			},
		}
	}
	if _, exists := lighthouseResult["renderedStructuredData"]; !exists {
		finalURL, _ := lighthouseResult["finalUrl"].(string)
		lighthouseResult["renderedStructuredData"] = map[string]interface{}{
			"status":   "complete",
			"source":   "chrome-rendered-dom",
			"finalUrl": finalURL,
			"page": map[string]interface{}{
				"title":        "Carbon Rims",
				"canonicalUrl": finalURL,
			},
			"jsonLd": []map[string]interface{}{
				{
					"index":    0,
					"selector": "html > head > script:nth-of-type(1)",
					"raw":      `{"@context":"https://schema.org","@type":"Organization","name":"TANZANITE","url":"https://example.com/"}`,
					"nodes": []map[string]interface{}{
						{
							"types":     []string{"Organization"},
							"type":      "Organization",
							"name":      "TANZANITE",
							"url":       "https://example.com/",
							"graphPath": "script[0]",
							"data": map[string]interface{}{
								"@context": "https://schema.org",
								"@type":    "Organization",
								"name":     "TANZANITE",
								"url":      "https://example.com/",
							},
						},
					},
				},
			},
			"microdata": []map[string]interface{}{},
			"rdfa":      []map[string]interface{}{},
		}
	}
	encoded, err := json.Marshal(payload)
	require.NoError(t, err)
	return string(encoded)
}
