# sexAI 面试问题汇总（2026-09-08）

> **公司方向**：成人视频站 + APP + Web3（外部 SEO 为主）
> **面试时长**：约 26 分钟
> **技术栈定位**：Golang 后端 + ES + Redis + MySQL，考虑用 K3S/K3SD 部署
> **工作地点**：远程 / 东南亚（泰国、老挝、越南）
> **薪资结算**：USDT

---

## 一、自我介绍 & 项目技术栈（开场）

### Q1：请简单做下自我介绍，还有最近用的项目和技术栈

**面试官原话**：
> 先简单做一下自我介绍，还有最近用的项目和技术栈吧。

**我的回答要点**：
- 广东财经大学毕业，7 年全栈开发经验
- 最近在中旅集团后台管理系统
- 前端：React
- 后端：Go
- 做过 Agent 工作流（如合同照片自动录入信息到后台）

---

## 二、AI & Agent 相关（核心考察）

### Q2：AI 这块做了多少开发经验？多少个项目？

**面试官原话**：
> 目前的话是做了多少 AI 的开发经验？多少个项目啊？

**我的回答要点**：
- 多 agent 协作、向量数据库、记忆优化等
- 海南航空项目已经开始做这块

---

### Q3：多 Agent 协作的场景怎么设计？通信和协调怎么做？

**面试官原话**：
> 那问你几个简单问题吧。那多 aj 的协作的场景的话，你会怎么设计？然后 aj 之间的通信和协调你会怎么做？

**我的回答要点**：
- 中央消息总线（消息发布 / 事件订阅）
- 解耦，新增/替换 Agent 干净

**参考标准答案**（建议补充）：

```
多 Agent 架构三层设计：
1. 编排层（Orchestrator）——决定 Agent 调用顺序和触发条件
2. 通信层（Message Bus）——Agent 间通过事件 / 消息解耦
3. 执行层（Agents）——具体业务 Agent（意图分析、槽位抽取、向量检索等）

通信模式：
- 消息队列（Kafka / Redis Streams）：异步、解耦、可追溯
- 事件总线（Pub/Sub）：一对多广播
- 直接 RPC / gRPC：同步、强一致场景

协调机制：
- 状态机 + 超时熔断（防止死循环）
- 优先级 + 置信度投票（解决矛盾决策）
- 全局记忆同步节点（避免记忆断层）
```

---

### Q4：Agent 之间的消息键模式怎么设计？

**面试官原话**：
> 键模式怎么设计吗？

**我的回答要点**：
- 加组件 + 时间戳

**参考标准答案**（建议补充）：

```json
{
  "msg_id": "uuid-v4",                  // 消息唯一标识
  "trace_id": "trace-xxx",              // 全链路追踪
  "parent_msg_id": "uuid-xxx",          // 父消息（支持会话树）
  "session_id": "session-xxx",          // 会话 ID
  "agent_id": "intent-analysis-agent",  // 发送方 Agent
  "target_agent": "slot-extract-agent", // 目标 Agent
  "msg_type": "request | response | event",
  "timestamp": 1725753600000,           // 毫秒时间戳
  "payload": { ... },                   // 业务负载
  "confidence": 0.92                    // 置信度（用于仲裁）
}
```

**设计原则**：
- **msg_id**：去重 + 幂等性
- **trace_id**：全链路日志串联
- **session_id**：会话隔离
- **timestamp**：消息过期和顺序判断
- **parent_msg_id**：支持多轮对话和分支

---

### Q5：多 Agent 死循环和矛盾决策怎么处理？

**面试官原话**：
> 那你这里面多维度的出现了出现了死循环和相互矛盾的决策，你会怎么处理？

**我的回答要点**：
- 状态机 + 超时熔断（最大重试次数、超时时间）
- 优先级排序
- 置信度阈值（< 0.6 不采纳）
- 仲裁机制（按优先级 + 权重 + 置信度投票）

**参考标准答案**（建议补充）：

```python
# 三层防护机制
class AgentCoordinator:
    def __init__(self):
        self.max_retries = 3              # 最大重试次数
        self.timeout = 5                  # 单次超时（秒）
        self.confidence_threshold = 0.6   # 置信度阈值
        self.state_machine = {}           # 状态机记录每个 Agent 状态

    # 1. 防死循环：状态机 + 超时熔断
    def invoke_agent(self, agent, input_data, context):
        agent_id = agent.name
        # 状态机：检测循环
        if self.state_machine.get(agent_id, 0) >= self.max_retries:
            raise MaxRetriesExceeded(f"{agent_id} 超过最大重试次数")
        self.state_machine[agent_id] = self.state_machine.get(agent_id, 0) + 1

        # 超时熔断
        try:
            result = agent.run(input_data, context, timeout=self.timeout)
            self.state_machine[agent_id] = 0  # 成功后重置
            return result
        except TimeoutError:
            return self.fallback(agent, input_data)

    # 2. 防矛盾：置信度阈值
    def filter_by_confidence(self, results):
        return [r for r in results if r.confidence >= self.confidence_threshold]

    # 3. 仲裁投票：优先级 + 权重 + 置信度
    def arbitrate(self, results):
        # 加权投票：priority × weight × confidence
        scores = {}
        for r in results:
            score = r.priority * r.weight * r.confidence
            scores[r.agent_id] = scores.get(r.agent_id, 0) + score
        return max(scores, key=scores.get)
```

