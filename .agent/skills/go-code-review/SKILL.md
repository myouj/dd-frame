---
name: go-code-review
description: Review Go code changes against industry-standard best practices. Use after completing Go feature development, before merging code, or when the user asks to review/audit Go code quality. Covers error handling, concurrency safety, performance, naming conventions, API design, testing, and security.
---

# Go 代码审查

基于 Go 官方 [Code Review Comments](https://go.dev/wiki/CodeReviewComments)、[Effective Go](https://go.dev/doc/effective_go) 及业界最佳实践的代码审查规范。

## 审查原则

1. **只报真实问题**，不报主观偏好
2. **区分严重等级**：Critical > Major > Minor
3. **提供修复代码**，而不只是指出问题
4. **优先关注安全和正确性**，其次是性能和风格

## 审查规则

---

## 一、错误处理 (Error Handling)

### 规则 E-1：忽略错误返回值（Critical）

禁止使用 `_` 忽略 error 返回值，除非明确注释原因。

**违规示例：**

```go
result, _ := DoSomething()  // 违规！忽略了错误
```

**正确做法：**

```go
result, err := DoSomething()
if err != nil {
    return fmt.Errorf("do something failed: %w", err)
}
```

**例外情况（必须注释说明）：**

```go
// 忽略错误：日志写入失败不应影响主流程
_ = logger.Sync()
```

---

### 规则 E-2：错误包装丢失上下文（Major）

使用 `fmt.Errorf` 包装错误时，必须包含 `%w` 动词和上下文信息。

**违规示例：**

```go
if err != nil {
    return err  // 违规！丢失了调用栈上下文
}
```

```go
if err != nil {
    return fmt.Errorf("failed: %v", err)  // 违规！应使用 %w 而非 %v
}
```

**正确做法：**

```go
if err != nil {
    return fmt.Errorf("query user by id %d failed: %w", id, err)
}
```

---

### 规则 E-3：使用 == 比较错误（Major）

应使用 `errors.Is()` 或 `errors.As()` 比较错误，而非 `==`。

**违规示例：**

```go
if err == ErrNotFound {  // 违规！包装后的错误无法匹配
    return nil
}
```

**正确做法：**

```go
if errors.Is(err, ErrNotFound) {
    return nil
}

// 获取特定错误类型
var appErr *AppError
if errors.As(err, &appErr) {
    log.Printf("app error code: %s", appErr.Code)
}
```

---

### 规则 E-4：panic 用于业务错误（Critical）

禁止在业务逻辑中使用 `panic`，应返回 error。

**违规示例：**

```go
func CreateUser(name string) (*User, error) {
    if name == "" {
        panic("name is required")  // 违规！应返回 error
    }
    ...
}
```

**正确做法：**

```go
func CreateUser(name string) (*User, error) {
    if name == "" {
        return nil, fmt.Errorf("name is required")
    }
    ...
}
```

**`panic` 仅允许用于：**
- 程序启动时的配置校验失败
- 不可恢复的程序状态（如 `init()` 函数）

---

## 二、并发安全 (Concurrency Safety)

### 规则 C-1：Goroutine 泄漏（Critical）

所有 goroutine 必须有明确的退出机制。

**违规示例：**

```go
func Process() {
    go func() {
        for data := range ch {  // 违规！如果 ch 永不关闭，goroutine 永远不退出
            process(data)
        }
    }()
}
```

**正确做法：**

```go
func Process(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case data := <-ch:
                process(data)
            }
        }
    }()
}
```

---

### 规则 C-2：竞态条件（Critical）

共享状态必须有同步保护（mutex、channel 或 atomic）。

**违规示例：**

```go
type Counter struct {
    count int  // 违规！无同步保护
}

func (c *Counter) Increment() {
    c.count++  // 竞态条件
}
```

**正确做法：**

```go
type Counter struct {
    mu    sync.Mutex
    count int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// 或使用 atomic
type Counter struct {
    count atomic.Int64
}

func (c *Counter) Increment() {
    c.count.Add(1)
}
```

---

### 规则 C-3：Context 未传播（Major）

函数签名应接收 `context.Context` 作为第一个参数。

**违规示例：**

```go
func QueryUser(id int64) (*User, error) {  // 违规！缺少 context
    return db.Query(id)
}
```

**正确做法：**

```go
func QueryUser(ctx context.Context, id int64) (*User, error) {
    return db.QueryContext(ctx, id)
}
```

---

### 规则 C-4：WaitGroup 使用错误（Major）

`Add()` 必须在 goroutine 启动前调用。

**违规示例：**

```go
var wg sync.WaitGroup
go func() {
    wg.Add(1)  // 违规！Add 在 goroutine 内部，可能在 Wait 之后执行
    defer wg.Done()
    doWork()
}()
wg.Wait()
```

**正确做法：**

```go
var wg sync.WaitGroup
wg.Add(1)  // 在启动 goroutine 前调用
go func() {
    defer wg.Done()
    doWork()
}()
wg.Wait()
```

---

## 三、性能 (Performance)

### 规则 P-1：循环中使用 defer（Major）

`defer` 在函数返回时执行，循环中积累会导致资源泄漏或内存问题。

**违规示例：**

```go
func ProcessFiles(paths []string) error {
    for _, path := range paths {
        f, err := os.Open(path)
        if err != nil {
            return err
        }
        defer f.Close()  // 违规！所有文件在函数结束时才关闭
        process(f)
    }
    return nil
}
```

**正确做法：**

```go
func ProcessFiles(paths []string) error {
    for _, path := range paths {
        if err := processFile(path); err != nil {
            return err
        }
    }
    return nil
}

func processFile(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    defer f.Close()
    return process(f)
}
```

---

### 规则 P-2：字符串拼接使用 +（Minor）

循环或大量拼接应使用 `strings.Builder`。

**违规示例：**

```go
var result string
for _, s := range items {
    result += s  // 违规！每次拼接创建新字符串
}
```

**正确做法：**

```go
var sb strings.Builder
for _, s := range items {
    sb.WriteString(s)
}
result := sb.String()
```

---

### 规则 P-3：Slice 预分配（Minor）

已知长度时应预分配 slice 容量。

**违规示例：**

```go
var result []User
for _, u := range users {
    result = append(result, u)  // 多次扩容
}
```

**正确做法：**

```go
result := make([]User, 0, len(users))
for _, u := range users {
    result = append(result, u)
}
```

---

### 规则 P-4：Map 遍历中修改（Major）

禁止在 `range` 遍历 map 时修改该 map。

**违规示例：**

```go
for k, v := range m {
    if v == "" {
        delete(m, k)  // 违规！遍历中修改 map
    }
}
```

**正确做法：**

```go
var toDelete []string
for k, v := range m {
    if v == "" {
        toDelete = append(toDelete, k)
    }
}
for _, k := range toDelete {
    delete(m, k)
}
```

---

## 四、命名规范 (Naming Conventions)

### 规则 N-1：导出标识符缺少注释（Minor）

导出的函数、类型、常量必须有 godoc 注释。

**违规示例：**

```go
type UserService struct {  // 违规！缺少注释
    repo UserRepo
}

func NewUserService(repo UserRepo) *UserService {  // 违规！缺少注释
    return &UserService{repo: repo}
}
```

**正确做法：**

```go
// UserService 用户业务服务，处理用户创建、更新等操作。
type UserService struct {
    repo UserRepo
}

// NewUserService 创建用户服务实例。
func NewUserService(repo UserRepo) *UserService {
    return &UserService{repo: repo}
}
```

---

### 规则 N-2：Receiver 命名不一致（Minor）

同一类型的 receiver 名称应保持一致，使用类型首字母小写。

**违规示例：**

```go
func (u *User) Name() string { ... }
func (user *User) Age() int { ... }      // 违规！receiver 名不一致
func (self *User) Email() string { ... } // 违规！禁止使用 self
```

**正确做法：**

```go
func (u *User) Name() string { ... }
func (u *User) Age() int { ... }
func (u *User) Email() string { ... }
```

---

### 规则 N-3：包名使用大写或下划线（Minor）

包名应全小写，不使用下划线。

**违规示例：**

```go
package UserData      // 违规！大写
package user_service  // 违规！下划线
```

**正确做法：**

```go
package user
package userservice
```

---

### 规则 N-4：接口命名不以 I 开头（Minor）

Go 接口命名惯例：实现类用具体名，接口用行为描述。

**违规示例：**

```go
type IUserService interface {  // 违规！不应以 I 开头
    GetUser(id int64) (*User, error)
}
```

**正确做法：**

```go
// 接口：描述行为
type UserFinder interface {
    GetUser(id int64) (*User, error)
}

// 实现：具体名称
type UserService struct { ... }
```

---

## 五、API 设计 (API Design)

### 规则 A-1：返回具体类型而非接口（Major）

函数应返回具体实现类型，让调用方决定是否转换为接口。

**违规示例：**

```go
func NewUserRepo() Repository {  // 违规！返回接口
    return &userRepo{}
}
```

**正确做法：**

```go
func NewUserRepo() *userRepo {  // 返回具体类型
    return &userRepo{}
}
```

---

### 规则 A-2：接收接口而非具体类型（Major）

函数参数应接收接口，提高可测试性。

**违规示例：**

```go
func NewUserService(repo *userRepo) *UserService {  // 违规！接收具体类型
    return &UserService{repo: repo}
}
```

**正确做法：**

```go
func NewUserService(repo UserRepo) *UserService {  // 接收接口
    return &UserService{repo: repo}
}
```

---

### 规则 A-3：可变参数滥用（Minor）

禁止使用可变参数传递配置项，应使用 options struct。

**违规示例：**

```go
func CreateUser(name string, opts ...Option) error {  // 过度使用 options pattern
    ...
}
```

**正确做法（配置项较多时）：**

```go
type CreateUserConfig struct {
    Email    string
    Phone    string
    Role     Role
}

func CreateUser(name string, cfg CreateUserConfig) error {
    ...
}
```

---

## 六、测试 (Testing)

### 规则 T-1：表驱动测试缺失（Minor）

多场景测试应使用表驱动模式。

**违规示例：**

```go
func TestAdd(t *testing.T) {
    if Add(1, 2) != 3 { t.Error("1+2") }
    if Add(0, 0) != 0 { t.Error("0+0") }
    if Add(-1, 1) != 0 { t.Error("-1+1") }
}
```

**正确做法：**

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 1, 2, 3},
        {"zeros", 0, 0, 0},
        {"negative and positive", -1, 1, 0},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Add(tt.a, tt.b); got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}
