package session

import (
	"net/http"

	"github.com/geolav/QEasyApp/internal/port"
	"github.com/gin-gonic/gin"
)

type handler struct {
	sessionUC port.SessionUseCase
}

func New(sessionUC port.SessionUseCase) *handler {
	return &handler{sessionUC: sessionUC}
}

func (h *handler) CreateSession(c *gin.Context) {
	organizerID := c.GetString("user_id")
	var req struct {
		QuizID string `json:"quiz_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.sessionUC.CreateSession(req.QuizID, organizerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.ID,
		"room_code":  session.RoomCode,
	})
}

func (h *handler) JoinSession(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		RoomCode string `json:"room_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	participant, err := h.sessionUC.JoinSession(req.RoomCode, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"participant_id": participant.ID,
		"session_id":     participant.SessionID,
	})
}

func (h *handler) StartSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	organizerID := c.GetString("user_id")

	question, err := h.sessionUC.StartSession(sessionID, organizerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "session started",
		"question": question,
	})
}

func (h *handler) NextQuestion(c *gin.Context) {
	sessionID := c.Param("session_id")
	organizerID := c.GetString("user_id")

	question, err := h.sessionUC.NextQuestion(sessionID, organizerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "next question",
		"question": question,
	})
}

func (h *handler) GetLeaderboard(c *gin.Context) {
	sessionID := c.Param("session_id")

	participants, err := h.sessionUC.GetLeaderboard(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, participants)
}

func (h *handler) FinishSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	organizerID := c.GetString("user_id")

	if err := h.sessionUC.FinishSession(sessionID, organizerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session finished"})
}
