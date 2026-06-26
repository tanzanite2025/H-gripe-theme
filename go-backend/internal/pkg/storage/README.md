# 云存储服务实现指南

## 概述

本项目实现了多种云存储服务，包括本地存储、AWS S3和阿里云OSS。所有实现都基于统一的StorageService接口，提供一致的API调用体验。

## 支持的存储类型

### 1. 本地存储 (Local)
**文件**: `storage.go`

**功能**:
- ✅ 文件上传
- ✅ 从Reader上传
- ✅ 文件删除
- ✅ URL生成
- ✅ 按日期组织文件
- ✅ 路径遍历攻击防护

**特点**:
- 无需外部依赖
- 适合开发和测试环境
- 自动创建日期目录结构

### 2. AWS S3
**文件**: `s3.go`

**功能**:
- ✅ 文件上传到S3
- ✅ 从Reader上传
- ✅ 文件删除
- ✅ 预签名URL生成
- ✅ 列出对象
- ✅ 复制对象
- ✅ 自动内容类型检测

**特点**:
- 使用官方`github.com/aws/aws-sdk-go-v2` SDK
- 支持IAM角色和静态凭证
- 支持自定义端点（MinIO等）
- 支持CDN域名

### 3. 阿里云OSS
**文件**: `oss.go`

**功能**:
- ✅ 文件上传到OSS
- ✅ 从Reader上传
- ✅ 文件删除
- ✅ 预签名URL生成
- ✅ 分片上传（大文件）
- ✅ 对象元信息
- ✅ ACL权限设置
- ✅ 对象存在检查

**特点**:
- 使用官方`github.com/aliyun/aliyun-oss-go-sdk` SDK
- 支持分片上传大文件
- 完整的对象管理功能
- 支持CDN加速域名

## 安装依赖

```bash
# AWS S3
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/credentials
go get github.com/aws/aws-sdk-go-v2/service/s3

# 阿里云OSS
go get github.com/aliyun/aliyun-oss-go-sdk/oss

# UUID生成（所有存储都需要）
go get github.com/google/uuid
```

## 环境变量配置

### 通用配置

```env
STORAGE_TYPE=local          # local, s3, oss
STORAGE_BASE_URL=https://cdn.example.com  # 可选：CDN域名
```

### 本地存储

```env
STORAGE_TYPE=local
STORAGE_LOCAL_PATH=./uploads
STORAGE_BASE_URL=http://localhost:8080
```

### AWS S3

```env
STORAGE_TYPE=s3
STORAGE_BUCKET=my-bucket
STORAGE_REGION=us-west-2
STORAGE_ACCESS_KEY_ID=AKIAXXXXXXXXXXXXXXXX
STORAGE_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
# 可选：自定义端点（MinIO等）
STORAGE_ENDPOINT=https://minio.example.com
# 可选：CDN域名
STORAGE_BASE_URL=https://cdn.example.com
```

### 阿里云OSS

```env
STORAGE_TYPE=oss
STORAGE_BUCKET=my-bucket
STORAGE_REGION=cn-hangzhou
STORAGE_ACCESS_KEY_ID=LTAI4xxxxxxxxxx
STORAGE_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxx
# 可选：自定义端点
STORAGE_ENDPOINT=https://oss-cn-hangzhou.aliyuncs.com
# 可选：CDN域名
STORAGE_BASE_URL=https://cdn.example.com
```

## 使用示例

### 基本用法

```go
package main

import (
    "context"
    "log"
    "github.com/yourusername/tanzanite/go-backend/internal/pkg/storage"
)

func main() {
    // 从环境变量加载配置
    config := storage.LoadConfigFromEnv()
    
    // 创建存储服务
    storageService, err := storage.NewStorageService(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // 上传文件
    ctx := context.Background()
    url, err := storageService.Upload(ctx, fileHeader)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("File uploaded: %s", url)
}
```

### 从Reader上传

```go
func uploadFromReader(storageService storage.StorageService, reader io.Reader, filename string) {
    ctx := context.Background()
    url, err := storageService.UploadFromReader(ctx, reader, filename)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("File uploaded: %s\n", url)
}
```

### 文件验证

```go
// 验证图片文件
maxSize := int64(5 * 1024 * 1024) // 5MB
err := storage.ValidateFile(fileHeader, maxSize, storage.ImageTypes)
if err != nil {
    return fmt.Errorf("invalid file: %w", err)
}

// 验证文档文件
err = storage.ValidateFile(fileHeader, maxSize, storage.DocTypes)
```

### 删除文件

```go
func deleteFile(storageService storage.StorageService, url string) {
    ctx := context.Background()
    err := storageService.Delete(ctx, url)
    if err != nil {
        log.Printf("Failed to delete: %v", err)
    }
}
```

### S3特有功能

#### 预签名URL（临时访问私有文件）

