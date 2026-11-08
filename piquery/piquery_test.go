package piquery

import (
	"strings"
	"sync"
	"testing"
)

func TestMaxPosition(t *testing.T) {
	if got := MaxPosition(); got < 1000000 {
		t.Errorf("最大位置 = %d, 期望至少 1000000", got)
	}
}

func TestQueryFirstPosition(t *testing.T) {
	r, err := Query(1)
	if err != nil {
		t.Fatalf("Query(1) 失败: %v", err)
	}
	if r.Current != "3" {
		t.Errorf("第 1 位 = %q, 期望 3（整数位）", r.Current)
	}
	if r.Previous != "" {
		t.Errorf("第 1 位的前 5 位应为空, 实际 %q", r.Previous)
	}
	if len(r.Next) != 5 {
		t.Errorf("第 1 位的后 5 位长度 = %d, 期望 5", len(r.Next))
	}
}

func TestQueryKnownDigits(t *testing.T) {
	// π = 3.1415926535...
	want := map[int]string{2: "1", 3: "4", 4: "1", 5: "5", 6: "9", 7: "2"}
	for pos, digit := range want {
		r, err := Query(pos)
		if err != nil {
			t.Fatalf("Query(%d) 失败: %v", pos, err)
		}
		if r.Current != digit {
			t.Errorf("第 %d 位 = %q, 期望 %q", pos, r.Current, digit)
		}
	}
}

func TestQueryContext(t *testing.T) {
	// π = 3.1415926535 8979...，第 10 位是 index 9 的 "3"
	r, err := Query(10)
	if err != nil {
		t.Fatalf("Query(10) 失败: %v", err)
	}
	if r.Current != "3" {
		t.Errorf("第 10 位 = %q, 期望 3", r.Current)
	}
	if r.Previous != "59265" {
		t.Errorf("第 10 位前 5 位 = %q, 期望 59265", r.Previous)
	}
	if r.Next != "58979" {
		t.Errorf("第 10 位后 5 位 = %q, 期望 58979", r.Next)
	}
}

func TestQueryLastPosition(t *testing.T) {
	max := MaxPosition()
	r, err := Query(max)
	if err != nil {
		t.Fatalf("Query(%d) 失败: %v", max, err)
	}
	if r.Next != "" {
		t.Errorf("最后一位的后 5 位应为空, 实际 %q", r.Next)
	}
}

func TestQueryInvalidPosition(t *testing.T) {
	for _, pos := range []int{0, -1, MaxPosition() + 1} {
		if _, err := Query(pos); err == nil {
			t.Errorf("Query(%d) 期望返回错误", pos)
		}
	}
}

func TestNewPIStoreCheap(t *testing.T) {
	s, err := NewPIStore()
	if err != nil {
		t.Fatalf("NewPIStore 失败: %v", err)
	}
	if s.MaxPosition() != MaxPosition() {
		t.Error("NewPIStore 与包级 MaxPosition 不一致")
	}
	// 校验只做一次：第二次调用不应再扫一遍数据（这里只验证可用性）
	if _, err := NewPIStore(); err != nil {
		t.Fatalf("第二次 NewPIStore 失败: %v", err)
	}
}

func TestConcurrentQuery(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 500; j++ {
				r, err := Query((i*7+j)%1000 + 1)
				if err != nil || len(r.Current) != 1 {
					t.Errorf("并发查询异常: %+v %v", r, err)
					return
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestInvalidPositionErrorMentionsRange(t *testing.T) {
	_, err := Query(0)
	if err == nil || !strings.Contains(err.Error(), "1") {
		t.Errorf("错误信息应说明范围, 实际 %v", err)
	}
}

func BenchmarkQuery(b *testing.B) {
	if _, err := NewPIStore(); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		_, _ = Query(i%1000 + 1)
	}
}

func BenchmarkNewPIStore(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewPIStore()
	}
}
