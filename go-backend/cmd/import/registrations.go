package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"tanzanite/internal/domain/registration"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"

	"gorm.io/gorm"
)

// RegistrationExportData WordPress 导出的注册数据结构
type RegistrationExportData struct {
	Registrations []struct {
		ID              int        `json:"id"`
		UserID          int        `json:"user_id"`
		ProductID       int        `json:"product_id"`
		SerialNumber    string     `json:"serial_number"`
		PurchaseDate    time.Time  `json:"purchase_date"`
		PurchaseProof   string     `json:"purchase_proof"`
		Retailer        string     `json:"retailer"`
		WarrantyPeriod  int        `json:"warranty_period"`
		WarrantyExpires time.Time  `json:"warranty_expires"`
		Status          string     `json:"status"`
		Notes           string     `json:"notes"`
		CreatedAt       time.Time  `json:"created_at"`
		UpdatedAt       time.Time  `json:"updated_at"`
	} `json:"registrations"`
	WarrantyClaims []struct {
		ID             int        `json:"id"`
		RegistrationID int        `json:"registration_id"`
		UserID         int        `json:"user_id"`
		IssueType      string     `json:"issue_type"`
		Description    string     `json:"description"`
		Images         string     `json:"images"`
		Status         string     `json:"status"`
		Resolution     string     `json:"resolution"`
		ProcessedBy    *int       `json:"processed_by"`
		ProcessedAt    *time.Time `json:"processed_at"`
		CreatedAt      time.Time  `json:"created_at"`
		UpdatedAt      time.Time  `json:"updated_at"`
	} `json:"warranty_claims"`
	Stats struct {
		TotalRegistrations int    `json:"total_registrations"`
		TotalClaims        int    `json:"total_claims"`
		ExportDate         string `json:"export_date"`
	} `json:"stats"`
}

