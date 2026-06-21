package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/uuid"
)

// OSSStore 阿里云 OSS 存储实现
type OSSStore struct {
	client *oss.Client
	bucket *oss.Bucket
	prefix string // 对象 key 前缀（如 "uploads/"）
	cdnURL string // CDN 或 Bucket 公开访问 URL
}

// OSSConfig OSS 配置
type OSSConfig struct {
	Endpoint        string
	AccessKeyID     string
	AccessKeySecret string
	Bucket          string
	Prefix          string
	CDNURL          string // 可选，CDN 加速域名
}

// NewOSSStore 创建阿里云 OSS 存储实例
func NewOSSStore(cfg OSSConfig) (*OSSStore, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("create oss client failed: %w", err)
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("get oss bucket failed: %w", err)
	}

	cdnURL := cfg.CDNURL
	if cdnURL == "" {
		// 默认使用 OSS 公开 URL
		cdnURL = fmt.Sprintf("https://%s.%s", cfg.Bucket, cfg.Endpoint)
	}

	return &OSSStore{
		client: client,
		bucket: bucket,
		prefix: cfg.Prefix,
		cdnURL: cdnURL,
	}, nil
}

// Upload 上传文件到 OSS
func (s *OSSStore) Upload(_ context.Context, key string, reader io.Reader) (string, error) {
	if key == "" {
		key = s.generateKey("")
	}

	objectKey := s.objectKey(key)

	// 读取到 buffer（OSS SDK 需要 io.Reader 但有些场景需要 seek）
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("read upload data failed: %w", err)
	}

	if err := s.bucket.PutObject(objectKey, bytes.NewReader(data)); err != nil {
		return "", fmt.Errorf("upload to oss failed: %w", err)
	}

	return key, nil
}

// Download 从 OSS 下载文件
func (s *OSSStore) Download(_ context.Context, key string) (io.ReadCloser, error) {
	objectKey := s.objectKey(key)

	body, err := s.bucket.GetObject(objectKey)
	if err != nil {
		return nil, fmt.Errorf("download from oss failed: %w", err)
	}

	return body, nil
}

// Delete 删除 OSS 文件
func (s *OSSStore) Delete(_ context.Context, key string) error {
	objectKey := s.objectKey(key)

	if err := s.bucket.DeleteObject(objectKey); err != nil {
		return fmt.Errorf("delete from oss failed: %w", err)
	}

	return nil
}

// URL 返回文件的公开访问 URL
func (s *OSSStore) URL(key string) string {
	return s.cdnURL + "/" + s.objectKey(key)
}

func (s *OSSStore) objectKey(key string) string {
	if s.prefix != "" {
		return filepath.Join(s.prefix, key)
	}
	return key
}

func (s *OSSStore) generateKey(ext string) string {
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	name := uuid.New().String()
	if ext != "" {
		name += ext
	}
	return filepath.Join(dateDir, name)
}
