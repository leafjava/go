# 常见问题解答（FAQ）

## 学习相关

### Q1: 我有 Java 基础，学 Go 需要多久？

**A:** 根据你的情况：
- **基础语法**：3-5 天（Go 语法比 Java 简单）
- **Web 开发**：1 周（Gin 类似 Spring Boot）
- **Web3 集成**：1 周（主要是熟悉库）
- **实战项目**：1 周

**总计：4 周可以达到能独立开发的水平**

### Q2: Go 和 Java 最大的区别是什么？

**A:** 主要区别：
1. **并发模型**：Go 用 Goroutine + Channel，Java 用 Thread
2. **错误处理**：Go 用返回值，Java 用 try-catch
3. **继承**：Go 用组合，Java 用继承
4. **编译速度**：Go 快得多
5. **部署**：Go 编译成单个二进制文件，Java 需要 JVM

### Q3: 学完这个教程能找到工作吗？

**A:** 可以，但需要：
1. **完成毕业项目**（SkillsBay 或类似项目）
2. **刷 LeetCode**（Go 版本，50-100 题）
3. **准备面试题**（第24课的内容）
4. **有 GitHub 项目展示**
5. **结合你的 Web3 经验**（这是你的优势）

你的优势：
- 世界冠军背景
- TON 生态经验
- 行业人脉资源
- 有实战项目（SkillsBay、CINA）

## 技术问题

### Q4: 什么时候用指针，什么时候用值？

**A:** 简单规则：
- **方法接收者**：优先用指针（除非有特殊原因）
- **函数参数**：
  - 需要修改 → 用指针
  - 大结构体 → 用指针
  - 小数据（int, bool） → 用值
- **返回值**：
  - 可能不存在 → 用指针（可以返回 nil）
  - 确定存在 → 用值

### Q5: Slice 和 Array 有什么区别？

**A:**
| 特性 | Array | Slice |
|------|-------|-------|
| 长度 | 固定 | 可变 |
| 传递 | 值传递（复制） | 引用传递 |
| 使用 | 很少用 | 常用 |

**建议：99% 的情况用 Slice**

### Q6: Channel 什么时候需要缓冲？

**A:**
- **无缓冲**：需要同步（发送方等待接收方）
- **有缓冲**：允许异步（发送方不等待）

```go
// 无缓冲：必须有接收方才能发送
ch := make(chan int)

// 有缓冲：可以先发送，后接收
ch := make(chan int, 10)
```

### Q7: 如何避免 Goroutine 泄漏？

**A:** 常见原因和解决方案：
1. **Channel 没有关闭**
   ```go
   // 错误
   for {
       data := <-ch  // 如果 ch 永远不关闭，会一直阻塞
   }
   
   // 正确
   for data := range ch {  // ch 关闭后会自动退出
       // 处理 data
   }
   ```

2. **使用 Context 控制超时**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
   defer cancel()
   
   select {
   case <-ctx.Done():
       return
   case result := <-ch:
       // 处理结果
   }
   ```

### Q8: GORM 查询慢怎么优化？

**A:** 优化技巧：
1. **添加索引**
   ```go
   type User struct {
       Email string `gorm:"index"`  // 单列索引
   }
   ```

2. **预加载关联**
   ```go
   // 错误：N+1 查询
   db.Find(&users)
   for _, user := range users {
       db.Model(&user).Association("Orders").Find(&user.Orders)
   }
   
   // 正确：预加载
   db.Preload("Orders").Find(&users)
   ```

3. **只查询需要的字段**
   ```go
   db.Select("id", "name").Find(&users)
   ```

4. **使用原生 SQL**（复杂查询）
   ```go
   db.Raw("SELECT * FROM users WHERE ...").Scan(&users)
   ```

## Web3 相关

### Q9: go-ethereum 和 ethclient 有什么区别？

**A:**
- **go-ethereum**：完整的以太坊实现（包含节点）
- **ethclient**：go-ethereum 的客户端库（只用于连接节点）

**Web3 开发只需要 ethclient**

### Q10: 如何安全存储私钥？

**A:** 绝对不要：
- ❌ 硬编码在代码中
- ❌ 提交到 Git
- ❌ 明文存储在数据库

推荐方案：
1. **环境变量**（开发环境）
   ```go
   privateKey := os.Getenv("PRIVATE_KEY")
   ```

2. **加密存储**（生产环境）
   ```go
   // 使用 AES 加密
   encrypted := encrypt(privateKey, masterKey)
   db.Save(encrypted)
   ```

3. **硬件钱包**（高安全场景）

### Q11: Gas 估算不准确怎么办？

**A:** 解决方案：
1. **增加 Gas Limit 缓冲**
   ```go
   estimatedGas, _ := client.EstimateGas(ctx, msg)
   gasLimit := estimatedGas * 120 / 100  // 增加 20%
   ```

2. **使用历史数据**
   ```go
   // 查询最近 10 个区块的 Gas Price
   avgGasPrice := getAverageGasPrice(10)
   ```

3. **监控 Gas Price 波动**
   ```go
   // 使用 Gas Price Oracle
   gasPrice, _ := client.SuggestGasPrice(ctx)
   ```

### Q12: 如何处理区块链网络延迟？

**A:** 策略：
1. **设置超时**
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

2. **重试机制**
   ```go
   for i := 0; i < 3; i++ {
       result, err := queryBlockchain()
       if err == nil {
           break
       }
       time.Sleep(time.Second * time.Duration(i+1))
   }
   ```

3. **使用缓存**
   ```go
   // Redis 缓存余额（5分钟）
   balance, err := redis.Get("balance:" + address)
   if err != nil {
       balance = queryFromBlockchain()
       redis.Set("balance:"+address, balance, 5*time.Minute)
   }
   ```

## 项目相关

### Q13: 项目结构怎么组织？

**A:** 推荐结构（类似 Java 的分层架构）：
```
project/
├── main.go              # 入口
├── config/              # 配置
├── models/              # 数据模型（类似 Entity）
├── handlers/            # 控制器（类似 Controller）
├── services/            # 业务逻辑（类似 Service）
├── repositories/        # 数据访问（类似 DAO）
├── middleware/          # 中间件
├── utils/               # 工具函数
└── tests/               # 测试
```

### Q14: 如何写单元测试？

**A:** 示例：
```go
// wallet_test.go
package services

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestTransfer(t *testing.T) {
    // 准备
    wallet := &Wallet{Balance: 100}
    
    // 执行
    err := wallet.Transfer(50)
    
    // 断言
    assert.NoError(t, err)
    assert.Equal(t, 50.0, wallet.Balance)
}

