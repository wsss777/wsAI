package rag

// Chunk 表示文档中可索引且有序的内容片段。
type Chunk struct {
	Index        int
	SectionTitle string
	Content      string
	TokenCount   int
	StartOffset  int
	EndOffset    int
}
