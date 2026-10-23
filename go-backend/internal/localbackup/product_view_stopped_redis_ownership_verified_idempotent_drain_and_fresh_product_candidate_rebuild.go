package localbackup

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Private operator configuration. No endpoint accepts addresses or credentials.
type ProductViewRecoveryRedisConnectionConfiguration struct {
	Mode             string   `json:"mode"`
	Addresses        []string `json:"addresses"`
	Database         int      `json:"database"`
	MasterName       string   `json:"master_name,omitempty"`
	Username         string   `json:"username,omitempty"`
	Password         string   `json:"password,omitempty"`
	SentinelUsername string   `json:"sentinel_username,omitempty"`
	SentinelPassword string   `json:"sentinel_password,omitempty"`
}

type ProductViewStoppedDrainAndCandidateRebuildRequest struct {
	Confirmation                                    string `json:"confirmation"`
	AllListedWritersStoppedAndUseThisSingleHostGate bool   `json:"all_listed_writers_stopped_and_use_this_single_host_gate"`
}

// Durable retry intent, not a production handoff or permission to resume after
// a switch. A new maintenance window must obtain its own stopped verification.
type ProductViewStoppedDrainAndCandidateRebuildProgress struct {
	Version                int                                                 `json:"version"`
	TaskID                 string                                              `json:"task_id"`
	PackageID              string                                              `json:"package_id"`
	ConfigurationSHA256    string                                              `json:"configuration_sha256"`
	OwnershipSHA256        string                                              `json:"ownership_sha256"`
	WriterInventorySHA256  string                                              `json:"writer_inventory_sha256"`
	Operator               string                                              `json:"operator"`
	TargetDatabaseOID      int64                                               `json:"target_database_oid"`
	Phase                  string                                              `json:"phase"`
	StartedAt              time.Time                                           `json:"started_at"`
	StoppedLeaseVerifiedAt time.Time                                           `json:"stopped_lease_verified_at,omitempty"`
	FlushCalls             int                                                 `json:"flush_calls"`
	Before                 *service.ProductViewCountRecoveryBoundaryInspection `json:"before,omitempty"`
	After                  *service.ProductViewCountRecoveryBoundaryInspection `json:"after,omitempty"`
	RebuiltPreviewSHA256   string                                              `json:"rebuilt_preview_sha256,omitempty"`
	SHA256                 string                                              `json:"sha256"`
}

func (configuration ProductViewRecoveryRedisConnectionConfiguration) validateExplicitProductViewRecoveryRedisConnection() error {
	if (configuration.Mode != "standalone" && configuration.Mode != "sentinel") || len(configuration.Addresses) == 0 || configuration.Database < 0 || (configuration.Mode == "standalone" && (len(configuration.Addresses) != 1 || configuration.MasterName != "")) || (configuration.Mode == "sentinel" && strings.TrimSpace(configuration.MasterName) == "") {
		return errors.New("product view recovery requires explicit standalone or Sentinel Redis addresses and logical database")
	}
	for _, address := range configuration.Addresses {
		if strings.TrimSpace(address) != address || address == "" || strings.ContainsAny(address, "/@\r\n") {
			return errors.New("Redis recovery addresses must contain host and port only")
		}
	}
	return nil
}

func (manager *Manager) saveProductViewStoppedDrainProgress(task *Task) error {
	progress := task.ProductViewStoppedDrain
	progress.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(progress)
	if err != nil {
		return err
	}
	progress.SHA256 = digest
	return manager.saveTask(task)
}

func (manager *Manager) validateProductViewStoppedDrainProgress(task *Task) error {
	progress := task.ProductViewStoppedDrain
	if progress == nil {
		return nil
	}
	copy := *progress
	copy.SHA256 = ""
	digest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(copy)
	configurationDigest, configurationErr := customerServiceRecoveryPlanCanonicalJSONSHA256(manager.Configuration)
	if err != nil || configurationErr != nil || digest != progress.SHA256 || progress.Version != 1 || progress.TaskID != task.ID || task.Manifest == nil || progress.PackageID != task.Manifest.PackageID || progress.ConfigurationSHA256 != configurationDigest || progress.Operator == "" || progress.TargetDatabaseOID <= 0 {
		return errors.New("product view drain progress binding changed; retain maintenance for inspection")
	}
	switch progress.Phase {
	case "waiting_for_stop", "draining", "buffers_drained", "candidate_rebuilt":
	default:
		return errors.New("product view drain progress phase is invalid")
	}
	if progress.Phase == "candidate_rebuilt" && (progress.After == nil || !progress.After.RedisBuffersEmpty || progress.RebuiltPreviewSHA256 == "") {
		return errors.New("product view drain completion has no verified fresh candidate")
	}
	return nil
}

