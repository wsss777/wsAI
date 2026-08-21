package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"wsai/backend/infra/logger"
	"wsai/backend/infra/mysql"
	documentrepo "wsai/backend/infra/mysql/repository"
	sessiondocumentrepo "wsai/backend/infra/mysql/repository"
	sessionrepo "wsai/backend/infra/mysql/repository"
	"wsai/backend/model"
	"wsai/backend/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UploadDocument 上传文档并保存记录
func UploadDocument(userName string, fileHeader *multipart.FileHeader) (*model.Document, error) {
	userName = strings.TrimSpace(userName)
	if userName == "" {
		return nil, fmt.Errorf("username is empty")
	}
	if fileHeader == nil {
		return nil, fmt.Errorf("fileHeader is empty")
	}
	if strings.TrimSpace(fileHeader.Filename) == "" {
		return nil, fmt.Errorf("fileHeader filename is empty")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !isAllowedDocumentExt(ext) {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	safeUserName := sanitizePathSegment(userName)
	fileType := strings.TrimPrefix(ext, ".")

	tempDir := filepath.Join("backend", "data", "uploads", "rag", safeUserName, "_tmp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		logger.L().Error("UploadDocument create temp dir failed",
			zap.String("username", userName),
			zap.String("temp_dir", tempDir),
			zap.Error(err))
		return nil, err
	}

	src, err := fileHeader.Open()
	if err != nil {
		logger.L().Error("UploadDocument open failed",
			zap.String("username", userName),
			zap.String("file_name", fileHeader.Filename),
			zap.Error(err))
		return nil, err
	}
	defer src.Close()

	tempFile, err := os.CreateTemp(tempDir, "upload-*"+ext)
	if err != nil {
		logger.L().Error("UploadDocument create temp file failed",
			zap.String("username", userName),
			zap.String("temp_dir", tempDir),
			zap.Error(err))
		return nil, err
	}
	tempPath := tempFile.Name()

	defer func() {
		_ = tempFile.Close()
	}()

	defer func() {
		if _, statErr := os.Stat(tempPath); statErr == nil {
			_ = os.Remove(tempPath)
		}
	}()

	hasher := sha256.New()
	writer := io.MultiWriter(tempFile, hasher)
	if _, err := io.Copy(writer, src); err != nil {
		logger.L().Error("UploadDocument copy file failed",
			zap.String("username", userName),
			zap.String("temp_path", tempPath),
			zap.Error(err))
		return nil, err
	}

	contentHash := hex.EncodeToString(hasher.Sum(nil))

	existingDoc := &model.Document{}
	err = mysql.DB.
		Where("user_name = ? AND content_hash = ?", userName, contentHash).
		Order("updated_at desc").
		First(existingDoc).
		Error
	if err == nil {
		logger.L().Info("UploadDocument duplicated document detected",
			zap.String("username", userName),
			zap.String("file_name", fileHeader.Filename),
			zap.String("document_id", existingDoc.DocumentID),
			zap.String("content_hash", contentHash))
		return existingDoc, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.L().Error("UploadDocument query duplicated document failed",
			zap.String("username", userName),
			zap.String("file_name", fileHeader.Filename),
			zap.String("content_hash", contentHash),
			zap.Error(err))
		return nil, err
	}

	documentID := utils.GenerateUUID()
	finalDir := filepath.Join("backend", "data", "uploads", "rag", safeUserName, documentID)
	if err := os.MkdirAll(finalDir, 0o755); err != nil {
		logger.L().Error("UploadDocument create final dir failed",
			zap.String("username", userName),
			zap.String("document_id", documentID),
			zap.String("final_dir", finalDir),
			zap.Error(err))
		return nil, err
	}

	finalPath := filepath.Join(finalDir, "original"+ext)
	if err := tempFile.Close(); err != nil {
		logger.L().Error("UploadDocument close temp file failed",
			zap.String("username", userName),
			zap.String("temp_path", tempPath),
			zap.Error(err))
		return nil, err
	}

	if err := os.Rename(tempPath, finalPath); err != nil {
		logger.L().Error("UploadDocument move temp file failed",
			zap.String("username", userName),
			zap.String("temp_path", tempPath),
			zap.String("final_path", finalPath),
			zap.Error(err))
		return nil, err
	}

	doc := &model.Document{
		DocumentID:     documentID,
		UserName:       userName,
		FileName:       fileHeader.Filename,
		FileType:       fileType,
		FileSize:       fileHeader.Size,
		StoragePath:    finalPath,
		ContentHash:    contentHash,
		ParseStatus:    model.DocumentParseStatusPending,
		ChunkCount:     0,
		EmbeddingModel: "",
		ErrorMessage:   "",
	}

	createdDoc, err := documentrepo.CreateDocument(doc)
	if err != nil {
		_ = os.Remove(finalPath)
		_ = os.Remove(finalDir)

		logger.L().Error("UploadDocument create document record failed",
			zap.String("username", userName),
			zap.String("document_id", documentID),
			zap.String("storage_path", finalPath),
			zap.Error(err))
		return nil, err
	}

	go IngestDocument(createdDoc)
	return createdDoc, nil
}

// isAllowedDocumentExt 检查文档类型
func isAllowedDocumentExt(ext string) bool {
	switch ext {
	case ".pdf", ".txt", ".md":
		return true
	default:
		return false
	}
}

// sanitizePathSegment 清理路径片段
func sanitizePathSegment(s string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(strings.TrimSpace(s))
}

// GetDocumentDeatail 查询用户文档详情
func GetDocumentDeatail(userName, documentID string) (*model.Document, error) {
	userName = strings.TrimSpace(userName)
	documentID = strings.TrimSpace(documentID)
	if userName == "" {
		return nil, fmt.Errorf("username is empty")
	}
	if documentID == "" {
		return nil, fmt.Errorf("documentID is empty")
	}

	document, err := documentrepo.GetDocumentByUserNameAndDocumentID(userName, documentID)
	if err != nil {
		return nil, err
	}
	return document, nil
}

// ListUserDocuments 查询用户文档列表
func ListUserDocuments(userName string) ([]*model.Document, error) {
	userName = strings.TrimSpace(userName)
	if userName == "" {
		return nil, fmt.Errorf("username is empty")
	}

	documents, err := documentrepo.ListDocumentsByUserName(userName)
	if err != nil {
		return nil, err
	}
	return documents, nil
}

// DeleteDocument 删除用户拥有的文档及其上传源文件。
func DeleteDocument(userName, documentID string) error {
	userName = strings.TrimSpace(userName)
	documentID = strings.TrimSpace(documentID)
	if userName == "" || documentID == "" {
		return fmt.Errorf("username and documentID are required")
	}
	document, err := documentrepo.DeleteDocumentByUserNameAndDocumentID(userName, documentID)
	if err != nil {
		return err
	}
	if document.StoragePath != "" {
		baseDir, baseErr := filepath.Abs(filepath.Join("backend", "data", "uploads", "rag"))
		targetDir, targetErr := filepath.Abs(filepath.Dir(document.StoragePath))
		rel, relErr := filepath.Rel(baseDir, targetDir)
		if baseErr == nil && targetErr == nil && relErr == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			if err := os.RemoveAll(targetDir); err != nil {
				logger.L().Warn("DeleteDocument remove source directory failed", zap.String("document_id", documentID), zap.Error(err))
			}
		} else {
			logger.L().Warn("DeleteDocument skipped unsafe source directory", zap.String("document_id", documentID), zap.String("source_path", document.StoragePath))
		}
	}
	return nil
}

