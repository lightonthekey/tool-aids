// jieba中文分词
// 参考 https://github.com/huichen/sego
// package jieba 使用 sego 包实现了jieba中文分词
package jieba

import (
	"log"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/huichen/sego"
)

// 存放分词器
var segmenter sego.Segmenter

// 字符串分词的函数，将被调用的频率非常高
// 那么我们只在第一次调用这个函数的时候才加载字典文件，之后的调用将直接使用已加载的字典
func init() {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not get current file information（jieba.go）")
	}

	// 获取当前文件所在目录
	currentDir := filepath.Dir(currentFile)
	// 构建文件路径
	dictPath := filepath.Join(currentDir, "data", "jieba_dictionary.txt")
	segmenter.LoadDictionary(dictPath)
}

// JiebaSegment 用给定的模式对参数字符串进行分词并返回分词结果
// model 要使用的分词模式，普通模式=false, 搜索模式=true
func JiebaSego(content string, model bool) string {
	segments := segmenter.Segment([]byte(content))

	wordSlice := sego.SegmentsToSlice(segments, model)
	result := ""
	for _, word := range wordSlice {
		result += word + " "
	}
	return result
}

// JiebaSegmentPlus 对参数字符串进行分词并以字符串的形式返回分词结果
// model 要使用的分词模式，普通模式=false, 搜索模式=true
// 增加词性
func JiebaSegoPlusV(content string, model bool) string {
	segments := segmenter.Segment([]byte(content))
	return sego.SegmentsToString(segments, model)
}

// ReplaceSpecialChars 使用正则表达式将参数字符串中的特殊字符替换为空格并返回。
func ReplaceSpance(content, pattern, replace string) string {
	if pattern == "" {
		pattern = `[,.!;:'"? ，。？！：；‘“”’]+`
	}
	regexp := regexp.MustCompile(pattern)
	return regexp.ReplaceAllString(content, replace)
}