```

---

### 规则 T-2：测试未覆盖错误路径（Major）

测试必须覆盖正常路径和错误路径。

**违规示例：**

```go
func TestGetUser(t *testing.T) {
    user, err := service.GetUser(ctx, 1)
    if err != nil {
        t.Fatal(err)
    }
    if user.Name != "Alice" {
        t.Errorf("got %s, want Alice", user.Name)
    }
    // 违规！未测试用户不存在、参数错误等场景
}
```

---

### 规则 T-3：使用 t.Fatal 在子测试中（Minor）

`t.Fatal` 会停止当前 goroutine，子测试中应使用 `t.Error` + `return`。

**违规示例：**

```go
t.Run("case1", func(t *testing.T) {
    if err != nil {
        t.Fatal(err)  // 违规！会停止整个测试
    }
})
```

**正确做法：**

```go
t.Run("case1", func(t *testing.T) {
    if err != nil {
        t.Errorf("unexpected error: %v", err)
        return
    }
})
```

---

## 七、安全 (Security)

### 规则 S-1：SQL 注入（Critical）

禁止使用字符串拼接构造 SQL。

**违规示例：**

```go
query := "SELECT * FROM users WHERE name = '" + name + "'"  // 违规！SQL 注入
db.Query(query)
```

**正确做法：**

```go
db.Query("SELECT * FROM users WHERE name = ?", name)
```

---

### 规则 S-2：硬编码密钥（Critical）

禁止在代码中硬编码密码、密钥、Token。

**违规示例：**

```go
const apiKey = "sk-1234567890abcdef"  // 违规！硬编码密钥
```

**正确做法：**

```go
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    return nil, fmt.Errorf("API_KEY environment variable not set")
}
```

---

### 规则 S-3：未验证输入（Major）

外部输入必须验证边界条件。

**违规示例：**

```go
func GetUser(c *gin.Context) {
    id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
    user, _ := service.GetUser(c, id)  // 违规！未校验 id 范围
    c.JSON(200, user)
}
```

**正确做法：**

```go
func GetUser(c *gin.Context) {
    id, err := strconv.ParseInt(c.Param("id"), 10, 64)
    if err != nil || id <= 0 {
        c.JSON(400, gin.H{"error": "invalid id"})
        return
    }
    user, err := service.GetUser(c, id)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, user)
}
```

---

## 八、代码风格 (Code Style)

### 规则 ST-1：导出函数缺少注释（Minor）

所有导出函数、类型、常量必须有 godoc 注释。

### 规则 ST-2：魔法数字（Minor）

使用常量替代魔法数字。

**违规示例：**

```go
if status == 3 {  // 违规！3 是什么？
    ...
}
```

**正确做法：**

```go
const StatusCompleted = 3
if status == StatusCompleted {
    ...
}
```

---

### 规则 ST-3：空 switch case 缺少注释（Minor）

空的 switch case 必须注释说明原因。

**违规示例：**

```go
switch status {
case StatusDraft:
case StatusSubmitted:  // 违规！空 case 是 fallthrough 还是无操作？
    process()
}
```

**正确做法：**

```go
switch status {
case StatusDraft:
    // 草稿状态无需处理
case StatusSubmitted:
    process()
}
```

---

## 审查输出格式

```markdown
## Go 代码审查报告

