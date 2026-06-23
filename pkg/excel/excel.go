package excel

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// excelTag struct tag 名称
const excelTag = "excel"

// --- Option ---

// Option 生成选项
type Option func(*options)

type options struct {
	sheet    string
	template io.Reader
}

// WithSheet 指定工作表名称
func WithSheet(name string) Option {
	return func(o *options) { o.sheet = name }
}

// WithTemplate 使用已有 .xlsx 模板填充数据（保留样式）
func WithTemplate(r io.Reader) Option {
	return func(o *options) { o.template = r }
}

func buildOptions(opts []Option) options {
	o := options{sheet: "Sheet1"}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// --- Column mapping ---

type colMapping struct {
	index    int    // struct field index
	name     string // excel tag value (column header)
	fieldTyp reflect.Type
}

// parseStructColumns 从 struct 类型中提取 excel tag 列映射
func parseStructColumns(t reflect.Type) []colMapping {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var cols []colMapping
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get(excelTag)
		if tag == "" || tag == "-" {
			continue
		}
		cols = append(cols, colMapping{
			index:    i,
			name:     tag,
			fieldTyp: f.Type,
		})
	}
	return cols
}

// --- Parse ---

// Parse 解析 Excel 为结构体切片
//
// reader: .xlsx 文件流
// sheet: 可选工作表名称，默认取第一个 sheet
//
// 返回解析结果、行级错误和致命错误。
// 行级错误（如类型转换失败）不会中断解析，对应行将被跳过。
func Parse[T any](reader io.Reader, sheet ...string) ([]T, RowErrors, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("open excel file: %w", err)
	}
	defer f.Close()

	sheetName := ""
	if len(sheet) > 0 && sheet[0] != "" {
		sheetName = sheet[0]
	} else {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, nil, fmt.Errorf("no sheets found in excel file")
		}
		sheetName = sheets[0]
	}

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, fmt.Errorf("read sheet %q: %w", sheetName, err)
	}
	if len(rows) < 2 {
		return nil, nil, nil // 只有表头或空表
	}

	// 解析表头 → 列索引映射
	header := rows[0]
	colIndexMap := make(map[string]int, len(header))
	for i, h := range header {
		colIndexMap[strings.TrimSpace(h)] = i
	}

	// 提取 struct 列映射
	var zero T
	cols := parseStructColumns(reflect.TypeOf(zero))
	if len(cols) == 0 {
		return nil, nil, fmt.Errorf("no `excel` tags found in struct %T", zero)
	}

	// 校验列头是否匹配
	for _, col := range cols {
		if _, ok := colIndexMap[col.name]; !ok {
			return nil, nil, fmt.Errorf("column %q not found in excel header", col.name)
		}
	}

	var result []T
	var rowErrors RowErrors

	for rowIdx := 1; rowIdx < len(rows); rowIdx++ {
		row := rows[rowIdx]
		item := reflect.New(reflect.TypeOf(zero)).Elem()
		hasError := false

		for _, col := range cols {
			ci := colIndexMap[col.name]
			val := ""
			if ci < len(row) {
				val = strings.TrimSpace(row[ci])
			}

			field := item.Field(col.index)
			if err := setFieldValue(field, val); err != nil {
				rowErrors = append(rowErrors, RowError{
					Row:    rowIdx + 1, // Excel 行号从 1 开始
					Column: col.name,
					Reason: err.Error(),
				})
				hasError = true
			}
		}

		if !hasError {
			result = append(result, item.Interface().(T))
		}
	}

	return result, rowErrors, nil
}

// setFieldValue 将字符串值设置到 struct field（支持常见类型）
func setFieldValue(field reflect.Value, val string) error {
	if val == "" {
		return nil // 空值跳过，保持零值
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(val)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(val, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as int: %w", val, err)
		}
		field.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(val, 10, field.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as uint: %w", val, err)
		}
		field.SetUint(n)

	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as float: %w", val, err)
		}
		field.SetFloat(n)

	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return fmt.Errorf("cannot parse %q as bool: %w", val, err)
		}
		field.SetBool(b)

	case reflect.Struct:
		// 特殊处理 time.Time
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t, err := parseTime(val)
			if err != nil {
				return fmt.Errorf("cannot parse %q as time: %w", val, err)
			}
			field.Set(reflect.ValueOf(t))
		} else {
			return fmt.Errorf("unsupported struct type: %s", field.Type())
		}

	case reflect.Ptr:
		// 支持 *string, *int 等指针类型
		elem := reflect.New(field.Type().Elem())
		if err := setFieldValue(elem.Elem(), val); err != nil {
			return err
		}
		field.Set(elem)

	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}
	return nil
}

// parseTime 尝试多种常见时间格式解析
func parseTime(val string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02 15:04:05",
		"2006/01/02",
		"01/02/2006",
	}
	for _, layout := range formats {
		if t, err := time.Parse(layout, val); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized time format")
}

// --- Generate ---

// Generate 从结构体切片生成 Excel 文件
//
// 第一行为表头（取自 struct tag），后续行按序填充数据。
func Generate[T any](data []T, opts ...Option) (*excelize.File, error) {
	o := buildOptions(opts)

	var f *excelize.File
	if o.template != nil {
		var err error
		f, err = excelize.OpenReader(o.template)
		if err != nil {
			return nil, fmt.Errorf("open template: %w", err)
		}
	} else {
		f = excelize.NewFile()
	}

	sheetName := o.sheet
	if o.template == nil {
		// 新建文件时重命名默认 sheet
		idx, _ := f.GetSheetIndex("Sheet1")
		if idx >= 0 && sheetName != "Sheet1" {
			f.SetSheetName("Sheet1", sheetName)
		}
	}

	// 提取列映射
	var zero T
	cols := parseStructColumns(reflect.TypeOf(zero))
	if len(cols) == 0 {
		f.Close()
		return nil, fmt.Errorf("no `excel` tags found in struct %T", zero)
	}

	// 写入表头（模板模式下跳过，假设模板已有表头）
	if o.template == nil {
		for i, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheetName, cell, col.name)
		}
	}

	// 写入数据行
	for rowIdx, item := range data {
		v := reflect.ValueOf(item)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		excelRow := rowIdx + 2 // 第 2 行开始写数据
		for i, col := range cols {
			cell, _ := excelize.CoordinatesToCellName(i+1, excelRow)
			val := fieldToString(v.Field(col.index))
			f.SetCellValue(sheetName, cell, val)
		}
	}

	return f, nil
}

// fieldToString 将 struct field 值转为字符串
func fieldToString(v reflect.Value) string {
	// 处理指针
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	// 特殊处理 time.Time
	if v.Type() == reflect.TypeOf(time.Time{}) {
		return v.Interface().(time.Time).Format("2006-01-02 15:04:05")
	}

	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}
