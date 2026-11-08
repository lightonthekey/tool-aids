# tool-aids

Go 工具库：IP 归属地查询、中文分词、圆周率位数查询。

线上示例：
- IPv4 归属地：https://tooldeer.com/ipv4.html
- IPv6 归属地：https://tooldeer.com/ipv6.html
- 中文分词：https://tooldeer.com/fenci.html
- 圆周率查询：https://tooldeer.com/pifinder.html

## 安装

```bash
go get github.com/lightonthekey/tool-aids@latest
# 指定版本
go get github.com/lightonthekey/tool-aids@v1.0.15
```

## 包说明

| 包 | 功能 | 数据 |
| --- | --- | --- |
| `ipqqwry` | 纯真 IP 库（qqwry.dat）IPv4 归属地 | 内嵌 10MB（可关闭） |
| `ipv6` | IP2Location LITE DB3 地址库，支持 IPv6/IPv4 | 外部文件 94MB（不内嵌） |
| `jieba` | 基于 sego + 结巴词典的中文分词 | 内嵌 9MB（可关闭） |
| `piquery` | 圆周率 100 万位查询 | 内嵌 1MB |

最小示例见 [`examples/demo`](examples/demo/main.go)。

### ipqqwry

```go
ipqqwry.FindAddress("8.8.8.8")            // => "美国"
ipqqwry.FindAddressList([]string{"8.8.8.8", "114.22.63.70"}) // 保序、保留重复项
ipqqwry.FindAddressAll([]string{...})     // map 形式（会去重，需要保序请用 FindAddressList）

q := ipqqwry.NewQQwry()                   // 每个请求新建，单个实例不可并发
res, err := q.FindErr("8.8.8.8")          // 需要区分错误类型时用 FindErr
```

数据文件查找顺序：`LoadFile(path)` → `TOOL_AIDS_QQWRY_DAT` 环境变量 → 可执行文件同级 `data/ipqqwry.dat` → 源码目录 `data/` → 内嵌数据。

IP 库每月更新，更新后调用 `ipqqwry.Reload()` 或 `ipqqwry.LoadFile(path)` **即可热切换，无需重启进程**。

### ipv6

```go
info, err := ipv6.FindAddress("2001:4860:4860::8888")
info, err = ipv6.FindAddressToChinese("2001:4860:4860::8888")
info, err = ipv6.FindAddressWithLang("2001:4860:4860::8888", "cn") // 一次搞定，别查两遍
```

地址库 94MB，不内嵌。查找顺序：`Init(path)` → `TOOL_AIDS_IPV6_DB` → 可执行文件同级 `data/` → 源码目录 `data/`。
加载是延迟的，文件缺失只会让查询返回错误，不会让进程退出。

### jieba

```go
words, err := jieba.JiebaWords("我爱北京天安门", false)
s := jieba.JiebaSego("我爱北京天安门", false)         // 空格分隔，末尾带空格
s = jieba.JiebaSegoPlusV("我爱北京天安门", false)     // 附带词性
s = jieba.ReplaceSpance("你好，世界！", "", " ")      // 正则缓存，非法正则不 panic
```

首次调用会加载词典（约 1.3s、常驻约 180MB），加载由 `sync.Once` 保证只发生一次且并发安全。

### piquery

```go
r, err := piquery.Query(2)   // 第 2 个字符 => "1"
r, err = piquery.Query(1)    // => "3"（第 1 位是整数位）
piquery.MaxPosition()        // 1000001
```

`NewPIStore()` 返回共享的只读实例，首次调用校验数据，之后是 O(1)。

## 构建选项

内嵌数据会让二进制增大约 19MB（ipqqwry 10MB + jieba 9MB）。需要更小的二进制时：

```bash
go build -tags tool_aids_noembed ./...
```

此时两个包改为从磁盘查找数据文件，找不到只返回错误、不会崩溃。

## 测试

```bash
go test ./...
go vet ./...
```

## 发版

```bash
git tag v1.0.15
git push origin v1.0.15
```

## 数据来源与许可

- 纯真 IP 库：https://github.com/wisdomfusion/qqwry.dat 、https://www.cz88.net/
- IP2Location LITE：https://lite.ip2location.com/database-download
- 分词：https://github.com/huichen/sego （词典来自结巴分词）

本项目采用 [MIT License](LICENSE)。