---

### Q6：Agent 的记忆管理怎么做？怎么解决记忆断层？

**面试官原话**：
> 那你这里，Agent 的 Agent 的记忆管理，你会怎么做？就比如说你这里已经采用了多 Agent 然后这里出现了机遇，每个 Agent 之间出现了记忆层的断层或者是混乱。然后你这里要怎么去梳理这个，然后保留哪些历史记录该保留哪些历史记录该清理？

**我的回答要点**：
- 分层记忆（短期 + 长期）
- 每个 Agent 短期记忆用完就清理
- 长期记忆存核心意图、最终决策结果
- 全局记忆：所有 Agent 能读但不能随便写
- 记忆同步节点（意图分析 Agent 确认后写入全局记忆）
- 历史记录策略保留

**参考标准答案**（建议补充）：

```
记忆分层架构：
┌─────────────────────────────────────┐
│  全局长期记忆（Global Memory）        │  ← 所有 Agent 可读，专用 Agent 可写
│  - 用户核心意图                       │
│  - 关键决策结果                       │
│  - 用户偏好画像                       │
│  - 存储：向量数据库 + KV              │
└─────────────────────────────────────┘
            ↑ ↓ 同步节点
┌─────────────────────────────────────┐
│  Agent 局部短期记忆（Local Memory）   │  ← 单 Agent 私有，TTL 自动清理
│  - 当前任务上下文                     │
│  - 中间计算结果                       │
│  - 存储：Redis / 内存                 │
│  - TTL：任务结束自动清理              │
└─────────────────────────────────────┘

防断层机制：
1. 记忆同步节点：在关键节点（意图确认、决策生成）写入全局记忆
2. Agent 启动前先读全局记忆，避免重复梳理
3. 历史记录策略：
   - 保留：核心意图、决策结果、用户偏好
   - 清理：临时变量、调试日志、过期上下文
   - 压缩：超长对话做摘要存储
```

---

## 三、RAG 检索增强（核心考察）

### Q7：RAG 的完整流程是什么？为什么要分块？

**面试官原话**：
> 那在记忆里面，我们一般就都都会用到那个检索 rag 的那一块的，rag 的检索还有分块，聊一下那个 rag 的完整流程和为什么要进行这个分块。

**我的回答要点**：
- 两条链路：离线建库 + 在线检索
- 离线：文档解析 → 清洗 → 分块 → 向量化
- 在线：用户输入向量化 → 混合检索 → 关键词召回
- 分块原因：超出模型上下文 + 噪音太多 + 命中率更高

**参考标准答案**（建议补充）：

```
RAG 完整流程：

离线建库链路（一次性或增量）：
文档输入 → 解析（PDF/Word/MD） → 清洗（去噪/格式化） 
        → 分块（Chunking） → 向量化（Embedding） → 存入向量数据库

在线检索链路（每次请求）：
用户 Query → Query 向量化 → 混合检索（向量 + 关键词）
          → 重排（Rerank） → Top-K 召回 → 拼接 Prompt → LLM 生成答案

为什么要分块：
1. 上下文限制 —— LLM 有最大 token 限制（如 4K/8K/32K）
2. 噪音过滤 —— 整篇文档塞入会让无关信息干扰 LLM 注意力
3. 命中率提升 —— 精确的小块比大段全文更容易命中用户问题
4. 成本控制 —— 减少 token 消耗
5. 检索精度 —— 小的语义单元匹配更准确
```

---

### Q8：分块大小和重叠怎么定？检索不准一般是哪个环节的问题？

**面试官原话**：
> 那里面块的大小和重叠你会怎么定？然后有时候检索不准确一般是哪个环节出了问题了？

**我的回答要点**：

**分块大小和重叠**：
- 中文场景常用 256~512 字（token 约 400~800）
- 重叠 10%~20%（保留上下文连贯性）
- 取决于文档类型（合同 200~300，文章 500~800）

**检索不准的环节**：
1. 分块策略问题（切太碎 / 切断语义边界）
2. Embedding 模型问题（语义表达不准）
3. 检索策略问题（纯向量检索对专有名词弱）
4. 缺少重排（直接取 Top-K 可能漏相关）

**参考标准答案**（建议补充）：

```python
# 分块策略选择
chunking_strategies = {
    "fixed_size": {
        "chunk_size": 500,        # 固定 500 字
        "overlap": 50,             # 重叠 50 字（10%）
        "适用": "结构化文档"
    },
    "semantic": {
        "chunk_size": "动态",       # 按语义边界切分
        "overlap": "1~2 句",
        "适用": "文章、博客"
    },
    "recursive": {
        "chunk_size": 500,
        "overlap": 50,
        "适用": "通用场景（推荐）"
    }
}

# 检索不准排查清单
def diagnose_low_retrieval_accuracy():
    checks = [
        ("分块质量", "切碎了？切断语义？边界是否合理？"),
        ("Embedding 模型", "中文效果？是否用 M3E/BGE 等中文优化模型？"),
        ("检索策略", "纯向量？是否需要混合检索（BM25 + 向量）？"),
        ("重排", "是否使用 BGE-Reranker 等模型精排？"),
        ("Query 改写", "用户 Query 是否需要改写/扩展？"),
        ("元数据过滤", "是否需要按时间/分类过滤？"),
    ]
    return checks
```

