package maintenancegate

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/gofrs/flock"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const ProductViewRedisOwnershipFilename = "product-view-redis-runtime-ownership.json"
const ProductViewRedisOwnershipKey = "product:view_counts:recovery_ownership"

// This marker identifies the registered runtime's logical Redis database. It
// does not attest that external writers use the filesystem maintenance gate.
type ProductViewRedisRuntimeOwnership struct {
	Version           int    `json:"version"`
	SiteID            string `json:"site_id"`
	Environment       string `json:"environment"`
	Database          string `json:"database"`
	DestinationSHA256 string `json:"destination_sha256"`
	KeysSHA256        string `json:"keys_sha256"`
	MarkerSHA256      string `json:"marker_sha256"`
	ServerRunIDSHA256 string `json:"server_run_id_sha256"`
}

func FingerprintRedisRecoveryDestination(mode string, addresses []string, database int, master string) string {
	addresses = append([]string(nil), addresses...)
	sort.Strings(addresses)
	if mode != "sentinel" {
		master = ""
	}
	return FingerprintExternalProjectionConfiguration(struct {
		Mode      string
		Addresses []string
		Database  int
		Master    string
	}{mode, addresses, database, master})
}

var initializeOnlyEmptyProductViewRedisOwnershipScript = redis.NewScript(`
local marker = redis.call('GET', KEYS[1])
if marker then return marker end
for i=2,4 do
  if redis.call('EXISTS', KEYS[i]) ~= 0 then
    return redis.error_reply('existing product view buffers require manual provenance reconciliation before ownership enrollment')
  end
end
redis.call('SET', KEYS[1], ARGV[1])
return ARGV[1]
`)

func readProductViewRedisPrimaryServerIdentity(ctx context.Context, client redis.UniversalClient) (string, error) {
	role, err := client.Do(ctx, "ROLE").Slice()
	if err != nil || len(role) == 0 || role[0] != "master" {
		return "", errors.New("product view recovery requires a readable Redis primary identity")
	}
	info, err := client.Info(ctx, "server").Result()
	if err != nil {
		return "", errors.New("redis server identity is unavailable; recovery refused")
	}
	for _, line := range strings.Split(info, "\n") {
		if value, ok := strings.CutPrefix(strings.TrimSpace(line), "run_id:"); ok && len(value) == 40 {
			for _, character := range value {
				if !strings.ContainsRune("0123456789abcdef", character) {
					return "", errors.New("redis server identity is invalid")
				}
			}
			return FingerprintExternalProjectionConfiguration(value), nil
		}
	}
	return "", errors.New("redis server identity is missing; recovery refused")
}

func ReadRegisteredProductViewRedisRuntimeOwnership(directory string) (*ProductViewRedisRuntimeOwnership, error) {
	data, err := os.ReadFile(filepath.Join(directory, ProductViewRedisOwnershipFilename)) // #nosec G304 -- The directory is administrator-configured private maintenance storage; the ownership filename is fixed.
	if err != nil {
		return nil, errors.New("product view Redis ownership has not been registered by the business runtime")
	}
	var record ProductViewRedisRuntimeOwnership
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&record) != nil || decoder.Decode(new(any)) != io.EOF || record.Version != 1 || record.SiteID == "" || record.Environment == "" || record.Database == "" || !externalProjectionSHA256.MatchString(record.DestinationSHA256) || !externalProjectionSHA256.MatchString(record.KeysSHA256) || !externalProjectionSHA256.MatchString(record.MarkerSHA256) || !externalProjectionSHA256.MatchString(record.ServerRunIDSHA256) {
		return nil, errors.New("registered product view Redis ownership is invalid")
	}
	return &record, nil
}

