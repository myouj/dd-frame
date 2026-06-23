package excel

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// ParseFormFile 从 multipart/form-data 请求中解析上传的 Excel 文件
//
// fieldName: form 表单字段名
// sheet: 可选工作表名称
//
// 返回解析结果、行级错误和致命错误。
func ParseFormFile[T any](c *gin.Context, fieldName string, sheet ...string) ([]T, RowErrors, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, nil, fmt.Errorf("get form file %q: %w", fieldName, err)
	}

	// 校验文件扩展名
	name := strings.ToLower(file.Filename)
	if !strings.HasSuffix(name, ".xlsx") && !strings.HasSuffix(name, ".xls") {
		return nil, nil, fmt.Errorf("unsupported file type: %s, expected .xlsx or .xls", file.Filename)
	}

	reader, err := file.Open()
	if err != nil {
		return nil, nil, fmt.Errorf("open uploaded file: %w", err)
	}
	defer reader.Close()

	return Parse[T](reader, sheet...)
}

// Download 将 Excel 文件写入 HTTP 响应（浏览器自动下载）
//
// filename: 下载文件名（如 "orders.xlsx"）
// f: excelize.File 实例
func Download(c *gin.Context, filename string, f *excelize.File) {
	// 先写入 buffer，避免流式写入失败后无法发送错误响应
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "generate excel file failed",
		})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buf.Bytes())
}
