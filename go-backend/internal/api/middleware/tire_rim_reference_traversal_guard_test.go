package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func TestTireRimReferenceTraversalGuardRestoresBodyAndAllowsNormalLookups(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TireRimReferenceTraversalGuard())
	router.POST("/api/v1/engineering/tire-rim/solve", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil || !strings.Contains(string(body), `"tire_width_mm":40`) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "body was not restored"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/engineering/tire-rim/solve", strings.NewReader(`{"tire_width_mm":40,"rim_system":"hookless"}`))
	request.RemoteAddr = "198.51.100.10:1234"
	request.Header.Set("X-Device-Fingerprint", "normal-user")
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected body-restored lookup to succeed, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

func TestTireRimReferenceTraversalGuardBlocksAdjacentWidthSweep(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TireRimReferenceTraversalGuard())
	router.POST("/api/v1/engineering/tire-rim/solve", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for index := 0; index <= tireRimReferenceTraversalStreakLimit; index++ {
		width := 40 + index
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/engineering/tire-rim/solve",
			strings.NewReader(`{"tire_width_mm":`+strconv.Itoa(width)+`,"rim_system":"hookless"}`),
		)
		request.RemoteAddr = "198.51.100.11:1234"
		request.Header.Set("X-Device-Fingerprint", "sweeper")
		router.ServeHTTP(recorder, request)
		if index < tireRimReferenceTraversalStreakLimit && recorder.Code != http.StatusNoContent {
			t.Fatalf("expected adjacent lookup %d to pass, got %d", width, recorder.Code)
		}
		if index == tireRimReferenceTraversalStreakLimit {
			if recorder.Code != http.StatusTooManyRequests {
				t.Fatalf("expected adjacent sweep to be blocked, got %d: %s", recorder.Code, recorder.Body.String())
			}
			if recorder.Header().Get("Retry-After") != "60" {
				t.Fatalf("expected one-minute retry window, got %q", recorder.Header().Get("Retry-After"))
			}
		}
	}
}

func TestTireRimReferenceTraversalGuardDoesNotShareAdjacentSweepAcrossFingerprints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TireRimReferenceTraversalGuard())
	router.POST("/api/v1/engineering/tire-rim/solve", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for index := 0; index <= tireRimReferenceTraversalStreakLimit; index++ {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/engineering/tire-rim/solve",
			strings.NewReader(`{"tire_width_mm":`+strconv.Itoa(40+index)+`,"rim_system":"hookless"}`),
		)
		request.RemoteAddr = "198.51.100.12:1234"
		request.Header.Set("X-Device-Fingerprint", "fingerprint-"+strconv.Itoa(index))
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("expected distinct fingerprint lookup to pass, got %d", recorder.Code)
		}
	}
}

func TestTireRimReferenceTraversalGuardAllowsAlternatingWidthChecks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TireRimReferenceTraversalGuard())
	router.POST("/api/v1/engineering/tire-rim/solve", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	for index := 0; index <= tireRimReferenceTraversalStreakLimit*2; index++ {
		width := 40
		if index%2 == 1 {
			width = 41
		}
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/api/v1/engineering/tire-rim/solve",
			strings.NewReader(`{"tire_width_mm":`+strconv.Itoa(width)+`,"rim_system":"hookless"}`),
		)
		request.RemoteAddr = "198.51.100.13:1234"
		request.Header.Set("X-Device-Fingerprint", "manual-comparison")
		router.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusNoContent {
			t.Fatalf("expected alternating width check to pass, got %d: %s", recorder.Code, recorder.Body.String())
		}
	}
}

func TestTireRimReferenceTraversalTrackerUsesRedisForAdjacentSweepState(t *testing.T) {
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	tracker := newTireRimReferenceTraversalTracker(redisClient)

	for index := 0; index <= tireRimReferenceTraversalStreakLimit; index++ {
		streak := tracker.observe(nil, "redis-sweeper", 40+index, time.Now().UTC())
		if index == tireRimReferenceTraversalStreakLimit {
			if streak != tireRimReferenceTraversalStreakLimit {
				t.Fatalf("expected Redis-backed streak %d, got %d", tireRimReferenceTraversalStreakLimit, streak)
			}
			continue
		}
		if streak != index {
			t.Fatalf("expected Redis-backed streak %d, got %d", index, streak)
		}
	}
}
