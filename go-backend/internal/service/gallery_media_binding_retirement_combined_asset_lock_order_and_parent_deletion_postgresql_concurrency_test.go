package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func createIsolatedPostgreSQLGalleryMediaConcurrencyTestServices(ctx context.Context, t *testing.T) (*gorm.DB, *repository.GalleryRepository, *MediaService) {
	t.Helper()
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&gallery.Gallery{}, &gallery.GalleryImage{}))
	db := env.db.WithContext(ctx)
	assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	galleries := repository.NewGalleryRepository(db)
	galleries.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	return db, galleries, assets
}

func waitForPostgreSQLGalleryMutationToBlockOnTransaction(t *testing.T, db *gorm.DB, blockerPID int, table string) {
	t.Helper()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE ?)", blockerPID, "%"+table+"%").Scan(&waiting).Error
		return err == nil && waiting
	}, 4*time.Second, 20*time.Millisecond)
}

func TestPostgreSQLGalleryCoverOriginalAndThumbnailAllWriteEntrypointsSerializeBindingAndRetirement(t *testing.T) {
	for _, field := range []string{"cover", "original", "thumbnail"} {
		mutations := []string{"create gallery", "append image", "update image", "batch append", "import gallery"}
		if field == "cover" {
			mutations = []string{"create gallery", "update cover", "import gallery"}
		}
		for _, mutation := range mutations {
			for _, firstOperation := range []string{"bind", "retire"} {
				t.Run(field+"/"+mutation+"/"+firstOperation+" first", func(t *testing.T) {
					ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
					defer cancel()
					db, galleries, assets := createIsolatedPostgreSQLGalleryMediaConcurrencyTestServices(ctx, t)
					asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/gallery-race.png")
					parent := &gallery.Gallery{Name: "Existing", Slug: "existing"}
					require.NoError(t, galleries.CreateGallery(parent))
					existing := &gallery.GalleryImage{GalleryID: parent.ID, URL: "https://external.example.test/before.png", Thumbnail: "https://external.example.test/before-thumb.png"}
					if mutation == "update image" {
						require.NoError(t, galleries.CreateGalleryImage(existing))
					}
					bind := func() error {
						item := &gallery.Gallery{ID: parent.ID, Name: "Bound", Slug: parent.Slug}
						image := gallery.GalleryImage{GalleryID: parent.ID, URL: "https://external.example.test/image.png", Thumbnail: "https://external.example.test/thumb.png"}
						reference := "/uploads/" + asset.StorageKey
						switch field {
						case "cover":
							item.CoverImage = reference
						case "original":
							image.URL = reference
							if mutation != "import gallery" {
								image.MediaAssetID, image.URL = &asset.ID, ""
							}
						default:
							image.Thumbnail = reference
						}
						switch mutation {
						case "create gallery":
							item.ID = 0
							item.Slug = "created"
							if field != "cover" {
								item.Images = []gallery.GalleryImage{image}
							}
							return galleries.CreateGallery(item)
						case "update cover":
							return galleries.UpdateGallery(item)
						case "append image":
							return galleries.CreateGalleryImage(&image)
						case "update image":
							image.ID = existing.ID
							return galleries.UpdateGalleryImage(&image)
						case "batch append":
							return galleries.BatchCreateImages([]gallery.GalleryImage{image, {GalleryID: parent.ID, URL: "https://external.example.test/second.png"}})
						default:
							var images []gallery.GalleryImage
							if field != "cover" {
								images = []gallery.GalleryImage{image}
							}
							_, _, err := galleries.SaveImportedGalleryAndAppendPreviouslyUnplacedImages(item, images)
							return err
						}
					}
					firstLocked := make(chan int, 1)
					releaseFirst := make(chan struct{})
					var pauseOnce, releaseOnce sync.Once
					release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
					defer release()
					pause := func(tx *gorm.DB) {
						pauseOnce.Do(func() {
							var pid int
							if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
								tx.AddError(err)
								return
							}
							firstLocked <- pid
							select {
							case <-releaseFirst:
							case <-ctx.Done():
								tx.AddError(ctx.Err())
							}
						})
					}
					callbackName := "pause_gallery_media_race_" + uuid.NewString()
					if firstOperation == "retire" {
						require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "media_assets" {
								pause(tx)
							}
						}))
						defer db.Callback().Delete().Remove(callbackName)
					} else {
						table := "gallery_images"
						if mutation == "create gallery" || mutation == "update cover" || mutation == "import gallery" {
							table = "galleries"
						}
						callback := func(tx *gorm.DB) {
							if tx.Statement.Table == table {
								pause(tx)
							}
						}
						if mutation == "update image" || mutation == "update cover" || mutation == "import gallery" {
							require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, callback))
							defer db.Callback().Update().Remove(callbackName)
						} else {
							require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, callback))
							defer db.Callback().Create().Remove(callbackName)
						}
					}
					bindResult, retireResult := make(chan error, 1), make(chan error, 1)
					startBinding := func() { go func() { bindResult <- bind() }() }
					startRetirement := func() {
						go func() { retireResult <- assets.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }()
					}
					if firstOperation == "bind" {
						startBinding()
					} else {
						startRetirement()
					}
					var pid int
					select {
					case pid = <-firstLocked:
					case <-ctx.Done():
						t.Fatal("first operation did not acquire the media lock")
					}
					if firstOperation == "bind" {
						startRetirement()
					} else {
						startBinding()
					}
					waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "media_assets")
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
					var events int64
					require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).Count(&events).Error)
					if firstOperation == "bind" {
						require.NoError(t, bindErr)
						require.ErrorIs(t, retireErr, ErrMediaAssetInUse)
						require.Zero(t, events)
					} else {
						require.NoError(t, retireErr)
						require.ErrorIs(t, bindErr, repository.ErrGalleryMediaReferenceInvalid)
						require.EqualValues(t, 1, events)
						rebindError := bind()
						require.True(t, errors.Is(rebindError, repository.ErrGalleryMediaReferenceInvalid) || errors.Is(rebindError, repository.ErrGalleryMediaAssetNotFound), "physical cleanup must not permit ID or URL rebinding: %v", rebindError)
						current, err := galleries.FindGalleryByID(parent.ID)
						require.NoError(t, err)
						require.Equal(t, "Existing", current.Name)
						require.Empty(t, current.CoverImage)
						var created int64
						require.NoError(t, db.Model(&gallery.Gallery{}).Where("slug = ?", "created").Count(&created).Error)
						require.Zero(t, created)
						if mutation == "update image" {
							stored, err := galleries.FindGalleryImageByID(existing.ID)
							require.NoError(t, err)
							require.Equal(t, existing.URL, stored.URL)
							require.Equal(t, existing.Thumbnail, stored.Thumbnail)
						} else {
							require.Empty(t, current.Images)
						}
					}
				})
			}
		}
	}
}