**常见修复方案**：
- 换更好的 Embedding 模型（BGE-large-zh、M3E）
- 加入 BM25 关键词召回（混合检索）
- 增加 Rerank 重排（BGE-Reranker）
- Query 改写（让 LLM 重新表述用户问题）
- 元数据过滤（按时间、文档类型缩小范围）

---

## 四、Golang 基础（必考）

### Q9：Go 开发几年经验？

**面试官原话**：
> 狗狼开发的话，现在最近狗狼开发是有是开了有几年的开发经验啊？

**我的回答**：毕业就开始用 Go，差不多 7 年了。

---

### Q10：GMP 调度模型？G、M、P 分别是什么？

**面试官原话**：
> 那聊一下 Golang 的一些基础的吧，那个 GMP 的调度模型，这里 G 是指什么？M 是指什么？P 是指什么？然后它们的执行顺序。

**我的回答要点**：
- G：Goroutine（协程本身，存了很多指针，初始栈 2KB）
- M：Machine（内核线程）
- P：Processor（逻辑处理器）
- M 必须绑定 P 才能执行 G

**参考标准答案**（建议补充）：

```
GMP 调度模型（Go 1.5+ 引入）：

G（Goroutine）：
  - 用户态轻量线程
  - 初始栈 2KB（可动态增长到 1GB）
  - 包含：栈指针、程序计数器、寄存器、调度信息

M（Machine）：
  - OS 内核线程（真正干活的）
  - 由 OS 调度
  - M 必须绑定 P 才有执行权

P（Processor）：
  - 逻辑处理器（默认 = GOMAXPROCS = CPU 核数）
  - 维护一个本地 Goroutine 队列（LRQ）
  - M 必须持有 P 才能执行 G

执行流程：
  1. M 绑定 P，从 P 的本地队列取 G 执行
  2. 本地队列空 → 从全局队列（GRQ）取
  3. 全局队列也空 → 从其他 P 的本地队列偷一半（Work Stealing）
  4. G 阻塞时，M 释放 P，P 找其他 M 绑定继续执行其他 G

为什么 GMP 而非 GM：
  - GM 模型下，M 切换 G 需要加全局锁，性能差
  - GMP 把锁分散到每个 P 的本地队列，减少竞争
```

```go
// Goroutine 创建示例
go func() {
    fmt.Println("hello goroutine")
}()

// 设置 P 数量（默认 = CPU 核数）
runtime.GOMAXPROCS(8)
```

---

### Q11：Go error 处理的三种模式？

**面试官原话**：
> 这里 Golang 里面有那个 error，常见我们异常处理的话，Golang 的 error 处理三种模式是是哪三种模式啊？

**我的回答要点**：
1. 直接返回 error
2. panic / recover（处理严重错误）
3. 自定义错误类型 / 错误包装

**追问**：哨兵错误 + 包装错误（errors.Is / errors.As / fmt.Errorf %w）

**参考标准答案**（建议补充）：

```go
import (
    "errors"
    "fmt"
)

// 模式 1：哨兵错误（Sentinel Error）
// 预定义的全局 error 变量，用于 == 比较
var (
    ErrNotFound     = errors.New("资源未找到")
    ErrUnauthorized = errors.New("未授权")
    ErrInvalidInput = errors.New("参数无效")
)

func GetUser(id int) (*User, error) {
    if id <= 0 {
        return nil, ErrInvalidInput
    }
    // ...
    return nil, ErrNotFound
}

// 调用方判断
if errors.Is(err, ErrNotFound) {
    return 404
}

// 模式 2：自定义错误类型（Custom Error Type）
// 携带更多上下文信息
type BusinessError struct {
    Code    int
    Message string
    Err     error  // 包装底层错误
}

func (e *BusinessError) Error() string {
    return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
}

func (e *BusinessError) Unwrap() error {
    return e.Err
}

// errors.As 提取自定义错误
var be *BusinessError
if errors.As(err, &be) {
    return be.Code
}

// 模式 3：错误包装（Error Wrapping）
// 用 fmt.Errorf %w 包装错误链
func GetUserProfile(id int) (*Profile, error) {
    user, err := GetUser(id)
    if err != nil {
        return nil, fmt.Errorf("获取用户档案失败 (id=%d): %w", id, err)
    }
    // ...
}

// errors.Unwrap 拆解错误链
for err != nil {
    fmt.Println(err)
    err = errors.Unwrap(err)
}
```

**Go 1.13+ 三件套**：`errors.Is` + `errors.As` + `fmt.Errorf("...%w", err)`

---

### Q12：Channel 关闭后读和写会发生什么？

**面试官原话**：
> 在 Golang 里面我们常用到 channel 通道，我们在 channel 通道里面会有，就是关闭的时候，我们会遇见像已关闭的 channel 通道读和写会发生什么情况？

**我的回答要点**：
- **写**：向已关闭 channel 写数据 → **panic（不可 recover）**
- **读**：从已关闭 channel 读数据 → **不 panic，返回零值，不阻塞**
- 容易踩坑：读会一直返回零值
- 黄金法则：发送方负责 close

**参考标准答案**（建议补充）：

