package wheelsetlacing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func TestListWheelsetLacingTopologiesReturnsCacheableContract(t *testing.T) {
	router := newTestRouter()
	first := performRequest(router, http.MethodGet, "/api/v1/wheelset-lacing/topologies", "")
	if first.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", first.Code, http.StatusOK, first.Body.String())
	}
	if first.Header().Get("ETag") == "" || first.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Fatalf("missing cache headers: etag=%q cache=%q", first.Header().Get("ETag"), first.Header().Get("Cache-Control"))
	}
	for _, fragment := range []string{`"topologies"`, `"21h-g3-2to1"`, `"18h-uniform-2to1"`, `"uniform_18h_2to1"`, `"display_layout":"g3_21h_triplet_2to1"`, `"contract_version":"v1.4"`, `"length_unit":"mm"`} {
		if !strings.Contains(first.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, first.Body.String())
		}
	}
	if strings.Contains(first.Body.String(), `"spoke_length"`) || strings.Contains(first.Body.String(), `"tension"`) {
		t.Fatalf("topology contract must not expose physical calculations: %s", first.Body.String())
	}

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/wheelset-lacing/topologies", nil)
	request.Header.Set("If-None-Match", first.Header().Get("ETag"))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNotModified {
		t.Fatalf("cached status = %d, want %d", response.Code, http.StatusNotModified)
	}
}

func TestValidateReturnsG3Mapping(t *testing.T) {
	router := newTestRouter()
	response := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"21h-g3-2to1","hole_count":21,"cross":2,"distribution":"g3_2to1"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	for _, fragment := range []string{`"valid":true`, `"hole_count":21`, `"hub_holes_a"`, `"hub_holes_b"`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, response.Body.String())
		}
	}
}