### 概要

| 等级 | 数量 |
|------|------|
| Critical | N |
| Major | N |
| Minor | N |

### 问题列表

#### [Critical] 规则 E-1：忽略错误返回值
- **文件**: `internal/order/biz/service.go:45`
- **问题**: `inventoryClient.DeductStock()` 错误被忽略
- **修复**: 
  ```go
  if err := s.inventoryClient.DeductStock(ctx, item.ProductID, item.Quantity); err != nil {
      return fmt.Errorf("deduct stock failed: %w", err)
  }
  ```

#### [Major] 规则 C-3：Context 未传播
- **文件**: `internal/order/model/dao.go:32`
- **问题**: `QueryByID` 函数签名缺少 context 参数
- **修复**: 添加 `ctx context.Context` 作为第一个参数

### 通过项
- [x] 规则 E-2：错误包装包含上下文
- [x] 规则 S-1：无 SQL 注入风险
- ...
```

## 快速审查命令

```bash
# 1. 检查忽略的错误
grep -rn "_, _ =" --include="*.go" | grep -v "_test.go"
grep -rn "err := " --include="*.go" | grep -v "if err"

# 2. 检查 panic 使用
grep -rn "panic(" --include="*.go" | grep -v "_test.go"

# 3. 检查硬编码密钥
grep -rn "password\|secret\|token\|apikey" --include="*.go" -i | grep -v "os.Getenv"

# 4. 运行静态分析
go vet ./...
staticcheck ./...

# 5. 检查竞态条件
go test -race ./...
```

## 审查流程

1. **获取变更文件列表**（git diff 或用户指定范围）
2. **按 Critical → Major → Minor 顺序检查**
3. **对每个问题提供：文件路径 + 行号 + 问题描述 + 修复代码**
4. **最后列出所有通过的规则**
5. **如果存在 Critical 问题，明确标注"必须修复后才能合并"**
