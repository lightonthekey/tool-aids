// 该示例演示 tool-aids 各功能的基本用法：
//
//	go run ./examples/demo
package main

import (
	"fmt"
	"log"

	"github.com/lightonthekey/tool-aids/ipqqwry"
	"github.com/lightonthekey/tool-aids/ipv6"
	gojieba "github.com/lightonthekey/tool-aids/jieba"
	"github.com/lightonthekey/tool-aids/piquery"
)

func main() {
	fmt.Println("== IPv4 归属地 ==")
	fmt.Println(ipqqwry.FindAddress("111.207.139.90"))
	for _, r := range ipqqwry.FindAddressList([]string{"111.207.19.90", "114.22.63.70", "61.65.188.80"}) {
		fmt.Printf("%s\t%s\t%s\n", r.IP, r.Country, r.Area)
	}
	q := ipqqwry.NewQQwry()
	fmt.Println(q.Find("111.207.139.90"))

	fmt.Println("== 中文分词 ==")
	words, err := gojieba.JiebaWords("田野上方覆盖着一层烟雾，迷迷蒙蒙的，仿佛一幅大肆渲染的水墨画。", true)
	if err != nil {
		log.Println("分词失败:", err)
	}
	fmt.Println(words)
	fmt.Println(gojieba.JiebaSegoPlusV("我爱北京天安门", false))

	fmt.Println("== IPv6 归属地 ==")
	info, err := ipv6.FindAddressWithLang("2001:4860:4860::8888", "cn")
	fmt.Printf("%+v err=%v\n", info, err)

	fmt.Println("== 圆周率 ==")
	r, err := piquery.Query(2)
	fmt.Printf("第%d位=%s 前=%s 后=%s err=%v\n", r.Position, r.Current, r.Previous, r.Next, err)
	fmt.Println("最大可查位置:", piquery.MaxPosition())
}
