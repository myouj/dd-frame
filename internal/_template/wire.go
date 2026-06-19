package _template

import (
	"github.com/example/dd-frame/internal/_template/api"
	"github.com/example/dd-frame/internal/_template/biz"
	"github.com/example/dd-frame/internal/_template/model"
	"github.com/example/dd-frame/internal/_template/service"
)

// Wire 模块内 IoC 装配
//
// 创建本模块所有依赖并返回 API handler。
// 在 app/wire.go 中调用此函数注册模块。
func Wire() *api.EntityAPI {
	// 1. 数据层
	repo := model.NewEntityDAO()

	// 2. 业务编排层
	svc := biz.NewEntityService(repo)

	// 3. 应用边界层
	appSvc := service.NewEntityAppService(svc)

	// 4. API 层
	return api.NewEntityAPI(appSvc)
}
