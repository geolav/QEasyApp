package websocket

import (
	"encoding/json"
	"log"
)

func (h *WSHandler) handleStartSession(client *Client) {
	question, err := h.sessionUC.StartSession(client.SessionID, client.UserID)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}

	answers := make([]AnswerPayload, len(question.Answers))
	for i, a := range question.Answers {
		answers[i] = AnswerPayload{ID: a.ID, Text: a.Text}
	}

	data, _ := json.Marshal(Event{
		Type: EventSessionStarted,
		Payload: QuestionPayload{
			QuestionID:    question.ID,
			QuestionIndex: 0,
			Text:          question.Text,
			ImageURL:      question.ImageURL,
			Type:          string(question.Type),
			TimeLimit:     question.TimeLimit,
			Answers:       answers,
		},
	})

	h.hub.Broadcast <- Message{
		SessionID: client.SessionID,
		Data:      data,
	}
}

func (h *WSHandler) handleNextQuestion(client *Client, payload json.RawMessage) {
	question, err := h.sessionUC.NextQuestion(client.SessionID, client.UserID)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}

	answers := make([]AnswerPayload, len(question.Answers))
	for i, a := range question.Answers {
		answers[i] = AnswerPayload{ID: a.ID, Text: a.Text}
	}

	data, _ := json.Marshal(Event{
		Type: EventNextQuestion,
		Payload: QuestionPayload{
			QuestionID: question.ID,
			Text:       question.Text,
			ImageURL:   question.ImageURL,
			Type:       string(question.Type),
			TimeLimit:  question.TimeLimit,
			Answers:    answers,
		},
	})

	// рассылаем всем в сессии
	h.hub.Broadcast <- Message{
		SessionID: client.SessionID,
		Data:      data,
	}
}

func (h *WSHandler) handleSubmitAnswer(client *Client, payload json.RawMessage) {
	var req struct {
		QuestionID     string   `json:"question_id"`
		AnswerIDs      []string `json:"answer_ids"` // массив
		AnswerID       string   `json:"answer_id"`  // fallback для старых клиентов
		ResponseTimeMs int      `json:"response_time_ms"`
		ParticipantID  string   `json:"participant_id"`
	}

	if err := json.Unmarshal(payload, &req); err != nil {
		h.sendError(client, "invalid payload")
		return
	}

	// если answer_ids не пришёл — берём одиночный answer_id
	answerIDs := req.AnswerIDs
	if len(answerIDs) == 0 && req.AnswerID != "" {
		answerIDs = []string{req.AnswerID}
	}

	score, err := h.sessionUC.SubmitAnswer(
		client.SessionID,
		req.ParticipantID,
		req.QuestionID,
		answerIDs,
		req.ResponseTimeMs,
	)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}

	data, _ := json.Marshal(Event{
		Type:    EventAnswerSubmitted,
		Payload: map[string]int{"score": score},
	})
	client.Send <- data
}

func (h *WSHandler) handleFinishSession(client *Client) {
	if err := h.sessionUC.FinishSession(client.SessionID, client.UserID); err != nil {
		h.sendError(client, err.Error())
		return
	}

	leaderboard, err := h.sessionUC.GetLeaderboard(client.SessionID)
	if err != nil {
		h.sendError(client, err.Error())
		return
	}

	data, _ := json.Marshal(Event{
		Type:    EventLeaderboard,
		Payload: leaderboard,
	})

	// рассылаем лидерборд всем
	h.hub.Broadcast <- Message{
		SessionID: client.SessionID,
		Data:      data,
	}
}

func (h *WSHandler) sendError(client *Client, msg string) {
	data, _ := json.Marshal(Event{
		Type:    EventError,
		Payload: map[string]string{"message": msg},
	})
	client.Send <- data
	log.Printf("ws error for user %s: %s", client.UserID, msg)
}