```go
// 写：panic
ch := make(chan int)
close(ch)
ch <- 1  // panic: send on closed channel

// 读：返回零值（用 ok 模式判断）
ch := make(chan int)
close(ch)

val := <-ch        // val = 0（不报错！）
val, ok := <-ch     // val = 0, ok = false（推荐）

// 安全模式：用 defer recover 兜底（但不是最佳实践）
defer func() {
    if r := recover(); r != nil {
        fmt.Println("recovered:", r)
    }
}()

// 最佳实践：发送方负责 close
func producer(ch chan<- int) {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // 只有发送方 close
}

func consumer(ch <-chan int) {
    for v := range ch {  // 直到 close 才退出
        fmt.Println(v)
    }
}
```

---

### Q13：多发送方一个接收方，怎么安全关闭 channel？

**面试官原话**：
> 那我如果有多个发送方，一个接收方，我这里怎么安全的关闭这个 channel 通道呢？

**我的回答要点**：
- 用 `sync.WaitGroup` 等所有发送协程结束
- 由专门的 goroutine 统一关闭 channel

**参考标准答案**（建议补充）：

```go
// 方案1：sync.WaitGroup + 专门 goroutine 关闭
func multiSender() {
    ch := make(chan int)
    var wg sync.WaitGroup

    // 3 个发送方
    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 5; j++ {
                ch <- id*10 + j
            }
        }(i)
    }

    // 专门 goroutine 等所有发送完成后关闭
    go func() {
        wg.Wait()
        close(ch)
    }()

    // 接收方
    for v := range ch {
        fmt.Println(v)
    }
}

// 方案2：用 sync.Once 保证只关闭一次
func multiSenderWithOnce() {
    ch := make(chan int)
    var once sync.Once
    var wg sync.WaitGroup

    for i := 0; i < 3; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            for j := 0; j < 5; j++ {
                ch <- id*10 + j
            }
        }(i)
    }

    go func() {
        wg.Wait()
        once.Do(func() { close(ch) })  // 保证只 close 一次
    }()

    for v := range ch {
        fmt.Println(v)
    }
}

// 方案3：用 done channel 通知关闭
func multiSenderWithDone() {
    ch := make(chan int)
    done := make(chan struct{})

    for i := 0; i < 3; i++ {
        go func(id int) {
            for j := 0; j < 5; j++ {
                select {
                case ch <- id*10 + j:
                case <-done:
                    return
                }
            }
        }(i)
    }

    go func() {
        // 某种条件触发关闭
        time.Sleep(time.Second)
        close(done)
        close(ch)
    }()

    for v := range ch {
        fmt.Println(v)
    }
}
```

**核心原则**：多个发送方时，**不能在发送方内部 close**（会触发 panic），必须由独立的协调者关闭。

---

## 五、Elasticsearch（搜索核心）

### Q14：ES 的分片和副本是干什么的？

**面试官原话**：
> ES 里面我们的分片和副本是干什么的？

**我的回答要点**：
- 分片：水平扩展存储，提升性能
- 副本：提升数据冗余（高可用）+ 分担读压力

**参考标准答案**（建议补充）：

```
ES 分片（Shard）：
- 把一个索引（Index）的数据拆成多份，分布到不同节点
- 作用：水平扩展存储 + 并行查询提升性能
- 创建索引时指定（number_of_shards），创建后不可改
- 默认 5 个主分片

ES 副本（Replica）：
- 每个分片的拷贝，提供数据冗余
- 作用：高可用（主分片挂了副本顶上）+ 读扩展（搜索请求路由到副本）
- 可动态调整（number_of_replicas）
- 默认 1 个副本

关系：
- 一个索引 = N 个主分片 + N × M 个副本
- 副本数越多，可用性越高，存储成本也越高
- 副本提升读性能但不提升写性能（写仍要主分片处理）
```

```json
PUT /my_index
{
  "settings": {
    "number_of_shards": 5,      // 主分片数（创建后不可改）
    "number_of_replicas": 1     // 副本数（可动态调整）
  }
}
```

---

### Q15：ES 深分页为什么慢？怎么处理？

**面试官原话**：
> 那那我们在这里的时候，通常有深深那种深分页的查询，比如说我们一万页啊，这里上下页一万页的上下页在 ES 这里为什么会慢？然后你会怎么处理呢？

**我的回答要点**：
- point in time（PIT）机制
- search_after 配合游标

**参考标准答案**（建议补充）：

```
ES 深分页（from + size）为什么慢：

from + size 机制：
- 每个分片都要查 from + size 条数据
- 协调节点汇总所有分片结果后再做全局排序
- 取第 from ~ from+size 条
- from=10000, size=10 → 每个分片都要查 10010 条，排序后丢弃前 10000 条

性能问题：
- 分片数越多越慢（10 分片 × 10010 = 100100 条数据排序）
- 默认 index.max_result_window = 10000（超过报错）
- 深度翻页 = 全量扫描 + 全局排序

优化方案：

1. search_after（推荐）：
   - 基于上一页最后一条的 sort 值查下一页
   - 不需要 from，深翻页性能稳定
   - 必须有全局唯一的 sort 字段（如 _id + 时间戳）

2. Point In Time (PIT)：
   - 创建快照，避免翻页过程中数据变化
   - 配合 search_after 使用

3. Scroll API（已不推荐）：
   - 维护游标上下文
   - 适合全量导出，不适合用户翻页
```

