package quiz

import (
	"log"
	"net/http"

	"github.com/geolav/QEasyApp/internal/entity"
	"github.com/geolav/QEasyApp/internal/port"
	"github.com/gin-gonic/gin"
)

type handler struct {
	quizUC port.QuizUseCase
}

func New(quizUC port.QuizUseCase) *handler {
	return &handler{quizUC: quizUC}
}

func (h *handler) CreateQuiz(c *gin.Context) {
	creatorID := c.GetString("user_id")
	log.Printf("creatorID: '%s'", creatorID)

	var req struct {
		Title           string `json:"title" binding:"required"`
		Category        string `json:"category"`
		TimePerQuestion int    `json:"time_per_question" binding:"required,min=5"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quiz, err := h.quizUC.CreateQuiz(creatorID, req.Title, req.Category, req.TimePerQuestion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, quiz)
}

func (h *handler) AddQuestion(c *gin.Context) {
	quizID := c.Param("quiz_id")

	var req struct {
		Text      string `json:"text" binding:"required"`
		ImageURL  string `json:"image_url"`
		Type      string `json:"type" binding:"required"`
		TimeLimit int    `json:"time_limit" binding:"required,min=5"`
		Answers   []struct {
			Text      string `json:"text" binding:"required"`
			IsCorrect bool   `json:"is_correct"`
		} `json:"answers" binding:"required,min=2"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	answers := make([]entity.Answer, len(req.Answers))
	for i, a := range req.Answers {
		answers[i] = entity.Answer{
			Text:      a.Text,
			IsCorrect: a.IsCorrect,
		}
	}

	question := entity.Question{
		Text:      req.Text,
		ImageURL:  req.ImageURL,
		Type:      entity.QuestionType(req.Type),
		TimeLimit: req.TimeLimit,
		Answers:   answers,
	}

	if err := h.quizUC.AddQuestion(quizID, question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "question added"})
}

func (h *handler) PublishQuiz(c *gin.Context) {
	quizID := c.Param("quiz_id")
	creatorID := c.GetString("user_id")

	if err := h.quizUC.PublishQuiz(quizID, creatorID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "quiz published"})
}

func (h *handler) GetQuiz(c *gin.Context) {
	quizID := c.Param("quiz_id")

	quiz, err := h.quizUC.GetQuiz(quizID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quiz)
}

func (h *handler) GetMyQuizzes(c *gin.Context) {
	creatorID := c.GetString("user_id")

	quizzes, err := h.quizUC.GetMyQuizzes(creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quizzes)
}
