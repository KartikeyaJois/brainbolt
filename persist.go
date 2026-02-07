package main

import (
	"database/sql"
)

// UserRepository handles all database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetUserByID retrieves a user by id
func (r *UserRepository) GetUserByID(id int) (*User, error) {
	var user User
	var lastAnswerCorrect sql.NullBool
	var lastAnsweredAt sql.NullTime
	query := `SELECT id, username, score, streak, max_streak, total_correct, total_answered, 
	          COALESCE(current_difficulty, 0) as current_difficulty, last_answer_correct, last_answered_at 
	          FROM users WHERE id = ?`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Score, &user.Streak, &user.MaxStreak,
		&user.TotalCorrect, &user.TotalAnswered, &user.CurrentDifficulty, &lastAnswerCorrect, &lastAnsweredAt,
	)

	if err != nil {
		return nil, err
	}

	if lastAnswerCorrect.Valid {
		user.LastAnswerCorrect = &lastAnswerCorrect.Bool
	}
	if lastAnsweredAt.Valid {
		user.LastAnsweredAt = &lastAnsweredAt.Time
	}

	return &user, nil
}

// CreateUser creates a new user with default values and returns the user with generated ID
func (r *UserRepository) CreateUser(username string) (*User, error) {
	insertQuery := `INSERT INTO users (username, score, streak, max_streak, total_correct, total_answered, current_difficulty) 
	                VALUES (?, 0, 0, 0, 0, 0, 1)`
	result, err := r.db.Exec(insertQuery, username)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:                int(id),
		Username:          username,
		Score:             0,
		Streak:            0,
		MaxStreak:         0,
		TotalCorrect:      0,
		TotalAnswered:     0,
		CurrentDifficulty: 1,
	}, nil
}

// UpdateUserDifficulty updates the user's current difficulty
func (r *UserRepository) UpdateUserDifficulty(userID int, difficulty int) error {
	query := `UPDATE users SET current_difficulty = ? WHERE id = ?`
	_, err := r.db.Exec(query, difficulty, userID)
	return err
}

// UpdateUserAfterAnswer updates user stats after answering a question
func (r *UserRepository) UpdateUserAfterAnswer(userID int, user *User) error {
	query := `UPDATE users SET 
	          score = ?, streak = ?, max_streak = ?, total_correct = ?, 
	          total_answered = ?, current_difficulty = ?, last_answer_correct = ?, last_answered_at = ? 
	          WHERE id = ?`

	var lastAnswerCorrect interface{}
	if user.LastAnswerCorrect != nil {
		lastAnswerCorrect = *user.LastAnswerCorrect
	} else {
		lastAnswerCorrect = nil
	}

	_, err := r.db.Exec(query, user.Score, user.Streak, user.MaxStreak,
		user.TotalCorrect, user.TotalAnswered, user.CurrentDifficulty,
		lastAnswerCorrect, user.LastAnsweredAt, userID)
	return err
}

