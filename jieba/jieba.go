// Package jieba 使用 sego 实现中文分词（结巴分词词典）。
//
// 参考：https://github.com/huichen/sego
//
// # 词典加载
//
//	首次使用时加载 data/jieba_dictionary.txt（9MB，加载约 1.3s，常驻内存约 180MB）。
//	加载只发生一次，且并发安全；加载失败不会让宿主进程退出，只会让相关函数返回错误。
//	词典文件查找顺序（优先级从高到低）：
//	  TOOL_AIDS_JIEBA_DICT 环境变量 / 可执行文件同级 data 目录 / 源码目录 / 内嵌副本写入临时目录
package jieba

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/huichen/sego"
)

// defaultPattern ReplaceSpance 未指定 pattern 时使用的默认特殊字符集合
const defaultPattern = `[,.!;:'"? ，。？！：；‘“”’]+`

var (
	segOnce   sync.Once
	segmenter *sego.Segmenter
	segErr    error

	// 已编译正则缓存，避免每次调用都重新编译
	regexCache sync.Map // pattern -> compiledRegex
)

// load 保证分词器只加载一次。
//
// 注意：sego.LoadDictionary 在打开文件失败时会调用 log.Fatalf（os.Exit，无法 recover），
// 所以这里先自行打开一次确认可读，再交给 sego；词典内容异常导致的 panic 则由 recover 兜住。
func load() (*sego.Segmenter, error) {
	segOnce.Do(func() {
		path, err := resolveDict()
		if err != nil {
			segErr = err
			return
		}
		if err := checkReadable(path); err != nil {
			segErr = err
			return
		}
		seg := &sego.Segmenter{}
		func() {
			defer func() {
				if r := recover(); r != nil {
					segErr = fmt.Errorf("分词词典加载失败: %v", r)
				}
			}()
			// LoadDictionary 会向标准日志输出两行加载信息，属于预期行为
			seg.LoadDictionary(path)
		}()
		if segErr == nil {
			segmenter = seg
		}
	})
	return segmenter, segErr
}

func checkReadable(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("分词词典无法打开 %s: %w", path, err)
	}
	return f.Close()
}

// resolveDict 依次尝试各种可能位置，最后回退到内嵌词典。
func resolveDict() (string, error) {
	if p := strings.TrimSpace(os.Getenv("TOOL_AIDS_JIEBA_DICT")); p != "" {
		if fileExists(p) {
			return p, nil
		}
		// 环境变量写错时继续尝试其他位置，避免整个分词功能不可用
		log.Printf("jieba: 环境变量 TOOL_AIDS_JIEBA_DICT=%s 不存在，继续尝试其他位置", p)
	}

	// 可执行文件同级 data 目录（部署时把 data 目录和程序放在一起）
	if exe, err := os.Executable(); err == nil {
		if p := filepath.Join(filepath.Dir(exe), "data", "jieba_dictionary.txt"); fileExists(p) {
			return p, nil
		}
		// 兼容 go build 后在包目录内运行的场景
		if p := filepath.Join(filepath.Dir(exe), "jieba", "data", "jieba_dictionary.txt"); fileExists(p) {
			return p, nil
		}
	}

	// 源码目录（go test / go run）
	if _, file, _, ok := runtime.Caller(0); ok {
		if p := filepath.Join(filepath.Dir(file), "data", "jieba_dictionary.txt"); fileExists(p) {
			return p, nil
		}
	}

	// 兜底：sego 只支持从文件加载，把内嵌词典写到临时目录
	if len(embeddedDict) == 0 {
		return "", fmt.Errorf("找不到分词词典，请通过环境变量 TOOL_AIDS_JIEBA_DICT 指定 jieba_dictionary.txt 的路径")
	}
	return writeTempDict()
}

func writeTempDict() (string, error) {
	dir := filepath.Join(os.TempDir(), "tool-aids")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}
	path := filepath.Join(dir, "jieba_dictionary.txt")
	if st, err := os.Stat(path); err == nil && st.Size() == int64(len(embeddedDict)) {
		return path, nil
	}
	if err := os.WriteFile(path, embeddedDict, 0o644); err != nil {
		return "", fmt.Errorf("写入临时词典失败: %w", err)
	}
	return path, nil
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// LoadStatus 返回词典是否已加载以及加载错误，便于服务启动时自检。
func LoadStatus() error {
	_, err := load()
	return err
}

// JiebaWords 分词并返回词列表。
// model 为 true 时使用搜索模式（粒度更细），false 为普通模式。
func JiebaWords(content string, model bool) ([]string, error) {
	seg, err := load()
	if err != nil {
		return nil, err
	}
	if content == "" {
		return []string{}, nil
	}
	return sego.SegmentsToSlice(seg.Segment([]byte(content)), model), nil
}

// JiebaSego 分词后以空格分隔返回（末尾带一个空格，保持历史行为）。
// 加载失败时返回空字符串；需要区分错误请用 JiebaWords 或 JiebaSegoErr。
func JiebaSego(content string, model bool) string {
	s, _ := JiebaSegoErr(content, model)
	return s
}

// JiebaSegoErr 与 JiebaSego 相同，但会返回加载错误。
func JiebaSegoErr(content string, model bool) (string, error) {
	words, err := JiebaWords(content, model)
	if err != nil {
		return "", err
	}
	return joinWithSpace(words), nil
}

// JiebaSegoPlusV 分词并在结果中附带词性。
func JiebaSegoPlusV(content string, model bool) string {
	s, _ := JiebaSegoPlusVErr(content, model)
	return s
}

// JiebaSegoPlusVErr 与 JiebaSegoPlusV 相同，但会返回加载错误。
func JiebaSegoPlusVErr(content string, model bool) (string, error) {
	seg, err := load()
	if err != nil {
		return "", err
	}
	if content == "" {
		return "", nil
	}
	return sego.SegmentsToString(seg.Segment([]byte(content)), model), nil
}

// joinWithSpace 用 strings.Builder 拼接，避免 += 造成的 O(n²) 拷贝。
func joinWithSpace(words []string) string {
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	n := 0
	for _, w := range words {
		n += len(w) + 1
	}
	b.Grow(n)
	for _, w := range words {
		b.WriteString(w)
		b.WriteByte(' ')
	}
	return b.String()
}

// ReplaceSpance 用正则替换文本中的特殊字符。
// pattern 为空时使用默认标点集合。pattern 非法时原样返回内容（不 panic），并可通过
// ReplaceSpanceErr 获取错误。
func ReplaceSpance(content, pattern, replace string) string {
	s, _ := ReplaceSpanceErr(content, pattern, replace)
	return s
}

// ReplaceSpanceErr 与 ReplaceSpance 相同，但会返回正则编译错误。
func ReplaceSpanceErr(content, pattern, replace string) (string, error) {
	re, err := compileCached(pattern)
	if err != nil {
		return content, err
	}
	return re.ReplaceAllString(content, replace), nil
}

// compileCached 缓存已编译的正则；pattern 为空时使用默认值。
func compileCached(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		pattern = defaultPattern
	}
	if v, ok := regexCache.Load(pattern); ok {
		if c, ok := v.(compiledRegex); ok {
			return c.re, c.err
		}
	}
	re, err := regexp.Compile(pattern)
	regexCache.Store(pattern, compiledRegex{re: re, err: err})
	return re, err
}

type compiledRegex struct {
	re  *regexp.Regexp
	err error
}