// Enrollment must run before API listeners, flushers or other dependencies start.
// A missing marker is initialized only when all three delta keys are absent.
// A prior filesystem pin is never silently rebound after marker loss.
func RegisterProductViewRedisRuntimeOwnershipBeforeStartingWriters(ctx context.Context, directory, siteID, environment, database, destinationSHA256, keysSHA256 string, client redis.UniversalClient) error {
	if directory == "" || siteID == "" || environment == "" || database == "" || client == nil || !externalProjectionSHA256.MatchString(destinationSHA256) || !externalProjectionSHA256.MatchString(keysSHA256) {
		return errors.New("explicit storefront and Redis ownership identities are required")
	}
	lease := flock.New(filepath.Join(directory, "product-view-redis-ownership-registration.lock"))
	locked, err := lease.TryLockContext(ctx, 100*time.Millisecond)
	if err != nil || !locked {
		_ = lease.Close()
		return errors.New("redis ownership registration is busy")
	}
	defer func(closeResource func() error) { _ = closeResource() }(lease.Close)
	if fenced, err := IsApplicationFenced(directory); err != nil || fenced {
		return errors.New("redis ownership registration is forbidden during maintenance")
	}
	server, err := readProductViewRedisPrimaryServerIdentity(ctx, client)
	if err != nil {
		return err
	}
	var marker string
	if _, err := os.Stat(filepath.Join(directory, ProductViewRedisOwnershipFilename)); err == nil {
		prior, err := ReadRegisteredProductViewRedisRuntimeOwnership(directory)
		if err != nil || prior.SiteID != siteID || prior.Environment != environment || prior.Database != database || prior.DestinationSHA256 != destinationSHA256 || prior.KeysSHA256 != keysSHA256 {
			return errors.New("previous Redis ownership scope changed; controlled reconciliation is required")
		}
		marker, err = client.Get(ctx, ProductViewRedisOwnershipKey).Result()
		if err != nil || FingerprintExternalProjectionConfiguration(marker) != prior.MarkerSHA256 {
			return errors.New("pinned Redis ownership marker changed or disappeared; preserve buffers for inspection")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return errors.New("previous Redis ownership cannot be inspected")
	} else {
		candidate, _ := json.Marshal(struct{ SiteID, Environment, Database, DestinationSHA256, KeysSHA256, Nonce string }{siteID, environment, database, destinationSHA256, keysSHA256, uuid.NewString()})
		marker, err = initializeOnlyEmptyProductViewRedisOwnershipScript.Run(ctx, client, []string{ProductViewRedisOwnershipKey, "product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id"}, string(candidate)).Text()
		if err != nil {
			return errors.New("redis ownership enrollment refused; existing or invalid buffers require provenance reconciliation")
		}
	}
	var scope struct{ SiteID, Environment, Database, DestinationSHA256, KeysSHA256, Nonce string }
	if json.Unmarshal([]byte(marker), &scope) != nil || scope.SiteID != siteID || scope.Environment != environment || scope.Database != database || scope.DestinationSHA256 != destinationSHA256 || scope.KeysSHA256 != keysSHA256 || uuid.Validate(scope.Nonce) != nil {
		return errors.New("redis logical database belongs to a different or unknown product view owner")
	}
	record := ProductViewRedisRuntimeOwnership{1, siteID, environment, database, destinationSHA256, keysSHA256, FingerprintExternalProjectionConfiguration(marker), server}
	if err := VerifyRegisteredProductViewRedisRuntimeOwnership(ctx, client, &record); err != nil {
		return err
	}
	data, _ := json.Marshal(record)
	file, err := os.CreateTemp(directory, "product-view-redis-ownership-*.next")
	if err != nil {
		return err
	}
	defer func(cleanupPath string) { _ = os.Remove(cleanupPath) }(file.Name())
	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	_, writeErr := file.Write(data)
	if err := errors.Join(writeErr, file.Sync(), file.Close()); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), filepath.Join(directory, ProductViewRedisOwnershipFilename)); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		directoryFile, err := os.Open(directory) // #nosec G304 -- The directory is administrator-configured private maintenance storage; the ownership filename is fixed.
		if err != nil {
			return err
		}
		return errors.Join(directoryFile.Sync(), directoryFile.Close())
	}
	return nil
}

func VerifyRegisteredProductViewRedisRuntimeOwnership(ctx context.Context, client redis.UniversalClient, record *ProductViewRedisRuntimeOwnership) error {
	if record == nil || client == nil {
		return errors.New("registered Redis ownership is required")
	}
	server, err := readProductViewRedisPrimaryServerIdentity(ctx, client)
	if err != nil || server != record.ServerRunIDSHA256 {
		return errors.New("redis instance restarted or changed since runtime registration; recovery refused")
	}
	marker, err := client.Get(ctx, ProductViewRedisOwnershipKey).Result()
	if err != nil || FingerprintExternalProjectionConfiguration(marker) != record.MarkerSHA256 {
		return errors.New("redis ownership marker mismatch; recovery refused")
	}
	ttl, err := client.PTTL(ctx, ProductViewRedisOwnershipKey).Result()
	if err != nil || ttl != -1 {
		return errors.New("redis ownership marker must be persistent")
	}
	return nil
}
