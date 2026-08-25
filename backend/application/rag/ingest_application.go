package rag

import (
	"context"
	"fmt"
	"log"
	domain "wsai/backend/domain/rag"
	repo "wsai/backend/infra/mysql/repository"
	infrarag "wsai/backend/infra/rag"
	"wsai/backend/model"
	"wsai/backend/utils"
)

func IngestDocument(doc *model.Document) {
	ctx := context.Background()
	fail := func(err error) {
		log.Printf("document ingestion failed: document_id=%s file=%s error=%v", doc.DocumentID, doc.FileName, err)
		_ = repo.UpdateIngestResult(doc.DocumentID, model.DocumentParseStatusFailed, 0, "", err.Error())
	}
	_ = repo.UpdateIngestResult(doc.DocumentID, model.DocumentParseStatusProcessing, 0, "", "")
	text, err := domain.ParseTextFile(doc.StoragePath, doc.FileType)
	if err == nil && text == "" {
		err = fmt.Errorf("文档未提取到可索引文本")
	}
	if err != nil {
		fail(err)
		return
	}
	chunks := domain.ChunkText(text, 800)
	inputs := make([]string, len(chunks))
	for i := range chunks {
		inputs[i] = embeddingText(chunks[i])
	}
	vectors, err := infrarag.Embed(ctx, inputs)
	if err == nil && len(vectors) != len(chunks) {
		err = fmt.Errorf("向量数量不匹配")
	}
	if err != nil {
		fail(err)
		return
	}
	if err = infrarag.Ensure(ctx, len(vectors[0])); err != nil {
		fail(err)
		return
	}
	models := make([]*model.DocumentChunk, 0, len(chunks))
	points := make([]map[string]interface{}, 0, len(chunks))
	for i, c := range chunks {
		id := utils.GenerateUUID()
		models = append(models, &model.DocumentChunk{ChunkID: id, DocumentPK: doc.ID, DocumentID: doc.DocumentID, ChunkIndex: int64(c.Index), SectionTitle: c.SectionTitle, Content: c.Content, TokenCount: int64(c.TokenCount), EmbeddingModel: configModel(), VectorPointID: id})
		points = append(points, map[string]interface{}{"id": id, "vector": vectors[i], "payload": map[string]interface{}{"chunk_id": id, "document_id": doc.DocumentID, "user_name": doc.UserName, "file_name": doc.FileName, "chunk_index": c.Index, "section_title": c.SectionTitle}})
	}
	if err = infrarag.Upsert(ctx, points); err != nil {
		fail(err)
		return
	}
	if _, err = repo.BatchCreateChunks(models); err != nil {
		fail(err)
		return
	}
	_ = repo.UpdateIngestResult(doc.DocumentID, model.DocumentParseStatusReady, int64(len(models)), configModel(), "")
}
func configModel() string { return infrarag.EmbeddingModel() }
func embeddingText(chunk domain.Chunk) string {
	if chunk.SectionTitle == "" {
		return chunk.Content
	}
	return "标题：" + chunk.SectionTitle + "\n正文：" + chunk.Content
}