```json
// search_after + PIT 实现深翻页
POST /my_index/_pit?keep_alive=2m

POST /_search
{
  "pit": {
    "id": "PIT_ID",
    "keep_alive": "2m"
  },
  "size": 10,
  "sort": [
    { "timestamp": "asc" },
    { "_id": "asc" }  // 全局唯一 tiebreaker
  ],
  "search_after": [1725753600000, "doc_123"]  // 上一页最后一条
}
```

---

### Q16：什么是倒排索引？为什么 ES 适合全文检索而 MySQL 不适合？

**面试官原话**：
> 那嗯，ES，ES 有一个倒排索引的逻辑，然后这里什么是倒排索引？为什么 ES 适合全文解锁，而买这个不适合呢？

**我的回答要点**：
- 倒排索引：分词后建立关键词→文档的映射
- MySQL 全文检索弱：分词器弱（中文支持差）
- ES 倒排索引：性能好 + 可分片到多机器并行查询

**参考标准答案**（建议补充）：

```
倒排索引（Inverted Index）：

正排索引（MySQL 默认）：
  文档1 → [关键词1, 关键词2, 关键词3]
  文档2 → [关键词1, 关键词4]
  
  查"关键词1在哪些文档" → 全表扫描

倒排索引（ES）：
  关键词1 → [文档1, 文档2, 文档5, 文档8]
  关键词2 → [文档1, 文档3]
  关键词3 → [文档1, 文档4]
  
  查"关键词1在哪些文档" → 直接拿列表

ES 倒排索引结构：
  - 词项字典（Term Dictionary）：存储所有词项
  - 词项索引（Term Index）：FST 加速字典查找
  - 倒排列表（Posting List）：文档 ID 列表

为什么 ES 适合全文检索、MySQL 不适合：
1. 分词器：ES 内置多种分词器（IK 中文、Standard、SmartCN）
   MySQL 中文分词支持弱
2. 索引结构：ES 倒排索引为全文检索优化
   MySQL B+Tree 适合精确查找和范围，全文检索需全表扫
3. 分布式：ES 数据可分片到多节点并行查询
   MySQL 单机性能瓶颈
4. 相关性评分：ES 内置 TF-IDF / BM25 算法
   MySQL 全文检索评分简单

MySQL 全文索引（NGRAM）仅作为基础方案，
复杂搜索（高亮、聚合、模糊、相关性排序）必须用 ES。
```

---

### Q17：text 和 keyword 类型的区别？什么时候用哪个？

**面试官原话**：
> 这里 ES 的索引里面有 Tax 和 keyword 的类型，它这里有什么区别？什么时候用用 Tax 什么时候用 keyword？

**我的回答要点**：
- text：分词后支持全文检索
- keyword：精确匹配、聚合、排序
- 需要搜索 → text
- 需要精确过滤/分组 → keyword

**参考标准答案**（建议补充）：

```
text 类型：
- 会被分词器拆分（如 "iPhone 15 Pro Max" → ["iphone", "15", "pro", "max"]）
- 存储原文字段（store: true 时）+ 分词后的词项
- 支持全文检索（match query）
- 不支持聚合、排序（除非启用 fielddata，代价大）
- 适用：商品描述、文章正文、评论内容

keyword 类型：
- 不分词，原值存储
- 支持精确匹配（term query）、聚合（terms agg）、排序（sort）
- 不支持全文检索
- 最大长度 32,766 字节
- 适用：商品分类、标签、状态、邮箱、用户名
```

```json
// 典型映射设计
{
  "mappings": {
    "properties": {
      "title": {
        "type": "text",        // 文章标题支持全文检索
        "fields": {
          "keyword": {          // 同时存 keyword 用于精确匹配
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "category": {
        "type": "keyword"      // 分类用于聚合（terms agg）
      },
      "tags": {
        "type": "keyword"      // 标签数组
      },
      "content": {
        "type": "text",
        "analyzer": "ik_max_word"  // 中文分词
      },
      "status": {
        "type": "keyword"      // 状态字段（精确过滤）
      }
    }
  }
}

// text 检索
GET /articles/_search
{
  "query": {
    "match": {
      "title": "iPhone"  // 全文检索
    }
  }
}

// keyword 聚合
GET /articles/_search
{
  "aggs": {
    "by_category": {
      "terms": {
        "field": "category"  // 聚合
      }
    }
  }
}
```

---

## 六、容器编排（不熟悉）

### Q18：K3S 和 K3SD 有做过吗？

**面试官原话**：
> K3S 和 K3SD 有做过吗？你们这边。

**我的回答**：实际主要是 React 和全栈开发，这块接触不深，承认不熟悉。

---

## 七、MySQL 主从复制（实战偏弱）

### Q19：MySQL 主从复制原理？怎么保证主从数据一致？

**面试官原话**：
> mysql 这里我们通常遇见一种情况，就是 mysql 这里平有瓶颈嘛，它的存储那些，我们做主从复制这一块。mysql 的主从复制的原理是什么？怎么保证这个主从数据库

**我的回答要点**：
- 主库写入 binlog（二进制日志）
- 从库 IO 线程拉 binlog 写入 relay log
- 从库 SQL 线程回放 relay log 保持一致
- 核心：基于 binlog

