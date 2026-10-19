package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/database"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

// SettingImport 导入的设置结构
type SettingImport struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Group       string `json:"group"`
	Locale      string `json:"locale"`
	IsPublic    bool   `json:"is_public"`
	Description string `json:"description"`
}

func main() {
	// 命令行参数
	dryRun := flag.Bool("dry-run", false, "试运行模式，不实际写入数据库")
	inputFile := flag.String("input", "data/import/settings.json", "输入 JSON 文件路径")
	flag.Parse()

	fmt.Println("========================================")
	fmt.Println("设置数据导入工具")
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
	var importSettings []SettingImport
	if err := json.Unmarshal(data, &importSettings); err != nil {
		log.Fatalf("❌ 解析 JSON 失败: %v", err)
	}

	fmt.Printf("✓ 成功读取 %d 条设置\n", len(importSettings))
	fmt.Println()

	// 统计信息
	stats := map[string]int{
		"total":   0,
		"created": 0,
		"updated": 0,
		"skipped": 0,
	}
	groupStats := make(map[string]int)

	// 创建 repository
	settingRepo := repository.NewSettingRepository(db)

	if !*dryRun {
		storageConfig := storage.LoadConfigFromEnv()
		if _, configured := os.LookupEnv("STORAGE_BASE_URL"); !configured {
			storageConfig.BaseURL = cfg.Server.BaseURL
		}
		storageService, err := storage.NewStorageService(storageConfig)
		if err != nil {
			log.Fatalf("initialize settings import media identity validation: %v", err)
		}
		storefrontBaseURL := strings.TrimSpace(os.Getenv("STOREFRONT_BASE_URL"))
		if storefrontBaseURL == "" {
			storefrontBaseURL = cfg.Server.BaseURL
		}
		settingRepo.ConfigurePublicBusinessMediaReferenceIdentityResolver(service.NewMediaService(repository.NewMediaRepository(db), storageService, nil, storefrontBaseURL, cfg.MediaUpload.AccountStorageQuotaBytes))
	}

	// The entire file is one mutation, including all media-bearing settings.
	fmt.Println("开始导入...")
	fmt.Println()

	for i, importSetting := range importSettings {
		stats["total"]++
		groupStats[importSetting.Group]++

		// 显示进度
		if (i+1)%10 == 0 || i == len(importSettings)-1 {
			fmt.Printf("  进度: %d/%d (%.1f%%)\r", i+1, len(importSettings), float64(i+1)/float64(len(importSettings))*100)
		}

		if *dryRun {
			fmt.Printf("  [试运行] %s.%s = %s\n", importSetting.Group, importSetting.Key, truncate(importSetting.Value, 50))
		}
	}
	if !*dryRun {
		importStats, err := importSettingsWithAtomicRegisteredMediaValidation(settingRepo, importSettings)
		if err != nil {
			log.Fatalf("settings import rolled back: %v", err)
		}
		stats = importStats
	}

	fmt.Println()
	fmt.Println()

	// 输出统计信息
	fmt.Println("========================================")
	fmt.Println("导入完成")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Printf("总计: %d 条设置\n", stats["total"])
	fmt.Printf("  - 新建: %d\n", stats["created"])
	fmt.Printf("  - 更新: %d\n", stats["updated"])
	fmt.Printf("  - 跳过: %d\n", stats["skipped"])
	fmt.Println()

	fmt.Println("按分组统计:")
	for group, count := range groupStats {
		fmt.Printf("  - %s: %d\n", group, count)
	}
	fmt.Println()

	if *dryRun {
		fmt.Println("⚠️  这是试运行，没有实际写入数据库")
		fmt.Println("仅预览内容，不代表媒体登记、复用资格或整批保存校验通过")
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
