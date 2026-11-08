//go:build tool_aids_noembed

package ipqqwry

// 使用 tool_aids_noembed 构建时禁用内嵌数据，改为从磁盘查找 IP 库。
var embeddedData []byte