**参考标准答案**（建议补充）：

```
MySQL 主从复制原理：

1. 主库（Master）：
   - 写入操作记录到 binlog（Binary Log）
   - binlog 格式：STATEMENT / ROW / MIXED
   - 默认 Mysql 5.7+ 推荐 ROW 格式

2. 从库（Slave）—— 三个线程：
   - I/O Thread：连接主库，请求 binlog 变更
   - Dump Thread（主库）：读取 binlog 推送给从库
   - SQL Thread：读取 relay log 并回放

3. 复制流程：
   主库写入 → binlog → 从库 I/O Thread 拉取 → relay log 
   → 从库 SQL Thread 回放 → 从库数据更新

复制模式：
- 异步复制（默认）：主库不等待从库确认，性能最好但可能丢数据
- 半同步复制：至少一个从库确认收到，性能折中
- 全同步复制：所有从库都执行完才返回（几乎不用）
```

```sql
-- 主库配置
[mysqld]
server-id = 1
log_bin = /var/log/mysql/mysql-bin.log
binlog_format = ROW
sync_binlog = 1

-- 从库配置
[mysqld]
server-id = 2
relay_log = /var/log/mysql/relay-bin.log
read_only = ON

-- 主库创建复制账号
CREATE USER 'repl'@'%' IDENTIFIED BY 'password';
GRANT REPLICATION SLAVE ON *.* TO 'repl'@'%';

-- 从库配置复制
CHANGE MASTER TO
  MASTER_HOST='master_ip',
  MASTER_USER='repl',
  MASTER_PASSWORD='password',
  MASTER_LOG_FILE='mysql-bin.000001',
  MASTER_LOG_POS=0;

START SLAVE;
SHOW SLAVE STATUS\G
```

---

### Q20：主从切换时怎么保证数据不丢失？

**面试官原话**：
> 那那那么这个主从切换，我因为主，可能主库压力大了，或者是主库宕机了，我主从进行主从切换的时候，我怎么保证数据不丢失？

**我的回答要点**：
- 半同步复制（至少一个从库确认）
- 配合定位断点（GTID）
- 监控延迟

**参考标准答案**（建议补充）：

```
保证数据不丢失的方案：

1. 半同步复制（Semi-Sync Replication）：
   - 主库写入 binlog 后，等待至少一个从库确认收到
   - 主库才返回客户端成功
   - 性能略差（多一次网络往返），但保证至少一份冗余

2. GTID（Global Transaction Identifier）：
   - 每个事务有全局唯一 ID
   - 切换时从库自动定位断点，无需手动指定 binlog 位点
   - MySQL 5.6+ 必用

3. 监控主从延迟：
   - Seconds_Behind_Master（SHOW SLAVE STATUS）
   - 延迟 > 阈值时禁止切换

4. 切换流程：
   ① 停止主库写入
   ② 等待从库追平（Seconds_Behind_Master = 0）
   ③ 提升从库为主（RESET MASTER; 或基于 GTID）
   ④ 应用层切换连接

5. MHA / Orchestrator：
   - 自动选主、切换、补偿
   - 业界成熟方案
```

```sql
-- 启用半同步复制
-- 主库
INSTALL PLUGIN rpl_semi_sync_master SONAME 'semisync_master.so';
SET GLOBAL rpl_semi_sync_master_enabled = 1;
SET GLOBAL rpl_semi_sync_master_timeout = 1000;  -- 1秒超时

-- 从库
INSTALL PLUGIN rpl_semi_sync_slave SONAME 'semisync_slave.so';
SET GLOBAL rpl_semi_sync_slave_enabled = 1;
STOP SLAVE; START SLAVE;
```

---

## 八、Redis（高频考点）

### Q21：Redis 淘汰策略有哪些？

**面试官原话**：
> 应用这方面，那那问几个 redis 常见的问题吧。我们现在，比如说你现在 redis 有大大内存，或者是那个会导致 redis 的内存满了，redis 的淘汰策略有哪几个策略？

**我的回答要点**：
- 8 种淘汰策略
- 默认 noeviction（直接报错）

**参考标准答案**（建议补充）：

```
Redis 8 种淘汰策略（maxmemory-policy）：

分类1：只针对有过期时间的键
1. volatile-lru       —— 在过期键中淘汰最近最少使用
2. volatile-lfu       —— 在过期键中淘汰最不经常使用（Redis 4.0+）
3. volatile-ttl       —— 淘汰最早过期的
4. volatile-random    —— 随机淘汰过期键

分类2：针对所有键
5. allkeys-lru        —— 所有键中淘汰最近最少使用（最常用）
6. allkeys-lfu        —— 所有键中淘汰最不经常使用
7. allkeys-random     —— 所有键中随机淘汰

分类3：不淘汰
8. noeviction         —— 默认策略，写满直接报错（OOM）

生产建议：
- 缓存场景：allkeys-lru（允许丢失部分缓存）
- 强一致场景：noeviction（写满报错由应用层处理）
- 热点数据明显：allkeys-lfu（比 LRU 更精准）
```

```bash
# 配置
CONFIG SET maxmemory 2gb
CONFIG SET maxmemory-policy allkeys-lru
```

---

