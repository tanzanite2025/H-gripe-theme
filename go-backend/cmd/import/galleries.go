package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"tanzanite/internal/domain/gallery"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"

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
	inputFile  = flag.String("input", "scripts/wordpress-export/export/galleries.json", "输入文件路径")
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
		fmt.Println("=== 试运行模式 - 预览数据 ===\n")
		previewData(data)
		fmt.Println("\n试运行完成，未写入数据库")
		return
	}

	// 导入数据
	stats := importData(db, data)

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
	defer file.Close()

	content, err := ioutil.ReadAll(file)
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

// importData 导入数据
func importData(db *gorm.DB, data *GalleryExportData) ImportStats {
	stats := ImportStats{}

	// WordPress ID 到 Go ID 的映射
	galleryIDMap := make(map[int]uint)

	// 导入图片库
	fmt.Println("导入图片库...")
	for _, g := range data.Galleries {
		fmt.Printf("  处理: %s (slug: %s)...", g.Name, g.Slug)

		// 检查是否已存在
		var existing gallery.Gallery
		err := db.Where("slug = ?", g.Slug).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新图片库
			newGallery := gallery.Gallery{
				Name:        g.Name,
				Slug:        g.Slug,
				Description: g.Description,
				CoverImage:  g.CoverImage,
				ViewCount:   g.ViewCount,
				Status:      g.Status,
				CreatedAt:   g.CreatedAt,
				UpdatedAt:   g.UpdatedAt,
			}

			if err := db.Create(&newGallery).Error; err != nil {
				fmt.Printf(" 失败: %v\n", err)
				stats.GalleriesSkipped++
				continue
			}

			galleryIDMap[g.ID] = newGallery.ID
			stats.GalleriesCreated++
			fmt.Printf(" 创建 (ID: %d)\n", newGallery.ID)

		} else if err == nil {
			// 更新现有图片库
			existing.Name = g.Name
			existing.Description = g.Description
			existing.CoverImage = g.CoverImage
			existing.ViewCount = g.ViewCount
			existing.Status = g.Status
			existing.UpdatedAt = g.UpdatedAt

			if err := db.Save(&existing).Error; err != nil {
				fmt.Printf(" 更新失败: %v\n", err)
				stats.GalleriesSkipped++
				continue
			}

			galleryIDMap[g.ID] = existing.ID
			stats.GalleriesUpdated++
			fmt.Printf(" 更新 (ID: %d)\n", existing.ID)

		} else {
			fmt.Printf(" 查询失败: %v\n", err)
			stats.GalleriesSkipped++
		}
	}

	// 导入图片
	fmt.Println("\n导入图片...")
	for _, img := range data.Images {
		// 获取对应的 Go 图片库 ID
		goGalleryID, ok := galleryIDMap[img.GalleryID]
		if !ok {
			fmt.Printf("  跳过图片 (图片库 ID %d 不存在): %s\n", img.GalleryID, img.Title)
			stats.ImagesSkipped++
			continue
		}

		fmt.Printf("  添加图片: %s...", truncate(img.Title, 30))

		// 检查是否已存在（通过 URL）
		var existing gallery.GalleryImage
		err := db.Where("gallery_id = ? AND url = ?", goGalleryID, img.URL).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新图片
			newImage := gallery.GalleryImage{
				GalleryID:   goGalleryID,
				URL:         img.URL,
				Thumbnail:   img.Thumbnail,
				Title:       img.Title,
				Description: img.Description,
				Alt:         img.Alt,
				Width:       img.Width,
				Height:      img.Height,
				Size:        img.Size,
				Tags:        img.Tags,
				Order:       img.Order,
			}

			if err := db.Create(&newImage).Error; err != nil {
				fmt.Printf(" 失败: %v\n", err)
				stats.ImagesSkipped++
				continue
			}

			stats.ImagesCreated++
			fmt.Printf(" 创建 (ID: %d)\n", newImage.ID)

		} else if err == nil {
			fmt.Printf(" 已存在，跳过\n")
			stats.ImagesSkipped++
		} else {
			fmt.Printf(" 查询失败: %v\n", err)
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
