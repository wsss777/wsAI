package session

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateRunesKeepsUTF8Valid(t *testing.T) {
	input := strings.Repeat("你好", 50)

	got := truncateRunes(input, 80)

	if !utf8.ValidString(got) {
		t.Fatalf("截断后的标题不是有效 UTF-8：%q", got)
	}
	if count := len([]rune(got)); count != 80 {
		t.Fatalf("标题字符数 = %d，期望 80", count)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("截断后的标题未保留省略号：%q", got)
	}
}

func TestTruncateRunesDoesNotUseByteLength(t *testing.T) {
	input := "你好啊,你可以给我讲一下go八股值得注意的点吗?尤其是一些边界情况"
	if len(input) <= 80 || len([]rune(input)) >= 80 {
		t.Fatalf("测试数据未覆盖字节数超限、字符数未超限的场景")
	}

	got := truncateRunes(input, 80)

	if got != input {
		t.Fatalf("标题被错误截断：got %q, want %q", got, input)
	}
}
