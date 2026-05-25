package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"tanzanite/internal/domain/faq"
	"tanzanite/internal/domain/post"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/database"

	"gorm.io/gorm"
)

// WordPress导出的数据结构
type WPUser struct {
	ID         int    `json:"id"`
	Email      string `json:"email"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Role       string `json:"role"`
	Locale     string `json:"locale"`
	Registered string `json:"registered"`
}

type WPPost struct {
	ID              int     `json:"id"`
	Title           string  `json:"title"`
	Slug            string  `json:"slug"`
	Content         string  `json:"content"`
	Excerpt         string  `json:"excerpt"`
	Status          string  `json:"status"`
	AuthorID        int     `json:"author_id"`
	Locale          string  `json:"locale"`
	ParentID        *int    `json:"parent_id"`
	FeaturedImage   string  `json:"featured_image"`
	MetaTitle       string  `json:"meta_title"`
	MetaDescription string  `json:"meta_description"`
	Tags            string  `json:"tags"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
	PublishedAt     *string `json:"published_at"`
}

type WPProduct struct {
	ID               int                  `json:"id"`
	SKU              string               `json:"sku"`
	Name             string               `json:"name"`
	Slug             string               `json:"slug"`
	Description      string               `json:"description"`
	ShortDescription string               `json:"short_description"`
	Price            float64              `json:"price"`
	SalePrice        *float64             `json:"sale_price"`
	Stock            int                  `json:"stock"`
	WeightGrams      int                  `json:"weight_grams"`
	Status           string               `json:"status"`
	Locale           string               `json:"locale"`
	ParentID         *int                 `json:"parent_id"`
	Featured         bool                 `json:"featured"`
	MetaTitle        string               `json:"meta_title"`
	MetaDescription  string               `json:"meta_description"`
	Images           []WPProductImage     `json:"images"`
	CreatedAt        string               `json:"created_at"`
	UpdatedAt        string               `json:"updated_at"`
}

type WPProductImage struct {
	URL   string `json:"url"`
	Alt   string `json:"alt"`
	Order int    `json:"order"`
}

type WPSetting struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Type   string `json:"type"`
	Group  string `json:"group"`
	Locale string `json:"locale"`
}

