package repository

import (
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

func BindSessionDocuments(bindings []*model.SessionDocument) ([]*model.SessionDocument, error) {
	if len(bindings) == 0 {
		return bindings, nil
	}

	err := mysql.DB.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "session_id"},
				{Name: "document_pk"},
			},
			DoNothing: true,
		}).
		Create(bindings).
		Error
	if err != nil {
		logger.L().Error("BindSessionDocuments error",
			zap.Int("count", len(bindings)),
			zap.String("session_id", bindings[0].SessionID),
			zap.Error(err))
		return nil, err
	}

	return bindings, nil
}

func ListSessionDocumentsBySessionID(sessionID string) ([]*model.SessionDocument, error) {
	var bindings []*model.SessionDocument
	if sessionID == "" {
		return bindings, nil
	}

	err := mysql.DB.
		Where("session_id = ?", sessionID).
		Order("created_at asc").
		Find(&bindings).
		Error
	if err != nil {
		logger.L().Error("ListSessionDocumentsBySessionID error",
			zap.String("session_id", sessionID),
			zap.Error(err))
		return nil, err
	}

	return bindings, nil
}

func DeleteSessionDocument(sessionID string, documentPK int64) error {
	if sessionID == "" || documentPK <= 0 {
		return nil
	}

	err := mysql.DB.
		Where("session_id = ? AND document_pk = ?", sessionID, documentPK).
		Delete(&model.SessionDocument{}).
		Error
	if err != nil {
		logger.L().Error("DeleteSessionDocument error",
			zap.String("session_id", sessionID),
			zap.Int64("document_pk", documentPK),
			zap.Error(err))
		return err
	}

	return nil
}