// BindDocumentsToSession 绑定会话和文档
func BindDocumentsToSession(userName string, sessionID string, documentIDs []string) error {
	userName = strings.TrimSpace(userName)
	sessionID = strings.TrimSpace(sessionID)
	if userName == "" {
		return fmt.Errorf("username is empty")
	}
	if sessionID == "" {
		return fmt.Errorf("sessionID is empty")
	}
	if len(documentIDs) == 0 {
		return nil
	}

	sessionInfo, err := sessionrepo.GetSessionByID(sessionID)
	if err != nil {
		return err
	}
	if sessionInfo.UserName != userName {
		return fmt.Errorf("session does not belong to user")
	}

	seen := make(map[string]struct{}, len(documentIDs))
	bindings := make([]*model.SessionDocument, 0, len(documentIDs))
	for _, documentID := range documentIDs {
		documentID = strings.TrimSpace(documentID)
		if documentID == "" {
			return fmt.Errorf("documentID is empty")
		}
		if _, ok := seen[documentID]; ok {
			continue
		}

		document, err := documentrepo.GetDocumentByUserNameAndDocumentID(userName, documentID)
		if err != nil {
			return err
		}

		bindings = append(bindings, &model.SessionDocument{
			SessionID:  sessionID,
			DocumentPK: document.ID,
			DocumentID: document.DocumentID,
		})
		seen[documentID] = struct{}{}
	}

	if len(bindings) == 0 {
		return nil
	}

	_, err = sessiondocumentrepo.BindSessionDocuments(bindings)
	return err
}

// UnbindDocumentFromSession 仅移除用户文档与用户会话之间的关联。
// 源文档及其向量索引仍可供其他会话使用。
func UnbindDocumentFromSession(userName, sessionID, documentID string) error {
	userName = strings.TrimSpace(userName)
	sessionID = strings.TrimSpace(sessionID)
	documentID = strings.TrimSpace(documentID)
	if userName == "" || sessionID == "" || documentID == "" {
		return fmt.Errorf("username, sessionID and documentID are required")
	}

	sessionInfo, err := sessionrepo.GetSessionByID(sessionID)
	if err != nil {
		return err
	}
	if sessionInfo.UserName != userName {
		return fmt.Errorf("session does not belong to user")
	}

	document, err := documentrepo.GetDocumentByUserNameAndDocumentID(userName, documentID)
	if err != nil {
		return err
	}
	return sessiondocumentrepo.DeleteSessionDocument(sessionID, document.ID)
}

// ListSessionDocuments 查询会话已绑定文档
func ListSessionDocuments(userName string, sessionID string) ([]*model.Document, error) {
	userName = strings.TrimSpace(userName)
	sessionID = strings.TrimSpace(sessionID)
	if userName == "" {
		return nil, fmt.Errorf("username is empty")
	}
	if sessionID == "" {
		return nil, fmt.Errorf("sessionID is empty")
	}

	sessionInfo, err := sessionrepo.GetSessionByID(sessionID)
	if err != nil {
		return nil, err
	}
	if sessionInfo.UserName != userName {
		return nil, fmt.Errorf("session does not belong to user")
	}

	bindings, err := sessiondocumentrepo.ListSessionDocumentsBySessionID(sessionID)
	if err != nil {
		return nil, err
	}
	if len(bindings) == 0 {
		return []*model.Document{}, nil
	}

	documents := make([]*model.Document, 0, len(bindings))
	for _, binding := range bindings {
		document, err := documentrepo.GetDocumentByUserNameAndDocumentID(userName, binding.DocumentID)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}

	return documents, nil
}
