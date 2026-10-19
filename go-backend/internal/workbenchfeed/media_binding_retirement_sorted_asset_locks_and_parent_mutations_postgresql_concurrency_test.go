package workbenchfeed

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	mediaservice "commerce-platform/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func createIsolatedPostgreSQLWorkbenchFeedMediaConcurrencyDatabase(t *testing.T, ctx context.Context) *gorm.DB {
	t.Helper()
	host, username, databaseName := os.Getenv("DB_HOST"), os.Getenv("DB_USERNAME"), os.Getenv("DB_NAME")
	if host == "" || username == "" {
		t.Skip("PostgreSQL integration environment is not configured")
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	if databaseName == "" {
		databaseName = "postgres"
	}
	connectionURL := &url.URL{Scheme: "postgres", Host: net.JoinHostPort(host, port), User: url.UserPassword(username, os.Getenv("DB_PASSWORD")), Path: "/" + databaseName, RawQuery: "sslmode=disable&connect_timeout=5"}
	bootstrap, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	bootstrapConnection, err := bootstrap.DB()
	require.NoError(t, err)
	schema := "test_workbench_media_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	require.NoError(t, bootstrap.Exec("CREATE SCHEMA "+schema).Error)
	t.Cleanup(func() {
		require.NoError(t, bootstrap.Exec("DROP SCHEMA "+schema+" CASCADE").Error)
		_ = bootstrapConnection.Close()
	})
	query := connectionURL.Query()
	query.Set("search_path", schema)
	connectionURL.RawQuery = query.Encode()
	db, err := gorm.Open(postgres.Open(connectionURL.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	connection, err := db.DB()
	require.NoError(t, err)
	connection.SetMaxOpenConns(8)
	t.Cleanup(func() { _ = connection.Close() })
	return db.WithContext(ctx)
}

func pauseFirstPostgreSQLWorkbenchMediaMutationAfterRowLock(t *testing.T, db *gorm.DB, table, sqlFragment string) (<-chan int, func()) {
	t.Helper()
	locked, release := make(chan int, 1), make(chan struct{})
	var pauseOnce, releaseOnce sync.Once
	unlock := func() { releaseOnce.Do(func() { close(release) }) }
	name := "pause_workbench_transaction_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table != table || !strings.Contains(tx.Statement.SQL.String(), sqlFragment) {
			return
		}
		pauseOnce.Do(func() {
			var pid int
			if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
				tx.AddError(err)
				return
			}
			locked <- pid
			select {
			case <-release:
			case <-tx.Statement.Context.Done():
				tx.AddError(tx.Statement.Context.Err())
			}
		})
	}))
	t.Cleanup(func() { unlock(); _ = db.Callback().Query().Remove(name) })
	return locked, unlock
}

func waitForPostgreSQLWorkbenchMutationToWaitOnRowLock(t *testing.T, db *gorm.DB, blocker int, table string) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE ?)", blocker, "%"+table+"%").Scan(&waiting).Error
		return err == nil && waiting
	}, 5*time.Second, 20*time.Millisecond)
}

func TestPostgreSQLWorkbenchFeedCreateAndUpdateSerializeNewMediaBindingAgainstRetirementInBothOrders(t *testing.T) {
	for _, mutation := range []string{"create", "update"} {
		for _, first := range []string{"bind", "retire"} {
			t.Run(mutation+"/"+first+" first", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				db := createIsolatedPostgreSQLWorkbenchFeedMediaConcurrencyDatabase(t, ctx)
				db, feed, assets, root := createWorkbenchFeedMediaBindingTestServices(t, db)
				asset := createWorkbenchFeedRegisteredImageAndPhysicalObject(t, db, root, "workbench-feed/race.png")
				entry, err := feed.Create(workbenchFeedEntryInputWithImagePaths())
				require.NoError(t, err)
				bind := func() error {
					input := workbenchFeedEntryInputWithImagePaths("/uploads/" + asset.StorageKey)
					input.Content = "Committed binding"
					if mutation == "create" {
						_, err := feed.Create(input)
						return err
					}
					_, err := feed.Update(entry.ID, input)
					return err
				}
				locked, release := pauseFirstPostgreSQLWorkbenchMediaMutationAfterRowLock(t, db, "media_assets", "FOR UPDATE")
				bindResult, retireResult := make(chan error, 1), make(chan error, 1)
				startBinding := func() { go func() { bindResult <- bind() }() }
				startRetirement := func() {
					go func() {
						retireResult <- assets.DeleteAsset(ctx, asset.ID, mediaservice.MediaAssetDeleteConfirmation(asset.ID))
					}()
				}
				if first == "bind" {
					startBinding()
				} else {
					startRetirement()
				}
				var pid int
				select {
				case pid = <-locked:
				case <-ctx.Done():
					t.Fatal("first media lock was not acquired")
				}
				if first == "bind" {
					startRetirement()
				} else {
					startBinding()
				}
				waitForPostgreSQLWorkbenchMutationToWaitOnRowLock(t, db, pid, "media_assets")
				release()
				var bindErr, retireErr error
				select {
				case bindErr = <-bindResult:
				case <-ctx.Done():
					t.Fatal("binding timed out")
				}
				select {
				case retireErr = <-retireResult:
				case <-ctx.Done():
					t.Fatal("retirement timed out")
				}
				var events, placements int64
				require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
				require.NoError(t, db.Model(&FeedMedia{}).Count(&placements).Error)
				if first == "bind" {
					require.NoError(t, bindErr)
					require.ErrorIs(t, retireErr, mediaservice.ErrMediaAssetInUse)
					require.Zero(t, events)
					require.EqualValues(t, 1, placements)
				} else {
					require.NoError(t, retireErr)
					require.ErrorIs(t, bindErr, ErrMediaReferenceInvalid)
					require.EqualValues(t, 1, events)
					require.Zero(t, placements)
					require.ErrorIs(t, bind(), ErrMediaReferenceInvalid)
					stored, err := feed.Get(entry.ID)
					require.NoError(t, err)
					require.Equal(t, entry.Content, stored.Content)
				}
			})
		}
	}
}

