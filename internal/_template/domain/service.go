package domain

// DomainServiceInterface 领域服务接口示例
//
// 当逻辑不适合放在单个聚合根上时（跨聚合操作），使用领域服务。
type DomainServiceInterface interface {
	// DoSomething 执行跨聚合操作
	DoSomething(param string) (string, error)
}