### Q22：Redis 大 Key 和热 Key 怎么发现和处理？

**面试官原话**：
> 下一个下一个就是我们大 k 热 k 你会怎么发现 redis 里面的大 k 和热 k 然后并进行处理。

**我的回答要点**：

**大 Key 发现**：
- `redis-cli --bigkeys` 扫描整个实例
- 用 `redis-cli -h host -p port --bigkeys`

**大 Key 处理**：
- `UNLINK` 非阻塞删除
- 同步优化访问频率 + 设置过期时间

**热 Key**：
- 拆（三方面：分散到多个 key）
- 本地缓存
- 读写分离

**参考标准答案**（建议补充）：

```bash
# 大 Key 发现
redis-cli --bigkeys                    # 扫描整个实例
redis-cli --memkeys                    # 按内存排序（Redis 7.0+）
redis-cli -h host -p port --bigkeys -i 0.1  # 0.1秒间隔避免阻塞

# 自定义扫描大 Key
redis-cli --scan --pattern '*' | xargs -L 1 -I {} sh -c 'redis-cli OBJECT IDLETIME {} && redis-cli STRLEN {}'

# 大 Key 处理
UNLINK bigkey                          # 非阻塞删除（Redis 4.0+）
DEL bigkey                             # 阻塞删除（生产慎用）

# 热 Key 发现
redis-cli --hotkeys                    # Redis 7.0+ 支持
redis-cli INFO commandstats            # 看命令统计
monitor 命令采样（生产慎用，会降低 QPS 30%）
```

```go
// 热 Key 处理三板斧
// 1. Key 拆分
// 原：user:profile:123 → 改为多 key 分散
user:profile:123:basic → 基础信息
user:profile:123:ext   → 扩展信息

// 2. 本地缓存（应用层）
type HotKeyCache struct {
    localCache *sync.Map  // 或 LRU
    redis      *redis.Client
}

func (c *HotKeyCache) Get(key string) (string, error) {
    // 先查本地
    if v, ok := c.localCache.Load(key); ok {
        return v.(string), nil
    }
    // 再查 Redis
    v, err := c.redis.Get(ctx, key).Result()
    if err == nil {
        c.localCache.Store(key, v)
    }
    return v, err
}

// 3. 读写分离（一主多从）+ 读副本分担
```

---

### Q23：为什么线上禁止用 `KEYS *`？

**面试官原话**：
> 然后我我们为什么线上会要禁止用 k 的星呢？k 的星有什么不好的地方？

**我的回答要点**：
- 全量扫描 → CPU 飙升
- 内存暴涨
- 主从同步阻塞
- 阻塞 + 性能风险

**参考标准答案**（建议补充）：

```bash
# 错误用法：全量扫描（生产禁用）
KEYS *                 # O(N) 全库扫描，阻塞 Redis
KEYS user:*            # 同样阻塞

# 正确替代：SCAN 增量迭代
SCAN 0 MATCH user:* COUNT 1000  # 非阻塞，分批返回游标

# SCAN 返回值
# 1) "0"      ← 下次迭代起点，0 表示结束
# 2) 1) "user:1"
#    2) "user:2"
#    3) "user:100"
```

```
KEYS * 的危害：
1. CPU 飙升 —— 单线程遍历所有 key
2. 阻塞所有请求 —— KEYS 执行期间 Redis 不响应其他命令
3. 主从同步阻塞 —— 主库 KEYS 会让从库也跟着阻塞
4. 内存暴涨 —— 一次性返回大量数据撑爆客户端内存

SCAN 优势：
- 增量迭代，每次返回少量
- 不阻塞 Redis
- 可配合 MATCH 模式匹配
- COUNT 控制每次返回数量
```

---

## 九、外部三（SEO 考察）

### Q24：SEO 优化 sitemap 和 robots 了解过吗？

**面试官原话**：
> 外部三外部三一般的话，我们作为开发的话，虽然公司有配 seo 专员，但是我们做做项目的可能自己要学习和领悟这些 seo 的相关知识。setmap 和 nmap 这一块主要做什么？

**面试官原话**：
> setmap 和 roombot 站点地图这一块有了解过吗？

**我的回答要点**：
- 海南航空项目做过：服务端渲染、meta 标签、Google/Bingbot 适配
- 区分爬虫、提交 sitemap、JSON-LD 结构化数据
- 配置 robots、统一 canonical

**参考标准答案**（建议补充）：

```
robots.txt（爬虫规则）：
- 放在网站根目录 /robots.txt
- 告诉搜索引擎哪些路径可以/不可以抓取
- User-agent: * 适用所有爬虫
- Allow / Disallow

sitemap.xml（站点地图）：
- 列出网站所有重要页面
- 包含 lastmod、changefreq、priority
- 提交给 Google Search Console / Bing Webmaster
- 帮助搜索引擎发现和索引页面

meta 标签（首页 SEO）：
- title：60 字符以内，含核心关键词
- description：155 字符以内，吸引点击
- keywords：核心关键词（现代搜索引擎权重低）
- og:title / og:description / og:image（Open Graph，分享卡片）
- twitter:card（Twitter 分享卡片）
- canonical：规范 URL，避免重复内容

robots meta 标签（单页面）：
- <meta name="robots" content="index, follow">
- noindex / nofollow / noarchive
```

