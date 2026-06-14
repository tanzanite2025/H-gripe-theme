package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// WordPressTranslation WordPress 导出的翻译数据结构
type WordPressTranslation struct {
	WordPressPostID    int    `json:"wordpress_post_id"`
	TranslationGroupID int    `json:"translation_group_id"`
	Locale             string `json:"locale"`
	Slug               string `json:"slug"`
	Title              string `json:"title"`
	Status             string `json:"status"`
	PublishedAt        string `json:"published_at"`
	ModifiedAt         string `json:"modified_at"`
	MetaKeywords       string `json:"meta_keywords"`
	CanonicalURL       string `json:"canonical_url"`
}

// ExportData WordPress 导出的完整数据结构
type ExportData struct {
	ExportDate       string                            `json:"export_date"`
	WordPressVersion string                            `json:"wordpress_version"`
	SiteURL          string                            `json:"site_url"`
	Stats            map[string]interface{}            `json:"stats"`
	LocaleStats      map[string]int                    `json:"locale_stats"`
	Translations     []WordPressTranslation            `json:"translations"`
	Groups           map[string][]WordPressTranslation `json:"groups"`
}

// Post Go 后端的 Post 模型（简化版）
type Post struct {
	ID                 uint   `gorm:"primarykey"`
	Slug               string `gorm:"uniqueIndex:idx_slug_locale"`
	Locale             string `gorm:"uniqueIndex:idx_slug_locale"`
	TranslationGroupID *uint  `gorm:"index"`
	MetaKeywords       string
	CanonicalURL       string
	UpdatedAt          time.Time
}

func main() {
	// 命令行参数
	var (
		jsonFile   = flag.String("file", "exports/blog-translations.json", "导出的 JSON 文件路径")
		dbHost     = flag.String("db-host", "localhost", "数据库主机")
		dbPort     = flag.Int("db-port", 5432, "数据库端口")
		dbUser     = flag.String("db-user", "tanzanite", "数据库用户名")
		dbPassword = flag.String("db-password", "", "数据库密码")
		dbName     = flag.String("db-name", "tanzanite", "数据库名称")
		dryRun     = flag.Bool("dry-run", false, "试运行模式（不实际写入数据库）")
	)
	flag.Parse()

	log.Println("========================================")
	log.Println("博客翻译数据导入工具")
	log.Println("========================================")
	log.Println()

	// 读取 JSON 文件
	log.Printf("正在读取文件: %s\n", *jsonFile)
	data, err := ioutil.ReadFile(*jsonFile)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}

	var exportData ExportData
	if err := json.Unmarshal(data, &exportData); err != nil {
		log.Fatalf("解析 JSON 失败: %v", err)
	}

	log.Printf("导出信息:\n")
	log.Printf("  - 导出时间: %s\n", exportData.ExportDate)
	log.Printf("  - WordPress 版本: %s\n", exportData.WordPressVersion)
	log.Printf("  - 站点 URL: %s\n", exportData.SiteURL)
	log.Printf("  - 文章总数: %.0f\n", exportData.Stats["total_posts"])
	log.Printf("  - 翻译组数: %.0f\n", exportData.Stats["total_groups"])
	log.Printf("  - 多语言组: %.0f\n", exportData.Stats["multi_lang_groups"])
	log.Println()

	log.Println("语言分布:")
	for locale, count := range exportData.LocaleStats {
		log.Printf("  - %s: %d 篇\n", locale, count)
	}
	log.Println()

	if *dryRun {
		log.Println("⚠️  试运行模式：不会实际写入数据库")
		log.Println()
	}

	// 连接数据库
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		*dbHost, *dbPort, *dbUser, *dbPassword, *dbName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}

	log.Println("数据库连接成功")
	log.Println()

	// 开始导入
	if err := importTranslations(db, exportData.Translations, *dryRun); err != nil {
		log.Fatalf("导入失败: %v", err)
	}

	log.Println()
	log.Println("========================================")
	log.Println("导入完成！")
	log.Println("========================================")
}

func importTranslations(db *gorm.DB, translations []WordPressTranslation, dryRun bool) error {
	log.Println("开始导入翻译数据...")
	log.Println()

	var (
		successCount = 0
		skipCount    = 0
		errorCount   = 0
	)

	// 使用事务
	return db.Transaction(func(tx *gorm.DB) error {
		for i, trans := range translations {
			// 显示进度
			if (i+1)%10 == 0 || i == 0 {
				log.Printf("进度: %d/%d\n", i+1, len(translations))
			}

			// 查找对应的文章（通过 slug 和 locale）
			var post Post
			err := tx.Table("posts").
				Select("id, slug, locale, translation_group_id").
				Where("slug = ? AND locale = ?", trans.Slug, trans.Locale).
				First(&post).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					log.Printf("  ⚠️  未找到文章: slug=%s locale=%s (WordPress ID=%d)\n",
						trans.Slug, trans.Locale, trans.WordPressPostID)
					skipCount++
					continue
				}
				log.Printf("  ❌ 查询失败: %v\n", err)
				errorCount++
				continue
			}

			// 准备更新数据
			groupID := uint(trans.TranslationGroupID)
			updates := map[string]interface{}{
				"translation_group_id": groupID,
			}

			// 如果有 SEO 元数据，也更新
			if trans.MetaKeywords != "" {
				updates["meta_keywords"] = trans.MetaKeywords
			}
			if trans.CanonicalURL != "" {
				updates["canonical_url"] = trans.CanonicalURL
			}

			// 试运行模式：只显示，不实际更新
			if dryRun {
				log.Printf("  [DRY-RUN] 将更新: post_id=%d, translation_group_id=%d\n",
					post.ID, groupID)
				successCount++
				continue
			}

			// 更新数据库
			err = tx.Table("posts").
				Where("id = ?", post.ID).
				Updates(updates).Error

			if err != nil {
				log.Printf("  ❌ 更新失败: post_id=%d, error=%v\n", post.ID, err)
				errorCount++
				continue
			}

			successCount++
		}

		log.Println()
		log.Println("导入统计:")
		log.Printf("  ✅ 成功: %d\n", successCount)
		log.Printf("  ⚠️  跳过: %d\n", skipCount)
		log.Printf("  ❌ 失败: %d\n", errorCount)

		// 如果是试运行，回滚事务
		if dryRun {
			log.Println()
			log.Println("试运行完成，未实际写入数据库")
			return gorm.ErrInvalidTransaction // 回滚
		}

		return nil
	})
}
