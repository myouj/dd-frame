package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LocalStore 本地磁盘存储实现
type LocalStore struct {
	BaseDir string // 存储根目录（如 ./uploads）
	BaseURL string // 访问 URL 前缀（如 /uploads）
}

// NewLocalStore 创建本地存储实例
func NewLocalStore(baseDir, baseURL string) (*LocalStore, error) {
	if baseDir == "" {
		baseDir = "./uploads"
	}
	if baseURL == "" {
		baseURL = "/uploads"
	}

	// 确保根目录存在
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir failed: %w", err)
	}

	return &LocalStore{BaseDir: baseDir, BaseURL: baseURL}, nil
}

// safePath 将 key 解析为绝对路径，并验证其位于 BaseDir 内。
// 防止路径遍历攻击（如 key = "../../etc/cron.d/malicious"）。
func (s *LocalStore) safePath(key string) (string, error) {
	fullPath := filepath.Join(s.BaseDir, key)
	absBase, err := filepath.Abs(s.BaseDir)
	if err != nil {
		return "", fmt.Errorf("resolve base dir failed: %w", err)
	}
	absFull, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve file path failed: %w", err)
	}
	if !strings.HasPrefix(absFull, absBase+string(os.PathSeparator)) && absFull != absBase {
		return "", fmt.Errorf("invalid key: path traversal detected")
	}
	return absFull, nil
}

// Upload 上传文件到本地磁盘
//
// key 为相对路径（如 "avatar/photo.jpg"），存储到 BaseDir/key。
// 如果 key 为空，自动生成 UUID 文件名。
func (s *LocalStore) Upload(_ context.Context, key string, reader io.Reader) (string, error) {
	if key == "" {
		key = s.generateKey("")
	}

	fullPath, err := s.safePath(key)
	if err != nil {
		return "", err
	}

	// 确保父目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create dir failed: %w", err)
	}

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("create file failed: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("write file failed: %w", err)
	}

	return key, nil
}

// Download 从本地磁盘读取文件
func (s *LocalStore) Download(_ context.Context, key string) (io.ReadCloser, error) {
	fullPath, err := s.safePath(key)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %w", err)
	}
	return f, nil
}

// Delete 删除本地文件
func (s *LocalStore) Delete(_ context.Context, key string) error {
	fullPath, err := s.safePath(key)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file failed: %w", err)
	}
	return nil
}

// URL 返回文件的 HTTP 访问路径
func (s *LocalStore) URL(key string) string {
	return s.BaseURL + "/" + key
}

// generateKey 生成唯一文件路径：日期子目录 + UUID
func (s *LocalStore) generateKey(ext string) string {
	now := time.Now()
	dateDir := now.Format("2006/01/02")
	name := uuid.New().String()
	if ext != "" {
		name += ext
	}
	return filepath.Join(dateDir, name)
}
