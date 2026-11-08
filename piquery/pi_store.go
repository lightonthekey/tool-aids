// Package piquery 提供圆周率指定位数的快速查询。
//
// 数据文件 pi_digits.txt 内嵌在二进制中（约 1MB，含 100 万位小数），
// 首次使用时校验一次，之后所有查询都是纯内存切片操作。
package piquery

import (
	_ "embed"
	"errors"
	"fmt"
	"sync"
)

//go:embed pi_digits.txt
var piDigits []byte

// PIStore 存储圆周率数据并提供查询功能。
// 实例是只读的，可安全并发使用。
type PIStore struct {
	digits []byte
	maxPos int
}

// Result 包含查询结果的结构体
type Result struct {
	Position int    // 查询位置（1 表示数据串第 1 个字符，即整数位 "3"）
	Current  string // 当前位置的数字
	Previous string // 前 5 位
	Next     string // 后 5 位
}

var (
	storeOnce sync.Once
	store     *PIStore
	storeErr  error
)

// load 只做一次全量校验，避免每个请求都扫一遍 100 万字节。
func load() (*PIStore, error) {
	storeOnce.Do(func() {
		for i, b := range piDigits {
			if b < '0' || b > '9' {
				storeErr = fmt.Errorf("圆周率数据文件第 %d 个字节不是数字（0x%02x）", i, b)
				return
			}
		}
		store = &PIStore{digits: piDigits, maxPos: len(piDigits)}
	})
	return store, storeErr
}

// NewPIStore 返回共享的 PIStore 实例（首次调用会校验数据，之后为 O(1)）。
// 返回值只读，不要修改；并发调用安全。
func NewPIStore() (*PIStore, error) {
	return load()
}

// Query 查询指定位置的圆周率数字及前后各 5 位。
func Query(position int) (Result, error) {
	s, err := load()
	if err != nil {
		return Result{}, err
	}
	return s.Query(position)
}

// MaxPosition 返回最大可查询的位置。
func MaxPosition() int {
	s, err := load()
	if err != nil {
		return 0
	}
	return s.maxPos
}

// Query 查询指定位置的圆周率数字及前后各 5 位。
// position 从 1 开始，对应数据串下标 position-1（第 1 位是整数位 "3"，
// 如需“小数点后第 N 位”请传 N+1）。
func (p *PIStore) Query(position int) (Result, error) {
	if p == nil || p.maxPos == 0 {
		return Result{}, errors.New("圆周率数据未初始化")
	}
	if position < 1 || position > p.maxPos {
		return Result{}, fmt.Errorf("位置超出范围，必须在 1 到 %d 之间", p.maxPos)
	}

	index := position - 1

	startPrev := index - 5
	if startPrev < 0 {
		startPrev = 0
	}
	endNext := index + 6
	if endNext > p.maxPos {
		endNext = p.maxPos
	}

	return Result{
		Position: position,
		Current:  string(p.digits[index]),
		Previous: string(p.digits[startPrev:index]),
		Next:     string(p.digits[index+1 : endNext]),
	}, nil
}

// MaxPosition 返回最大可查询的位置。
func (p *PIStore) MaxPosition() int {
	if p == nil {
		return 0
	}
	return p.maxPos
}
