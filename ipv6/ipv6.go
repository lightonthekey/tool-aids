// 怎么更新？
// 下载最新版本替换文件： IP2LOCATION-LITE-DB3.IPV6.BIN
// 数据库下载地址 https://lite.ip2location.com/database-download

// Package ipv6 提供一些用于处理 ipv6 地址的工具函数。
package ipv6

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"

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

var ipv6db *ip2location.DB

func init() {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not get current file information（jieba.go）")
	}

	currentDir := filepath.Dir(currentFile)
	dictPath := filepath.Join(currentDir, "data", "IP2LOCATION-LITE-DB3.IPV6.BIN")

	var err error
	ipv6db, err = ip2location.OpenDB(dictPath)
	if err != nil {
		log.Fatalf("Failed to open ipv6 db: %v", err)
	}
}

// 查询地址，返回英语
func FindAddress(ipv6 string) (LocationInfo, error) {
	results, err := ipv6db.Get_all(ipv6)
	if err != nil {
		return LocationInfo{}, fmt.Errorf("Failed to get location for %s: %v", ipv6, err)
	}

	data := LocationInfo{
		CountryShort: results.Country_short,
		CountryLong:  results.Country_long,
		Region:       results.Region,
		City:         results.City,
		Ip:           ipv6,
	}
	return data, nil
}

// 查询地址，返回中文
func FindAddressToChinese(ipv6 string) (LocationInfo, error) {
	results, err := ipv6db.Get_all(ipv6)
	if err != nil {
		return LocationInfo{}, fmt.Errorf("Failed to get location for %s: %v", ipv6, err)
	}

	data := LocationInfo{
		CountryShort: results.Country_short,
		CountryLong:  results.Country_long,
		Region:       results.Region,
		City:         results.City,
		Ip:           ipv6,
	}
	data = langTranslate(data)
	return data, nil
}

// 语言翻译
func langTranslate(info LocationInfo) LocationInfo {
	var countryMap = data.CountryShortMap
	var cityMap = data.CityyMap
	var regionMap = data.RegionMap
	if val, ok := cityMap[info.City]; ok {
		info.City = val
	}
	if val, ok := countryMap[info.CountryShort]; ok {
		info.CountryLong = val
	}
	if val, ok := regionMap[info.Region]; ok {
		info.Region = val
	}
	return info
}
