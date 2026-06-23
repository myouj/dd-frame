package excel

import "fmt"

// RowError 行级解析错误
//
// 记录 Excel 中某一行的解析失败信息，包含行号、列名和原因。
type RowError struct {
	Row    int    // Excel 行号（从 2 开始，第 1 行为表头）
	Column string // 列名（struct tag 中定义的名称）
	Reason string // 错误原因
}

func (e *RowError) Error() string {
	return fmt.Sprintf("row %d, column %q: %s", e.Row, e.Column, e.Reason)
}

// RowErrors 行级错误集合
type RowErrors []RowError

func (errs RowErrors) Error() string {
	if len(errs) == 0 {
		return ""
	}
	if len(errs) == 1 {
		return errs[0].Error()
	}
	return fmt.Sprintf("%d errors: first: %s", len(errs), errs[0].Error())
}

// HasErrors 是否有解析错误
func (errs RowErrors) HasErrors() bool {
	return len(errs) > 0
}
