package main

import (
	"fmt"

	"github.com/lightonthekey/tool-aids/ipqqwry"
	"github.com/lightonthekey/tool-aids/ipv6"
	gojieba "github.com/lightonthekey/tool-aids/jieba"
)

func main() {
	var data = ipqqwry.FindAddress("111.207.139.90")
	fmt.Println(data)

	var data2 = ipqqwry.FindAddressAll([]string{"111.207.19.90", "114.22.63.70", "61.65.188.80"})
	fmt.Println(data2)

	xx := ipqqwry.NewQQwry()
	var data3 = xx.Find("111.207.139.90")
	fmt.Println(data3)

	var jieba = gojieba.JiebaSego("田野上方覆盖着一层烟雾，迷迷蒙蒙的，仿佛一幅大肆渲染的水墨画。", true)
	fmt.Println(jieba)

	data6, _ := ipv6.FindAddress("111.207.139.90")
	fmt.Println(data6)

}
