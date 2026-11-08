// Package ipv6 基于 IP2Location LITE 数据库（IP2LOCATION-LITE-DB3.IPV6.BIN）
// 提供 IPv6/IPv4 归属地查询。
//
// # 数据来源
//
//	https://lite.ip2location.com/database-download 下载 IP2LOCATION-LITE-DB3.IPV6.BIN，
//	替换 data 目录下的同名文件即可。
//
// # 数据文件
//
//	地址库约 94MB，不内嵌进二进制。查找顺序（优先级从高到低）：
//	  Init(path) / TOOL_AIDS_IPV6_DB 环境变量 / 可执行文件同级 data 目录 / 源码目录 data 目录
//	加载是延迟的：只有真正查询时才打开文件。文件缺失不会让进程退出，
//	只会让查询返回错误。
package ipv6

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ip2location/ip2location-go"
	"github.com/lightonthekey/tool-aids/ipv6/data"
)

// LocationInfo 表示一个地理位置的信息。
type LocationInfo struct {
	Ip           string
	CountryShort string
	CountryLong  string
	Region       string
	City         string
}

const dbFileName = "IP2LOCATION-LITE-DB3.IPV6.BIN"

var (
	openMu   sync.Mutex
	db       *ip2location.DB
	dbPath   string // 已加载 / 显式指定的路径
	openErr  error
	override string // Init 指定的路径
)

// Init 显式指定地址库文件路径并立即加载。传空字符串表示恢复自动查找。
// 加载失败时返回错误，且不会影响进程其他功能。
func Init(path string) error {
	openMu.Lock()
	defer openMu.Unlock()
	if db != nil {
		db.Close()
		db = nil
	}
	openErr = nil
	override = path
	if path == "" {
		dbPath = ""
		return nil
	}
	d, err := ip2location.OpenDB(path)
	if err != nil {
		openErr = fmt.Errorf("打开 IPv6 地址库失败(%s): %w", path, err)
		return openErr
	}
	db, dbPath = d, path
	return nil
}

// DBPath 返回当前使用的地址库路径，未加载时返回空字符串。
func DBPath() string {
	openMu.Lock()
	defer openMu.Unlock()
	return dbPath
}

// Close 释放地址库句柄，下次查询会自动重新打开（同时清空失败缓存）。
func Close() {
	openMu.Lock()
	defer openMu.Unlock()
	openErr = nil
	if db != nil {
		db.Close()
		db = nil
		dbPath = ""
	}
}

// open 延迟打开地址库（并发安全，失败结果会被缓存避免反复读盘）。
func open() (*ip2location.DB, error) {
	openMu.Lock()
	defer openMu.Unlock()
	if db != nil {
		return db, nil
	}
	if openErr != nil {
		return nil, openErr
	}
	path, err := resolveDB()
	if err != nil {
		openErr = err
		return nil, err
	}
	d, err := ip2location.OpenDB(path)
	if err != nil {
		openErr = fmt.Errorf("打开 IPv6 地址库失败(%s): %w", path, err)
		return nil, openErr
	}
	db, dbPath = d, path
	return db, nil
}

// resolveDB 依次查找地址库文件。
func resolveDB() (string, error) {
	if override != "" {
		if fileExists(override) {
			return override, nil
		}
		return "", fmt.Errorf("Init 指定的 IPv6 地址库不存在: %s", override)
	}
	if p := strings.TrimSpace(os.Getenv("TOOL_AIDS_IPV6_DB")); p != "" {
		if fileExists(p) {
			return p, nil
		}
		return "", fmt.Errorf("环境变量 TOOL_AIDS_IPV6_DB 指定的地址库不存在: %s", p)
	}
	if exe, err := os.Executable(); err == nil {
		if p := filepath.Join(filepath.Dir(exe), "data", dbFileName); fileExists(p) {
			return p, nil
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		if p := filepath.Join(filepath.Dir(file), "data", dbFileName); fileExists(p) {
			return p, nil
		}
	}
	return "", fmt.Errorf("找不到 IPv6 地址库 %s，请通过 Init 或环境变量 TOOL_AIDS_IPV6_DB 指定", dbFileName)
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// FindAddress 查询地址，返回英文。
func FindAddress(ip string) (LocationInfo, error) {
	return lookup(ip, false)
}

// FindAddressToChinese 查询地址，返回中文（国家/地区/城市做了翻译）。
func FindAddressToChinese(ip string) (LocationInfo, error) {
	return lookup(ip, true)
}

// FindAddressWithLang 按语言查询，一次搞定，避免分别调用造成重复查询。
// lang 为 "cn"/"zh"/"zh-cn"/"zh_cn" 时返回中文，其余返回英文。
func FindAddressWithLang(ip, lang string) (LocationInfo, error) {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "cn", "zh", "zh-cn", "zh_cn", "zh-hans":
		return lookup(ip, true)
	}
	return lookup(ip, false)
}

func lookup(ip string, chinese bool) (LocationInfo, error) {
	ip = strings.TrimSpace(ip)
	info := LocationInfo{Ip: ip}
	if net.ParseIP(ip) == nil {
		return info, fmt.Errorf("无效的 IP 地址: %q", ip)
	}

	d, err := open()
	if err != nil {
		return info, err
	}
	results, err := d.Get_all(ip)
	if err != nil {
		return info, fmt.Errorf("查询 %s 失败: %w", ip, err)
	}

	info.CountryShort = normalized(results.Country_short)
	info.CountryLong = normalized(results.Country_long)
	info.Region = normalized(results.Region)
	info.City = normalized(results.City)
	if chinese {
		info = langTranslate(info)
	}
	return info, nil
}

// normalized 把 ip2location 的占位返回（无效地址、数据库无该字段）统一成空字符串，
// 避免把 "Invalid IP address." 之类的字符串透给前端。
func normalized(s string) string {
	s = strings.TrimSpace(s)
	switch s {
	case "-", "N/A", "Invalid IP address.", "This parameter is unavailable for selected data file.":
		return ""
	}
	return s
}

// langTranslate 语言翻译
func langTranslate(info LocationInfo) LocationInfo {
	if val, ok := data.CityyMap[info.City]; ok {
		info.City = val
	}
	if val, ok := data.CountryShortMap[info.CountryShort]; ok {
		info.CountryLong = val
	}
	if val, ok := data.RegionMap[info.Region]; ok {
		info.Region = val
	}
	return info
}
