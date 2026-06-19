package api

// GRPCHandler Connect RPC handler 模板
//
// 实际使用时实现 proto 生成的接口：
//
// type EntityGRPCHandler struct {
//     svc *service.EntityAppService
// }
//
// func (h *EntityGRPCHandler) CreateEntity(
//     ctx context.Context,
//     req *connect.Request[pb.CreateEntityRequest],
// ) (*connect.Response[pb.CreateEntityResponse], error) {
//     // 1. proto DTO → service DTO
//     // 2. 调用 svc
//     // 3. service DTO → proto DTO
// }
