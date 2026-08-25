package rag

import (
	"context"
	"strings"
	"wsai/backend/config"
	repo "wsai/backend/infra/mysql/repository"
	infrarag "wsai/backend/infra/rag"
)

// Citation 标识用于支撑检索回答的文档片段。
// 启用对应集成后，检索适配器由 infra/rag 提供。
type Citation struct {
	DocumentID string  `json:"document_id"`
	FileName   string  `json:"file_name"`
	ChunkID    string  `json:"chunk_id"`
	Score      float64 `json:"score"`
}

func RetrieveContext(ctx context.Context, user, sessionID, question string) (string, []Citation, error) {
	bindings, err := repo.ListSessionDocumentsBySessionID(sessionID)
	if err != nil {
		return "", nil, err
	}
	ids := make([]string, 0, len(bindings))
	for _, b := range bindings {
		ids = append(ids, b.DocumentID)
	}
	if len(ids) == 0 {
		return "", []Citation{}, nil
	}
	vectors, err := infrarag.Embed(ctx, []string{question})
	if err != nil {
		return "", nil, err
	}
	hits, err := infrarag.Search(ctx, vectors[0], user, ids, config.C.RAGConfig.TopK)
	if err != nil {
		return "", nil, err
	}
	chunkIDs := make([]string, len(hits))
	for i, h := range hits {
		chunkIDs[i] = h.ChunkID
	}
	chunks, err := repo.GetChunksByChunkIDs(chunkIDs)
	if err != nil {
		return "", nil, err
	}
	byID := map[string]string{}
	for _, c := range chunks {
		byID[c.ChunkID] = c.Content
	}
	var b strings.Builder
	citations := make([]Citation, 0, len(hits))
	for i, h := range hits {
		if text := byID[h.ChunkID]; text != "" {
			b.WriteString("【资料")
			b.WriteString(string(rune('1' + i)))
			b.WriteString("】\n")
			b.WriteString(text)
			b.WriteString("\n\n")
			citations = append(citations, Citation{DocumentID: h.DocumentID, FileName: h.FileName, ChunkID: h.ChunkID, Score: h.Score})
		}
	}
	return b.String(), citations, nil
}