type WPFAQ struct {
	ID        int     `json:"id"`
	Question  string  `json:"question"`
	Answer    string  `json:"answer"`
	Category  string  `json:"category"`
	Locale    string  `json:"locale"`
	ParentID  *int    `json:"parent_id"`
	Order     int     `json:"order"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

func main() {
	fmt.Println("🔄 Starting WordPress data import...")

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 连接数据库
	db, err := database.Init(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	exportDir := "./scripts/export"

	// 导入数据
	if err := importUsers(db, exportDir); err != nil {
		log.Printf("⚠️  Failed to import users: %v", err)
	}

	if err := importPosts(db, exportDir); err != nil {
		log.Printf("⚠️  Failed to import posts: %v", err)
	}

	if err := importProducts(db, exportDir); err != nil {
		log.Printf("⚠️  Failed to import products: %v", err)
	}

	if err := importSettings(db, exportDir); err != nil {
		log.Printf("⚠️  Failed to import settings: %v", err)
	}

	if err := importFAQs(db, exportDir); err != nil {
		log.Printf("⚠️  Failed to import FAQs: %v", err)
	}

	fmt.Println("✅ Data import completed!")
}

func importUsers(db *gorm.DB, exportDir string) error {
	fmt.Println("📥 Importing users...")

	data, err := ioutil.ReadFile(exportDir + "/users.json")
	if err != nil {
		return err
	}

	var wpUsers []WPUser
	if err := json.Unmarshal(data, &wpUsers); err != nil {
		return err
	}

	for _, wu := range wpUsers {
		u := &user.User{
			ID:        uint(wu.ID),
			Email:     wu.Email,
			Username:  wu.Username,
			FirstName: wu.FirstName,
			LastName:  wu.LastName,
			Role:      wu.Role,
			Locale:    wu.Locale,
			Status:    "active",
			Password:  "$2a$10$placeholder", // 需要用户重置密码
		}

		if err := db.Create(u).Error; err != nil {
			log.Printf("Failed to import user %s: %v", wu.Email, err)
		}
	}

	fmt.Printf("✅ Imported %d users\n", len(wpUsers))
	return nil
}

func importPosts(db *gorm.DB, exportDir string) error {
	fmt.Println("📥 Importing posts...")

	data, err := ioutil.ReadFile(exportDir + "/posts.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("⏭️  No posts file found, skipping...")
			return nil
		}
		return err
	}

	var wpPosts []WPPost
	if err := json.Unmarshal(data, &wpPosts); err != nil {
		return err
	}

	for _, wp := range wpPosts {
		createdAt, _ := time.Parse("2006-01-02 15:04:05", wp.CreatedAt)
		updatedAt, _ := time.Parse("2006-01-02 15:04:05", wp.UpdatedAt)

		var publishedAt *time.Time
		if wp.PublishedAt != nil {
			t, _ := time.Parse("2006-01-02 15:04:05", *wp.PublishedAt)
			publishedAt = &t
		}

		var parentID *uint
		if wp.ParentID != nil {
			pid := uint(*wp.ParentID)
			parentID = &pid
		}

		p := &post.Post{
			ID:          uint(wp.ID),
			Title:       wp.Title,
			Slug:        wp.Slug,
			Content:     wp.Content,
			Excerpt:     wp.Excerpt,
			Status:      wp.Status,
			AuthorID:    uint(wp.AuthorID),
			Locale:      wp.Locale,
			ParentID:    parentID,
			FeaturedImg: wp.FeaturedImage,
			MetaTitle:   wp.MetaTitle,
			MetaDesc:    wp.MetaDescription,
			Tags:        wp.Tags,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			PublishedAt: publishedAt,
		}

		if err := db.Create(p).Error; err != nil {
			log.Printf("Failed to import post %s: %v", wp.Slug, err)
		}
	}

	fmt.Printf("✅ Imported %d posts\n", len(wpPosts))
	return nil
}

func importProducts(db *gorm.DB, exportDir string) error {
	fmt.Println("📥 Importing products...")

	data, err := ioutil.ReadFile(exportDir + "/products.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("⏭️  No products file found, skipping...")
			return nil
		}
		return err
	}

	var wpProducts []WPProduct
	if err := json.Unmarshal(data, &wpProducts); err != nil {
		return err
	}

	for _, wp := range wpProducts {
		var parentID *uint
		if wp.ParentID != nil {
			pid := uint(*wp.ParentID)
			parentID = &pid
		}

		p := &product.Product{
			ID:          uint(wp.ID),
			SKU:         wp.SKU,
			Name:        wp.Name,
			Slug:        wp.Slug,
			Description: wp.Description,
			ShortDesc:   wp.ShortDescription,
			Price:       wp.Price,
			SalePrice:   wp.SalePrice,
			Stock:       wp.Stock,
			Weight:      wp.WeightGrams,
			Status:      wp.Status,
			Locale:      wp.Locale,
			ParentID:    parentID,
			Featured:    wp.Featured,
			MetaTitle:   wp.MetaTitle,
			MetaDesc:    wp.MetaDescription,
		}

		if err := db.Create(p).Error; err != nil {
			log.Printf("Failed to import product %s: %v", wp.SKU, err)
			continue
		}

		// 导入产品图片
		for _, img := range wp.Images {
			pi := &product.ProductImage{
				ProductID: p.ID,
				URL:       img.URL,
				Alt:       img.Alt,
				Order:     img.Order,
			}
			db.Create(pi)
		}
	}

	fmt.Printf("✅ Imported %d products\n", len(wpProducts))
	return nil
}

func importSettings(db *gorm.DB, exportDir string) error {
	fmt.Println("📥 Importing settings...")

	data, err := ioutil.ReadFile(exportDir + "/settings.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("⏭️  No settings file found, skipping...")
			return nil
		}
		return err
	}

	var wpSettings []WPSetting
	if err := json.Unmarshal(data, &wpSettings); err != nil {
		return err
	}

	for _, ws := range wpSettings {
		s := &setting.Setting{
			Key:    ws.Key,
			Value:  ws.Value,
			Type:   ws.Type,
			Group:  ws.Group,
			Locale: ws.Locale,
		}

		if err := db.Create(s).Error; err != nil {
			log.Printf("Failed to import setting %s: %v", ws.Key, err)
		}
	}

	fmt.Printf("✅ Imported %d settings\n", len(wpSettings))
	return nil
}

func importFAQs(db *gorm.DB, exportDir string) error {
	fmt.Println("📥 Importing FAQs...")

	data, err := ioutil.ReadFile(exportDir + "/faqs.json")
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("⏭️  No FAQs file found, skipping...")
			return nil
		}
		return err
	}

	var wpFAQs []WPFAQ
	if err := json.Unmarshal(data, &wpFAQs); err != nil {
		return err
	}

	for _, wf := range wpFAQs {
		var parentID *uint
		if wf.ParentID != nil {
			pid := uint(*wf.ParentID)
			parentID = &pid
		}

		f := &faq.FAQ{
			ID:       uint(wf.ID),
			Question: wf.Question,
			Answer:   wf.Answer,
			Category: wf.Category,
			Locale:   wf.Locale,
			ParentID: parentID,
			Order:    wf.Order,
			Status:   wf.Status,
		}

		if err := db.Create(f).Error; err != nil {
			log.Printf("Failed to import FAQ %d: %v", wf.ID, err)
		}
	}

	fmt.Printf("✅ Imported %d FAQs\n", len(wpFAQs))
	return nil
}