func (manager *Manager) StartStoppedProductViewDrainAndFreshCandidateRebuild(id, actor string, request ProductViewStoppedDrainAndCandidateRebuildRequest) (*Task, error) {
	task, err := manager.GetTask(id)
	if err != nil {
		return nil, err
	}
	if taskHasUnresolvedProductCatalogStoppedHandoffPreparation(task) {
		return nil, errors.New("resume or cancel the stopped product handoff preparation instead")
	}
	cfg := manager.Configuration
	if request.Confirmation != "DRAIN PRODUCT VIEWS "+cfg.SiteID+"/"+cfg.Environment+"/"+id || !request.AllListedWritersStoppedAndUseThisSingleHostGate || strings.TrimSpace(actor) == "" {
		return nil, errors.New("explicit single-host writer inventory attestation and stopped drain confirmation are required")
	}
	if cfg.ProductViewRecoveryRedis == nil || cfg.ProductViewRecoveryRedis.validateExplicitProductViewRecoveryRedisConnection() != nil || len(cfg.ProductViewRecoveryWriterInventory) == 0 {
		return nil, errors.New("configure the private Redis connection and complete single-host writer inventory first")
	}
	for _, writer := range cfg.ProductViewRecoveryWriterInventory {
		if strings.TrimSpace(writer) == "" {
			return nil, errors.New("writer inventory has an empty entry")
		}
	}
	if task.ProductViewStoppedDrain == nil || task.ProductViewStoppedDrain.Phase == "candidate_rebuilt" {
		if err := validateFreshProductCatalogCandidateTaskForHandoffInspection(task); err != nil {
			return nil, err
		}
	} else if task.State != "failed" && task.State != "interrupted" {
		return nil, errors.New("only an interrupted or failed stopped drain can be resumed")
	}
	unresolvedDrain := task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt"
	// Expiry cannot strand a partially committed drain. Its retained package can
	// finish this window, but the expired result still cannot authorize handoff.
	if task.Journal != nil || task.WorkCopyExpired || (!unresolvedDrain && !time.Now().Before(task.ExpiresAt)) {
		return nil, errors.New("live switch or expired product candidate cannot start a stopped drain")
	}
	if err := manager.validateProductViewStoppedDrainProgress(task); err != nil {
		return nil, err
	}
	if err := cfg.validateRuntimeBoundary(); err != nil {
		return nil, err
	}
	if err := manager.startTaskOperation(task, "waiting_for_product_view_writers_to_stop", func(ctx context.Context, task *Task) error {
		return manager.drainProductViewsAndRebuildCandidateUnderOneMaintenanceLease(ctx, task, actor)
	}); err != nil {
		return nil, err
	}
	return manager.GetTask(id)
}

func (manager *Manager) connectAndVerifyRuntimeOwnedProductViewRedis(ctx context.Context) (redis.UniversalClient, *maintenancegate.ProductViewRedisRuntimeOwnership, error) {
	cfg := manager.Configuration
	connection := cfg.ProductViewRecoveryRedis
	if connection == nil || connection.validateExplicitProductViewRecoveryRedisConnection() != nil {
		return nil, nil, errors.New("explicit private product view Redis recovery configuration is required")
	}
	data, err := os.ReadFile(filepath.Join(cfg.MaintenanceDirectory, "application-boundary.json"))
	var boundary maintenancegate.ApplicationBoundary
	expectedProjection := service.DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(true)
	if err != nil || decodeStrictJSON(data, &boundary) != nil || boundary.ValidateRegisteredRecoveryProjectionProtocol() != nil || boundary.ExternalProjections == nil || !reflect.DeepEqual(boundary.ExternalProjections.ProductCatalog, &expectedProjection) {
		return nil, nil, errors.New("business runtime must register this exact immutable product view SQL receipt contract")
	}
	ownership, err := maintenancegate.ReadRegisteredProductViewRedisRuntimeOwnership(cfg.MaintenanceDirectory)
	destination := maintenancegate.FingerprintRedisRecoveryDestination(connection.Mode, connection.Addresses, connection.Database, connection.MasterName)
	if err != nil || ownership.SiteID != cfg.SiteID || ownership.Environment != cfg.Environment || ownership.Database != cfg.DatabaseName || ownership.DestinationSHA256 != destination || boundary.ExternalProjections.RedisDestinationSHA256 != destination || ownership.KeysSHA256 != expectedProjection.ViewCountKeysSHA256 {
		return nil, nil, errors.New("private Redis destination, site or logical database differs from runtime ownership")
	}
	// One connection lets CLIENT LIST distinguish the executor's own socket.
	client := redis.NewUniversalClient(&redis.UniversalOptions{Addrs: connection.Addresses, DB: connection.Database, Username: connection.Username, Password: connection.Password, MasterName: connection.MasterName, SentinelUsername: connection.SentinelUsername, SentinelPassword: connection.SentinelPassword, PoolSize: 1, MaxActiveConns: 1, ContextTimeoutEnabled: true, DialTimeout: 5 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second})
	if err := maintenancegate.VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, client, ownership); err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, ownership, nil
}

