# piquery - 高效的圆周率查询库

piquery是一个Go语言开源库，用于高效查询圆周率的指定位数及其前后数字。该库采用预存数据的方式，支持查询1到100万位的圆周率数字，提供毫秒级的查询响应。

## 特点

- 采用嵌入式数据文件，无需外部依赖
- 支持查询1到100万位的圆周率数字
- 每次查询返回指定位置的数字及其前后各5位数字
- 简单易用的API接口
- 完善的错误处理

## 安装go get github.com/yourusername/piquery
## 使用示例 package main
```
import (
	"fmt"
	"log"
	"piquery"
)

func main() {
	// 初始化PIStore
	store, err := piquery.NewPIStore()
	if err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	fmt.Printf("支持查询的最大位数: %d\n", store.MaxPosition())

	// 查询第100位
	result, err := store.Query(100)
	if err != nil {
		log.Fatalf("查询失败: %v", err)
	}

	// 输出结果
	fmt.Printf("第%d位: %c\n", result.Position, result.CurrentDigit)
	fmt.Printf("前5位: %s\n", result.PreviousDigits)
	fmt.Printf("后5位: %s\n", result.NextDigits)
}
```
## API文档

### PIStore

`PIStore`是圆周率查询的核心结构体，用于存储圆周率数据并提供查询功能。

#### NewPIStore() (*PIStore, error)

创建一个新的PIStore实例，自动加载嵌入式的圆周率数据。

#### Query(position int) (Result, error)

查询指定位置的圆周率数字。`position`是要查询的位置，从1开始。返回一个`Result`结构体和可能的错误。

#### MaxPosition() int

返回最大可查询的位置（即预存的圆周率总位数）。

### Result

`Result`结构体包含查询结果：

- `Position`: 查询的位置
- `CurrentDigit`: 当前位置的数字（byte类型，可直接转换为字符）
- `PreviousDigits`: 前5位数字（如果不足5位则返回实际可获取的数量）
- `NextDigits`: 后5位数字（如果不足5位则返回实际可获取的数量）

## 数据文件补充

该库包含一个`pi_digits.txt`文件，目前包含圆周率的前若干位数字。作为创作者，你可以通过以下步骤补充完整至100万位：

1. 从可靠来源获取圆周率前100万位数字（纯数字，不含小数点）
2. 替换`piquery/pi_digits.txt`文件内容
3. 重新编译使用该库的项目

## 许可证

本项目采用MIT许可证，详情参见LICENSE文件。