```go
s3Storage := storageService.(*storage.S3StorageImpl)

// 生成1小时有效的预签名URL
url, err := s3Storage.GetPresignedURL(ctx, filename, time.Hour)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Presigned URL: %s\n", url)
```

#### 列出对象

```go
keys, err := s3Storage.ListObjects(ctx, "uploads/2024/", 100)
if err != nil {
    log.Fatal(err)
}

for _, key := range keys {
    fmt.Println(key)
}
```

### OSS特有功能

#### 分片上传大文件

```go
ossStorage := storageService.(*storage.OSSStorageImpl)

// 5MB每片
chunkSize := int64(5 * 1024 * 1024)
url, err := ossStorage.UploadMultipart(ctx, reader, filename, chunkSize)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Multipart upload completed: %s\n", url)
```

#### 设置对象ACL

```go
// 设置为公开读
err := ossStorage.SetObjectACL(ctx, key, oss.ACLPublicRead)
if err != nil {
    log.Fatal(err)
}
```

#### 检查对象是否存在

```go
exists, err := ossStorage.IsObjectExist(ctx, key)
if err != nil {
    log.Fatal(err)
}

if exists {
    fmt.Println("Object exists")
}
```

## 文件结构

```
go-backend/internal/pkg/storage/
├── storage.go      # 核心接口和本地存储
├── s3.go           # AWS S3实现
├── oss.go          # 阿里云OSS实现
└── README.md       # 本文档
```

## 支持的文件类型

### 图片
`.jpg`, `.jpeg`, `.png`, `.gif`, `.webp`, `.svg`, `.ico`

### 视频
`.mp4`, `.avi`, `.mov`, `.wmv`, `.flv`, `.webm`

### 音频
`.mp3`, `.wav`, `.ogg`

### 文档
`.pdf`, `.doc`, `.docx`, `.xls`, `.xlsx`, `.ppt`, `.pptx`

### 压缩包
`.zip`, `.rar`, `.7z`, `.tar`, `.gz`

## 最佳实践

### 1. 文件验证

始终在上传前验证文件：

```go
// 定义最大文件大小
const MaxImageSize = 10 * 1024 * 1024 // 10MB
const MaxDocSize = 50 * 1024 * 1024   // 50MB

// 验证上传文件
if err := storage.ValidateFile(file, MaxImageSize, storage.ImageTypes); err != nil {
    return fmt.Errorf("invalid file: %w", err)
}
```

### 2. 错误处理

妥善处理上传失败：

```go
url, err := storageService.Upload(ctx, file)
if err != nil {
    log.Printf("Upload failed: %v", err)
    return err
}

// 保存URL到数据库
if err := saveToDatabase(url); err != nil {
    // 上传成功但保存失败，清理已上传的文件
    storageService.Delete(ctx, url)
    return err
}
```

### 3. 使用CDN

在生产环境配置CDN域名：

```env
STORAGE_BASE_URL=https://cdn.yourdomain.com
```

### 4. 安全性

- 使用IAM角色而非硬编码密钥（AWS）
- 设置合适的对象ACL权限
- 启用HTTPS传输
- 定期清理过期文件

### 5. 性能优化

- 对大文件使用分片上传
- 使用预签名URL避免代理下载
- 启用CDN加速静态资源
- 配置合适的缓存策略

## 迁移指南

### 从本地存储迁移到S3

1. 更新环境变量：
```env
STORAGE_TYPE=s3
STORAGE_BUCKET=my-bucket
STORAGE_REGION=us-west-2
```

2. 迁移现有文件（可选）：
```bash
aws s3 sync ./uploads s3://my-bucket/
```

3. 重启应用，新文件将自动上传到S3

### 从S3迁移到OSS

1. 使用OSS迁移工具
2. 更新环境变量
3. 验证迁移结果

## 故障排查

### 问题：上传失败

**检查项**:
- 环境变量配置是否正确
- 网络连接是否正常
- 密钥权限是否足够
- Bucket是否存在

### 问题：无法访问上传的文件

**解决方案**:
- 检查对象ACL设置
- 验证CDN配置
- 确认CORS设置（跨域访问）

### 问题：S3预签名URL无法访问

**可能原因**:
- URL已过期
- 时钟不同步
- 权限设置错误

## 参考文档

- [AWS S3文档](https://docs.aws.amazon.com/s3/)
- [阿里云OSS文档](https://help.aliyun.com/product/31815.html)
- [MinIO文档](https://min.io/docs/minio/linux/index.html)

## 更新日志

### 2026-06-26
- ✅ 实现AWS S3存储
- ✅ 实现阿里云OSS存储
- ✅ 添加预签名URL支持
- ✅ 添加分片上传功能
- ✅ 完善文档

---

**状态**: ✅ 生产就绪

**维护者**: Tanzanite Development Team
