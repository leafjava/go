# Go 语言快速参考手册

> 适合有 Java 基础的开发者快速查阅

## 基础语法对比

### 变量声明

```go
// Go
var name string = "林燊"
age := 23

// Java
String name = "林燊";
int age = 23;
```

### 函数定义

```go
// Go
func add(a, b int) int {
    return a + b
}

// Java
public int add(int a, int b) {
    return a + b;
}
```

### 错误处理

```go
// Go
result, err := divide(10, 0)
if err != nil {
    fmt.Println("错误:", err)
}

// Java
try {
    result = divide(10, 0);
} catch (Exception e) {
    System.out.println("错误: " + e.getMessage());
}
```

## 常用数据结构

### Slice（切片）

```go
// 创建
s := []int{1, 2, 3}
s := make([]int, 5)
s := make([]int, 5, 10)  // 长度5，容量10

// 追加
s = append(s, 4, 5, 6)

// 切片
s[1:3]  // [2, 3]
s[:3]   // [1, 2, 3]
s[3:]   // [4, 5, 6]

// 遍历
for i, v := range s {
    fmt.Println(i, v)
}
```

### Map（映射）

```go
// 创建
m := make(map[string]int)
m := map[string]int{"a": 1, "b": 2}

// 添加/修改
m["c"] = 3

// 读取
value := m["a"]
value, exists := m["d"]  // 检查是否存在

// 删除
delete(m, "a")

// 遍历
for key, value := range m {
    fmt.Println(key, value)
}
```

## 结构体和方法

```go
// 定义结构体
type User struct {
    ID   int
    Name string
}

// 值接收者（不修改原数据）
func (u User) GetName() string {
    return u.Name
}

// 指针接收者（修改原数据）
func (u *User) SetName(name string) {
    u.Name = name
}

// 使用
user := User{ID: 1, Name: "Alice"}
user.SetName("Bob")
```

## 接口

```go
// 定义接口
type Writer interface {
    Write(data []byte) error
}

// 实现接口（隐式）
type FileWriter struct{}

func (f *FileWriter) Write(data []byte) error {
    // 实现
    return nil
}

// 使用
var w Writer = &FileWriter{}
w.Write([]byte("hello"))
```

## 并发编程

### Goroutine

```go
// 启动 goroutine
go func() {
    fmt.Println("异步执行")
}()

// 带参数
go processData(data)
```

### Channel

```go
// 创建 channel
ch := make(chan int)
bufferedCh := make(chan int, 10)

// 发送
ch <- 42

// 接收
value := <-ch

// 关闭
close(ch)

// 遍历
for value := range ch {
    fmt.Println(value)
}
```

### Select

```go
select {
case msg := <-ch1:
    fmt.Println("ch1:", msg)
case msg := <-ch2:
    fmt.Println("ch2:", msg)
case <-time.After(1 * time.Second):
    fmt.Println("超时")
}
```

### WaitGroup

```go
var wg sync.WaitGroup

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        // 处理任务
    }(i)
}

wg.Wait()  // 等待所有 goroutine 完成
```

## Gin 框架

### 基本路由

```go
r := gin.Default()

// GET
r.GET("/users", func(c *gin.Context) {
    c.JSON(200, gin.H{"users": []string{"Alice", "Bob"}})
})

// POST
r.POST("/users", func(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, user)
})

// 路径参数
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})

// 查询参数
r.GET("/search", func(c *gin.Context) {
    keyword := c.Query("keyword")
    page := c.DefaultQuery("page", "1")
    c.JSON(200, gin.H{"keyword": keyword, "page": page})
})
```

### 中间件

```go
// 全局中间件
r.Use(gin.Logger())
r.Use(gin.Recovery())

// 路由组中间件
authorized := r.Group("/api", AuthMiddleware())
{
    authorized.GET("/users", getUsers)
}

// 自定义中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        // 验证 token
        c.Next()
    }
}
```

## GORM

### 模型定义

```go
type User struct {
    gorm.Model
    Name  string
    Email string `gorm:"uniqueIndex"`
    Age   int
}
```

### CRUD 操作

```go
// 创建
db.Create(&user)

// 查询
db.First(&user, 1)  // 主键查询
db.Where("name = ?", "Alice").First(&user)
db.Find(&users)  // 查询所有

// 更新
db.Model(&user).Update("name", "Bob")
db.Model(&user).Updates(User{Name: "Bob", Age: 25})

// 删除
db.Delete(&user, 1)
```

## Web3 开发

### 以太坊

```go
import (
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
)

// 连接节点
client, _ := ethclient.Dial("https://eth.llamarpc.com")

// 查询余额
address := common.HexToAddress("0x742d35Cc...")
balance, _ := client.BalanceAt(context.Background(), address, nil)

// 查询区块
block, _ := client.BlockByNumber(context.Background(), nil)
```

## 常用命令

```bash
# 初始化项目
go mod init project-name

# 安装依赖
go get github.com/gin-gonic/gin

# 运行
go run main.go

# 编译
go build

# 测试
go test ./...

# 格式化代码
go fmt ./...

# 查看依赖
go mod tidy
```

## 常见错误

### 1. 未使用的变量

```go
// 错误
name := "Alice"  // declared and not used

// 解决：使用 _ 忽略
_ = name
```

### 2. nil 指针

```go
// 错误
var p *int
fmt.Println(*p)  // panic

// 解决：检查 nil
if p != nil {
    fmt.Println(*p)
}
```

### 3. Slice 越界

```go
// 错误
s := []int{1, 2, 3}
fmt.Println(s[5])  // panic

// 解决：检查长度
if len(s) > 5 {
    fmt.Println(s[5])
}
```

## 性能优化技巧

1. **使用指针接收者**（避免复制大结构体）
2. **预分配 Slice 容量**：`make([]int, 0, 100)`
3. **使用 sync.Pool** 复用对象
4. **避免不必要的字符串拼接**（使用 strings.Builder）
5. **使用 Context 控制超时**

## 调试技巧

```go
// 打印变量类型
fmt.Printf("%T\n", variable)

// 打印详细信息
fmt.Printf("%+v\n", struct)

// 打印调用栈
debug.PrintStack()
```

---

**更多内容请查看完整教程** → [课程大纲](./COURSE_OUTLINE.md)
