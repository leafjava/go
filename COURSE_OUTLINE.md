# Go 语言 Web3 全栈速成教程 - 完整大纲

## 第一周：Go 基础语法 + 核心特性 ✅

- [x] 第1课：环境搭建 + Hello World
- [x] 第2课：变量、常量、基础类型
- [x] 第3课：函数、错误处理
- [x] 第4课：结构体、方法、接口
- [x] 第5课：指针（重点）
- [x] 第6课：Slice 和 Map
- [x] 第7课：并发编程：Goroutine + Channel

## 第二周：Web 开发核心（Gin + GORM）

### 已完成
- [x] 第8课：Gin 框架入门

### 第9课：Gin 进阶 - 中间件、验证、错误处理

**学习内容：**
- 中间件（Logger、Recovery、CORS）
- 自定义中间件（JWT 认证、限流）
- 参数验证（validator）
- 统一错误处理
- 响应格式标准化

**实战项目：**
```go
// 统一响应格式
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}

// JWT 中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        // 验证 token
        c.Next()
    }
}

// 限流中间件
func RateLimitMiddleware(limit int) gin.HandlerFunc {
    // 实现限流逻辑
}
```

### 第10课：GORM 数据库操作

**学习内容：**
- GORM 安装和配置
- 模型定义和迁移
- CRUD 操作
- 关联查询（一对一、一对多、多对多）
- 事务处理
- 查询优化

**实战项目：**
```go
// 用户模型
type User struct {
    gorm.Model
    Address  string `gorm:"uniqueIndex"`
    Username string
    Email    string
    Wallets  []Wallet `gorm:"foreignKey:UserID"`
}

// 钱包模型
type Wallet struct {
    gorm.Model
    UserID  uint
    Address string  `gorm:"uniqueIndex"`
    Balance float64
    Network string
}

// 交易模型
type Transaction struct {
    gorm.Model
    Hash   string `gorm:"uniqueIndex"`
    From   string
    To     string
    Amount float64
    Status string
}
```

### 第11课：JWT 认证 + 权限控制

**学习内容：**
- JWT 原理和实现
- 用户注册和登录
- Token 生成和验证
- 刷新 Token
- RBAC 权限控制

**实战项目：**
```go
// 注册
POST /api/v1/auth/register
{
    "username": "linshen",
    "email": "linshen@example.com",
    "password": "password123"
}

// 登录
POST /api/v1/auth/login
{
    "email": "linshen@example.com",
    "password": "password123"
}

// 响应
{
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "...",
    "expires_in": 3600
}
```

### 第12课：配置管理 + 日志

**学习内容：**
- Viper 配置管理
- 环境变量
- 配置文件（YAML/JSON）
- Zap 日志库
- 日志分级和轮转

**实战项目：**
```yaml
# config.yaml
server:
  port: 8080
  mode: debug

database:
  host: localhost
  port: 5432
  name: skillsbay
  user: postgres
  password: password

blockchain:
  ethereum:
    rpc_url: https://eth.llamarpc.com
  ton:
    rpc_url: https://toncenter.com/api/v2

jwt:
  secret: your-secret-key
  expires_in: 3600
```

### 第13-14课：实战项目 - SkillsBay API

**项目功能：**
1. 用户系统
   - 注册/登录
   - 个人信息管理
   - 钱包绑定

2. 钱包管理
   - 多链钱包支持
   - 余额查询
   - 转账记录

3. 交易系统
   - 交易查询
   - 交易统计
   - Gas 估算

4. NFT 系统
   - NFT 列表
   - NFT 详情
   - 所有权转移

**项目结构：**
```
skillsbay-api/
├── main.go
├── config/
│   ├── config.go
│   └── config.yaml
├── models/
│   ├── user.go
│   ├── wallet.go
│   ├── transaction.go
│   └── nft.go
├── handlers/
│   ├── auth.go
│   ├── wallet.go
│   ├── transaction.go
│   └── nft.go
├── middleware/
│   ├── auth.go
│   ├── cors.go
│   └── logger.go
├── services/
│   ├── user_service.go
│   ├── wallet_service.go
│   └── blockchain_service.go
├── database/
│   └── db.go
└── utils/
    ├── jwt.go
    ├── response.go
    └── validator.go
```

