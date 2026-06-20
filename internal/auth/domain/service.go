package domain

// PasswordHasher 密码加密端口接口
//
// 实现层在 biz 层（bcrypt 实现），领域层仅定义契约。
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}
