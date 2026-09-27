package agent

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	app "wsai/backend/application/agent"
	ragapp "wsai/backend/application/rag"
	sessionapp "wsai/backend/application/session"
	repo "wsai/backend/infra/mysql/repository"
)

type request struct {
	SessionID   string   `json:"session_id"`
	Question    string   `json:"question" binding:"required"`
	ModelType   string   `json:"model_type"`
	DocumentIDs []string `json:"document_ids"`
}

func StreamRun(c *gin.Context) {
	var r request
	if c.ShouldBindJSON(&r) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid request"})
		return
	}
	u := c.GetString("username")
	if r.SessionID == "" {
		id, code := sessionapp.CreateStreamSessionOnly(u, r.Question)
		if code != 1000 {
			c.JSON(500, gin.H{"message": "failed to create session"})
			return
		}
		r.SessionID = id
		if len(r.DocumentIDs) > 0 {
			if e := ragapp.BindDocumentsToSession(u, id, r.DocumentIDs); e != nil {
				c.JSON(400, gin.H{"message": "failed to bind documents"})
				return
			}
		}
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	emit := func(event string, data any) {
		b, _ := json.Marshal(data)
		_, _ = c.Writer.WriteString("event: " + event + "\ndata: " + string(b) + "\n\n")
		c.Writer.Flush()
	}
	_, _ = app.Run(c.Request.Context(), app.Input{UserName: u, SessionID: r.SessionID, Question: r.Question, ModelType: r.ModelType}, emit)
}
func GetRun(c *gin.Context) {
	runID, e := strconv.ParseUint(c.Param("run_id"), 10, 0)
	if e != nil || runID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid run id"})
		return
	}
	run, e := repo.GetAgentRun(c.GetString("username"), uint(runID))
	if e != nil {
		c.JSON(404, gin.H{"message": "run not found"})
		return
	}
	steps, e := repo.ListAgentSteps(run.ID)
	if e != nil {
		c.JSON(500, gin.H{"message": "failed to load steps"})
		return
	}
	c.JSON(200, gin.H{"run": run, "steps": steps})
}

func GetArtifact(c *gin.Context) {
	artifact, err := repo.GetAgentArtifact(c.GetString("username"), c.Param("artifact_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "artifact not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"artifact": artifact})
}
