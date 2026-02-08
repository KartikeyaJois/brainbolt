package models

import "time"

// Question represents a quiz question
type Question struct {
	ID         int
	Difficulty int
	Question   string
	Options    []string
	Answer     string
}

// User represents a user in the quiz system
type User struct {
	ID                int        `json:"id" db:"id"`
	Username          string     `json:"username" db:"username"`
	Score             int64      `json:"score" db:"score"`
	Streak            int        `json:"streak" db:"streak"`
	MaxStreak         int        `json:"maxStreak" db:"max_streak"`
	TotalCorrect      int        `json:"totalCorrect" db:"total_correct"`
	TotalAnswered     int        `json:"totalAnswered" db:"total_answered"`
	CurrentDifficulty int        `json:"currentDifficulty" db:"current_difficulty"`
	LastAnsweredAt    *time.Time `json:"lastAnsweredAt,omitempty" db:"last_answered_at"`
}
