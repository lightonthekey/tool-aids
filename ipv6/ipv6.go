// 怎么更新？
// 下载最新版本替换文件： IP2LOCATION-LITE-DB3.IPV6.BIN
// 数据库下载地址 https://lite.ip2location.com/database-download

// Package ipv6 提供一些用于处理 ipv6 地址的工具函数。
package ipv6

import (
	"fmt"
	"os"

	"github.com/ip2location/ip2location-go"
)

// LocationInfo 表示一个地理位置的信息。
type LocationInfo struct {
	CountryShort string `json:"Country_short"`
	CountryLong  string `json:"Country_long"`
	Region       string `json:"Region"`
	City         string `json:"City"`
}

// FindAddress 接受一个 ipv6 地址作为输入，返回该地址对应的地理位置信息。
// 因为它依赖于一个地理位置数据库，路径需要提前设置好。
// 如果没有找到对应的地理位置信息，或者发生其他错误，它会返回一个空的 LocationInfo 和一个非nil的error.
func FindAddress(ipv6 string) (LocationInfo, error) {
	demo_path, _ := os.Getwd()
	var data_path = demo_path + "/ipv6/data/IP2LOCATION-LITE-DB3.IPV6.BIN"

	db, err := ip2location.OpenDB(data_path)
	if err != nil {
		return LocationInfo{}, fmt.Errorf("open DB failed: %v", err)
	}

	results, err := db.Get_all(ipv6)
	if err != nil {
		return LocationInfo{}, fmt.Errorf("get all failed: %v", err)
	}

	data := LocationInfo{
		CountryShort: results.Country_short,
		CountryLong:  results.Country_long,
		Region:       results.Region,
		City:         results.City,
	}
	return data, nil
}