func requireNoOtherProductViewRedisLogicalDatabaseClients(ctx context.Context, client redis.UniversalClient, database int) error {
	currentID, err := client.Do(ctx, "CLIENT", "ID").Int64()
	if err != nil {
		return errors.New("cannot verify stopped Redis clients")
	}
	clients, err := client.Do(ctx, "CLIENT", "LIST").Text()
	if err != nil {
		return errors.New("Redis CLIENT LIST permission is required for stopped verification")
	}
	for _, line := range strings.Split(strings.TrimSpace(clients), "\n") {
		fields := map[string]string{}
		for _, field := range strings.Fields(line) {
			key, value, ok := strings.Cut(field, "=")
			if ok {
				fields[key] = value
			}
		}
		if fields["id"] == "" || fields["db"] == "" {
			return errors.New("Redis client inspection shape is invalid")
		}
		// Replication sockets do not issue application writes. Include subscribed
		// clients: RESP3 permits ordinary commands on a subscribed connection.
		replicationSocket := strings.ContainsAny(fields["flags"], "SM")
		if !replicationSocket && fields["db"] == strconv.Itoa(database) && fields["id"] != strconv.FormatInt(currentID, 10) {
			return errors.New("other Redis clients still use the target logical database; stop all writers and readers")
		}
	}
	return nil
}

func (manager *Manager) drainProductViewsAndRebuildCandidateUnderOneMaintenanceLease(ctx context.Context, task *Task, actor string) (resultErr error) {
	lease, err := manager.acquireMaintenance(ctx, task)
	if err != nil {
		return err
	}
	defer lease.Close()
	if err := manager.drainProductViewsAndRebuildCandidateWithHeldMaintenanceLease(ctx, task, actor); err != nil {
		return err
	}
	return manager.releaseMaintenance(task)
}

