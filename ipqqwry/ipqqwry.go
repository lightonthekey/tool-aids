// Package ipqqwry 基于纯真 IP 库（qqwry.dat）提供 IPv4 归属地查询。
//
// # 数据来源
//
//	IP 库每月更新一次，下载地址：https://github.com/wisdomfusion/qqwry.dat
//	或从 https://www.cz88.net/ 安装目录中取 qqwry.dat。
//	算法参考：https://github.com/freshcn/qqwry
//
// # 数据文件
//
//	默认使用编译期内嵌的 data/ipqqwry.dat，部署时只需要一个二进制文件
//	（可用 `go build -tags tool_aids_noembed` 关闭内嵌以减小体积）。
//	外部文件优先级从高到低：LoadFile(path) / 环境变量 TOOL_AIDS_QQWRY_DAT /
//	可执行文件同级 data 目录 / 源码目录 data 目录 / 内嵌数据。
//	外部文件更新后调用 Reload() 即可热切换，无需重启进程。
package ipqqwry

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
)

const (
	// IndexLen 单条索引记录的字节长度
	IndexLen = 7
	// RedirectMode1 国家的类型, 指向另一个指向
	RedirectMode1 = 0x01
	// RedirectMode2 国家的类型, 指向一个指向
	RedirectMode2 = 0x02

	// maxStringLen 单条归属地字符串的最大长度，防止脏数据导致越界扫描
	maxStringLen = 1024
	// maxSearchSteps 索引二分查找的最大步数（52 万条记录最多 20 步）
	maxSearchSteps = 64

	envDataPath = "TOOL_AIDS_QQWRY_DAT"
	dbFileName  = "ipqqwry.dat"
)

// ResultQQwry 归属地信息
type ResultQQwry struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Area    string `json:"area"`
}

// dbSnapshot 是一次加载得到的只读快照。
// 用原子指针替换整个快照，查询侧无需加锁，Reload 也不会踩到正在读取的数据。
type dbSnapshot struct {
	data   []byte
	ipNum  int64
	source string
}

var (
	current atomic.Pointer[dbSnapshot]
	loadMu  sync.Mutex
	loadErr error // 仅供诊断，不阻塞使用
)

