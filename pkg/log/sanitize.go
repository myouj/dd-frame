package log

import "strings"

// SensitiveKeys 敏感字段名列表（小写匹配）
var SensitiveKeys = []string{
	"password",
	"token",
	"secret",
	"authorization",
	"access_key",
	"access_key_secret",
}

// Sanitize 对 keysAndValues 中的敏感字段值进行脱敏
//
// 将 "password" 等 key 对应的 value 替换为 "***"。
// 用于日志输出前过滤敏感信息。
func Sanitize(keysAndValues []interface{}) []interface{} {
	if len(keysAndValues) == 0 {
		return keysAndValues
	}

	result := make([]interface{}, len(keysAndValues))
	copy(result, keysAndValues)

	for i := 0; i < len(result)-1; i += 2 {
		key, ok := result[i].(string)
		if !ok {
			continue
		}
		lowerKey := strings.ToLower(key)
		for _, sk := range SensitiveKeys {
			if strings.Contains(lowerKey, sk) {
				result[i+1] = "***"
				break
			}
		}
	}

	return result
}

// MaskToken 将 Bearer Token 截断为前 6 位 + "***"
//
// 用于日志中记录 Token 时脱敏。
func MaskToken(token string) string {
	if len(token) <= 6 {
		return "***"
	}
	return token[:6] + "***"
}
