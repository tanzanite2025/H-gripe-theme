package shipping

import "strings"

// YanwenTrackingStage is the five-stage operational view of an official
// Yanwen tracking status. Exception statuses are exposed separately by
// IsException and do not get a guessed stage.
type YanwenTrackingStage string

const (
	YanwenTrackingStageUnknown      YanwenTrackingStage = "UNKNOWN"
	YanwenTrackingStageCollected    YanwenTrackingStage = "COLLECTED"
	YanwenTrackingStageInTransitAir YanwenTrackingStage = "IN_TRANSIT_AIR"
	YanwenTrackingStageCustoms      YanwenTrackingStage = "CUSTOMS"
	YanwenTrackingStageLastMile     YanwenTrackingStage = "LAST_MILE"
	YanwenTrackingStageDelivered    YanwenTrackingStage = "DELIVERED"
)

// YanwenTrackingStageClassification is a local display/read-model result.
// It preserves the official status code separately and never becomes a
// generic tracking event or an order status.
type YanwenTrackingStageClassification struct {
	Stage       YanwenTrackingStage
	StageRank   int
	IsException bool
}

// ClassifyYanwenTrackingStatus maps a documented Yanwen status code to one
// of the five operational stages. Exception codes set IsException but keep
// the stage unknown because their grouped official codes do not identify a
// reliable base stage by themselves.
func ClassifyYanwenTrackingStatus(status string) YanwenTrackingStageClassification {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "OR10", "PU10":
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageCollected, StageRank: 1}
	case "LH20":
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageInTransitAir, StageRank: 2}
	case "S303", "IC50", "IC60":
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageCustoms, StageRank: 3}
	case "LM10", "LM20", "LM25":
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageLastMile, StageRank: 4}
	case "LM40":
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageDelivered, StageRank: 5}
	case "PU30", "EC30":
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageUnknown, IsException: true}
	case "IC51", "IC70", "LM50", "LM90":
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageUnknown, IsException: true}
	default:
		return YanwenTrackingStageClassification{Stage: YanwenTrackingStageUnknown}
	}
}

// ClassifyYanwenTrackingProgress uses the official current status and its
// official checkpoint statuses to build a local stage read model. A grouped
// exception code may use the highest known normal stage from its checkpoints;
// when no such fact exists, the result remains UNKNOWN.
func ClassifyYanwenTrackingProgress(currentStatus string, checkpointStatuses []string) YanwenTrackingStageClassification {
	current := ClassifyYanwenTrackingStatus(currentStatus)
	if !current.IsException || current.Stage != YanwenTrackingStageUnknown {
		return current
	}
	best := YanwenTrackingStageClassification{Stage: YanwenTrackingStageUnknown, IsException: true}
	for _, checkpointStatus := range checkpointStatuses {
		candidate := ClassifyYanwenTrackingStatus(checkpointStatus)
		if candidate.IsException || candidate.StageRank <= best.StageRank {
			continue
		}
		best.Stage = candidate.Stage
		best.StageRank = candidate.StageRank
	}
	return best
}
