package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/dd-frame/pkg/storage"
)

// UploadResult 文件上传结果
type UploadResult struct {
	Key string `json:"key"` // 存储 key
	URL string `json:"url"` // 访问 URL
}

// defaultMaxFileSize 默认最大文件大小：10MB
const defaultMaxFileSize = 10 << 20

// FileUpload 文件上传中间件
//
// 校验文件大小和 MIME 类型，调用 Store 上传文件，
// 成功后将 UploadResult 注入 gin.Context（key: "upload_result"）。
//
// formFieldName: 表单字段名（如 "file"）
// store: 存储后端实例
// maxSize: 最大文件大小（字节），0 使用默认 10MB
// allowedTypes: 允许的 MIME 类型白名单，空表示不限制
func FileUpload(store storage.Store, maxSize int64, allowedTypes []string) gin.HandlerFunc {
	if maxSize <= 0 {
		maxSize = defaultMaxFileSize
	}

	typeSet := make(map[string]bool, len(allowedTypes))
	for _, t := range allowedTypes {
		typeSet[strings.ToLower(t)] = true
	}

	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    40001,
				"message": "file field 'file' is required",
			})
			c.Abort()
			return
		}
		defer file.Close()

		// 校验文件大小
		if header.Size > maxSize {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    40002,
				"message": fmt.Sprintf("file too large (max %d MB)", maxSize>>20),
			})
			c.Abort()
			return
		}

		// 校验 MIME 类型
		contentType := header.Header.Get("Content-Type")
		if len(typeSet) > 0 {
			if !typeSet[strings.ToLower(contentType)] {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    40003,
					"message": fmt.Sprintf("file type '%s' not allowed", contentType),
				})
				c.Abort()
				return
			}
		}

		// 留空让 Store 自动生成唯一文件名
		uploadedKey, err := store.Upload(c.Request.Context(), "", file)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    50001,
				"message": "file upload failed",
			})
			c.Abort()
			return
		}

		// 注入上传结果
		c.Set("upload_result", &UploadResult{
			Key: uploadedKey,
			URL: store.URL(uploadedKey),
		})

		c.Next()
	}
}

// GetUploadResult 从 gin.Context 获取上传结果
func GetUploadResult(c *gin.Context) *UploadResult {
	v, ok := c.Get("upload_result")
	if !ok {
		return nil
	}
	result, ok := v.(*UploadResult)
	if !ok {
		return nil
	}
	return result
}
