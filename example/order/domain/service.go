package order

// OrderPricingService 订单定价领域服务接口
//
// 定价涉及跨聚合计算（订单 + 商品 + 优惠），适合作为领域服务独立存在。
type OrderPricingService interface {
	CalculatePrice(item OrderItem, quantity int) (Money, error)
}

// OrderNumberGenerator 订单号生成器接口
//
// 订单号生成策略属于领域关注点，但具体实现依赖基础设施。
type OrderNumberGenerator interface {
	Generate() (string, error)
}
