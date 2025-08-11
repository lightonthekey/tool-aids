package piquery

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
)

// 嵌入圆周率数据文件
//
//go:embed pi_digits.txt
var piFS embed.FS

// PIStore 存储圆周率数据并提供查询功能
type PIStore struct {
	digits []byte
	maxPos int
}

// Result 包含查询结果的结构体
type Result struct {
	Position int
	Current  string
	Previous string
	Next     string
}

// NewPIStore 初始化PIStore，加载嵌入式的圆周率数据
func NewPIStore() (*PIStore, error) {
	// 读取嵌入式文件
	data, err := fs.ReadFile(piFS, "pi_digits.txt")
	if err != nil {
		return nil, errors.New("Failed to read pi digits file: " + err.Error())
	}

	// 验证数据格式（确保只包含数字）
	for _, b := range data {
		if b < '0' || b > '9' {
			return nil, errors.New("Data file contains non-digit characters")
		}
	}

	return &PIStore{
		digits: data,
		maxPos: len(data),
	}, nil
}

// Query 查询指定位置的圆周率数字及前后各5位
func (p *PIStore) Query(position int) (Result, error) {
	// 验证位置有效性（位置从1开始）
	if position < 1 || position > p.maxPos {
		return Result{}, fmt.Errorf("Invalid position, must be between 1 and %d", p.maxPos)
	}

	// 转换为0基索引
	index := position - 1

	// 获取当前数字
	current := string(p.digits[index])

	// 获取前5位数字（转换为字符串）
	startPrev := index - 5
	if startPrev < 0 {
		startPrev = 0
	}
	previous := string(p.digits[startPrev:index]) // 字节切片转字符串

	// 获取后5位数字（转换为字符串）
	endNext := index + 6
	if endNext > p.maxPos {
		endNext = p.maxPos
	}
	next := string(p.digits[index+1 : endNext]) // 字节切片转字符串

	return Result{
		Position: position,
		Current:  current,  // 赋值字符串
		Previous: previous, // 赋值字符串
		Next:     next,     // 赋值字符串
	}, nil

}

// MaxPosition 返回最大可查询的位置
func (p *PIStore) MaxPosition() int {
	return p.maxPos
}
