//go:build tool_aids_noembed

package jieba

// 使用 tool_aids_noembed 构建时禁用内嵌词典，改为从磁盘查找。
var embeddedDict []byte
