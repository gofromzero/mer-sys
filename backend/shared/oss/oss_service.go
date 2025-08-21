package oss

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

// OSSService OSS文件服务
type OSSService struct {
	endpoint    string
	accessKeyID string
	secretKey   string
	bucketName  string
}

// NewOSSService 创建OSS服务实例
func NewOSSService() *OSSService {
	cfg := g.Cfg()
	return &OSSService{
		endpoint:    cfg.MustGet(context.Background(), "oss.endpoint").String(),
		accessKeyID: cfg.MustGet(context.Background(), "oss.accessKeyId").String(),
		secretKey:   cfg.MustGet(context.Background(), "oss.accessKeySecret").String(),
		bucketName:  cfg.MustGet(context.Background(), "oss.bucket").String(),
	}
}

// UploadFileInfo 上传文件信息
type UploadFileInfo struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
	URL         string `json:"url"`
	ETag        string `json:"etag"`
}

// UploadOptions 上传选项
type UploadOptions struct {
	Directory string            // 目录路径，如 "products/images"
	FileName  string            // 自定义文件名，如果为空则生成UUID
	Metadata  map[string]string // 元数据
}

// UploadFile 上传文件
func (s *OSSService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, options *UploadOptions) (*UploadFileInfo, error) {
	if options == nil {
		options = &UploadOptions{}
	}

	// 读取文件内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	// 重置文件指针以便后续使用
	file.Seek(0, io.SeekStart)

	// 生成文件名
	fileName := options.FileName
	if fileName == "" {
		ext := filepath.Ext(header.Filename)
		fileName = fmt.Sprintf("%s%s", guid.S(), ext)
	}

	// 构建完整的对象键
	objectKey := fileName
	if options.Directory != "" {
		objectKey = strings.TrimSuffix(options.Directory, "/") + "/" + fileName
	}

	// 检测文件类型
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = detectContentType(fileName, fileContent)
	}

	// 验证文件类型
	if !isAllowedFileType(contentType) {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	// 验证文件大小 (5MB限制)
	if header.Size > 5*1024*1024 {
		return nil, fmt.Errorf("file size exceeds 5MB limit")
	}

	// TODO: 实际的OSS上传逻辑
	// 这里使用模拟的上传过程，在实际部署中需要替换为真实的阿里云OSS SDK调用
	url := s.simulateUpload(objectKey, contentType)

	g.Log().Infof(ctx, "File uploaded successfully: %s -> %s", header.Filename, url)

	return &UploadFileInfo{
		FileName:    fileName,
		ContentType: contentType,
		Size:        header.Size,
		URL:         url,
		ETag:        generateETag(fileContent),
	}, nil
}

// UploadFromBytes 从字节数组上传文件
func (s *OSSService) UploadFromBytes(ctx context.Context, data []byte, fileName, contentType string, options *UploadOptions) (*UploadFileInfo, error) {
	if options == nil {
		options = &UploadOptions{}
	}

	// 生成文件名
	if options.FileName != "" {
		fileName = options.FileName
	}

	// 构建完整的对象键
	objectKey := fileName
	if options.Directory != "" {
		objectKey = strings.TrimSuffix(options.Directory, "/") + "/" + fileName
	}

	// 检测文件类型
	if contentType == "" {
		contentType = detectContentType(fileName, data)
	}

	// 验证文件类型和大小
	if !isAllowedFileType(contentType) {
		return nil, fmt.Errorf("unsupported file type: %s", contentType)
	}

	if len(data) > 5*1024*1024 {
		return nil, fmt.Errorf("file size exceeds 5MB limit")
	}

	// 模拟上传
	url := s.simulateUpload(objectKey, contentType)

	return &UploadFileInfo{
		FileName:    fileName,
		ContentType: contentType,
		Size:        int64(len(data)),
		URL:         url,
		ETag:        generateETag(data),
	}, nil
}

// DeleteFile 删除文件
func (s *OSSService) DeleteFile(ctx context.Context, objectKey string) error {
	// TODO: 实际的OSS删除逻辑
	g.Log().Infof(ctx, "File deleted: %s", objectKey)
	return nil
}

// GetFileURL 获取文件访问URL
func (s *OSSService) GetFileURL(objectKey string) string {
	// TODO: 根据实际OSS配置生成访问URL
	return fmt.Sprintf("https://%s.%s/%s", s.bucketName, s.endpoint, objectKey)
}

// GetSignedURL 获取签名URL（用于临时访问）
func (s *OSSService) GetSignedURL(ctx context.Context, objectKey string, expiration time.Duration) (string, error) {
	// TODO: 生成签名URL
	baseURL := s.GetFileURL(objectKey)
	signature := fmt.Sprintf("?Expires=%d&OSSAccessKeyId=%s&Signature=mock_signature", 
		time.Now().Add(expiration).Unix(), s.accessKeyID)
	
	return baseURL + signature, nil
}

// simulateUpload 模拟上传过程（实际应替换为OSS SDK调用）
func (s *OSSService) simulateUpload(objectKey, contentType string) string {
	// 在实际实现中，这里应该是真正的阿里云OSS上传逻辑
	return fmt.Sprintf("https://%s.%s/%s", s.bucketName, s.endpoint, objectKey)
}

// detectContentType 检测文件类型
func detectContentType(fileName string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	default:
		// 基于文件内容检测
		if len(content) >= 3 {
			// JPEG
			if content[0] == 0xFF && content[1] == 0xD8 && content[2] == 0xFF {
				return "image/jpeg"
			}
			// PNG
			if len(content) >= 8 && content[0] == 0x89 && content[1] == 0x50 && content[2] == 0x4E && content[3] == 0x47 {
				return "image/png"
			}
			// GIF
			if len(content) >= 6 && content[0] == 0x47 && content[1] == 0x49 && content[2] == 0x46 {
				return "image/gif"
			}
		}
		return "application/octet-stream"
	}
}

// isAllowedFileType 检查是否为允许的文件类型
func isAllowedFileType(contentType string) bool {
	allowedTypes := []string{
		"image/jpeg",
		"image/jpg", 
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}
	
	for _, allowed := range allowedTypes {
		if strings.HasPrefix(contentType, allowed) {
			return true
		}
	}
	
	return false
}

// generateETag 生成ETag（简化实现）
func generateETag(content []byte) string {
	return fmt.Sprintf("\"%x\"", len(content)) // 简化实现，实际应使用MD5
}