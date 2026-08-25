package rag

import (
	"regexp"
	"strings"
)

var heading = regexp.MustCompile(`^(#{1,6}\s+.+|第[一二三四五六七八九十百千万0-9]+[章节部分].*|[一二三四五六七八九十]+、.+|\d+(?:\.\d+)*[、.．]\s*.+)$`)

type section struct{ title, body string }

func ChunkText(text string, size int) []Chunk {
	if size <= 0 {
		size = 800
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return []Chunk{}
	}
	out := []Chunk{}
	for _, s := range sections(text) {
		prev := ""
		for _, b := range split(s.body, size) {
			body := b
			if prev != "" {
				b = prev + "\n" + b
			}
			out = append(out, Chunk{Index: len(out), SectionTitle: s.title, Content: b, TokenCount: len([]rune(b))})
			prev = tail(body)
		}
	}
	return out
}
func sections(text string) []section {
	out := []section{}
	title := ""
	var b strings.Builder
	flush := func() {
		if v := strings.TrimSpace(b.String()); v != "" {
			out = append(out, section{title, v})
		}
		b.Reset()
	}
	for _, raw := range strings.Split(text, "\n") {
		line := strings.TrimSpace(raw)
		if heading.MatchString(line) {
			flush()
			title = strings.TrimSpace(strings.TrimLeft(line, "# "))
		} else {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	flush()
	return out
}
func split(text string, size int) []string {
	out := []string{}
	cur := ""
	for _, s := range sentences(text) {
		if len([]rune(s)) > size {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			runes := []rune(s)
			for len(runes) > size {
				out = append(out, string(runes[:size]))
				runes = runes[size:]
			}
			if len(runes) > 0 {
				out = append(out, string(runes))
			}
			continue
		}
		if len([]rune(cur))+len([]rune(s)) > size && cur != "" {
			out = append(out, cur)
			cur = ""
		}
		cur += s
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}
func sentences(text string) []string {
	r := []rune(strings.TrimSpace(text))
	out := []string{}
	start := 0
	for i, c := range r {
		if strings.ContainsRune("。！？；.!?;", c) {
			out = append(out, string(r[start:i+1]))
			start = i + 1
		}
	}
	if start < len(r) {
		out = append(out, string(r[start:]))
	}
	return out
}

// tail 固定返回上一块最后一个完整句子；没有完整句号时不产生 overlap。
func tail(text string) string {
	s := sentences(text)
	if len(s) == 0 {
		return ""
	}
	last := s[len(s)-1]
	if !strings.ContainsRune("。！？；.!?;", []rune(last)[len([]rune(last))-1]) {
		return ""
	}
	return last
}
