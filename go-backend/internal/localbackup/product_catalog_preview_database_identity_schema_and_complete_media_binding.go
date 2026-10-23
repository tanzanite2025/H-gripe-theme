package localbackup

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/lib/pq"
)

type productCatalogRecoveryDatabaseBinding struct {
	OID                    int64  `json:"oid"`
	Owner                  string `json:"owner"`
	Fingerprint            string `json:"fingerprint"`
	SchemaSHA256           string `json:"schema_sha256"`
	ColumnConstraintSHA256 string `json:"column_constraint_sha256"`
	IndexSHA256            string `json:"index_sha256"`
	WriteSideEffectsSHA256 string `json:"write_side_effects_sha256"`
}

type productCatalogRecoveryPreviewBinding struct {
	Version                        int                                           `json:"version"`
	ConfigurationSHA256            string                                        `json:"configuration_sha256"`
	Target                         productCatalogRecoveryDatabaseBinding         `json:"target"`
	Candidate                      productCatalogRecoveryDatabaseBinding         `json:"candidate"`
	TargetMedia                    *moduleRecoveryCompleteMediaInventory         `json:"target_media"`
	CandidateMedia                 *moduleRecoveryCompleteMediaInventory         `json:"candidate_media"`
	RequiredImageCleanupProtection *productCatalogRequiredImageCleanupProtection `json:"required_image_cleanup_protection,omitempty"`
}

func (cfg Configuration) captureProductCatalogRecoveryDatabaseBinding(ctx context.Context, name string) (*productCatalogRecoveryDatabaseBinding, error) {
	db, err := cfg.connectDatabase(ctx, name)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	result := &productCatalogRecoveryDatabaseBinding{}
	if err := db.QueryRowContext(ctx, `SELECT oid::bigint,pg_get_userbyid(datdba) FROM pg_database WHERE datname=current_database()`).Scan(&result.OID, &result.Owner); err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	observation, err := inspectModuleRecoveryDatabaseObservation(ctx, tx)
	if err != nil {
		return nil, err
	}
	if observation.Migration != cfg.ExpectedMigrationVersion {
		return nil, errors.New("product recovery migration changed")
	}
	result.Fingerprint = observation.Fingerprint
	result.ColumnConstraintSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(observation.Shapes)
	if err != nil {
		return nil, err
	}
	result.IndexSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(observation.Indexes)
	if err != nil {
		return nil, err
	}
	result.WriteSideEffectsSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(observation.WriteSideEffects)
	if err != nil {
		return nil, err
	}
	result.SchemaSHA256, err = customerServiceRecoveryPlanCanonicalJSONSHA256(struct {
		Shapes           any
		Indexes          any
		WriteSideEffects any
	}{observation.Shapes, observation.Indexes, observation.WriteSideEffects})
	return result, err
}

func (manager *Manager) productCatalogCandidateMediaConfiguration(task *Task) Configuration {
	cfg := manager.Configuration
	directory, _ := manager.taskDirectory(task.ID)
	cfg.StorageRoots = map[string]string{}
	for _, name := range mediaRoots {
		cfg.StorageRoots[name] = filepath.Join(directory, "product-isolated-target", "media", name)
	}
	return cfg
}

func (manager *Manager) captureProductCatalogRecoveryPreviewBaseline(ctx context.Context) (*productCatalogRecoveryPreviewBinding, error) {
	cfg := manager.Configuration
	configurationDigest, err := customerServiceRecoveryPlanCanonicalJSONSHA256(cfg)
	if err != nil {
		return nil, err
	}
	target, err := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, cfg.DatabaseName)
	if err != nil {
		return nil, err
	}
	media, err := cfg.inventoryModuleRecoveryCompleteMedia(ctx)
	if err != nil {
		return nil, err
	}
	return &productCatalogRecoveryPreviewBinding{Version: 1, ConfigurationSHA256: configurationDigest, Target: *target, TargetMedia: media}, nil
}

func (manager *Manager) completeProductCatalogRecoveryPreviewBinding(ctx context.Context, task *Task, binding *productCatalogRecoveryPreviewBinding, originalClone *productCatalogRecoveryDatabaseBinding) error {
	cfg := manager.Configuration
	current, err := manager.captureProductCatalogRecoveryPreviewBaseline(ctx)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, binding) {
		return errors.New("product target database identity, schema, sequence or complete media changed during validation")
	}
	stage, err := cfg.captureProductCatalogRecoveryDatabaseBinding(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	if originalClone == nil || stage.OID != originalClone.OID || stage.Owner != "local_backup_isolated_owner" || stage.Fingerprint != task.StageFingerprint || stage.SchemaSHA256 != originalClone.SchemaSHA256 {
		return errors.New("product candidate identity, schema or write effects changed during module merge")
	}
	candidateMedia, err := manager.productCatalogCandidateMediaConfiguration(task).inventoryModuleRecoveryCompleteMedia(ctx)
	if err != nil {
		return err
	}
	if err := validateProductCatalogCandidateCompleteMediaPreservation(binding.TargetMedia, candidateMedia, task.Manifest); err != nil {
		return err
	}
	protection, err := cfg.inspectProductCatalogRequiredImageCleanupProtectionInDatabase(ctx, task.StageDatabase)
	if err != nil {
		return err
	}
	binding.RequiredImageCleanupProtection = protection
	binding.Candidate, binding.CandidateMedia = *stage, candidateMedia
	task.ProductCatalogPreviewBinding = binding
	// Use the same database-plus-media definition as whole-storefront rollback.
	task.PreviewTargetFingerprint, _, err = cfg.targetFingerprint(ctx)
	return err
}