func TestPostgreSQLGalleryCoverOriginalAndThumbnailUseOneAscendingAssetLockPlan(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	db, galleries, _ := createIsolatedPostgreSQLGalleryMediaConcurrencyTestServices(ctx, t)
	low := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/gallery-low.png")
	middle := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/gallery-middle.png")
	high := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/gallery-high.png")
	first := &gallery.Gallery{Name: "First", Slug: "first", CoverImage: high.URL, Images: []gallery.GalleryImage{{URL: middle.URL, Thumbnail: low.URL}}}
	second := &gallery.Gallery{Name: "Second", Slug: "second", CoverImage: low.URL, Images: []gallery.GalleryImage{{URL: high.URL, Thumbnail: middle.URL}}}
	firstLocked, releaseFirst := make(chan int, 1), make(chan struct{})
	var pauseOnce, releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
	defer release()
	name := "pause_gallery_combined_lock_order_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(name, func(tx *gorm.DB) {
		if tx.Statement.Table != "media_assets" || !strings.Contains(tx.Statement.SQL.String(), "FOR UPDATE") {
			return
		}
		pauseOnce.Do(func() {
			asset, ok := tx.Statement.Dest.(*media.MediaAsset)
			if !ok || asset.ID != low.ID {
				tx.AddError(fmt.Errorf("combined gallery plan did not lock lowest asset ID first"))
				return
			}
			var pid int
			if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
				tx.AddError(err)
				return
			}
			firstLocked <- pid
			select {
			case <-releaseFirst:
			case <-ctx.Done():
				tx.AddError(ctx.Err())
			}
		})
	}))
	defer db.Callback().Query().Remove(name)
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go func() { firstResult <- galleries.CreateGallery(first) }()
	var pid int
	select {
	case pid = <-firstLocked:
	case <-ctx.Done():
		t.Fatal("combined plan never locked the lowest asset")
	}
	go func() { secondResult <- galleries.CreateGallery(second) }()
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "media_assets")
	release()
	for _, result := range []chan error{firstResult, secondResult} {
		select {
		case err := <-result:
			require.NoError(t, err)
		case <-ctx.Done():
			t.Fatal("combined gallery lock plans deadlocked")
		}
	}
}

