# piquery - 圆周率位数查询

查询圆周率指定位数的数字及前后各 5 位，数据内嵌在二进制中，毫秒级（实际是纳秒级）响应。

## 特点

- 数据文件 `pi_digits.txt` 通过 `go:embed` 内嵌，无需外部依赖
- 支持查询第 1 到 1000001 位（第 1 位是整数位 `3`）
- 首次调用校验一次数据（约 0.7ms），之后查询是纯内存切片，约 40ns
- `NewPIStore()` 返回共享只读实例，并发安全

## 使用

```go
import "github.com/lightonthekey/tool-aids/piquery"

// 包级快捷方法
r, err := piquery.Query(100)
fmt.Printf("第%d位: %s，前5位: %s，后5位: %s\n", r.Position, r.Current, r.Previous, r.Next)
fmt.Println("最大位置:", piquery.MaxPosition())

// 或者自行持有一个 Store
store, err := piquery.NewPIStore()
r, err = store.Query(100)
```

## API

### func Query(position int) (Result, error)

从共享实例查询；`position` 从 1 开始，对应数据串下标 `position-1`。
注意第 1 位是整数位 `3`，如需“小数点后第 N 位”请传 `N+1`。

### func NewPIStore() (*PIStore, error)

返回共享实例，首次调用会校验 100 万个字节，之后为 O(1)。返回值只读，可并发使用。

### func MaxPosition() int / (*PIStore).MaxPosition() int

最大可查询位置（当前数据为 1000001）。

### type Result

| 字段 | 说明 |
| --- | --- |
| `Position` | 查询的位置 |
| `Current` | 该位置的数字（字符串） |
| `Previous` | 前 5 位（不足则为空串） |
| `Next` | 后 5 位（不足则为空串） |

## 更新数据

1. 从可靠来源获取圆周率前 100 万位小数（纯数字，不含小数点，前面拼接整数位 `3`）
2. 替换 `piquery/pi_digits.txt`
3. 重新编译使用该库的项目

## 许可证

MIT，详见仓库根目录 [LICENSE](../LICENSE)。