var (
	dryRun     = flag.Bool("dry-run", false, "试运行模式，不实际写入数据库")
	inputFile  = flag.String("input", "scripts/wordpress-export/export/registrations.json", "输入文件路径")
	configFile = flag.String("config", "config/config.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	fmt.Println("=== 产品注册数据导入工具 ===")
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

	fmt.Printf("读取到 %d 个产品注册，%d 个保修申请\n", data.Stats.TotalRegistrations, data.Stats.TotalClaims)
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
	fmt.Printf("产品注册: 创建 %d, 更新 %d, 跳过 %d\n",
		stats.RegistrationsCreated, stats.RegistrationsUpdated, stats.RegistrationsSkipped)
	fmt.Printf("保修申请: 创建 %d, 跳过 %d\n",
		stats.ClaimsCreated, stats.ClaimsSkipped)
}

// readExportFile 读取导出文件
func readExportFile(filename string) (*RegistrationExportData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var data RegistrationExportData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	return &data, nil
}

// previewData 预览数据
func previewData(data *RegistrationExportData) {
	for i, reg := range data.Registrations {
		fmt.Printf("%d. 注册: %s (用户ID: %d, 产品ID: %d)\n", i+1, reg.SerialNumber, reg.UserID, reg.ProductID)
		fmt.Printf("   购买日期: %s, 保修到期: %s\n", reg.PurchaseDate.Format("2006-01-02"), reg.WarrantyExpires.Format("2006-01-02"))
		fmt.Printf("   状态: %s\n\n", reg.Status)
	}
}

// ImportStats 导入统计
type ImportStats struct {
	RegistrationsCreated int
	RegistrationsUpdated int
	RegistrationsSkipped int
	ClaimsCreated        int
	ClaimsSkipped        int
}

// importData 导入数据
func importData(db *gorm.DB, data *RegistrationExportData) ImportStats {
	stats := ImportStats{}

	// WordPress ID 到 Go ID 的映射
	registrationIDMap := make(map[int]uint)

	// 导入产品注册
	fmt.Println("导入产品注册...")
	for _, reg := range data.Registrations {
		fmt.Printf("  处理: %s...", reg.SerialNumber)

		// 检查是否已存在
		var existing registration.ProductRegistration
		err := db.Where("serial_number = ?", reg.SerialNumber).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新注册
			newReg := registration.ProductRegistration{
				UserID:          uint(reg.UserID),
				ProductID:       uint(reg.ProductID),
				SerialNumber:    reg.SerialNumber,
				PurchaseDate:    reg.PurchaseDate,
				PurchaseProof:   reg.PurchaseProof,
				Retailer:        reg.Retailer,
				WarrantyPeriod:  reg.WarrantyPeriod,
				WarrantyExpires: reg.WarrantyExpires,
				Status:          reg.Status,
				Notes:           reg.Notes,
				CreatedAt:       reg.CreatedAt,
				UpdatedAt:       reg.UpdatedAt,
			}

			if err := db.Create(&newReg).Error; err != nil {
				fmt.Printf(" 失败: %v\n", err)
				stats.RegistrationsSkipped++
				continue
			}

			registrationIDMap[reg.ID] = newReg.ID
			stats.RegistrationsCreated++
			fmt.Printf(" 创建 (ID: %d)\n", newReg.ID)

		} else if err == nil {
			// 更新现有注册
			existing.PurchaseDate = reg.PurchaseDate
			existing.PurchaseProof = reg.PurchaseProof
			existing.Retailer = reg.Retailer
			existing.WarrantyPeriod = reg.WarrantyPeriod
			existing.WarrantyExpires = reg.WarrantyExpires
			existing.Status = reg.Status
			existing.Notes = reg.Notes
			existing.UpdatedAt = reg.UpdatedAt

			if err := db.Save(&existing).Error; err != nil {
				fmt.Printf(" 更新失败: %v\n", err)
				stats.RegistrationsSkipped++
				continue
			}

			registrationIDMap[reg.ID] = existing.ID
			stats.RegistrationsUpdated++
			fmt.Printf(" 更新 (ID: %d)\n", existing.ID)

		} else {
			fmt.Printf(" 查询失败: %v\n", err)
			stats.RegistrationsSkipped++
		}
	}

	// 导入保修申请
	fmt.Println("\n导入保修申请...")
	for _, claim := range data.WarrantyClaims {
		// 获取对应的 Go 注册 ID
		goRegistrationID, ok := registrationIDMap[claim.RegistrationID]
		if !ok {
			fmt.Printf("  跳过申请 (注册 ID %d 不存在)\n", claim.RegistrationID)
			stats.ClaimsSkipped++
			continue
		}

		fmt.Printf("  添加保修申请 (注册 ID: %d)...", claim.RegistrationID)

		// 检查是否已存在（通过注册ID和创建时间）
		var existing registration.WarrantyClaim
		err := db.Where("registration_id = ? AND created_at = ?", goRegistrationID, claim.CreatedAt).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新申请
			newClaim := registration.WarrantyClaim{
				RegistrationID: goRegistrationID,
				UserID:         uint(claim.UserID),
				IssueType:      claim.IssueType,
				Description:    claim.Description,
				Images:         claim.Images,
				Status:         claim.Status,
				Resolution:     claim.Resolution,
				ProcessedAt:    claim.ProcessedAt,
				CreatedAt:      claim.CreatedAt,
				UpdatedAt:      claim.UpdatedAt,
			}

			if claim.ProcessedBy != nil {
				newClaim.ProcessedBy = uint(*claim.ProcessedBy)
			}

			if err := db.Create(&newClaim).Error; err != nil {
				fmt.Printf(" 失败: %v\n", err)
				stats.ClaimsSkipped++
				continue
			}

			stats.ClaimsCreated++
			fmt.Printf(" 创建 (ID: %d)\n", newClaim.ID)

		} else if err == nil {
			fmt.Printf(" 已存在，跳过\n")
			stats.ClaimsSkipped++
		} else {
			fmt.Printf(" 查询失败: %v\n", err)
			stats.ClaimsSkipped++
		}
	}

	return stats
}
