# MyAnts

![Static Badge](https://img.shields.io/badge/my_ants-v1.0-blue)

基于 [ants](https://github.com/panjf2000/ants/v2) 协程池的泛型任务并发执行框架，支持**速率限制**和 **context 超时控制**。

## 特性

- **泛型支持** — 输入参数 `T` 和返回结果 `R` 均为泛型，适配任意业务类型
- **协程池** — 底层使用 `ants` 高性能协程池，控制并发 goroutine 数量
- **速率限制** — 集成 `golang.org/x/time/rate`，可精确控制每秒任务提交频率
- **Context 超时** — 支持通过 `context.WithTimeout` 控制整体执行时长
- **流式结果** — 通过 channel 实时返回已完成任务的结果

## 依赖

```shell
go get -u github.com/panjf2000/ants/v2
go get -u golang.org/x/time
```

## 核心 API

### `NewMyAnts[T, R](size int, hz int, ctx context.Context) (MyAnts[T, R], error)`

创建一个 MyAnts 实例。

| 参数 | 类型 | 说明 |
|------|------|------|
| `size` | `int` | 协程池大小（最大并发 goroutine 数） |
| `hz` | `int` | 速率限制（每秒允许提交的任务数） |
| `ctx` | `context.Context` | 用于超时/取消控制的 context |

### `UseMyAnts(task TaskFunc[T, R], param T) (chan R, error)`

提交任务并返回结果 channel。

| 参数 | 类型 | 说明 |
|------|------|------|
| `task` | `TaskFunc[T, R]` | 任务函数，签名为 `func(ctx context.Context, param T, page int) (R, error)` |
| `param` | `T` | 输入参数，需实现 `SizeAble` 接口（提供 `Len() int` 方法） |

**返回值：** `chan R` — 从该 channel 中读取每个任务的执行结果，所有任务完成后 channel 自动关闭。

### `SizeAble` 接口

输入参数类型 `T` 必须实现此接口，用于确定任务总数：

```go
type SizeAble interface {
    Len() int
}
```

## 使用示例

### 1. 定义输入和输出类型

```go
// 输入类型 — 必须实现 SizeAble 接口
type InFunc struct {
    DictName []map[int]string
}

func (inf InFunc) Len() int {
    return len(inf.DictName)
}

// 输出类型
type OutFunc struct {
    string
}
```

### 2. 创建实例

```go
// 创建协程池：5个并发 goroutine，每秒最多提交500个任务，10秒超时
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

pool, err := myants.NewMyAnts[InFunc, OutFunc](5, 500, ctx)
if err != nil {
    panic(err)
}
```

### 3. 定义任务函数

```go
taskfn := myants.TaskFunc[InFunc, OutFunc](func(ctx context.Context, param InFunc, page int) (OutFunc, error) {
    select {
    case <-ctx.Done():
        return OutFunc{}, ctx.Err()
    default:
        // 根据 page 下标处理对应数据
        value := param.DictName[page]
        return OutFunc{string: fmt.Sprintf("结果: %d - %v", page, value)}, nil
    }
})
```

### 4. 提交任务并消费结果

```go
input := InFunc{
    DictName: []map[int]string{
        {1: "张飞"},
        {2: "李逵"},
        {3: "吕布"},
    },
}

resultCh, err := pool.UseMyAnts(taskfn, input)
if err != nil {
    panic(err)
}

// 从 channel 中读取结果（channel 关闭时循环自动结束）
for res := range resultCh {
    log.Println(res)
}
```

## 参数调优建议

| 场景 | `size` | `hz` | 说明 |
|------|--------|------|------|
| 调用外部 API（有限流） | 3~10 | 按 API 限流设定 | 避免触发对方限流 |
| CPU 密集型计算 | `runtime.NumCPU()` | 不限 (大值) | 充分利用 CPU |
| IO 密集型（数据库/文件） | 20~100 | 视数据库连接池大小 | 避免超出连接池 |

## 注意事项

- `param.Len()` 返回 0 时，会直接返回已关闭的空 channel，不会提交任何任务
- 任务函数中应检查 `ctx.Done()`，以便在超时时及时退出
- 结果 channel 带有 10 的缓冲区，消费端应及时读取避免阻塞 worker
- 所有 worker 完成后 channel 自动关闭，使用 `for range` 即可安全消费

## License

MIT