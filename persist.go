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

// GetUser retrieves a user by username
func (r *UserRepository) GetUser(username string) (*User, error) {
	var user User
	var lastAnswerCorrect sql.NullBool
	query := `SELECT username, score, streak, max_streak, total_correct, total_answered, 
	          COALESCE(current_difficulty, 0) as current_difficulty, last_answer_correct 
	          FROM users WHERE username = ?`

	err := r.db.QueryRow(query, username).Scan(
		&user.Username, &user.Score, &user.Streak, &user.MaxStreak,
		&user.TotalCorrect, &user.TotalAnswered, &user.CurrentDifficulty, &lastAnswerCorrect,
	)

	if err != nil {
		return nil, err
	}

	if lastAnswerCorrect.Valid {
		user.LastAnswerCorrect = &lastAnswerCorrect.Bool
	}

	return &user, nil
}

// CreateUser creates a new user with default values
func (r *UserRepository) CreateUser(username string) (*User, error) {
	insertQuery := `INSERT INTO users (username, score, streak, max_streak, total_correct, total_answered, current_difficulty) 
	                VALUES (?, 0, 0, 0, 0, 0, 1)`
	_, err := r.db.Exec(insertQuery, username)
	if err != nil {
		return nil, err
	}

	return &User{
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
func (r *UserRepository) UpdateUserDifficulty(username string, difficulty int) error {
	query := `UPDATE users SET current_difficulty = ? WHERE username = ?`
	_, err := r.db.Exec(query, difficulty, username)
	return err
}

// UpdateUserAfterAnswer updates user stats after answering a question
func (r *UserRepository) UpdateUserAfterAnswer(username string, user *User) error {
	query := `UPDATE users SET 
	          score = ?, streak = ?, max_streak = ?, total_correct = ?, 
	          total_answered = ?, current_difficulty = ?, last_answer_correct = ? 
	          WHERE username = ?`

	var lastAnswerCorrect interface{}
	if user.LastAnswerCorrect != nil {
		lastAnswerCorrect = *user.LastAnswerCorrect
	} else {
		lastAnswerCorrect = nil
	}

	_, err := r.db.Exec(query, user.Score, user.Streak, user.MaxStreak,
		user.TotalCorrect, user.TotalAnswered, user.CurrentDifficulty,
		lastAnswerCorrect, username)
	return err
}

// GetLeaderboardByScore returns top N users by score
func (r *UserRepository) GetLeaderboardByScore(limit int) ([]User, error) {
	query := `SELECT username, score, streak, max_streak, total_correct, total_answered, 
	          COALESCE(current_difficulty, 0) as current_difficulty, last_answer_correct 
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
		err := rows.Scan(
			&user.Username, &user.Score, &user.Streak, &user.MaxStreak,
			&user.TotalCorrect, &user.TotalAnswered, &user.CurrentDifficulty, &lastAnswerCorrect,
		)
		if err != nil {
			return nil, err
		}
		if lastAnswerCorrect.Valid {
			user.LastAnswerCorrect = &lastAnswerCorrect.Bool
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetLeaderboardByStreak returns top N users by max streak
func (r *UserRepository) GetLeaderboardByStreak(limit int) ([]User, error) {
	query := `SELECT username, score, streak, max_streak, total_correct, total_answered, 
	          COALESCE(current_difficulty, 0) as current_difficulty, last_answer_correct 
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
		err := rows.Scan(
			&user.Username, &user.Score, &user.Streak, &user.MaxStreak,
			&user.TotalCorrect, &user.TotalAnswered, &user.CurrentDifficulty, &lastAnswerCorrect,
		)
		if err != nil {
			return nil, err
		}
		if lastAnswerCorrect.Valid {
			user.LastAnswerCorrect = &lastAnswerCorrect.Bool
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetUserRankByScore returns the rank of a user by score (1-indexed)
func (r *UserRepository) GetUserRankByScore(username string) (int, error) {
	query := `SELECT COUNT(*) + 1 FROM users WHERE score > (SELECT score FROM users WHERE username = ?)`
	var rank int
	err := r.db.QueryRow(query, username).Scan(&rank)
	return rank, err
}

// GetUserRankByStreak returns the rank of a user by max streak (1-indexed)
func (r *UserRepository) GetUserRankByStreak(username string) (int, error) {
	query := `SELECT COUNT(*) + 1 FROM users WHERE max_streak > (SELECT max_streak FROM users WHERE username = ?)`
	var rank int
	err := r.db.QueryRow(query, username).Scan(&rank)
	return rank, err
}