func TestPostgreSQLWorkbenchFeedMultipleImagesUseOneAscendingMediaAssetLockPlan(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	db := createIsolatedPostgreSQLWorkbenchFeedMediaConcurrencyDatabase(t, ctx)
	db, feed, _, root := createWorkbenchFeedMediaBindingTestServices(t, db)
	low := createWorkbenchFeedRegisteredImageAndPhysicalObject(t, db, root, "content/low.png")
	high := createWorkbenchFeedRegisteredImageAndPhysicalObject(t, db, root, "content/high.png")
	first, err := feed.Create(workbenchFeedEntryInputWithImagePaths())
	require.NoError(t, err)
	second, err := feed.Create(workbenchFeedEntryInputWithImagePaths())
	require.NoError(t, err)
	var firstAssetID uint
	var observeOnce sync.Once
	name := "observe_workbench_first_asset_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table == "media_assets" && strings.Contains(tx.Statement.SQL.String(), "FOR UPDATE") {
			observeOnce.Do(func() {
				if item, ok := tx.Statement.Dest.(*media.MediaAsset); ok {
					firstAssetID = item.ID
				}
			})
		}
	}))
	defer db.Callback().Query().Remove(name)
	locked, release := pauseFirstPostgreSQLWorkbenchMediaMutationAfterRowLock(t, db, "media_assets", "FOR UPDATE")
	results := make(chan error, 2)
	go func() {
		_, err := feed.Update(first.ID, workbenchFeedEntryInputWithImagePaths(high.URL, low.URL))
		results <- err
	}()
	var pid int
	select {
	case pid = <-locked:
	case <-ctx.Done():
		t.Fatal("first media lock was not acquired")
	}
	require.Equal(t, low.ID, firstAssetID)
	go func() {
		_, err := feed.Update(second.ID, workbenchFeedEntryInputWithImagePaths(low.URL, high.URL))
		results <- err
	}()
	waitForPostgreSQLWorkbenchMutationToWaitOnRowLock(t, db, pid, "media_assets")
	release()
	for range 2 {
		select {
		case err := <-results:
			require.NoError(t, err)
		case <-ctx.Done():
			t.Fatal("combined image lock plan timed out")
		}
	}
}

func TestPostgreSQLWorkbenchFeedStatusAndDeletionSerializeWithEntryAndChildrenReplacement(t *testing.T) {
	for _, scenario := range []string{"update-before-status", "update-before-delete", "delete-before-update"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			db := createIsolatedPostgreSQLWorkbenchFeedMediaConcurrencyDatabase(t, ctx)
			db, feed, _, _ := createWorkbenchFeedMediaBindingTestServices(t, db)
			entry, err := feed.Create(workbenchFeedEntryInputWithImagePaths("https://external.example.test/before.png"))
			require.NoError(t, err)
			locked, release := pauseFirstPostgreSQLWorkbenchMediaMutationAfterRowLock(t, db, "workbench_feed_entries", "FOR NO KEY UPDATE")
			update := func() error {
				input := workbenchFeedEntryInputWithImagePaths("https://external.example.test/after.png")
				input.Content = "Latest content"
				_, err := feed.Update(entry.ID, input)
				return err
			}
			other := func() error {
				if scenario == "update-before-status" {
					_, err := feed.UpdateStatus(entry.ID, StatusPublished, 81)
					return err
				}
				return feed.Delete(entry.ID)
			}
			firstResult, secondResult := make(chan error, 1), make(chan error, 1)
			if scenario == "delete-before-update" {
				go func() { firstResult <- other() }()
			} else {
				go func() { firstResult <- update() }()
			}
			var pid int
			select {
			case pid = <-locked:
			case <-ctx.Done():
				t.Fatal("first parent lock was not acquired")
			}
			if scenario == "delete-before-update" {
				go func() { secondResult <- update() }()
			} else {
				go func() { secondResult <- other() }()
			}
			waitForPostgreSQLWorkbenchMutationToWaitOnRowLock(t, db, pid, "workbench_feed_entries")
			release()
			select {
			case err := <-firstResult:
				require.NoError(t, err)
			case <-ctx.Done():
				t.Fatal("first parent mutation timed out")
			}
			select {
			case err := <-secondResult:
				if scenario == "delete-before-update" {
					require.ErrorIs(t, err, ErrEntryNotFound)
				} else {
					require.NoError(t, err)
				}
			case <-ctx.Done():
				t.Fatal("second parent mutation timed out")
			}
			if scenario == "update-before-status" {
				stored, err := feed.Get(entry.ID)
				require.NoError(t, err)
				require.Equal(t, "Latest content", stored.Content)
				require.Equal(t, StatusPublished, stored.Status)
				require.Len(t, stored.Media, 1)
				require.Equal(t, "https://external.example.test/after.png", stored.Media[0].FilePath)
			} else {
				_, err := feed.Get(entry.ID)
				require.ErrorIs(t, err, ErrEntryNotFound)
			}
			var placements int64
			require.NoError(t, db.Unscoped().Model(&FeedMedia{}).Count(&placements).Error)
			expected := 2
			if scenario == "delete-before-update" {
				expected = 1
			}
			require.EqualValues(t, expected, placements, fmt.Sprint("unexpected association writes in ", scenario))
		})
	}
}