func TestTransferInsufficientBalance(t *testing.T) {
    wallet := &Wallet{Balance: 10}
    err := wallet.Transfer(50)
    assert.Error(t, err)
}
```

### Q15: 如何部署 Go 项目？

**A:** 步骤：
1. **编译**
   ```bash
   CGO_ENABLED=0 GOOS=linux go build -o app
   ```

2. **Docker 部署**
   ```dockerfile
   FROM golang:1.21 AS builder
   WORKDIR /app
   COPY . .
   RUN go build -o main .
   
   FROM alpine:latest
   WORKDIR /app
   COPY --from=builder /app/main .
   CMD ["./main"]
   ```

3. **运行**
   ```bash
   docker build -t skillsbay-api .
   docker run -p 8080:8080 skillsbay-api
   ```

## 面试相关

### Q16: Go 面试常考什么？

**A:** 高频题：
1. **Goroutine 和 Thread 的区别**
2. **Channel 的实现原理**
3. **Slice 的底层结构**
4. **Map 的底层实现**
5. **Interface 的原理**
6. **GC 垃圾回收机制**
7. **内存逃逸分析**
8. **Context 的使用场景**

**建议：第24课会详细讲解**

### Q17: 简历上怎么写 Go 项目？

**A:** 示例：
```
SkillsBay 区块链交易平台后端（Go 重构）
- 使用 Gin + GORM 构建 RESTful API，支持 10000+ QPS
- 集成 TON 和 Ethereum 区块链，实现多链钱包管理
- 使用 Redis 缓存优化查询性能，响应时间从 500ms 降至 50ms
- 实现 JWT 认证和 RBAC 权限控制
- 使用 Goroutine + Channel 实现并发交易处理，提升 5 倍吞吐量
- 编写单元测试，覆盖率达 80%
- 使用 Docker 部署，支持 CI/CD 自动化
```

## 学习资源

### Q18: 有哪些好的 Go 学习资源？

**A:** 推荐：
1. **官方文档**：https://go.dev/doc/
2. **Go by Example**：https://gobyexample.com/
3. **Go 语言圣经**：https://gopl-zh.github.io/
4. **Effective Go**：https://go.dev/doc/effective_go
5. **B站视频**：黑马程序员、老男孩

### Q19: 有哪些优秀的 Go 开源项目可以学习？

**A:** 推荐：
1. **gin-vue-admin**：完整的后台管理系统
2. **go-admin**：权限管理系统
3. **kratos**：微服务框架（B站开源）
4. **go-zero**：微服务框架
5. **go-ethereum**：以太坊 Go 实现

## 其他问题

### Q20: 遇到问题去哪里找答案？

**A:** 顺序：
1. **官方文档**（最权威）
2. **Stack Overflow**（搜索 `[go] your question`）
3. **GitHub Issues**（库的问题）
4. **Go 论坛**：https://forum.golangbridge.org/
5. **问 AI**（ChatGPT、Claude）

---

**还有问题？** 
- 查看 [快速参考](./QUICK_REFERENCE.md)
- 查看 [课程大纲](./COURSE_OUTLINE.md)
- 开始学习 → [第1课](./week1/day1-setup.md)
