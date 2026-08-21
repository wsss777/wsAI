package rag

import (
	"net/http"
	application "wsai/backend/application/rag"
	response "wsai/backend/response"
	"wsai/backend/response/code"

	"github.com/gin-gonic/gin"
)

// UploadDocument 上传知识库文档。
// @Summary      上传知识库文档
// @Description  上传 PDF、Markdown 或纯文本文件。当前接口仅保存原始文件和文档记录，初始解析状态为待处理。
// @Tags         RAG 文档管理
// @Accept       multipart/form-data
// @Produce      json
// @Param        file  formData  file  true  "待上传的文档，仅支持 .pdf、.md、.txt"
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "成功，包含 document 和 response 字段"
// @Failure      200  {object}  common.Response         "参数错误或服务异常；本项目业务错误仍使用 HTTP 200 返回"
// @Router       /api/v1/rag/documents [post]
func UploadDocument(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeInvalidParams))
		return
	}
	doc, err := application.UploadDocument(c.GetString("username"), file)
	if err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeServerBusy))
		return
	}
	c.JSON(http.StatusOK, gin.H{"document": doc, "response": (&response.Response{}).CodeOf(code.CodeSuccess)})
}

// ListDocuments 获取当前用户的知识库文档列表。
// @Summary      获取知识库文档列表
// @Description  按更新时间倒序返回当前登录用户上传的全部文档及其解析状态。
// @Tags         RAG 文档管理
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "成功，包含 documents 和 response 字段"
// @Failure      200  {object}  common.Response         "服务异常；本项目业务错误仍使用 HTTP 200 返回"
// @Router       /api/v1/rag/documents [get]
func ListDocuments(c *gin.Context) {
	documents, err := application.ListUserDocuments(c.GetString("username"))
	if err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeServerBusy))
		return
	}
	c.JSON(http.StatusOK, gin.H{"documents": documents, "response": (&response.Response{}).CodeOf(code.CodeSuccess)})
}

// GetDocument 获取指定知识库文档详情。
// @Summary      获取知识库文档详情
// @Description  只能查看当前登录用户拥有的文档；响应包含文件信息、解析状态和切片数量。
// @Tags         RAG 文档管理
// @Produce      json
// @Param        document_id  path  string  true  "文档 ID"
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "成功，包含 document 和 response 字段"
// @Failure      200  {object}  common.Response         "参数错误、文档不存在或无访问权限；本项目业务错误仍使用 HTTP 200 返回"
// @Router       /api/v1/rag/documents/{document_id} [get]
func GetDocument(c *gin.Context) {
	doc, err := application.GetDocumentDeatail(c.GetString("username"), c.Param("document_id"))
	if err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeInvalidParams))
		return
	}
	c.JSON(http.StatusOK, gin.H{"document": doc, "response": (&response.Response{}).CodeOf(code.CodeSuccess)})
}

// DeleteDocument 删除指定知识库文档。
// @Summary      删除知识库文档
// @Description  删除当前用户拥有的文档记录、会话关联、已保存的切片，以及上传的原始文件目录。
// @Tags         RAG 文档管理
// @Produce      json
// @Param        document_id  path  string  true  "文档 ID"
// @Security     ApiKeyAuth
// @Success      200  {object}  common.Response  "删除成功"
// @Failure      200  {object}  common.Response  "参数错误、文档不存在或无访问权限；本项目业务错误仍使用 HTTP 200 返回"
// @Router       /api/v1/rag/documents/{document_id} [delete]
func DeleteDocument(c *gin.Context) {
	if err := application.DeleteDocument(c.GetString("username"), c.Param("document_id")); err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeInvalidParams))
		return
	}
	c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeSuccess))
}

// BindDocuments 为会话绑定知识库文档。
// @Summary      为会话绑定知识库文档
// @Description  为当前登录用户拥有的会话关联一个或多个本人文档；重复的关联会被忽略。
// @Tags         RAG 会话关联
// @Accept       json
// @Produce      json
// @Param        session_id  path  string                   true  "会话 ID"
// @Param        request     body  map[string][]string      true  "请求体，例如：{\"document_ids\":[\"文档 ID\"]}"
// @Security     ApiKeyAuth
// @Success      200  {object}  common.Response  "绑定成功"
// @Failure      200  {object}  common.Response  "参数错误、会话或文档不存在、无访问权限或服务异常；本项目业务错误仍使用 HTTP 200 返回"
// @Router       /api/v1/rag/sessions/{session_id}/documents [post]
func BindDocuments(c *gin.Context) {
	var request struct {
		DocumentIDs []string `json:"document_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil || len(request.DocumentIDs) == 0 {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeInvalidParams))
		return
	}
	if err := application.BindDocumentsToSession(c.GetString("username"), c.Param("session_id"), request.DocumentIDs); err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeServerBusy))
		return
	}
	c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeSuccess))
}

// UnbindDocument 仅从当前会话移除一个文档。
func UnbindDocument(c *gin.Context) {
	if err := application.UnbindDocumentFromSession(c.GetString("username"), c.Param("session_id"), c.Param("document_id")); err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeInvalidParams))
		return
	}
	c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeSuccess))
}

// ListSessionDocuments 获取会话已绑定的知识库文档。
// @Summary      获取会话已绑定的知识库文档
// @Description  仅返回当前登录用户拥有的指定会话关联的文档，按绑定时间正序排列。
// @Tags         RAG 会话关联
// @Produce      json
// @Param        session_id  path  string  true  "会话 ID"
// @Security     ApiKeyAuth
// @Success      200  {object}  map[string]interface{}  "成功，包含 documents 和 response 字段"
// @Failure      200  {object}  common.Response         "会话不存在、无访问权限或服务异常；本项目业务错误仍使用 HTTP 200 返回"
// @Router       /api/v1/rag/sessions/{session_id}/documents [get]
func ListSessionDocuments(c *gin.Context) {
	documents, err := application.ListSessionDocuments(c.GetString("username"), c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusOK, (&response.Response{}).CodeOf(code.CodeServerBusy))
		return
	}
	c.JSON(http.StatusOK, gin.H{"documents": documents, "response": (&response.Response{}).CodeOf(code.CodeSuccess)})
}
