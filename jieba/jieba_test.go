package jieba

import (
	"strings"
	"sync"
	"testing"
)

// TestConcurrentFirstLoad 必须放在最前面：验证首次并发加载不会 panic。
// （旧实现用无锁的懒加载，8 个 goroutine 同时首次调用会有数个 goroutine panic）
func TestConcurrentFirstLoad(t *testing.T) {
	const n = 16
	text := "田野上方覆盖着一层烟雾，迷迷蒙蒙的，仿佛一幅大肆渲染的水墨画。"

	var wg sync.WaitGroup
	results := make([]string, n)
	errs := make([]error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("goroutine %d panic: %v", i, r)
				}
			}()
			results[i], errs[i] = JiebaSegoErr(text, true)
		}(i)
	}
	wg.Wait()

	for i := 0; i < n; i++ {
		if errs[i] != nil {
			t.Fatalf("第 %d 个 goroutine 返回错误: %v", i, errs[i])
		}
		if results[i] != results[0] {
			t.Fatalf("并发调用结果不一致:\n%q\n%q", results[0], results[i])
		}
	}
}

func TestJiebaWords(t *testing.T) {
	words, err := JiebaWords("我爱北京天安门", false)
	if err != nil {
		t.Fatalf("分词失败: %v", err)
	}
	if len(words) == 0 {
		t.Fatal("分词结果为空")
	}
	got := strings.Join(words, "/")
	t.Logf("普通模式: %s", got)
	if !strings.Contains(got, "北京") {
		t.Errorf("结果中缺少“北京”: %s", got)
	}

	search, err := JiebaWords("中华人民共和国", true)
	if err != nil {
		t.Fatalf("搜索模式分词失败: %v", err)
	}
	t.Logf("搜索模式: %s", strings.Join(search, "/"))
	if len(search) < 2 {
		t.Errorf("搜索模式结果过少: %v", search)
	}
}

func TestJiebaEmptyContent(t *testing.T) {
	if got := JiebaSego("", false); got != "" {
		t.Errorf("空字符串应返回空串, 实际 %q", got)
	}
	words, err := JiebaWords("", false)
	if err != nil || len(words) != 0 {
		t.Errorf("空字符串应返回空切片, 实际 %v %v", words, err)
	}
}

func TestJiebaSegoPlusV(t *testing.T) {
	got, err := JiebaSegoPlusVErr("我爱北京天安门", false)
	if err != nil {
		t.Fatalf("带词性分词失败: %v", err)
	}
	if !strings.Contains(got, "/") {
		t.Errorf("带词性结果格式异常: %q", got)
	}
}

func TestReplaceSpance(t *testing.T) {
	got := ReplaceSpance("你好，世界！Hello, world.", "", " ")
	if strings.ContainsAny(got, "，！,.") {
		t.Errorf("特殊字符未被替换: %q", got)
	}
	// 自定义 pattern + 缓存复用
	for i := 0; i < 3; i++ {
		if got := ReplaceSpance("a-b-c", "-", "_"); got != "a_b_c" {
			t.Errorf("自定义 pattern 结果 = %q, 期望 a_b_c", got)
		}
	}
	// 非法正则不应 panic
	out, err := ReplaceSpanceErr("abc", "[", " ")
	if err == nil {
		t.Error("非法正则应返回错误")
	}
	if out != "abc" {
		t.Errorf("非法正则应原样返回内容, 实际 %q", out)
	}
}

func TestConcurrentReplace(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = ReplaceSpance("你好，世界！", "", " ")
			}
		}()
	}
	wg.Wait()
}

func BenchmarkJiebaSego(b *testing.B) {
	text := "田野上方覆盖着一层烟雾，迷迷蒙蒙的，仿佛一幅大肆渲染的水墨画。"
	if _, err := JiebaWords(text, true); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = JiebaSego(text, true)
	}
}

func BenchmarkReplaceSpance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ReplaceSpance("你好，世界！Hello, world.", "", " ")
	}
}