func TestValidateFailsLoudlyForUnknownAndMismatchedSelections(t *testing.T) {
	router := newTestRouter()
	unknown := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"24h-g3-2to1"}`)
	if unknown.Code != http.StatusNotFound || !strings.Contains(unknown.Body.String(), "UNKNOWN_WHEELSET_LACING_TOPOLOGY") {
		t.Fatalf("unknown response = %d %s", unknown.Code, unknown.Body.String())
	}
	mismatch := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"24h-uniform-2to1","cross":3}`)
	if mismatch.Code != http.StatusUnprocessableEntity || !strings.Contains(mismatch.Body.String(), "TOPOLOGY_SELECTION_MISMATCH") {
		t.Fatalf("mismatch response = %d %s", mismatch.Code, mismatch.Body.String())
	}
	unknownField := performRequest(router, http.MethodPost, "/api/v1/wheelset-lacing/validate", `{"topology_id":"21h-g3-2to1","erd":600}`)
	if unknownField.Code != http.StatusBadRequest || !strings.Contains(unknownField.Body.String(), "INVALID_WHEELSET_LACING_REQUEST") {
		t.Fatalf("unknown field response = %d %s", unknownField.Code, unknownField.Body.String())
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationReturnsBackendCoordinatesAndMetrics(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"24h-symmetric-1to1-2x","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":20,"flange_offset_b_mm":35}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	for _, fragment := range []string{`"contract_version":"v1.15-backend-display-geometry"`, `"display_layout":"symmetric_1to1"`, `"spoke_head_style":"j_bend"`, `"rim_holes"`, `"spokes"`, `"flange_profile"`, `"aggregate_mean_absolute_projection_angle_degrees"`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, response.Body.String())
		}
	}
	if strings.Contains(response.Body.String(), `"straight_pull_projection"`) {
		t.Fatalf("J-bend geometry must not expose the straight-pull projection: %s", response.Body.String())
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationReturnsUniform18HTwoToOneGeometry(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"18h-uniform-2to1","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":20,"flange_offset_b_mm":35}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	for _, fragment := range []string{`"contract_version":"v1.15-backend-display-geometry"`, `"display_layout":"uniform_18h_2to1"`, `"spoke_head_style":"j_bend"`, `"hole_count":18`, `"drive_side_spoke_count":12`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, response.Body.String())
		}
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationSharesStraightPullHubHoleAnchors(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"24h-symmetric-1to1-2x-straight-pull","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":20,"flange_offset_b_mm":35}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	var body struct {
		Data struct {
			Topology struct {
				ID             string `json:"topology_id"`
				SpokeHeadStyle string `json:"spoke_head_style"`
				HubHolesA      []any  `json:"hub_holes_a"`
				HubHolesB      []any  `json:"hub_holes_b"`
			} `json:"topology"`
			StraightPullProjection struct {
				Spokes []struct {
					Side string `json:"side"`
					Hub  struct {
						ID int     `json:"id"`
						X  float64 `json:"x"`
						Y  float64 `json:"y"`
					} `json:"hub"`
				} `json:"spokes"`
			} `json:"straight_pull_projection"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode straight-pull geometry response: %v; body=%s", err, response.Body.String())
	}
	if body.Data.Topology.ID != "24h-symmetric-1to1-2x-straight-pull" || body.Data.Topology.SpokeHeadStyle != "straight_pull" {
		t.Fatalf("unexpected straight-pull topology identity: %+v", body.Data.Topology)
	}
	if len(body.Data.Topology.HubHolesA) != 6 || len(body.Data.Topology.HubHolesB) != 6 || len(body.Data.StraightPullProjection.Spokes) != 24 {
		t.Fatalf("straight-pull topology must expose 6 paired holes per flange and 24 spokes: A=%d B=%d spokes=%d",
			len(body.Data.Topology.HubHolesA), len(body.Data.Topology.HubHolesB), len(body.Data.StraightPullProjection.Spokes))
	}
	type sharedHubHoleProjection struct {
		spokeCount int
		x          float64
		y          float64
	}
	sharedHubHoles := make(map[string]sharedHubHoleProjection)
	for _, spoke := range body.Data.StraightPullProjection.Spokes {
		key := fmt.Sprintf("%s:%d", spoke.Side, spoke.Hub.ID)
		projection := sharedHubHoles[key]
		if projection.spokeCount > 0 && (projection.x != spoke.Hub.X || projection.y != spoke.Hub.Y) {
			t.Fatalf("straight-pull spokes from hole %s do not share one projected point", key)
		}
		projection.spokeCount++
		projection.x = spoke.Hub.X
		projection.y = spoke.Hub.Y
		sharedHubHoles[key] = projection
	}
	for key, projection := range sharedHubHoles {
		if projection.spokeCount != 2 {
			t.Fatalf("straight-pull hub hole %s is connected to %d spokes, want 2", key, projection.spokeCount)
		}
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationDerivesG3FlangeSizeFromParallelHoleSpacing(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"21h-g3-2to1","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":20,"flange_offset_b_mm":35,"g3_parallel_hole_spacing_mm":40}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	for _, fragment := range []string{`"spoke_head_style":"straight_pull"`, `"straight_pull_projection":{`, `"parallel_hole_spacing_mm":40`, `"side_a_flange_hole_circle_radius_mm":46.1`, `"side_a_flange_pcd_mm":92.19`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, response.Body.String())
		}
	}
	if strings.Contains(response.Body.String(), `"spoke_over_id"`) {
		t.Fatalf("straight-pull G3 response must not expose J-bend over-under crossings: %s", response.Body.String())
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationRejectsG3FlangeOutsideRimHoleCircle(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"21h-g3-2to1","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":20,"flange_offset_b_mm":35,"g3_parallel_hole_spacing_mm":210}`)
	if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), "WHEELSET_LACING_GEOMETRY_CALCULATION_FAILED") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationReturnsG3GroupSpacing(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"21h-g3-2to1","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":20,"flange_offset_b_mm":35,"g3_rim_hole_spacing_a_to_b_degrees":2.5,"g3_rim_hole_spacing_b_to_a_degrees":8.5,"g3_rim_hole_spacing_a_to_next_group_a_degrees":40.428571}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	for _, fragment := range []string{`"g3_group_spacing"`, `"spacing_a_to_b_degrees":2.5`, `"spacing_b_to_a_degrees":8.5`, `"spacing_a_to_next_group_a_degrees":40.43`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, response.Body.String())
		}
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationRejectsOutOfRangeG3GroupSpacing(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"21h-g3-2to1","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":20,"flange_offset_b_mm":35,"g3_rim_hole_spacing_a_to_b_degrees":52,"g3_rim_hole_spacing_b_to_a_degrees":4.87}`)
	if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), "WHEELSET_LACING_GEOMETRY_CALCULATION_FAILED") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationRejectsUnknownFields(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"24h-symmetric-1to1-2x","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"erd":622}`)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "INVALID_WHEELSET_LACING_GEOMETRY_REQUEST") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestHandleWheelsetLacingDisplayGeometryCalculationRejectsOutOfRangeFlangeOffset(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodPost, "/api/v1/wheelset-lacing/display-geometry", `{"topology_id":"24h-symmetric-1to1-2x","rim_radius":232,"flange_radius_a":66,"flange_radius_b":54,"flange_offset_a_mm":-1,"flange_offset_b_mm":35}`)
	if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), "WHEELSET_LACING_GEOMETRY_CALCULATION_FAILED") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestGetWheelsetLacingDisplayGeometryProjectionForServerRenderedReferencePageUsesCanonicalDisplayCoordinates(t *testing.T) {
	response := performRequest(newTestRouter(), http.MethodGet, "/api/v1/wheelset-lacing/display-geometry?topology_id=24h-symmetric-1to1-2x", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusOK, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Fatalf("cache header = %q", response.Header().Get("Cache-Control"))
	}
	for _, fragment := range []string{`"contract_version":"v1.15-backend-display-geometry"`, `"display_layout":"symmetric_1to1"`, `"spoke_head_style":"j_bend"`, `"rim_holes"`, `"spokes"`, `"flange_profile"`, `"metrics"`} {
		if !strings.Contains(response.Body.String(), fragment) {
			t.Fatalf("body missing %s: %s", fragment, response.Body.String())
		}
	}
}

func TestGetWheelsetLacingDisplayGeometryProjectionForServerRenderedReferencePageRejectsMissingOrUnknownQueryParameters(t *testing.T) {
	missing := performRequest(newTestRouter(), http.MethodGet, "/api/v1/wheelset-lacing/display-geometry", "")
	if missing.Code != http.StatusBadRequest || !strings.Contains(missing.Body.String(), "INVALID_WHEELSET_LACING_GEOMETRY_QUERY") {
		t.Fatalf("missing query response = %d %s", missing.Code, missing.Body.String())
	}
	unknown := performRequest(newTestRouter(), http.MethodGet, "/api/v1/wheelset-lacing/display-geometry?topology_id=24h-symmetric-1to1-2x&erd=622", "")
	if unknown.Code != http.StatusBadRequest || !strings.Contains(unknown.Body.String(), "INVALID_WHEELSET_LACING_GEOMETRY_QUERY") {
		t.Fatalf("unknown query response = %d %s", unknown.Code, unknown.Body.String())
	}
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewWheelsetLacingEngineeringHTTPHandler(service.NewWheelsetLacingService())
	handler.RegisterWheelsetLacingEngineeringRoutes(router.Group("/api/v1/wheelset-lacing"))
	return router
}

func performRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
