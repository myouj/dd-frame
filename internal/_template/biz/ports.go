package biz

import "context"

// ExternalServiceClient 出站端口接口示例
//
// 定义对外部系统的依赖，由 ACL 防腐层或 infrastructure 实现。
type ExternalServiceClient interface {
	// CallExternal 调用外部服务
	CallExternal(ctx context.Context, param string) (string, error)
}
