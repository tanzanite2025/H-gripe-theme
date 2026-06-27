package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"

	"gorm.io/gorm"
)

// TicketExportData WordPress 导出的工单数据结构
type TicketExportData struct {
	Tickets []struct {
		ID           int        `json:"id"`
		TicketNumber string     `json:"ticket_number"`
		UserID       int        `json:"user_id"`
		Subject      string     `json:"subject"`
		Category     string     `json:"category"`
		Priority     string     `json:"priority"`
		Status       string     `json:"status"`
		AssignedTo   *int       `json:"assigned_to"`
		Tags         string     `json:"tags"`
		CreatedAt    time.Time  `json:"created_at"`
		UpdatedAt    time.Time  `json:"updated_at"`
		ResolvedAt   *time.Time `json:"resolved_at"`
		ClosedAt     *time.Time `json:"closed_at"`
	} `json:"tickets"`
	Messages []struct {
		TicketID    int       `json:"ticket_id"`
		UserID      int       `json:"user_id"`
		IsStaff     bool      `json:"is_staff"`
		Content     string    `json:"content"`
		Attachments string    `json:"attachments"`
		IsInternal  bool      `json:"is_internal"`
		CreatedAt   time.Time `json:"created_at"`
	} `json:"messages"`
	Stats struct {
		TotalTickets  int    `json:"total_tickets"`
		TotalMessages int    `json:"total_messages"`
		ExportDate    string `json:"export_date"`
	} `json:"stats"`
}

var (
	dryRun     = flag.Bool("dry-run", false, "试运行模式，不实际写入数据库")
	inputFile  = flag.String("input", "scripts/wordpress-export/export/tickets.json", "输入文件路径")
	configFile = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	fmt.Println("=== 工单数据导入工具 ===")
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

	fmt.Printf("读取到 %d 个工单，%d 条消息\n", data.Stats.TotalTickets, data.Stats.TotalMessages)
	fmt.Printf("导出时间: %s\n\n", data.Stats.ExportDate)

	if *dryRun {
		fmt.Println("=== 试运行模式 - 预览数据 ===")
		fmt.Println()
		previewData(data)
		fmt.Println("\n试运行完成，未写入数据库")
		return
	}

	// 导入数据
	stats := importData(db, data)

	fmt.Println("\n=== 导入完成 ===")
	fmt.Printf("工单: 创建 %d, 更新 %d, 跳过 %d\n",
		stats.TicketsCreated, stats.TicketsUpdated, stats.TicketsSkipped)
	fmt.Printf("消息: 创建 %d, 跳过 %d\n",
		stats.MessagesCreated, stats.MessagesSkipped)
}

// readExportFile 读取导出文件
func readExportFile(filename string) (*TicketExportData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer func() { _ = file.Close() }()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var data TicketExportData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &data, nil
}

// previewData 预览数据
func previewData(data *TicketExportData) {
	for i, t := range data.Tickets {
		// 统计该工单的消息数量
		messageCount := 0
		for _, msg := range data.Messages {
			if msg.TicketID == t.ID {
				messageCount++
			}
		}

		fmt.Printf("%d. 工单: %s - %s\n", i+1, t.TicketNumber, t.Subject)
		fmt.Printf("   分类: %s, 优先级: %s, 状态: %s\n", t.Category, t.Priority, t.Status)
		fmt.Printf("   消息数量: %d\n\n", messageCount)
	}
}

// ImportStats 导入统计
type ImportStats struct {
	TicketsCreated  int
	TicketsUpdated  int
	TicketsSkipped  int
	MessagesCreated int
	MessagesSkipped int
}

// importData 导入数据
func importData(db *gorm.DB, data *TicketExportData) ImportStats {
	stats := ImportStats{}

	// WordPress ID 到 Go ID 的映射
	ticketIDMap := make(map[int]uint)

	// 导入工单
	fmt.Println("导入工单...")
	for _, t := range data.Tickets {
		fmt.Printf("  处理: %s...", t.TicketNumber)

		// 检查是否已存在
		var existing ticket.Ticket
		err := db.Where("ticket_number = ?", t.TicketNumber).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新工单
			newTicket := ticket.Ticket{
				TicketNumber: t.TicketNumber,
				UserID:       uint(t.UserID),
				Subject:      t.Subject,
				Category:     t.Category,
				Priority:     t.Priority,
				Status:       t.Status,
				Tags:         t.Tags,
				CreatedAt:    t.CreatedAt,
				UpdatedAt:    t.UpdatedAt,
				ResolvedAt:   t.ResolvedAt,
				ClosedAt:     t.ClosedAt,
			}

			if t.AssignedTo != nil {
				newTicket.AssignedTo = uint(*t.AssignedTo)
			}

			if err := db.Create(&newTicket).Error; err != nil {
				fmt.Printf(" 失败: %v\n", err)
				stats.TicketsSkipped++
				continue
			}

			ticketIDMap[t.ID] = newTicket.ID
			stats.TicketsCreated++
			fmt.Printf(" 创建 (ID: %d)\n", newTicket.ID)

		} else if err == nil {
			// 更新现有工单
			existing.Subject = t.Subject
			existing.Category = t.Category
			existing.Priority = t.Priority
			existing.Status = t.Status
			existing.Tags = t.Tags
			existing.UpdatedAt = t.UpdatedAt
			existing.ResolvedAt = t.ResolvedAt
			existing.ClosedAt = t.ClosedAt

			if t.AssignedTo != nil {
				existing.AssignedTo = uint(*t.AssignedTo)
			}

			if err := db.Save(&existing).Error; err != nil {
				fmt.Printf(" 更新失败: %v\n", err)
				stats.TicketsSkipped++
				continue
			}

			ticketIDMap[t.ID] = existing.ID
			stats.TicketsUpdated++
			fmt.Printf(" 更新 (ID: %d)\n", existing.ID)

		} else {
			fmt.Printf(" 查询失败: %v\n", err)
			stats.TicketsSkipped++
		}
	}

	// 导入消息
	fmt.Println("\n导入工单消息...")
	for _, msg := range data.Messages {
		// 获取对应的 Go 工单 ID
		goTicketID, ok := ticketIDMap[msg.TicketID]
		if !ok {
			fmt.Printf("  跳过消息 (工单 ID %d 不存在)\n", msg.TicketID)
			stats.MessagesSkipped++
			continue
		}

		fmt.Printf("  添加消息 (工单 ID: %d)...", msg.TicketID)

		// 检查是否已存在（通过工单ID、用户ID和创建时间）
		var existing ticket.TicketMessage
		err := db.Where("ticket_id = ? AND user_id = ? AND created_at = ?",
			goTicketID, msg.UserID, msg.CreatedAt).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新消息
			newMessage := ticket.TicketMessage{
				TicketID:    goTicketID,
				UserID:      uint(msg.UserID),
				IsStaff:     msg.IsStaff,
				Content:     msg.Content,
				Attachments: msg.Attachments,
				IsInternal:  msg.IsInternal,
				CreatedAt:   msg.CreatedAt,
			}

			if err := db.Create(&newMessage).Error; err != nil {
				fmt.Printf(" 失败: %v\n", err)
				stats.MessagesSkipped++
				continue
			}

			stats.MessagesCreated++
			fmt.Printf(" 创建 (ID: %d)\n", newMessage.ID)

		} else if err == nil {
			fmt.Printf(" 已存在，跳过\n")
			stats.MessagesSkipped++
		} else {
			fmt.Printf(" 查询失败: %v\n", err)
			stats.MessagesSkipped++
		}
	}

	return stats
}
