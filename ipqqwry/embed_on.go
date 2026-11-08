//go:build !tool_aids_noembed

package ipqqwry

import _ "embed"

// 默认把纯真 IP 库内嵌进二进制，部署时只需一个文件。
// 如需减小体积，可用 `go build -tags tool_aids_noembed` 关闭，
// 届时会按 环境变量 TOOL_AIDS_QQWRY_DAT / 可执行文件同级 data 目录 查找。
//
//go:embed data/ipqqwry.dat
var embeddedData []byte