func validateProductCatalogCandidateCompleteMediaPreservation(baseline, candidate *moduleRecoveryCompleteMediaInventory, manifest *Manifest) error {
	if baseline == nil || candidate == nil || manifest == nil {
		return errors.New("product complete media preservation binding is missing")
	}
	expected := map[string]FileIdentity{}
	for _, file := range baseline.Files {
		expected[file.Identity.Path] = file.Identity
	}
	for _, file := range manifest.Files {
		if !strings.HasPrefix(file.Path, "media/") {
			continue
		}
		if current, exists := expected[file.Path]; exists && current != file {
			return errors.New("product candidate image conflicts with retained media")
		}
		expected[file.Path] = file
	}
	if len(expected) != len(candidate.Files) {
		return errors.New("product candidate has missing or unregistered media files")
	}
	aliases := map[string]string{}
	for _, file := range candidate.Files {
		if expected[file.Identity.Path] != file.Identity {
			return errors.New("product candidate media content changed")
		}
		aliases[strings.ToLower(file.Identity.Path)] = file.Identity.Path
	}
	directories := map[string]bool{}
	for _, directory := range candidate.Directories {
		path := "media/" + directory.RootName + "/" + directory.Key
		if file, exists := aliases[strings.ToLower(path)]; exists {
			return errors.New("product candidate file/directory or portable path collision: " + file)
		}
		aliases[strings.ToLower(path)] = path
		directories[path] = true
	}
	if len(aliases) != len(candidate.Files)+len(candidate.Directories) {
		return errors.New("product candidate has a portable case collision")
	}
	for _, directory := range baseline.Directories {
		if !directories["media/"+directory.RootName+"/"+directory.Key] {
			return errors.New("product candidate lost a retained empty directory")
		}
	}
	return nil
}

func (manager *Manager) prepareProductCatalogCandidateCompleteMediaDirectories(task *Task, baseline *productCatalogRecoveryPreviewBinding) error {
	cfg := manager.productCatalogCandidateMediaConfiguration(task)
	for _, root := range cfg.StorageRoots {
		if err := os.MkdirAll(root, 0700); err != nil {
			return err
		}
	}
	for _, directory := range baseline.TargetMedia.Directories {
		if _, err := customerServiceCandidateMediaPathComponents(directory.RootName, directory.Key); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(cfg.StorageRoots[directory.RootName], filepath.FromSlash(directory.Key)), 0700); err != nil {
			return err
		}
	}
	return nil
}

func (manager *Manager) validateProductCatalogBundledStaticImagesAlreadyMatchTarget(ctx context.Context, task *Task) error {
	for _, file := range task.Manifest.Files {
		if !strings.HasPrefix(file.Path, "product-static/") {
			continue
		}
		key := strings.TrimPrefix(file.Path, "product-static/")
		identity, err := inspectBoundedModuleRecoveryRegularFileIdentity(ctx, manager.Configuration.ProductStaticAssetsDirectory, key, manager.Configuration.MaxExpandedBytes)
		identity.Path = file.Path
		if err != nil || identity != file {
			return errors.New("bundled product image is absent or differs in the deployed static assets; deploy matching assets before switching")
		}
	}
	return nil
}

func (manager *Manager) dropProductCatalogStageOnlyWithBoundIdentity(ctx context.Context, task *Task) error {
	if task.ProductCatalogPreviewBinding == nil {
		return manager.Configuration.dropOwnedStageDatabase(ctx, task.StageDatabase)
	}
	control, err := manager.Configuration.connectDatabase(ctx, "postgres")
	if err != nil {
		return err
	}
	defer control.Close()
	var oid int64
	if err := readDatabaseOID(ctx, control, task.StageDatabase, &oid); err != nil {
		return err
	}
	if oid == 0 {
		return nil
	}
	var owner string
	if err := control.QueryRowContext(ctx, `SELECT pg_get_userbyid(datdba) FROM pg_database WHERE oid=$1`, oid).Scan(&owner); err != nil {
		return err
	}
	if task.StageDatabase != "local_backup_stage_"+task.ID[:16] || oid != task.ProductCatalogPreviewBinding.Candidate.OID || owner != task.ProductCatalogPreviewBinding.Candidate.Owner {
		return errors.New("product stage cleanup identity changed; database retained for inspection")
	}
	_, err = control.ExecContext(ctx, `DROP DATABASE `+pq.QuoteIdentifier(task.StageDatabase)+` WITH (FORCE)`)
	return err
}