// The caller owns the exclusive application lease and the persistent fence.
// Completion never releases either, so a handoff can continue in this interval.
func (manager *Manager) drainProductViewsAndRebuildCandidateWithHeldMaintenanceLease(ctx context.Context, task *Task, actor string) (resultErr error) {
	defer func() {
		if resultErr != nil && task.ProductViewStoppedDrain != nil && task.ProductViewStoppedDrain.Phase != "candidate_rebuilt" {
			task.ProductCatalogRecovery = nil
		}
	}()
	cfg := manager.Configuration
	if err := manager.validateProductViewStoppedDrainProgress(task); err != nil {
		return err
	}
	if err := manager.verifyPreparedRecoveryFiles(task); err != nil {
		return err
	}
	extracted := filepath.Join(cfg.Workspace, task.ID, "verified")
	snapshot, err := readAndValidateProductCatalogModuleSnapshot(extracted, *task.Manifest)
	if err != nil {
		return err
	}
	if err := validateProductCatalogPackageFileInventory(snapshot, *task.Manifest, cfg); err != nil {
		return err
	}
	// From this point all errors retain the persistent fence. The operator may
	// retry the durable intent; neither TTL nor "release unmodified" can bypass it.
	client, ownership, err := manager.connectAndVerifyRuntimeOwnedProductViewRedis(ctx)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := requireNoOtherProductViewRedisLogicalDatabaseClients(ctx, client, cfg.ProductViewRecoveryRedis.Database); err != nil {
		return err
	}
	control, err := cfg.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	if err := assertNoUnfencedDatabaseSessions(ctx, control, cfg.DatabaseName); err != nil {
		return err
	}
	current, err := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	configurationDigest, _ := customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
	ownershipDigest := maintenancegate.FingerprintExternalProjectionConfiguration(ownership)
	inventoryDigest := maintenancegate.FingerprintExternalProjectionConfiguration(cfg.ProductViewRecoveryWriterInventory)
	if task.ProductViewStoppedDrain == nil || task.ProductViewStoppedDrain.Phase == "candidate_rebuilt" {
		// Current rows may have advanced since preview: this operation explicitly
		// rebuilds from them. Still require the original physical target, schema,
		// isolated candidate and exact executor configuration before dropping it.
		binding := task.ProductCatalogPreviewBinding
		candidate, candidateErr := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, task.StageDatabase)
		if binding == nil || binding.ConfigurationSHA256 != configurationDigest || binding.Target.OID != current.OID || binding.Target.Owner != current.Owner || binding.Target.SchemaSHA256 != current.SchemaSHA256 || candidateErr != nil || *candidate != binding.Candidate {
			return errors.New("product drain requires the original bound target schema and untouched isolated candidate")
		}
		task.ProductViewStoppedDrain = &ProductViewStoppedDrainAndCandidateRebuildProgress{Version: 1, TaskID: task.ID, PackageID: task.Manifest.PackageID, ConfigurationSHA256: configurationDigest, OwnershipSHA256: ownershipDigest, WriterInventorySHA256: inventoryDigest, Operator: actor, TargetDatabaseOID: current.OID, Phase: "waiting_for_stop", StartedAt: time.Now().UTC()}
	} else if task.ProductViewStoppedDrain.OwnershipSHA256 != ownershipDigest || task.ProductViewStoppedDrain.WriterInventorySHA256 != inventoryDigest || task.ProductViewStoppedDrain.TargetDatabaseOID != current.OID {
		return errors.New("stopped drain Redis ownership, writer inventory or target database changed")
	}
	progress := task.ProductViewStoppedDrain
	progress.StoppedLeaseVerifiedAt = time.Now().UTC()
	progress.Phase, task.Phase = "draining", "draining_idempotent_product_view_batches"
	// Old UI reports become ineligible before the first Redis claim or SQL write.
	task.ProductCatalogRecovery = nil
	if err := manager.saveProductViewStoppedDrainProgress(task); err != nil {
		return err
	}
	sqlDB, err := cfg.connectDatabase(ctx, cfg.DatabaseName)
	if err != nil {
		return err
	}
	database, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), DisableAutomaticPing: true})
	if err != nil {
		sqlDB.Close()
		return errors.New("cannot construct product view receipt repository")
	}
	buffer := service.NewProductViewCountBuffer(client, repository.NewProductRepository(database), 500)
	err = func() error {
		defer sqlDB.Close()
		before, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(ctx)
		if err != nil {
			return errors.New("product view buffer or SQL receipt validation failed; preserve state for reconciliation")
		}
		if progress.Before == nil {
			progress.Before = before
			if err := manager.saveProductViewStoppedDrainProgress(task); err != nil {
				return err
			}
		}
		for iteration := 0; iteration < 32; iteration++ {
			if err := maintenancegate.VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, client, ownership); err != nil {
				return err
			}
			if err := requireNoOtherProductViewRedisLogicalDatabaseClients(ctx, client, cfg.ProductViewRecoveryRedis.Database); err != nil {
				return err
			}
			if err := assertNoUnfencedDatabaseSessions(ctx, control, cfg.DatabaseName); err != nil {
				return err
			}
			observation, err := buffer.InspectReadOnlyProductViewRecoveryBoundary(ctx)
			if err != nil {
				return errors.New("product view stopped boundary changed or is invalid; preserve original batch")
			}
			if observation.RedisBuffersEmpty {
				progress.After = observation
				return nil
			}
			if _, err := buffer.Flush(ctx); err != nil {
				return errors.New("product view drain failed; original batch and SQL receipts retained for idempotent retry")
			}
			progress.FlushCalls++
			if err := manager.saveProductViewStoppedDrainProgress(task); err != nil {
				return err
			}
		}
		return errors.New("stopped product view drain exceeded capacity; preserve state and inspect remaining writers")
	}()
	if err != nil {
		return err
	}
	progress.Phase, task.Phase = "buffers_drained", "rebuilding_product_candidate_after_view_drain"
	if err := manager.saveProductViewStoppedDrainProgress(task); err != nil {
		return err
	}
	if err := manager.restoreProductCatalogIntoIsolatedTargetCloneWithHeldMaintenanceLease(ctx, task, extracted, snapshot); err != nil {
		return err
	}
	if err := maintenancegate.VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, client, ownership); err != nil {
		return err
	}
	if err := requireNoOtherProductViewRedisLogicalDatabaseClients(ctx, client, cfg.ProductViewRecoveryRedis.Database); err != nil {
		return err
	}
	// A new active or processing key invalidates this stopped interval even when
	// ownership is unchanged; never clear it or silently extend the window.
	remaining, err := client.Exists(ctx, "product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id").Result()
	if err != nil || remaining != 0 {
		return errors.New("product view buffers changed during candidate rebuild; retain maintenance and retry")
	}
	progress.RebuiltPreviewSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(task.ProductCatalogPreviewBinding)
	if err != nil {
		return err
	}
	progress.Phase = "candidate_rebuilt"
	task.Phase = "product_view_drain_and_fresh_candidate_rebuild_verified"
	if err := manager.saveProductViewStoppedDrainProgress(task); err != nil {
		return err
	}
	return nil
}
