package repository

import (
	"time"
	"wsai/backend/infra/mysql"
	"wsai/backend/model"

	"gorm.io/gorm"
)

func CreateAgentRun(run *model.AgentRun) error                { return mysql.DB.Create(run).Error }
func CreateAgentStep(step *model.AgentStep) error             { return mysql.DB.Create(step).Error }
func CreateAgentToolCall(call *model.AgentToolCall) error     { return mysql.DB.Create(call).Error }
func CreateAgentArtifact(artifact *model.AgentArtifact) error { return mysql.DB.Create(artifact).Error }
func SetAgentInputMessageID(id uint, messageID string) error {
	return mysql.DB.Model(&model.AgentRun{}).Where("id = ?", id).Update("input_message_id", messageID).Error
}

func FinishAgentRun(id uint, status, answer, outputMessageID, runErr string) error {
	now := time.Now()
	return mysql.DB.Model(&model.AgentRun{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "final_answer": answer, "output_message_id": outputMessageID, "error": runErr, "finished_at": &now,
	}).Error
}

func FinishAgentToolCall(id uint, summary, errorMessage string) error {
	now := time.Now()
	return mysql.DB.Model(&model.AgentToolCall{}).Where("id = ?", id).Updates(map[string]any{"result_summary": summary, "error_message": errorMessage, "finished_at": &now}).Error
}

func SetAgentToolCallArtifact(id uint, artifactID string) error {
	return mysql.DB.Model(&model.AgentToolCall{}).Where("id = ?", id).Update("artifact_id", artifactID).Error
}

func GetAgentArtifact(username, id string) (*model.AgentArtifact, error) {
	var artifact model.AgentArtifact
	err := mysql.DB.Table("agent_artifacts AS a").Select("a.*").
		Joins("JOIN agent_runs AS r ON r.id = a.run_id").
		Where("a.id = ? AND r.user_name = ?", id, username).Scan(&artifact).Error
	if err == nil && artifact.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &artifact, err
}

func GetAgentRun(username string, id uint) (*model.AgentRun, error) {
	var run model.AgentRun
	err := mysql.DB.Where("id = ? AND user_name = ?", id, username).First(&run).Error
	return &run, err
}

func ListAgentSteps(runID uint) ([]model.AgentStep, error) {
	var steps []model.AgentStep
	err := mysql.DB.Where("run_id = ?", runID).Order("sequence asc, id asc").Find(&steps).Error
	return steps, err
}
