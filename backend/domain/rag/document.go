package rag

// Document 表示已上传知识库文档的领域对象。
type Document struct {
	ID       string
	UserName string
	FileName string
	FileType string
}
