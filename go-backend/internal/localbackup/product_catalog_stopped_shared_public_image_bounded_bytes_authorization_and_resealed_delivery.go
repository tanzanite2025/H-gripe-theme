package localbackup

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
)

const stoppedProductPublicImageMaximumBytes = 4 << 20
const stoppedProductPublicImageVerification = "registered_shared_public_image_bytes_bound_read_only_sql_and_resealed_before_delivery"

type StoppedProductReadOnlyPublicImageHTTPRequest struct {
	ContractSHA256 string `json:"contract_sha256"`
	ObjectKey      string `json:"object_key"`
}

type StoppedProductReadOnlyPublicImageHTTPResult struct {
	*StoppedProductReadOnlyMediaHTTPResult
	Request             StoppedProductReadOnlyPublicImageHTTPRequest `json:"request"`
	ResponseBytesBase64 string                                       `json:"response_bytes_base64"`
	ObjectOwner         string                                       `json:"object_owner,omitempty"`
}

// This separate public-only contract cannot accept business cookies or private
// attachment selectors. Bytes leave the manager only after the shared helper
// has rolled back, resealed SQL and verified the complete stopped baseline.
func (manager *Manager) ExecuteStoppedProductReadOnlyPublicImageHTTPQuery(ctx context.Context, id string, request StoppedProductReadOnlyPublicImageHTTPRequest) (*StoppedProductReadOnlyPublicImageHTTPResult, error) {
	result, err := manager.executeStoppedProductMediaHTTPQueryWithOptionalPublicImageCapture(ctx, id, StoppedProductReadOnlyMediaHTTPRequest{
		ContractSHA256: request.ContractSHA256, Operation: "public_upload", ObjectKey: request.ObjectKey,
	}, true)
	if err != nil {
		return nil, err
	}
	if result.MatchesPackageFile != nil && !*result.MatchesPackageFile {
		return nil, errors.New("public image bytes differ from their package identity; delivery refused")
	}
	result.Version = 1
	result.Verification = stoppedProductPublicImageVerification
	owner := stoppedProductPublicSiteIdentityImageOwnerForObjectKey(request.ObjectKey)
	if owner != "" {
		result.Verification = stoppedProductSiteIdentityImageVerification
	}
	result.RemainingChecks = []string{"other_dedicated_owner_images_svg_avif_and_large_images", "nuxt_ipx_and_actual_ssr_client_resource_requests", "actual_proxy_cdn_browser_and_final_release"}
	return &StoppedProductReadOnlyPublicImageHTTPResult{StoppedProductReadOnlyMediaHTTPResult: result, Request: request, ResponseBytesBase64: base64.StdEncoding.EncodeToString(result.capturedPublicImageBytes), ObjectOwner: owner}, nil
}

func validateStoppedProductCapturedPublicImageResponse(status int, headers map[string]string, body []byte) error {
	if status == http.StatusNotFound {
		// The genuine authorization/file handler's bounded denial body is retained
		// with its digest, but the bridge never treats it as an image.
		return nil
	}
	if status != http.StatusOK || len(body) == 0 || len(body) > stoppedProductPublicImageMaximumBytes {
		return errors.New("only bounded complete public images or genuine not-found responses are supported")
	}
	mimeType := headers["Content-Type"]
	switch mimeType {
	case "image/png", "image/jpeg", "image/gif", "image/webp":
		if http.DetectContentType(body) == mimeType {
			return nil
		}
	}
	return errors.New("public image MIME type or byte signature is unsupported; delivery refused")
}
