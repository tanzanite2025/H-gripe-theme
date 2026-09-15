package storage

import (
	"strings"
	"testing"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

func TestS3ObjectKeyAcceptsPrivateBaseURLWhenPublicBaseURLConfigured(t *testing.T) {
	service := &s3StorageImpl{config: &Config{
		Bucket:         "public-media",
		PrivateBucket:  "private-evidence",
		Region:         "us-east-1",
		BaseURL:        "https://cdn.example.test",
		PrivateBaseURL: "https://private.example.test",
	}}
	key, err := service.ObjectKey("https://private.example.test/order-evidence/42/file.pdf")
	if err != nil {
		t.Fatalf("ObjectKey() error = %v", err)
	}
	if key != "order-evidence/42/file.pdf" {
		t.Fatalf("key = %q, want private object key", key)
	}
	key, err = service.ObjectKey("https://private-evidence.s3.us-east-1.amazonaws.com/warranty/claim.mp4")
	if err != nil {
		t.Fatalf("private standard ObjectKey() error = %v", err)
	}
	if key != "warranty/claim.mp4" {
		t.Fatalf("private standard key = %q, want warranty/claim.mp4", key)
	}
}

func TestOSSObjectKeyAcceptsPrivateBaseURLWhenPublicBaseURLConfigured(t *testing.T) {
	service := &ossStorageImpl{config: &Config{
		Bucket:         "public-media",
		PrivateBucket:  "private-evidence",
		Endpoint:       "https://oss-cn-hangzhou.aliyuncs.com",
		BaseURL:        "https://cdn.example.test",
		PrivateBaseURL: "https://private.example.test",
	}}
	key, err := service.ObjectKey("https://private.example.test/order-evidence/42/file.pdf")
	if err != nil {
		t.Fatalf("ObjectKey() error = %v", err)
	}
	if key != "order-evidence/42/file.pdf" {
		t.Fatalf("key = %q, want private object key", key)
	}
	key, err = service.ObjectKey("https://private-evidence.oss-cn-hangzhou.aliyuncs.com/warranty/claim.mp4")
	if err != nil {
		t.Fatalf("private standard ObjectKey() error = %v", err)
	}
	if key != "warranty/claim.mp4" {
		t.Fatalf("private standard key = %q, want warranty/claim.mp4", key)
	}
}

func TestS3PrivateURLNeverFallsBackToPublicBaseURL(t *testing.T) {
	service := &s3StorageImpl{config: &Config{
		Bucket:        "public-media",
		PrivateBucket: "private-evidence",
		Region:        "us-east-1",
		BaseURL:       "https://cdn.example.test",
	}}
	url := service.GetURL("warranty/claim.mp4")
	if strings.Contains(url, "cdn.example.test") || !strings.Contains(url, "private-evidence") {
		t.Fatalf("private URL = %q, want private bucket URL", url)
	}
}

func TestOSSPrivateURLNeverFallsBackToPublicBaseURL(t *testing.T) {
	service := &ossStorageImpl{
		config:        &Config{Bucket: "public-media", PrivateBucket: "private-evidence", Endpoint: "https://oss-cn-hangzhou.aliyuncs.com", BaseURL: "https://cdn.example.test"},
		privateBucket: &oss.Bucket{},
	}
	url := service.GetURL("warranty/claim.mp4")
	if strings.Contains(url, "cdn.example.test") || !strings.Contains(url, "private-evidence") {
		t.Fatalf("private URL = %q, want private bucket URL", url)
	}
}
