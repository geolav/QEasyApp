CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE quizzes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    time_per_question INT NOT NULL DEFAULT 30,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    order_index INT NOT NULL,
    type VARCHAR(20) NOT NULL,
    text TEXT NOT NULL,
    image_url TEXT,
    time_limit INT NOT NULL DEFAULT 30
);

CREATE TABLE answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id),
    room_code VARCHAR(10) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'waiting',
    current_question_index INT NOT NULL DEFAULT 0,
    started_at TIMESTAMP,
    ended_at TIMESTAMP
);

CREATE TABLE session_participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    is_organizer BOOLEAN NOT NULL DEFAULT FALSE,
    total_score INT NOT NULL DEFAULT 0,
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(session_id, user_id)
);

CREATE TABLE user_answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES session_participants(id),
    question_id UUID NOT NULL REFERENCES questions(id),
    answer_id UUID NOT NULL REFERENCES answers(id),
    response_time_ms INT NOT NULL,
    score INT NOT NULL DEFAULT 0,
    answered_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(participant_id, question_id)
);

CREATE INDEX idx_quizzes_creator ON quizzes(creator_id);
CREATE INDEX idx_questions_quiz ON questions(quiz_id);
CREATE INDEX idx_answers_question ON answers(question_id);
CREATE INDEX idx_session_room_code ON sessions(room_code);
CREATE INDEX idx_participants_session ON session_participants(session_id);
CREATE INDEX idx_user_answers_participant ON user_answers(participant_id);