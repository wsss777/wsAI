package rag

import (
	"fmt"
	"github.com/ledongthuc/pdf"
	"os"
	"strings"
)

// ParseTextFile 读取支持的 RAG 文本来源。PDF 解析属于基础设施职责，
// 因此不与当前领域包耦合。
func ParseTextFile(path, fileType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(fileType)) {
	case "txt", "md", "markdown":
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(content)), nil
	case "pdf":
		file, reader, err := pdf.Open(path)
		if err != nil {
			return "", err
		}
		defer file.Close()
		var b strings.Builder
		for page := 1; page <= reader.NumPage(); page++ {
			p := reader.Page(page)
			if p.V.IsNull() {
				continue
			}
			text, err := p.GetPlainText(nil)
			if err != nil {
				return "", err
			}
			b.WriteString(text)
			b.WriteByte('\n')
		}
		return strings.TrimSpace(b.String()), nil
	default:
		return "", fmt.Errorf("unsupported document type: %s", fileType)
	}
}
