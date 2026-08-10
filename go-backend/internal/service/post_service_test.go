package service

import (
	"errors"
	"testing"

	"tanzanite/internal/domain/post"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPostServicePublicAccessOnlyReturnsPublishedPosts(t *testing.T) {
	db, postService := newTestPostService(t)

	publishedPost := post.Post{
		Title:    "Published",
		Slug:     "published",
		Content:  "<p>Published</p>",
		Status:   "published",
		AuthorID: 1,
		Locale:   "en",
	}
	require.NoError(t, db.Create(&publishedPost).Error)

	draftPost := post.Post{
		Title:    "Draft",
		Slug:     "draft",
		Content:  "<p>Draft</p>",
		Status:   "draft",
		AuthorID: 1,
		Locale:   "en",
	}
	require.NoError(t, db.Create(&draftPost).Error)

	_, err := postService.GetPublicByID(draftPost.ID)
	require.ErrorIs(t, err, ErrPostNotFound)

	posts, total, err := postService.ListPublic("en", 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, posts, 1)
	assert.Equal(t, publishedPost.ID, posts[0].ID)
}

func TestPostServicePublicTranslationsRequirePublishedPosts(t *testing.T) {
	db, postService := newTestPostService(t)
	groupID := uint(77)

	sourcePost := post.Post{
		Title:              "English",
		Slug:               "english",
		Content:            "<p>English</p>",
		Status:             "published",
		AuthorID:           1,
		Locale:             "en",
		TranslationGroupID: &groupID,
	}
	require.NoError(t, db.Create(&sourcePost).Error)

	publishedTranslation := post.Post{
		Title:              "Chinese",
		Slug:               "chinese",
		Content:            "<p>Chinese</p>",
		Status:             "published",
		AuthorID:           1,
		Locale:             "zh",
		TranslationGroupID: &groupID,
	}
	require.NoError(t, db.Create(&publishedTranslation).Error)

	draftTranslation := post.Post{
		Title:              "Draft",
		Slug:               "draft-translation",
		Content:            "<p>Draft</p>",
		Status:             "draft",
		AuthorID:           1,
		Locale:             "fr",
		TranslationGroupID: &groupID,
	}
	require.NoError(t, db.Create(&draftTranslation).Error)

	translations, err := postService.GetPublicTranslations(sourcePost.ID)
	require.NoError(t, err)
	require.Len(t, translations, 2)
	assert.Equal(t, "en", translations[0].Locale)
	assert.Equal(t, "zh", translations[1].Locale)

	_, err = postService.GetPublicTranslations(draftTranslation.ID)
	require.ErrorIs(t, err, ErrPostNotFound)
}

func TestPostServiceUpdateAdminPostRejectsLocaleChange(t *testing.T) {
	db, postService := newTestPostService(t)

	existingPost := post.Post{
		Title:    "English",
		Slug:     "english-post",
		Content:  "<p>English</p>",
		Status:   "published",
		AuthorID: 1,
		Locale:   "en",
	}
	require.NoError(t, db.Create(&existingPost).Error)

	nextLocale := "fr"
	updatedPost, err := postService.UpdateAdminPost(existingPost.ID, PostUpdateInput{
		Locale: &nextLocale,
	})
	require.Nil(t, updatedPost)
	require.True(t, errors.Is(err, ErrPostLocaleImmutable))

	storedPost, err := postService.GetAdminPost(existingPost.ID)
	require.NoError(t, err)
	assert.Equal(t, "en", storedPost.Locale)
}

func TestPostServiceUpdateAdminPostAcceptsSameLocaleAlias(t *testing.T) {
	db, postService := newTestPostService(t)

	existingPost := post.Post{
		Title:    "English",
		Slug:     "english-alias-post",
		Content:  "<p>English</p>",
		Status:   "published",
		AuthorID: 1,
		Locale:   "en",
	}
	require.NoError(t, db.Create(&existingPost).Error)

	nextLocale := "en-US"
	updatedPost, err := postService.UpdateAdminPost(existingPost.ID, PostUpdateInput{
		Locale: &nextLocale,
	})
	require.NoError(t, err)
	require.NotNil(t, updatedPost)
	assert.Equal(t, "en", updatedPost.Locale)
}

func TestPostServiceUpdatePostSEOUsesDedicatedBoundary(t *testing.T) {
	db, postService := newTestPostService(t)

	existingPost := post.Post{
		Title:    "SEO Boundary",
		Slug:     "seo-boundary",
		Content:  "<p>Content</p>",
		Status:   "published",
		AuthorID: 1,
		Locale:   "en",
	}
	require.NoError(t, db.Create(&existingPost).Error)

	title := "SEO Boundary | Tanzanite"
	description := "A description maintained by the SEO control plane."
	canonical := "https://store.example.test/blog/seo-boundary"
	updatedPost, err := postService.UpdatePostSEO(existingPost.ID, PostSEOUpdateInput{
		MetaTitle:       &title,
		MetaDescription: &description,
		CanonicalURL:    &canonical,
	})
	require.NoError(t, err)
	require.NotNil(t, updatedPost)
	assert.Equal(t, title, updatedPost.MetaTitle)
	assert.Equal(t, description, updatedPost.MetaDesc)
	assert.Equal(t, canonical, updatedPost.CanonicalURL)
}

func newTestPostService(t *testing.T) (*gorm.DB, *PostService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&post.Post{}))
	return db, NewPostService(repository.NewPostRepository(db), nil, 0)
}
