package storage

import (
	"context"
	"io"
)

// Store 文件存储抽象接口
//
// 所有存储后端（本地磁盘、OSS、S3 等）均需实现此接口。
type Store interface {
	// Upload 上传文件，key 为存储路径（如 "avatar/2026/06/21/xxx.jpg"）
	// 返回实际存储的 key 和 error
	Upload(ctx context.Context, key string, reader io.Reader) (string, error)

	// Download 下载文件，返回 io.ReadCloser（调用者需关闭）
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, key string) error

	// URL 返回文件的访问 URL
	URL(key string) string
}
