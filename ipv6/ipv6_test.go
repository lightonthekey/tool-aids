package ipv6

import (
	"strings"
	"testing"
)

func TestFindAddressIPv6(t *testing.T) {
	cases := []struct{ ip, wantShort string }{
		{"2001:4860:4860::8888", "US"},
		{"2400:3200::1", "CN"},
	}
	for _, c := range cases {
		info, err := FindAddress(c.ip)
		if err != nil {
			t.Fatalf("FindAddress(%s) 失败: %v（地址库路径: %s）", c.ip, err, DBPath())
		}
		if info.CountryShort != c.wantShort {
			t.Errorf("FindAddress(%s).CountryShort = %q, 期望 %q", c.ip, info.CountryShort, c.wantShort)
		}
		if info.CountryLong == "" {
			t.Errorf("FindAddress(%s).CountryLong 为空", c.ip)
		}
	}
}

func TestFindAddressIPv4(t *testing.T) {
	info, err := FindAddress("8.8.8.8")
	if err != nil {
		t.Fatalf("查询 IPv4 失败: %v", err)
	}
	if info.CountryShort != "US" {
		t.Errorf("8.8.8.8 CountryShort = %q, 期望 US", info.CountryShort)
	}
}

func TestFindAddressInvalidIP(t *testing.T) {
	for _, ip := range []string{"", "not-an-ip", "999.999.999.999"} {
		if _, err := FindAddress(ip); err == nil {
			t.Errorf("FindAddress(%q) 期望返回错误", ip)
		}
	}
}

func TestFindAddressWithLang(t *testing.T) {
	info, err := FindAddressWithLang("2400:3200::1", "cn")
	if err != nil {
		t.Fatalf("中文查询失败: %v", err)
	}
	if !strings.Contains(info.CountryLong, "中国") {
		t.Errorf("中文国家名 = %q, 期望包含“中国”", info.CountryLong)
	}

	en, err := FindAddressWithLang("2400:3200::1", "en")
	if err != nil {
		t.Fatalf("英文查询失败: %v", err)
	}
	if en.CountryLong == info.CountryLong {
		t.Errorf("中英文结果相同: %q", en.CountryLong)
	}
}

func TestInitWithBadPath(t *testing.T) {
	defer func() { _ = Init("") }()
	if err := Init("Z:/definitely/not/exists.bin"); err == nil {
		t.Error("Init 不存在的文件应返回错误")
	}
	// 失败不应影响后续查询（_ = Init("") 之后会恢复自动查找）
}

func TestMissingDBReturnsError(t *testing.T) {
	orig := override
	defer func() {
		override = orig
		Close()
	}()
	Close()
	override = "Z:/definitely/not/exists.bin"
	if _, err := FindAddress("8.8.8.8"); err == nil {
		t.Error("地址库缺失时应返回错误，而不是 panic 或退出进程")
	}
}
