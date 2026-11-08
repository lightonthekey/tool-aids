package ipqqwry

import (
	"strings"
	"sync"
	"testing"
)

func TestEmbeddedDBLoaded(t *testing.T) {
	if err := LastError(); err != nil {
		t.Fatalf("内嵌 IP 库加载失败: %v", err)
	}
	if n := Count(); n < 400000 {
		t.Fatalf("IP 库记录数异常: %d", n)
	}
	t.Logf("数据来源=%s 记录数=%d", Source(), Count())
}

func TestFindAddress(t *testing.T) {
	for _, ip := range []string{"8.8.8.8", "111.207.139.90", "114.22.63.70", "223.5.5.5"} {
		got := FindAddress(ip)
		if got == "" || got == "未知" {
			t.Errorf("FindAddress(%s) = %q, 期望有效地区", ip, got)
		}
		t.Logf("FindAddress(%s) = %q", ip, got)
	}
}

func TestFindAddressInvalid(t *testing.T) {
	for _, ip := range []string{"not-an-ip", "1.2.3", "999.1.1.1", "2001:db8::1"} {
		res, err := NewQQwryFindErr(ip)
		if err == nil {
			t.Errorf("FindErr(%q) 期望返回错误", ip)
		}
		if res.Country != "未知" || res.Area != "未知" {
			t.Errorf("FindErr(%q) 国家/地区 = %q/%q, 期望 未知/未知", ip, res.Country, res.Area)
		}
	}

	res, _ := NewQQwryFindErr("")
	if res.Country != "未知" {
		t.Errorf("空 IP 应返回 未知, 实际 %q", res.Country)
	}
}

func TestFindAddressListKeepsOrderAndDuplicates(t *testing.T) {
	ips := []string{"8.8.8.8", "", "114.22.63.70", "8.8.8.8", "  "}
	got := FindAddressList(ips)
	if len(got) != 3 {
		t.Fatalf("FindAddressList 返回 %d 条, 期望 3 条（忽略空白项）: %+v", len(got), got)
	}
	if got[0].IP != "8.8.8.8" || got[2].IP != "8.8.8.8" {
		t.Errorf("顺序或重复项未保留: %+v", got)
	}
	if got[1].IP != "114.22.63.70" {
		t.Errorf("第 2 条应为 114.22.63.70, 实际 %s", got[1].IP)
	}
}

func TestFindAddressAllSkipsBlank(t *testing.T) {
	got := FindAddressAll([]string{"8.8.8.8", "", " "})
	if len(got) != 1 {
		t.Fatalf("FindAddressAll 返回 %d 条, 期望 1 条", len(got))
	}
	if _, ok := got["8.8.8.8"]; !ok {
		t.Error("结果中缺少 8.8.8.8")
	}
}

func TestConcurrentFind(t *testing.T) {
	var wg sync.WaitGroup
	ips := []string{"8.8.8.8", "114.22.63.70", "223.5.5.5", "111.207.139.90"}
	for w := 0; w < 16; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			q := NewQQwry()
			for i := 0; i < 200; i++ {
				if got := q.Find(ips[(i+w)%len(ips)]).Country; got == "" {
					t.Errorf("并发查询返回空地区")
					return
				}
			}
		}(w)
	}
	wg.Wait()
}

func TestReloadAndLoadFileError(t *testing.T) {
	before := Count()
	if err := Reload(); err != nil {
		t.Fatalf("Reload 失败: %v", err)
	}
	if after := Count(); after != before {
		t.Errorf("Reload 后记录数变化: %d -> %d", before, after)
	}
	if err := LoadFile("Z:/definitely/not/exists.dat"); err == nil {
		t.Error("LoadFile 不存在的文件应返回错误")
	}
	if Count() != before {
		t.Error("LoadFile 失败后数据被破坏")
	}
	if got := FindAddress("8.8.8.8"); got == "未知" || got == "" {
		t.Error("LoadFile 失败后查询异常")
	}
}

func TestTrimsWhitespace(t *testing.T) {
	// 前端按 \n 切分时，Windows 换行会残留 \r
	if got := FindAddress("8.8.8.8\r"); got == "未知" || got == "" {
		t.Errorf("带 \\r 的 IP 未被正确处理: %q", got)
	}
}

func TestByteToUInt32(t *testing.T) {
	if got := byteToUInt32([]byte{0x01, 0x02, 0x03}); got != 0x030201 {
		t.Errorf("byteToUInt32 = 0x%x, 期望 0x030201", got)
	}
	if got := byteToUInt32([]byte{0x01}); got != 0 {
		t.Errorf("长度不足应返回 0, 实际 %d", got)
	}
}

func TestParseDBRejectsGarbage(t *testing.T) {
	if _, err := parseDB([]byte{}, "x"); err == nil {
		t.Error("空数据应返回错误")
	}
	if _, err := parseDB(make([]byte, 20), "x"); err == nil || !strings.Contains(err.Error(), "损坏") {
		t.Errorf("全零数据应被判为损坏, 实际 %v", err)
	}
}

// NewQQwryFindErr 测试辅助：避免对不可寻址的临时值取地址
func NewQQwryFindErr(ip string) (ResultQQwry, error) {
	q := NewQQwry()
	return q.FindErr(ip)
}
