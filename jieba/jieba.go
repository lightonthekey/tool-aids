// jieba中文分词
// 参考 https://github.com/huichen/sego

package jieba

import (
	"log"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/huichen/sego"
)

type Jieba struct{}

// 载入词典
var goseg sego.Segmenter

func startInit() {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not get current file information（jieba.go）")
	}

	// 获取当前文件所在目录
	currentDir := filepath.Dir(currentFile)
	// 构建文件路径
	dictionaryPath := filepath.Join(currentDir, "data", "jieba_dictionary.txt")
	goseg.LoadDictionary(dictionaryPath)
}

// 字符串
// model 普通模式false, 搜索模式true
func JiebaSego(str string, model bool) string {
	if goseg == (sego.Segmenter{}) {
		startInit()
	}
	// 分词
	text := []byte(str)
	segments := goseg.Segment(text)

	// 处理分词结果
	// 支持普通模式和搜索模式两种分词，见代码中SegmentsToString函数的注释。
	x := sego.SegmentsToSlice(segments, model)
	str2 := ""
	for _, v := range x {
		str2 += v + " "
	}
	return str2
}

// 字符串
// model 普通模式false, 搜索模式true
func JiebaSegoPlusV(str string, model bool) string {
	if goseg == (sego.Segmenter{}) {
		startInit()
	}
	// 分词
	text := []byte(str)
	segments := goseg.Segment(text)

	// 处理分词结果
	// 支持普通模式和搜索模式两种分词，见代码中SegmentsToString函数的注释。
	x := sego.SegmentsToString(segments, model)
	return x
}

// 正则替换，将指定字符替换成空格
func ReplaceSpance(text, regexp1, replace string) string {
	str2 := ""
	if regexp1 == "" {
		reg2 := regexp.MustCompile(`[,.!;:'"? ，。？！：；‘“”’]+`)
		str2 = reg2.ReplaceAllString(text, replace)
	} else {
		reg2 := regexp.MustCompile(regexp1)
		str2 = reg2.ReplaceAllString(text, replace)
	}
	return str2
}
