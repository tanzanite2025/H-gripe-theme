package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"
	"tanzanite/internal/repository"
	"time"
)

// FAQImport 导入的FAQ结构
type FAQImport struct {
	ID        uint      `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Category  string    `json:"category"`
	Locale    string    `json:"locale"`
	ParentID  *uint     `json:"parent_id"`
	Order     int       `json:"order"`
	Status    string    `json:"status"`
	ViewCount int       `json:"view_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	// 命令行参数
	dryRun := flag.Bool("dry-run", false, "试运行模式，不实际写入数据库")
	inputFile := flag.String("input", "scripts/wordpress-export/export/faqs.json", "输入 JSON 文件路径")
	flag.Parse()

	fmt.Println("========================================")
	fmt.Println("FAQ 数据导入工具")
	fmt.Println("========================================")
	fmt.Println()

	if *dryRun {
		fmt.Println("⚠️  试运行模式 - 不会实际写入数据库")
		fmt.Println()
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ 加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("❌ 数据库连接失败: %v", err)
	}

	// 读取 JSON 文件
	fmt.Printf("📖 读取文件: %s\n", *inputFile)
	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("❌ 读取文件失败: %v", err)
	}

	// 解析 JSON
	var importFAQs []FAQImport
	if err := json.Unmarshal(data, &importFAQs); err != nil {
		log.Fatalf("❌ 解析 JSON 失败: %v", err)
	}

	fmt.Printf("✓ 成功读取 %d 条 FAQ\n", len(importFAQs))
	fmt.Println()

	// 统计信息
	stats := map[string]int{
		"total":   0,
		"created": 0,
		"updated": 0,
		"skipped": 0,
	}
	categoryStats := make(map[string]int)
	localeStats := make(map[string]int)

	// 创建 repository
	faqRepo := repository.NewFAQRepository(db)

	// 导入 FAQ
	fmt.Println("开始导入...")
	fmt.Println()

	for i, importFAQ := range importFAQs {
		stats["total"]++
		categoryStats[importFAQ.Category]++
		localeStats[importFAQ.Locale]++

		// 显示进度
		if (i+1)%10 == 0 || i == len(importFAQs)-1 {
			fmt.Printf("  进度: %d/%d (%.1f%%)\r", i+1, len(importFAQs), float64(i+1)/float64(len(importFAQs))*100)
		}

		// 试运行模式，跳过实际写入
		if *dryRun {
			if i < 5 { // 只显示前5条
				fmt.Printf("  [试运行] [%s] %s\n", importFAQ.Category, truncate(importFAQ.Question, 50))
			}
			continue
		}

		// 检查是否已存在（通过 question 和 locale）
		existing, err := faqRepo.FindByID(importFAQ.ID)

		// 创建 FAQ 对象
		f := &faq.FAQ{
			Question:  importFAQ.Question,
			Answer:    importFAQ.Answer,
			Category:  importFAQ.Category,
			Locale:    importFAQ.Locale,
			ParentID:  importFAQ.ParentID,
			Order:     importFAQ.Order,
			Status:    importFAQ.Status,
			ViewCount: importFAQ.ViewCount,
		}

		if err != nil {
			// 不存在，创建新记录
			if err := faqRepo.Create(f); err != nil {
				log.Printf("❌ 创建 FAQ 失败 [%s]: %v", importFAQ.Question, err)
				stats["skipped"]++
			} else {
				stats["created"]++
			}
		} else {
			// 已存在，更新
			existing.Question = importFAQ.Question
			existing.Answer = importFAQ.Answer
			existing.Category = importFAQ.Category
			existing.Locale = importFAQ.Locale
			existing.ParentID = importFAQ.ParentID
			existing.Order = importFAQ.Order
			existing.Status = importFAQ.Status
			existing.ViewCount = importFAQ.ViewCount

			if err := faqRepo.Update(existing); err != nil {
				log.Printf("❌ 更新 FAQ 失败 [%s]: %v", importFAQ.Question, err)
				stats["skipped"]++
			} else {
				stats["updated"]++
			}
		}
	}

	fmt.Println()
	fmt.Println()

	// 输出统计信息
	fmt.Println("========================================")
	fmt.Println("导入完成")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("总计: %d 条 FAQ\n", stats["total"])
	fmt.Printf("  - 新建: %d\n", stats["created"])
	fmt.Printf("  - 更新: %d\n", stats["updated"])
	fmt.Printf("  - 跳过: %d\n", stats["skipped"])
	fmt.Println()

	fmt.Println("按分类统计:")
	for category, count := range categoryStats {
		fmt.Printf("  - %s: %d\n", category, count)
	}
	fmt.Println()

	fmt.Println("按语言统计:")
	for locale, count := range localeStats {
		fmt.Printf("  - %s: %d\n", locale, count)
	}
	fmt.Println()

	if *dryRun {
		fmt.Println("⚠️  这是试运行，没有实际写入数据库")
		fmt.Println("移除 --dry-run 参数以执行实际导入")
	} else {
		fmt.Println("✓ 数据已成功导入到数据库")
	}
	fmt.Println("========================================")
}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