func init() {
	// 1) 外部文件优先：环境变量 → 可执行文件同级 data/ → 源码目录 data/
	//    这样运维只要替换磁盘上的 qqwry.dat（再调 Reload）就能更新数据，无需重新编译。
	if p, err := resolveExternal(); err == nil {
		if d, err := loadFile(p); err != nil {
			// 外部文件损坏时不要直接失败，回退内嵌数据，但要留下明显日志
			loadErr = err
			log.Printf("ipqqwry: 外部 IP 库 %s 加载失败: %v（回退内嵌数据）", p, err)
		} else {
			current.Store(d)
			loadErr = nil
			return
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		// 环境变量写了但文件不存在之类的配置问题，需要让运维看见
		log.Printf("ipqqwry: %v", err)
	}

	// 2) 内嵌数据兜底（默认构建即有）
	if len(embeddedData) > 0 {
		d, err := parseDB(embeddedData, "embedded")
		if err != nil {
			loadErr = err
			log.Printf("ipqqwry: 内嵌 IP 库不可用: %v", err)
			return
		}
		current.Store(d)
		loadErr = nil
		return
	}

	// 3) 既没有外部文件也没有内嵌数据（-tags tool_aids_noembed 构建且文件缺失）
	loadErr = fmt.Errorf("找不到 IP 库 %s，请通过 LoadFile 或环境变量 %s 指定", dbFileName, envDataPath)
	log.Printf("ipqqwry: %v", loadErr)
}

// resolveExternal 按优先级查找外部 IP 库文件：
// 环境变量 TOOL_AIDS_QQWRY_DAT → 可执行文件同级 data/ → 源码目录 data/。
// 都没找到时返回 os.ErrNotExist，调用方据此回退内嵌数据。
func resolveExternal() (string, error) {
	if p := strings.TrimSpace(os.Getenv(envDataPath)); p != "" {
		if fileExists(p) {
			return p, nil
		}
		return "", fmt.Errorf("环境变量 %s 指定的文件不存在: %s", envDataPath, p)
	}
	if exe, err := os.Executable(); err == nil {
		if p := filepath.Join(filepath.Dir(exe), "data", dbFileName); fileExists(p) {
			return p, nil
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		if p := filepath.Join(filepath.Dir(file), "data", dbFileName); fileExists(p) {
			return p, nil
		}
	}
	return "", os.ErrNotExist
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// parseDB 校验并包装一份 IP 库数据。
func parseDB(data []byte, source string) (*dbSnapshot, error) {
	if len(data) < 2*IndexLen {
		return nil, fmt.Errorf("IP 库数据过短（%d 字节），不是有效的 qqwry.dat", len(data))
	}
	start := binary.LittleEndian.Uint32(data[:4])
	end := binary.LittleEndian.Uint32(data[4:])
	if start < 8 || end <= start || int64(end)+IndexLen > int64(len(data)) {
		return nil, errors.New("IP 库数据已损坏或格式不正确")
	}
	return &dbSnapshot{
		data:   data,
		ipNum:  int64((end-start)/IndexLen + 1),
		source: source,
	}, nil
}

func loadFile(path string) (*dbSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseDB(data, path)
}

// LoadFile 使用指定的 qqwry.dat 替换当前 IP 库（热更新，无需重启进程）。
// 加载失败时保留原有数据不变。
func LoadFile(path string) error {
	d, err := loadFile(path)
	if err != nil {
		return fmt.Errorf("加载 IP 库 %s 失败: %w", path, err)
	}
	loadMu.Lock()
	current.Store(d)
	loadErr = nil
	loadMu.Unlock()
	return nil
}

// Reload 重新加载 IP 库：先按优先级查找外部文件（替换 qqwry.dat 后热更新即可，
// 无需重启进程），没有外部文件时重新解析内嵌数据。
func Reload() error {
	if p, err := resolveExternal(); err == nil {
		if err := LoadFile(p); err == nil {
			return nil
		}
	}
	if len(embeddedData) > 0 {
		d, err := parseDB(embeddedData, "embedded")
		if err != nil {
			return err
		}
		loadMu.Lock()
		current.Store(d)
		loadErr = nil
		loadMu.Unlock()
		return nil
	}
	cur := current.Load()
	if cur == nil {
		return errors.New("IP 库尚未加载，请先调用 LoadFile")
	}
	return LoadFile(cur.source)
}

// Count 返回当前 IP 库的记录条数。
func Count() int64 {
	if cur := current.Load(); cur != nil {
		return cur.ipNum
	}
	return 0
}

// Source 返回当前数据来源（内嵌为 "embedded"，否则为文件路径）。
func Source() string {
	if cur := current.Load(); cur != nil {
		return cur.source
	}
	return ""
}

// LastError 返回最近一次加载失败的原因，没有失败则为 nil。
func LastError() error { return loadErr }

// QQwry 纯真 IP 库查询器。
// 每个查询器持有各自的游标，同一实例不可并发使用；请为每个请求新建（NewQQwry 很廉价）。
type QQwry struct {
	Offset int64
	snap   *dbSnapshot
}

// NewQQwry 创建一个查询器，指向当前已加载的 IP 库快照。
func NewQQwry() QQwry {
	return QQwry{snap: current.Load()}
}

// SetOffset 设置读取游标
func (q *QQwry) SetOffset(offset int64) {
	q.Offset = offset
}

// FindAddressAll 批量查询，返回 原IP -> 结果 的映射。
// 注意：map 会丢失顺序，同一 IP 重复出现时只会保留一条；需要保序请用 FindAddressList。
func FindAddressAll(ips []string) map[string]ResultQQwry {
	rs := make(map[string]ResultQQwry, len(ips))
	obj := NewQQwry()
	for _, v := range ips {
		if strings.TrimSpace(v) == "" {
			continue
		}
		rs[v] = obj.Find(v)
	}
	return rs
}

// FindAddressList 批量查询，按入参顺序返回结果，保留重复项，忽略空白项。
func FindAddressList(ips []string) []ResultQQwry {
	rs := make([]ResultQQwry, 0, len(ips))
	obj := NewQQwry()
	for _, v := range ips {
		if strings.TrimSpace(v) == "" {
			continue
		}
		rs = append(rs, obj.Find(v))
	}
	return rs
}

// FindAddress 返回 IP 所在国家或地区（查询失败返回 "未知"）。
func FindAddress(ip string) string {
	q := NewQQwry()
	return q.Find(ip).Country
}

// Find 查询 IP 对应的归属地信息。查询失败时 Country/Area 为 "未知"。
func (q *QQwry) Find(ip string) (res ResultQQwry) {
	res, _ = q.FindErr(ip)
	return res
}

// FindErr 与 Find 相同，但会返回具体错误，便于调用方区分
// “IP 格式不对”“只支持 IPv4”“IP 库未加载” 等情况。
func (q *QQwry) FindErr(ip string) (ResultQQwry, error) {
	ip = strings.TrimSpace(ip)
	res := ResultQQwry{IP: ip}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		res.Country, res.Area = "未知", "未知"
		return res, fmt.Errorf("无法识别的 IP 地址: %q", ip)
	}
	v4 := parsed.To4()
	if v4 == nil {
		res.Country, res.Area = "未知", "未知"
		return res, fmt.Errorf("%s 不是 IPv4 地址（纯真库只支持 IPv4）", ip)
	}
	if q.snap == nil {
		res.Country, res.Area = "未知", "未知"
		return res, errors.New("IP 库未加载")
	}

	offset := q.searchIndex(binary.BigEndian.Uint32(v4))
	if offset <= 0 {
		res.Country, res.Area = "未知", "未知"
		return res, nil
	}

	country, area := q.readLocation(offset)
	dec := simplifiedchinese.GBK.NewDecoder()
	res.Country = decode(dec, country)
	areas := decode(dec, area)
	if areas == "CZ88.NET" || areas == " CZ88.NET" || areas == "" {
		res.Area = "未知"
	} else {
		res.Area = areas
	}
	if res.Country == "" {
		res.Country = "未知"
	}
	return res, nil
}

func decode(dec *encoding.Decoder, b []byte) string {
	if len(b) == 0 {
		return ""
	}
	// GBK 解码遇到非法字节时返回已解码的部分，直接采用
	s, _ := dec.Bytes(b)
	return string(s)
}

// readLocation 解析一条索引记录指向的归属地信息。
func (q *QQwry) readLocation(offset uint32) (country, area []byte) {
	mode := q.readMode(offset + 4)
	switch mode {
	case RedirectMode1:
		countryOffset := q.readUInt24()
		mode = q.readMode(countryOffset)
		if mode == RedirectMode2 {
			c := q.readUInt24()
			country = q.readString(c)
			countryOffset += 4
		} else {
			country = q.readString(countryOffset)
			countryOffset += uint32(len(country) + 1)
		}
		area = q.readArea(countryOffset)
	case RedirectMode2:
		countryOffset := q.readUInt24()
		country = q.readString(countryOffset)
		area = q.readArea(offset + 8)
	default:
		country = q.readString(offset + 4)
		area = q.readArea(offset + uint32(5+len(country)))
	}
	return country, area
}

// readData 从当前游标读取 num 个字节，返回切片共享底层数据（只读）。
// 越界时返回 nil，绝不 panic。
func (q *QQwry) readData(num int, offset ...int64) []byte {
	if len(offset) > 0 {
		q.Offset = offset[0]
	}
	if q.snap == nil {
		return nil
	}
	data := q.snap.data
	if q.Offset < 0 || q.Offset > int64(len(data)) {
		return nil
	}
	end := q.Offset + int64(num)
	if end > int64(len(data)) {
		end = int64(len(data))
	}
	rs := data[q.Offset:end]
	q.Offset = end
	return rs
}

// readMode 获取偏移值类型
func (q *QQwry) readMode(offset uint32) byte {
	mode := q.readData(1, int64(offset))
	if len(mode) == 0 {
		return 0
	}
	return mode[0]
}

// readArea 读取区域
func (q *QQwry) readArea(offset uint32) []byte {
	mode := q.readMode(offset)
	if mode == RedirectMode1 || mode == RedirectMode2 {
		areaOffset := q.readUInt24()
		if areaOffset == 0 {
			return nil
		}
		return q.readString(areaOffset)
	}
	return q.readString(offset)
}

// readString 从 offset 开始读取以 \0 结尾的字符串（不含 \0）。
func (q *QQwry) readString(offset uint32) []byte {
	if q.snap == nil {
		return nil
	}
	data := q.snap.data
	start := int64(offset)
	if start < 0 || start >= int64(len(data)) {
		return nil
	}
	limit := start + maxStringLen
	if limit > int64(len(data)) {
		limit = int64(len(data))
	}
	for i := start; i < limit; i++ {
		if data[i] == 0 {
			return data[start:i]
		}
	}
	return data[start:limit]
}

// searchIndex 查找 IP 对应的索引记录偏移。
func (q *QQwry) searchIndex(ip uint32) uint32 {
	header := q.readData(8, 0)
	if len(header) < 8 {
		return 0
	}
	start := binary.LittleEndian.Uint32(header[:4])
	end := binary.LittleEndian.Uint32(header[4:])
	if q.snap == nil || end < start || int64(end) >= int64(len(q.snap.data)) {
		return 0
	}

	for i := 0; i < maxSearchSteps; i++ {
		mid := getMiddleOffset(start, end)
		buf := q.readData(IndexLen, int64(mid))
		if len(buf) < IndexLen {
			return 0
		}
		cur := binary.LittleEndian.Uint32(buf[:4])

		if end-start == IndexLen {
			offset := byteToUInt32(buf[4:])
			next := q.readData(IndexLen) // 紧跟当前记录的下一条
			if len(next) < IndexLen {
				return offset
			}
			if ip < binary.LittleEndian.Uint32(next[:4]) {
				return offset
			}
			return 0
		}

		switch {
		case cur > ip:
			end = mid
		case cur < ip:
			start = mid
		default:
			return byteToUInt32(buf[4:])
		}
	}
	return 0
}

// readUInt24 读取 3 字节小端整数
func (q *QQwry) readUInt24() uint32 {
	buf := q.readData(3)
	return byteToUInt32(buf)
}

// getMiddleOffset 取区间中点（按索引记录对齐）
func getMiddleOffset(start, end uint32) uint32 {
	records := ((end - start) / IndexLen) >> 1
	return start + records*IndexLen
}

// byteToUInt32 将 3 字节小端序列转换为 uint32
func byteToUInt32(data []byte) uint32 {
	if len(data) < 3 {
		return 0
	}
	i := uint32(data[0]) & 0xff
	i |= (uint32(data[1]) << 8) & 0xff00
	i |= (uint32(data[2]) << 16) & 0xff0000
	return i
}