## 第三周：Web3 后端开发

### 第15课：Go 调用以太坊合约

**学习内容：**
- go-ethereum 库
- 连接以太坊节点
- 查询余额
- 发送交易
- 调用智能合约
- 事件监听

**实战代码：**
```go
import (
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
)

// 连接节点
client, err := ethclient.Dial("https://eth.llamarpc.com")

// 查询余额
address := common.HexToAddress("0x742d35Cc...")
balance, err := client.BalanceAt(context.Background(), address, nil)

// 发送交易
tx := types.NewTransaction(...)
signedTx, err := types.SignTx(tx, signer, privateKey)
err = client.SendTransaction(context.Background(), signedTx)
```

### 第16课：TON 区块链集成

**学习内容：**
- tonutils-go 库
- TON 钱包操作
- TON 交易
- Jetton 代币
- NFT 操作

### 第17课：交易构建、签名、发送

**学习内容：**
- 交易结构
- 私钥管理
- 交易签名
- Gas 估算
- Nonce 管理
- 交易重试

### 第18课：事件监听 + Gas 估算

**学习内容：**
- 合约事件监听
- 区块扫描
- 实时通知
- Gas Price 预测
- 交易加速

### 第19课：Redis 缓存 + 消息队列

**学习内容：**
- Redis 基础操作
- 缓存策略
- Asynq 消息队列
- 异步任务处理
- 定时任务

### 第20-21课：实战 - 区块链交互服务

**项目功能：**
1. 多链余额查询服务
2. 交易监控和通知
3. Gas Price 预测
4. 交易加速服务

## 第四周：工程化 + 面试准备

### 第22课：单元测试 + 性能优化

**学习内容：**
- testing 包
- testify 断言库
- Mock 测试
- 基准测试
- pprof 性能分析
- 内存优化

### 第23课：Docker + CI/CD

**学习内容：**
- Dockerfile 编写
- Docker Compose
- GitHub Actions
- 自动化部署

### 第24课：Go 面试高频题

**核心知识点：**
1. Goroutine 和 Channel 原理
2. GC 垃圾回收
3. Slice 底层实现
4. Map 底层实现
5. Interface 原理
6. Context 使用
7. 内存逃逸分析
8. 并发安全

### 第25-28课：毕业项目

**选择一个完整项目：**
1. SkillsBay 后端服务（Go 重写）
2. 多链钱包管理系统
3. DeFi 数据聚合平台
4. NFT 交易市场后端

**项目要求：**
- 完整的 RESTful API
- JWT 认证
- 数据库设计
- 区块链集成
- 单元测试
- Docker 部署
- API 文档

## 学习资源

### 官方文档
- [Go 官方文档](https://go.dev/doc/)
- [Gin 文档](https://gin-gonic.com/docs/)
- [GORM 文档](https://gorm.io/docs/)
- [go-ethereum 文档](https://geth.ethereum.org/docs/)

### 推荐书籍
- 《Go 语言圣经》
- 《Go 语言实战》
- 《Go 并发编程实战》

### 视频教程
- 黑马程序员 Go 教程（B站）
- 老男孩 Go 教程（B站）

### 实战项目参考
- [gin-vue-admin](https://github.com/flipped-aurora/gin-vue-admin)
- [go-admin](https://github.com/go-admin-team/go-admin)

## 学习建议

1. **每天坚持 2-4 小时**
2. **一定要动手写代码**
3. **完成每天的作业**
4. **第2周开始就要做项目**
5. **多看优秀开源项目的代码**

## 毕业标准

完成以下任一项目即可毕业：
1. SkillsBay 后端服务（Go 重写）
2. 多链钱包管理系统
3. DeFi 数据聚合平台
4. 自定义 Web3 项目

**项目必须包含：**
- ✅ 完整的 API 接口
- ✅ 数据库设计和实现
- ✅ JWT 认证
- ✅ 区块链集成
- ✅ 单元测试（覆盖率 > 60%）
- ✅ Docker 部署
- ✅ README 文档

---

**现在就开始学习吧！** → [第1课：环境搭建](./week1/day1-setup.md)
