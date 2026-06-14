package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"tanzanite/internal/domain/subscription"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"
)

type SubscriptionImport struct {
	ID             int        `json:"id"`
	Email          string     `json:"email"`
	Status         string     `json:"status"`
	Locale         string     `json:"locale"`
	Source         string     `json:"source"`
	Tags           string     `json:"tags"`
	UnsubToken     string     `json:"unsub_token"`
	SubscribedAt   *time.Time `json:"subscribed_at"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at"`
	CreatedAt      *time.Time `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at"`
}

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 读取 JSON 文件
	dataFile := "data/subscriptions.json"
	if len(os.Args) > 1 {
		dataFile = os.Args[1]
	}

	fmt.Printf("正在从 %s 导入订阅数据...\n", dataFile)

	data, err := ioutil.ReadFile(dataFile)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	var imports []SubscriptionImport
	if err := json.Unmarshal(data, &imports); err != nil {
		log.Fatalf("Failed to parse JSON: %v", err)
	}

	fmt.Printf("找到 %d 条订阅记录\n", len(imports))

	// 统计
	stats := map[string]int{
		"total":   len(imports),
		"created": 0,
		"updated": 0,
		"skipped": 0,
		"errors":  0,
	}

	// 导入数据
	for i, imp := range imports {
		if i > 0 && i%100 == 0 {
			fmt.Printf("已处理 %d/%d 条记录...\n", i, len(imports))
		}

		// 检查是否已存在
		var existing subscription.Subscription
		result := db.Where("email = ?", imp.Email).First(&existing)

		if result.Error == nil {
			// 记录已存在，更新
			existing.Status = imp.Status
			existing.Locale = imp.Locale
			existing.Source = imp.Source
			existing.Tags = imp.Tags
			existing.UnsubToken = imp.UnsubToken

			if imp.SubscribedAt != nil {
				existing.SubscribedAt = *imp.SubscribedAt
			}
			if imp.UnsubscribedAt != nil {
				existing.UnsubscribedAt = imp.UnsubscribedAt
			}
			if imp.UpdatedAt != nil {
				existing.UpdatedAt = *imp.UpdatedAt
			}

			if err := db.Save(&existing).Error; err != nil {
				log.Printf("Error updating subscription %s: %v", imp.Email, err)
				stats["errors"]++
				continue
			}
			stats["updated"]++
		} else {
			// 创建新记录
			sub := subscription.Subscription{
				Email:      imp.Email,
				Status:     imp.Status,
				Locale:     imp.Locale,
				Source:     imp.Source,
				Tags:       imp.Tags,
				UnsubToken: imp.UnsubToken,
			}

			if imp.SubscribedAt != nil {
				sub.SubscribedAt = *imp.SubscribedAt
			} else {
				sub.SubscribedAt = time.Now()
			}

			if imp.UnsubscribedAt != nil {
				sub.UnsubscribedAt = imp.UnsubscribedAt
			}

			if imp.CreatedAt != nil {
				sub.CreatedAt = *imp.CreatedAt
			} else {
				now := time.Now()
				sub.CreatedAt = now
			}

			if imp.UpdatedAt != nil {
				sub.UpdatedAt = *imp.UpdatedAt
			} else {
				now := time.Now()
				sub.UpdatedAt = now
			}

			if err := db.Create(&sub).Error; err != nil {
				log.Printf("Error creating subscription %s: %v", imp.Email, err)
				stats["errors"]++
				continue
			}
			stats["created"]++
		}
	}

	// 输出统计
	fmt.Println("\n导入完成!")
	fmt.Printf("总计: %d\n", stats["total"])
	fmt.Printf("创建: %d\n", stats["created"])
	fmt.Printf("更新: %d\n", stats["updated"])
	fmt.Printf("跳过: %d\n", stats["skipped"])
	fmt.Printf("错误: %d\n", stats["errors"])
}
