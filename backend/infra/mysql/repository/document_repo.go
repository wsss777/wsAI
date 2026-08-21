package repository

import (
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func CreateDocument(document *model.Document) (*model.Document, error) {
	err := mysql.DB.Create(document).Error
	if err != nil {
		logger.L().Error("Create document error",
			zap.Int64("id", document.ID),
			zap.String("document_id", document.DocumentID),
			zap.String("username", document.UserName),
			zap.String("file_name", document.FileName),
			zap.Error(err))
		return nil, err
	}
	return document, nil
}

func GetDocumentByDocumentID(documentID string) (*model.Document, error) {
	document := &model.Document{}
	if documentID == "" {
		return document, nil
	}
	err := mysql.DB.
		Where("document_id = ?", documentID).
		Order("updated_at desc").
		First(document).
		Error
	if err != nil {
		logger.L().Error("GetDocumentByDocumentID error",
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, err
	}
	return document, nil
}

func GetDocumentByUserNameAndDocumentID(userName, documentID string) (*model.Document, error) {
	document := &model.Document{}
	if userName == "" || documentID == "" {
		return document, nil
	}
	err := mysql.DB.
		Where("user_name = ? AND document_id = ?", userName, documentID).
		Order("updated_at desc").
		First(document).
		Error
	if err != nil {
		logger.L().Error("GetDocumentByUserNameAndDocumentID error",
			zap.String("user_name", userName),
			zap.String("document_id", documentID),
			zap.Error(err))
		return nil, err
	}
	return document, nil
}

func ListDocumentsByUserName(userName string) ([]*model.Document, error) {
	var documents []*model.Document
	err := mysql.DB.
		Where("user_name = ?", userName).
		Order("updated_at desc").
		Find(&documents).
		Error
	if err != nil {
		logger.L().Error("ListDocumentsByUserName error",
			zap.String("user_name", userName),
			zap.Error(err))
		return nil, err
	}
	return documents, nil
}

// DeleteDocumentByUserNameAndDocumentID 删除文档记录及其关联记录。
// 上传文件由调用方负责删除。
func DeleteDocumentByUserNameAndDocumentID(userName, documentID string) (*model.Document, error) {
	document, err := GetDocumentByUserNameAndDocumentID(userName, documentID)
	if err != nil {
		return nil, err
	}

	return document, mysql.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("document_id = ?", documentID).Delete(&model.SessionDocument{}).Error; err != nil {
			return err
		}
		if err := tx.Where("document_id = ?", documentID).Delete(&model.DocumentChunk{}).Error; err != nil {
			return err
		}
		return tx.Delete(document).Error
	})
}

func UpdateParseStatus(documentID string, parseStatus string) (*model.Document, error) {
	document := &model.Document{}
	if documentID == "" {
		return document, nil
	}
	err := mysql.DB.
		Model(&model.Document{}).
		Where("document_id = ?", documentID).
		Update("parse_status", parseStatus).
		Error
	if err != nil {
		logger.L().Error("UpdateParseStatus error",
			zap.String("document_id", documentID),
			zap.String("parse_status", parseStatus),
			zap.Error(err))
		return nil, err
	}
	return document, nil
}

func UpdateIngestResult(documentID, status string, count int64, embeddingModel, message string) error {
	return mysql.DB.Model(&model.Document{}).Where("document_id = ?", documentID).Updates(map[string]interface{}{"parse_status": status, "chunk_count": count, "embedding_model": embeddingModel, "error_message": message}).Error
}
