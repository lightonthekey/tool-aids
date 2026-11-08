//go:build !tool_aids_noembed

package jieba

import _ "embed"

// 默认把分词词典内嵌进二进制，部署时无需额外 data 目录。
// 如需减小体积，可用 `go build -tags tool_aids_noembed` 关闭，
// 届时会按 TOOL_AIDS_JIEBA_DICT / 可执行文件同级 data 目录 / 源码目录 查找。
//
//go:embed data/jieba_dictionary.txt
var embeddedDict []byte
