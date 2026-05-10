package websocket

const (
	EventSessionStarted  = "session_started"
	EventNextQuestion    = "next_question"
	EventAnswerSubmitted = "answer_submitted" // организатору — сколько ответили
	EventSessionFinished = "session_finished"
	EventLeaderboard     = "leaderboard"
	EventError           = "error"
)

// базовая обёртка для любого сообщения
type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// payload для нового вопроса
type QuestionPayload struct {
	QuestionID    string          `json:"question_id"`
	QuestionIndex int             `json:"question_index"`
	Text          string          `json:"text"`
	ImageURL      string          `json:"image_url,omitempty"`
	Type          string          `json:"type"`
	TimeLimit     int             `json:"time_limit"`
	Answers       []AnswerPayload `json:"answers"`
}

type AnswerPayload struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// payload для лидерборда
type LeaderboardPayload struct {
	Participants []ParticipantResult `json:"participants"`
}

type ParticipantResult struct {
	Username   string `json:"username"`
	TotalScore int    `json:"total_score"`
	Rank       int    `json:"rank"`
}
