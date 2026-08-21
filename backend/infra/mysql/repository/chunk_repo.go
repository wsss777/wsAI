package repository

import (
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"go.uber.org/zap"
)

func BatchCreateChunks(chunks []*model.DocumentChunk) ([]*model.DocumentChunk, error) {
	if len(chunks) == 0 {
		return chunks, nil
	}

	err := mysql.DB.Create(chunks).Error
	if err != nil {
		logger.L().Error("BatchCreateChunks error",
			zap.Int("count", len(chunks)),
			zap.Int64("document_pk", chunks[0].DocumentPK),
			zap.String("document_id", chunks[0].DocumentID),
			zap.Error(err))
		return nil, err
	}

	return chunks, nil
}

func ListChunksByDocumentPK(documentPK int64) ([]*model.DocumentChunk, error) {
	var chunks []*model.DocumentChunk
	if documentPK <= 0 {
		return chunks, nil
	}

	err := mysql.DB.
		Where("document_pk = ?", documentPK).
		Order("chunk_index asc").
		Find(&chunks).
		Error
	if err != nil {
		logger.L().Error("ListChunksByDocumentPK error",
			zap.Int64("document_pk", documentPK),
			zap.Error(err))
		return nil, err
	}

	return chunks, nil
}

func GetChunksByChunkIDs(chunkIDs []string) ([]*model.DocumentChunk, error) {
	var chunks []*model.DocumentChunk
	if len(chunkIDs) == 0 {
		return chunks, nil
	}

	err := mysql.DB.
		Where("chunk_id IN (?)", chunkIDs).
		Order("chunk_index asc").
		Find(&chunks).
		Error
	if err != nil {
		logger.L().Error("GetChunksByChunkIDs error",
			zap.Strings("chunk_ids", chunkIDs),
			zap.Error(err))
		return nil, err
	}

	return chunks, nil
}

func DeleteChunksByDocumentID(documentID string) error {
	return mysql.DB.Where("document_id = ?", documentID).Delete(&model.DocumentChunk{}).Error
}