// GetLeaderboardByScore returns top N users by score
func (r *UserRepository) GetLeaderboardByScore(limit int) ([]User, error) {
	query := `SELECT id, username, score, streak, max_streak, total_correct, total_answered, 
	          COALESCE(current_difficulty, 0) as current_difficulty, last_answer_correct, last_answered_at 
	          FROM users ORDER BY score DESC LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var lastAnswerCorrect sql.NullBool
		var lastAnsweredAt sql.NullTime
		err := rows.Scan(
			&user.ID, &user.Username, &user.Score, &user.Streak, &user.MaxStreak,
			&user.TotalCorrect, &user.TotalAnswered, &user.CurrentDifficulty, &lastAnswerCorrect, &lastAnsweredAt,
		)
		if err != nil {
			return nil, err
		}
		if lastAnswerCorrect.Valid {
			user.LastAnswerCorrect = &lastAnswerCorrect.Bool
		}
		if lastAnsweredAt.Valid {
			user.LastAnsweredAt = &lastAnsweredAt.Time
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetLeaderboardByStreak returns top N users by max streak
func (r *UserRepository) GetLeaderboardByStreak(limit int) ([]User, error) {
	query := `SELECT id, username, score, streak, max_streak, total_correct, total_answered, 
	          COALESCE(current_difficulty, 0) as current_difficulty, last_answer_correct, last_answered_at 
	          FROM users ORDER BY max_streak DESC LIMIT ?`

	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var lastAnswerCorrect sql.NullBool
		var lastAnsweredAt sql.NullTime
		err := rows.Scan(
			&user.ID, &user.Username, &user.Score, &user.Streak, &user.MaxStreak,
			&user.TotalCorrect, &user.TotalAnswered, &user.CurrentDifficulty, &lastAnswerCorrect, &lastAnsweredAt,
		)
		if err != nil {
			return nil, err
		}
		if lastAnswerCorrect.Valid {
			user.LastAnswerCorrect = &lastAnswerCorrect.Bool
		}
		if lastAnsweredAt.Valid {
			user.LastAnsweredAt = &lastAnsweredAt.Time
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetAskedQuestionIDs returns a set of question IDs that have been asked to a user
func (r *UserRepository) GetAskedQuestionIDs(userID int) (map[int]bool, error) {
	query := `SELECT question_id FROM user_questions WHERE user_id = ?`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	askedQuestions := make(map[int]bool)
	for rows.Next() {
		var questionID int
		if err := rows.Scan(&questionID); err != nil {
			return nil, err
		}
		askedQuestions[questionID] = true
	}

	return askedQuestions, rows.Err()
}

// RecordQuestionAsked records that a question was asked to a user
func (r *UserRepository) RecordQuestionAsked(userID int, questionID int) error {
	query := `INSERT IGNORE INTO user_questions (user_id, question_id) VALUES (?, ?)`
	_, err := r.db.Exec(query, userID, questionID)
	return err
}

// GetUserRankByScore returns the rank of a user by score (1-indexed)
func (r *UserRepository) GetUserRankByScore(userID int) (int, error) {
	query := `SELECT COUNT(*) + 1 FROM users WHERE score > (SELECT score FROM users WHERE id = ?)`
	var rank int
	err := r.db.QueryRow(query, userID).Scan(&rank)
	return rank, err
}

// GetUserRankByStreak returns the rank of a user by max streak (1-indexed)
func (r *UserRepository) GetUserRankByStreak(userID int) (int, error) {
	query := `SELECT COUNT(*) + 1 FROM users WHERE max_streak > (SELECT max_streak FROM users WHERE id = ?)`
	var rank int
	err := r.db.QueryRow(query, userID).Scan(&rank)
	return rank, err
}

// GetUsersByIDs retrieves multiple users by their IDs (for batch lookup)
func (r *UserRepository) GetUsersByIDs(ids []int) ([]User, error) {
	if len(ids) == 0 {
		return []User{}, nil
	}

	// Build placeholders for IN clause
	placeholders := ""
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args[i] = id
	}

	query := `SELECT id, username, score, streak, max_streak, total_correct, total_answered, 
	          COALESCE(current_difficulty, 0) as current_difficulty, last_answer_correct, last_answered_at 
	          FROM users WHERE id IN (` + placeholders + `)`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		var lastAnswerCorrect sql.NullBool
		var lastAnsweredAt sql.NullTime
		err := rows.Scan(
			&user.ID, &user.Username, &user.Score, &user.Streak, &user.MaxStreak,
			&user.TotalCorrect, &user.TotalAnswered, &user.CurrentDifficulty, &lastAnswerCorrect, &lastAnsweredAt,
		)
		if err != nil {
			return nil, err
		}
		if lastAnswerCorrect.Valid {
			user.LastAnswerCorrect = &lastAnswerCorrect.Bool
		}
		if lastAnsweredAt.Valid {
			user.LastAnsweredAt = &lastAnsweredAt.Time
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