func TestPostgreSQLGalleryParentDeletionSerializesSingleBatchUpdateAndImportMutations(t *testing.T) {
	for _, mutation := range []string{"append", "batch", "update", "import"} {
		for _, firstOperation := range []string{"write", "delete"} {
			t.Run(mutation+"/"+firstOperation+" first", func(t *testing.T) {
				ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
				defer cancel()
				db, galleries, _ := createIsolatedPostgreSQLGalleryMediaConcurrencyTestServices(ctx, t)
				parent := &gallery.Gallery{Name: "Parent", Slug: "parent"}
				require.NoError(t, galleries.CreateGallery(parent))
				existing := &gallery.GalleryImage{GalleryID: parent.ID, URL: "https://external.example.test/before.png"}
				if mutation == "update" {
					require.NoError(t, galleries.CreateGalleryImage(existing))
				}
				write := func() error {
					image := gallery.GalleryImage{GalleryID: parent.ID, URL: "https://external.example.test/after.png"}
					switch mutation {
					case "append":
						return galleries.CreateGalleryImage(&image)
					case "batch":
						return galleries.BatchCreateImages([]gallery.GalleryImage{image})
					case "update":
						image.ID = existing.ID
						return galleries.UpdateGalleryImage(&image)
					default:
						item := *parent
						_, _, err := galleries.SaveImportedGalleryAndAppendPreviouslyUnplacedImages(&item, []gallery.GalleryImage{image})
						return err
					}
				}
				firstLocked, releaseFirst := make(chan int, 1), make(chan struct{})
				var pauseOnce, releaseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
				defer release()
				pause := func(tx *gorm.DB) {
					pauseOnce.Do(func() {
						var pid int
						if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
							tx.AddError(err)
							return
						}
						firstLocked <- pid
						select {
						case <-releaseFirst:
						case <-ctx.Done():
							tx.AddError(ctx.Err())
						}
					})
				}
				name := "pause_gallery_parent_race_" + uuid.NewString()
				if firstOperation == "delete" {
					require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(name, func(tx *gorm.DB) {
						if tx.Statement.Table == "galleries" {
							pause(tx)
						}
					}))
					defer db.Callback().Delete().Remove(name)
				} else if mutation == "update" || mutation == "import" {
					require.NoError(t, db.Callback().Update().After("gorm:update").Register(name, func(tx *gorm.DB) {
						if tx.Statement.Table == "gallery_images" || tx.Statement.Table == "galleries" {
							pause(tx)
						}
					}))
					defer db.Callback().Update().Remove(name)
				} else {
					require.NoError(t, db.Callback().Create().After("gorm:create").Register(name, func(tx *gorm.DB) {
						if tx.Statement.Table == "gallery_images" {
							pause(tx)
						}
					}))
					defer db.Callback().Create().Remove(name)
				}
				writeResult, deleteResult := make(chan error, 1), make(chan error, 1)
				startWrite := func() { go func() { writeResult <- write() }() }
				startDelete := func() { go func() { deleteResult <- galleries.DeleteGallery(parent.ID) }() }
				if firstOperation == "write" {
					startWrite()
				} else {
					startDelete()
				}
				var pid int
				select {
				case pid = <-firstLocked:
				case <-ctx.Done():
					t.Fatal("first operation never locked the gallery parent")
				}
				if firstOperation == "write" {
					startDelete()
				} else {
					startWrite()
				}
				waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "galleries")
				release()
				select {
				case err := <-writeResult:
					if firstOperation == "write" {
						require.NoError(t, err)
					} else {
						require.ErrorIs(t, err, gorm.ErrRecordNotFound)
					}
				case <-ctx.Done():
					t.Fatal("image write timed out")
				}
				select {
				case err := <-deleteResult:
					require.NoError(t, err)
				case <-ctx.Done():
					t.Fatal("gallery deletion timed out")
				}
				var count int64
				require.NoError(t, db.Model(&gallery.GalleryImage{}).Where("gallery_id = ?", parent.ID).Count(&count).Error)
				require.Zero(t, count)
				_, err := galleries.FindGalleryByID(parent.ID)
				require.ErrorIs(t, err, gorm.ErrRecordNotFound)
			})
		}
	}
}
