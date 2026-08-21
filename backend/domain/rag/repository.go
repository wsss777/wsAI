package rag

import "context"

// Repository 定义 RAG 用例所需的文档持久化边界。
type Repository interface {
	FindDocument(ctx context.Context, documentID string) (*Document, error)
}
