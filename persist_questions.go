package main

import (
	"database/sql"
	"encoding/json"
)

// QuestionRepository handles DB access for questions
type QuestionRepository struct {
	db *sql.DB
}

// NewQuestionRepository creates a new question repository
func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

// GetQuestionByID returns a question by ID, or nil if not found
func (r *QuestionRepository) GetQuestionByID(id int) (*Question, error) {
	var q Question
	var optionsJSON []byte
	query := `SELECT id, difficulty, question, options, answer FROM questions WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&q.ID, &q.Difficulty, &q.Question, &optionsJSON, &q.Answer)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
		return nil, err
	}
	return &q, nil
}

// GetRandomQuestionForUser returns one random question for the given difficulty that the user
// has not been asked yet. If all at that difficulty were asked, returns any random question at that difficulty.
// Uses a single join query so the question is returned directly without a second lookup.
func (r *QuestionRepository) GetRandomQuestionForUser(userID int, difficulty int) (*Question, error) {
	if difficulty < 1 {
		difficulty = 1
	}
	if difficulty > 10 {
		difficulty = 10
	}

	// Prefer questions not yet asked; fallback to any at this difficulty
	query := `SELECT q.id, q.difficulty, q.question, q.options, q.answer
	          FROM questions q
	          WHERE q.difficulty = ?
	          AND NOT EXISTS (SELECT 1 FROM user_questions uq WHERE uq.user_id = ? AND uq.question_id = q.id)
	          ORDER BY RAND()
	          LIMIT 1`
	var q Question
	var optionsJSON []byte
	err := r.db.QueryRow(query, difficulty, userID).Scan(&q.ID, &q.Difficulty, &q.Question, &optionsJSON, &q.Answer)
	if err == sql.ErrNoRows {
		// All asked at this difficulty: allow repeats
		queryRepeat := `SELECT id, difficulty, question, options, answer FROM questions WHERE difficulty = ? ORDER BY RAND() LIMIT 1`
		err = r.db.QueryRow(queryRepeat, difficulty).Scan(&q.ID, &q.Difficulty, &q.Question, &optionsJSON, &q.Answer)
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
		return nil, err
	}
	return &q, nil
}
