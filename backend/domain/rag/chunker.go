package rag

import "strings"

// ChunkText 将文本切分为长度受限且顺序稳定的片段。
func ChunkText(text string, size int) []Chunk {
	if size <= 0 {
		size = 800
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return []Chunk{}
	}
	runes := []rune(text)
	chunks := make([]Chunk, 0, (len(runes)+size-1)/size)
	for start, index := 0, 0; start < len(runes); start, index = start+size, index+1 {
		end := start + size
		if end > len(runes) {
			end = len(runes)
		}
		content := string(runes[start:end])
		chunks = append(chunks, Chunk{Index: index, Content: content, TokenCount: len([]rune(content)), StartOffset: start, EndOffset: end})
	}
	return chunks
}
