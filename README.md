# Go 语言 Web3 全栈速成教程

> 专为有 Java 基础的 Web3 开发者设计 | 4周从零到实战

## 📚 课程概览

- **学习时长**: 4周（每天2-4小时）
- **目标**: 能独立开发 Go + Gin + GORM 的区块链交互服务
- **适合人群**: 有 Java/Spring Boot 基础，想快速上手 Go 做 Web3 后端

## 🎯 学习原则

1. 不求全，只学 Web3 开发最常用的 80% 内容
2. 每天理论 + 大量动手写代码
3. 利用 Java 经验做类比（Spring Boot → Gin）
4. 边学边做项目，第2周就开始写实战代码

## 📖 课程目录

### 第一周：Go 基础语法 + 核心特性
- [第1课：环境搭建 + Hello World](./week1/day1-setup.md)
- [第2课：变量、常量、基础类型](./week1/day2-basics.md)
- [第3课：函数、错误处理](./week1/day3-functions.md)
- [第4课：结构体、方法、接口](./week1/day4-structs.md)
- [第5课：指针（重点）](./week1/day5-pointers.md)
- [第6课：Slice 和 Map](./week1/day6-collections.md)
- [第7课：并发编程：Goroutine + Channel](./week1/day7-concurrency.md)

### 第二周：Web 开发核心（Gin + GORM）
- [第8课：Gin 框架入门](./week2/day8-gin-intro.md)
- [第9课：路由、中间件、参数绑定](./week2/day9-gin-advanced.md)
- [第10课：GORM 数据库操作](./week2/day10-gorm.md)
- [第11课：JWT 认证 + 权限控制](./week2/day11-jwt.md)
- [第12课：配置管理 + 日志](./week2/day12-config-log.md)
- [第13-14课：实战项目 - SkillsBay API](./week2/day13-14-project.md)

### 第三周：Web3 后端开发
- [第15课：Go 调用以太坊合约](./week3/day15-ethereum.md)
- [第16课：TON 区块链集成](./week3/day16-ton.md)
- [第17课：交易构建、签名、发送](./week3/day17-transactions.md)
- [第18课：事件监听 + Gas 估算](./week3/day18-events-gas.md)
- [第19课：Redis 缓存 + 消息队列](./week3/day19-redis-mq.md)
- [第20-21课：实战 - 区块链交互服务](./week3/day20-21-blockchain-service.md)

### 第四周：工程化 + 面试准备
- [第22课：单元测试 + 性能优化](./week4/day22-testing.md)
- [第23课：Docker + CI/CD](./week4/day23-docker.md)
- [第24课：Go 面试高频题](./week4/day24-interview.md)
- [第25-28课：毕业项目 - 完整 Web3 后端](./week4/day25-28-final-project.md)

## 🚀 快速开始

```bash
# 1. 克隆或下载本教程
cd D:\webProject\go

# 2. 安装 Go（1.21+）
# 下载：https://go.dev/dl/

# 3. 验证安装
go version

# 4. 开始第一课
cd week1
```

## 📦 推荐学习资源

### 官方文档
- [Go 官方教程](https://go.dev/tour/)
- [Go by Example](https://gobyexample.com/)
- [Go 语言圣经（中文）](https://gopl-zh.github.io/)

### 框架文档
- [Gin 官方文档](https://gin-gonic.com/zh-cn/docs/)
- [GORM 官方文档](https://gorm.io/zh_CN/docs/)

### Web3 相关
- [go-ethereum](https://github.com/ethereum/go-ethereum)
- [tonutils-go](https://github.com/xssnick/tonutils-go)

## 💡 学习建议

1. **每天固定时间学习** - 建议早上或晚上2-4小时
2. **一定要动手写代码** - 看懂 ≠ 会写
3. **完成每天的作业** - 作业是精心设计的，覆盖核心知识点
4. **遇到问题先搜索** - Go 社区很活跃，大部分问题都能找到答案
5. **第2周开始就要做项目** - 不要只学语法

## 🎓 学习路径

```
Week 1: 语法基础 → 能看懂 Go 代码
Week 2: Web 开发 → 能写 RESTful API
Week 3: Web3 集成 → 能调用区块链
Week 4: 工程化 → 能写生产级代码
```

## 📝 作业提交

每天的作业代码放在对应的目录下：
```
go/
├── week1/
│   ├── homework/
│   │   ├── day1/
│   │   ├── day2/
│   │   └── ...
├── week2/
└── ...
```

## 🏆 毕业标准

完成以下任一项目即可毕业：
1. SkillsBay 后端服务（Go 重写）
2. CINA 杠杆交易服务
3. TON 交易查询 + 分析服务

## 📞 获取帮助

- 遇到问题先查看 [常见问题](./FAQ.md)
- Go 官方论坛：https://forum.golangbridge.org/
- Stack Overflow Go 标签

---

**现在就开始第一课吧！** → [第1课：环境搭建](./week1/day1-setup.md)
