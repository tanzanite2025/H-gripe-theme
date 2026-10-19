package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/database"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"gorm.io/gorm"
)

// GalleryExportData WordPress 导出的图片库数据结构
type GalleryExportData struct {
	Galleries []struct {
		ID          int       `json:"id"`
		Name        string    `json:"name"`
		Slug        string    `json:"slug"`
		Description string    `json:"description"`
		CoverImage  string    `json:"cover_image"`
		ViewCount   int       `json:"view_count"`
		Status      string    `json:"status"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"galleries"`
	Images []struct {
		GalleryID   int    `json:"gallery_id"`
		URL         string `json:"url"`
		Thumbnail   string `json:"thumbnail"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Alt         string `json:"alt"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Size        int64  `json:"size"`
		Tags        string `json:"tags"`
		Order       int    `json:"order"`
	} `json:"images"`
	Stats struct {
		TotalGalleries int    `json:"total_galleries"`
		TotalImages    int    `json:"total_images"`
		ExportDate     string `json:"export_date"`
	} `json:"stats"`
}

var (
	dryRun     = flag.Bool("dry-run", false, "试运行模式，不实际写入数据库")
	inputFile  = flag.String("input", "data/import/galleries.json", "输入文件路径")
	configFile = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	fmt.Println("=== 图片库数据导入工具 ===")
	fmt.Printf("输入文件: %s\n", *inputFile)
	fmt.Printf("试运行模式: %v\n", *dryRun)
	fmt.Println()

	// 加载配置
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	// 读取导出文件
	data, err := readExportFile(*inputFile)
	if err != nil {
		log.Fatalf("读取导出文件失败: %v", err)
	}

	fmt.Printf("读取到 %d 个图片库，%d 张图片\n", data.Stats.TotalGalleries, data.Stats.TotalImages)
	fmt.Printf("导出时间: %s\n\n", data.Stats.ExportDate)

	if *dryRun {
		fmt.Println("=== 试运行模式 - 预览数据 ===")
		fmt.Println()
		previewData(data)
		fmt.Println("\n试运行完成，未写入数据库")
		return
	}

	// 导入数据
	galleryRepository := repository.NewGalleryRepository(db)
	storageConfig := storage.LoadConfigFromEnv()
	if _, configured := os.LookupEnv("STORAGE_BASE_URL"); !configured {
		storageConfig.BaseURL = cfg.Server.BaseURL
	}
	storageService, err := storage.NewStorageService(storageConfig)
	if err != nil {
		log.Fatalf("initialize gallery import media identity validation: %v", err)
	}
	storefrontBaseURL := strings.TrimSpace(os.Getenv("STOREFRONT_BASE_URL"))
	if storefrontBaseURL == "" {
		storefrontBaseURL = cfg.Server.BaseURL
	}
	galleryRepository.ConfigurePublicBusinessMediaReferenceIdentityResolver(service.NewMediaService(repository.NewMediaRepository(db), storageService, nil, storefrontBaseURL, cfg.MediaUpload.AccountStorageQuotaBytes))
	stats := importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db, galleryRepository, data)

	fmt.Println("\n=== 导入完成 ===")
	fmt.Printf("图片库: 创建 %d, 更新 %d, 跳过 %d\n",
		stats.GalleriesCreated, stats.GalleriesUpdated, stats.GalleriesSkipped)
	fmt.Printf("图片: 创建 %d, 跳过 %d\n",
		stats.ImagesCreated, stats.ImagesSkipped)
}

// readExportFile 读取导出文件
func readExportFile(filename string) (*GalleryExportData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var data GalleryExportData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &data, nil
}

// previewData 预览数据
func previewData(data *GalleryExportData) {
	for i, g := range data.Galleries {
		fmt.Printf("%d. 图片库: %s (slug: %s)\n", i+1, g.Name, g.Slug)
		fmt.Printf("   描述: %s\n", truncate(g.Description, 50))
		fmt.Printf("   状态: %s, 浏览: %d\n", g.Status, g.ViewCount)

		// 统计该图片库的图片数量
		imageCount := 0
		for _, img := range data.Images {
			if img.GalleryID == g.ID {
				imageCount++
			}
		}
		fmt.Printf("   图片数量: %d\n\n", imageCount)
	}
}

// ImportStats 导入统计
type ImportStats struct {
	GalleriesCreated int
	GalleriesUpdated int
	GalleriesSkipped int
	ImagesCreated    int
	ImagesSkipped    int
}

// Import each gallery and its new images in one transaction; reuse never copies assets.
func importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db *gorm.DB, galleries *repository.GalleryRepository, data *GalleryExportData) ImportStats {
	stats := ImportStats{}
	importedSourceIDs := make(map[int]bool)
	for _, incoming := range data.Galleries {
		if importedSourceIDs[incoming.ID] {
			stats.GalleriesSkipped++
			continue
		}
		importedSourceIDs[incoming.ID] = true
		var item gallery.Gallery
		err := db.Where("slug = ?", incoming.Slug).First(&item).Error
		creating := err == gorm.ErrRecordNotFound
		if err != nil && !creating {
			log.Printf("gallery lookup failed [%s]: %v", incoming.Slug, err)
			stats.GalleriesSkipped++
			for _, image := range data.Images {
				if image.GalleryID == incoming.ID {
					stats.ImagesSkipped++
				}
			}
			continue
		}
		if creating {
			item.Slug, item.CreatedAt = incoming.Slug, incoming.CreatedAt
		}
		item.Name, item.Description, item.CoverImage = incoming.Name, incoming.Description, incoming.CoverImage
		item.ViewCount, item.Status, item.UpdatedAt = incoming.ViewCount, incoming.Status, incoming.UpdatedAt
		var images []gallery.GalleryImage
		for _, image := range data.Images {
			if image.GalleryID != incoming.ID {
				continue
			}
			images = append(images, gallery.GalleryImage{URL: image.URL, Thumbnail: image.Thumbnail, Title: image.Title, Description: image.Description, Alt: image.Alt, Width: image.Width, Height: image.Height, Size: image.Size, Tags: image.Tags, Order: image.Order})
		}
		added, skipped, err := galleries.SaveImportedGalleryAndAppendPreviouslyUnplacedImages(&item, images)
		if err != nil {
			log.Printf("gallery import rolled back [%s]: %v", incoming.Slug, err)
			stats.GalleriesSkipped++
			stats.ImagesSkipped += len(images)
			continue
		}
		if creating {
			stats.GalleriesCreated++
		} else {
			stats.GalleriesUpdated++
		}
		stats.ImagesCreated += added
		stats.ImagesSkipped += skipped
	}
	for _, image := range data.Images {
		if !importedSourceIDs[image.GalleryID] {
			stats.ImagesSkipped++
		}
	}
	return stats
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