```html
<!-- robots.txt 示例 -->
User-agent: *
Allow: /
Disallow: /admin/
Disallow: /private/
Sitemap: https://example.com/sitemap.xml

<!-- 区分爬虫 -->
User-agent: Googlebot
Allow: /

User-agent: Bingbot
Allow: /

User-agent: AhrefsBot
Disallow: /
```

```xml
<!-- sitemap.xml 示例 -->
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/</loc>
    <lastmod>2026-09-08</lastmod>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://example.com/about</loc>
    <lastmod>2026-09-01</lastmod>
    <changefreq>monthly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>
```

```html
<!-- meta 标签配置 -->
<head>
  <title>主关键词 - 副关键词 - 品牌名</title>
  <meta name="description" content="页面摘要，包含核心关键词，吸引用户点击">
  <meta name="robots" content="index, follow">
  <link rel="canonical" href="https://example.com/page">
  
  <!-- Open Graph -->
  <meta property="og:title" content="分享标题">
  <meta property="og:description" content="分享描述">
  <meta property="og:image" content="https://example.com/og.jpg">
  <meta property="og:url" content="https://example.com/page">
  <meta property="og:type" content="website">
  
  <!-- Twitter Card -->
  <meta name="twitter:card" content="summary_large_image">
  <meta name="twitter:title" content="标题">
  <meta name="twitter:description" content="描述">
  <meta name="twitter:image" content="https://example.com/twitter.jpg">
  
  <!-- 结构化数据（JSON-LD） -->
  <script type="application/ld+json">
  {
    "@context": "https://schema.org",
    "@type": "WebSite",
    "name": "网站名",
    "url": "https://example.com"
  }
  </script>
</head>
```

---

## 十、反问环节 & 软性问题

### Q25：贵公司主要做什么业务？技术栈？

**我的提问**：
> 我想问一下，如果有幸进入贵公司的话，主要会做哪方面的一些任务和需求呢？

**面试官回答**：
- Golang 开发为主
- 项目性质：成人视频站 + APP + Web3（外部）
- 技术栈：Golang + ES + Redis + MySQL
- 部署：考虑用 K3S / K3SD 多 pod 部署

**面试官评价**：
> "你的 SEO 知识，说实话，一般吧，也不就是只是知道层面，而不是很了解。"

---

### Q26：工作时间是怎样的？

**我的回答**：
- 早上国内时间 10:00 ~ 22:00
- 中午休息 2 小时、晚上休息 1 小时
- 周一至周六
- 假期走国内假期

**面试官补充**：
> 工作时间可能长一点，人事有跟你说吗？

---

### Q27：薪资结算方式？

**面试官原话**：
> 薪资方面是走 USDT 结算，你这边要求薪资是多少？

**我的回答**：
- 第一次：20~30k（跨度太大，被要求具体）
- 第二次：20~25k

---

### Q28：签证情况？

**面试官原话**：
> 我想问一下，因为听候选人是在泰国，想知道你泰国现在目前是刷签还是什么签证吗？

**我的回答**：
- 目前是旅游签（泰国免签政策改后可待 30 天）
- 老挝有劳务签证可长期待
- 越南有签证可待 3 个月

---

## 十一、面试整体复盘

### 自我评估

**✅ 答得不错**：
- GMP 模型（基础概念正确）
- Go error 处理（哨兵+包装）
- Channel 关闭读写（正确且知道黄金法则）
- ES 分片副本、深分页（PIT）、倒排索引、text/keyword
- Redis 大 Key / 热 Key 发现
- 多 Agent 协作和记忆分层设计

**⚠️ 答得一般**：
- 多 Agent 死循环处理（提到了状态机+熔断+置信度，但不够具体）
- K3S/K3SD（直接承认不熟悉）
- MySQL 主从切换（半同步+GTID 概念对，细节弱）

**❌ 需要补强**：
- **SEO 知识**：面试官评价"只是知道层面，不是很了解"
  - 需深入：robots.txt 高级用法、sitemap 提交流程
  - 需深入：canonical、结构化数据（JSON-LD）各种 schema
  - 需深入：Core Web Vitals、移动适配
- **K3S / K3SD**：需要至少了解基本概念和部署
  - K3S = 轻量级 K8s（IoT/边缘场景）
  - K3SD = K3S 的 daemon 模式（单机部署）
- **MySQL 主从切换**：需要熟悉完整流程
  - MHA / Orchestrator 自动切换
  - GTID 模式 vs 传统 binlog 模式
  - 半同步复制的延迟影响

### 下次面试准备清单

1. **AI/Agent 部分**：补充置信度仲裁的具体算法、加权投票实现
2. **Go 部分**：补充 GMP 数据结构（runqueue、schedt）、调度触发时机
3. **ES 部分**：补充 PIT + search_after 完整代码示例
4. **MySQL**：补充 MHA 高可用架构、读写分离 + 分库分表
5. **SEO**：系统学习（推荐《SEO 实战密码》、Ahrefs 博客）
6. **K8S/K3S**：至少理解基本概念、yaml 部署、Service/Ingress

### 备注

- 工作时间偏长（10:00 ~ 22:00，周一至周六）
- 薪资 USDT 结算
- 公司做成人视频（Web3 方向），需评估是否接受
- 远程办公，候选人在泰国
