package main

import (
	"database/sql"
	"log"
)

// QuizService handles business logic for the quiz system
type QuizService struct {
	userRepo        *UserRepository
	leaderboardRepo *LeaderboardRepository
	lastAnswerRepo  *LastAnswerRepository
}

// NewQuizService creates a new quiz service
func NewQuizService(userRepo *UserRepository, leaderboardRepo *LeaderboardRepository, lastAnswerRepo *LastAnswerRepository) *QuizService {
	return &QuizService{
		userRepo:        userRepo,
		leaderboardRepo: leaderboardRepo,
		lastAnswerRepo:  lastAnswerRepo,
	}
}

// AdjustDifficulty adjusts the difficulty based on whether the last answer was correct
// Returns the new difficulty within bounds (1-10)
func (s *QuizService) AdjustDifficulty(currentDifficulty int, lastAnswerCorrect bool) int {
	if lastAnswerCorrect {
		// Increase difficulty, but cap at 10
		if currentDifficulty < 10 {
			return currentDifficulty + 1
		}
		return 10
	}
	// Decrease difficulty, but don't go below 1
	if currentDifficulty > 1 {
		return currentDifficulty - 1
	}
	return 1
}

// GetOrCreateUser gets a user or creates a new one if they don't exist
func (s *QuizService) GetOrCreateUser(username string) (*User, error) {
	user, err := s.userRepo.GetUser(username)
	if err == nil {
		return user, nil
	}

	if err == sql.ErrNoRows {
		// User doesn't exist, create a new one
		return s.userRepo.CreateUser(username)
	}

	return nil, err
}

// GetNextQuestionForUser gets the next question for a user.
// Difficulty is already updated when the user submits an answer (SubmitAnswer).
func (s *QuizService) GetNextQuestionForUser(username string) (*Question, int, error) {
	// Get or create user
	user, err := s.GetOrCreateUser(username)
	if err != nil {
		return nil, 0, err
	}

	// Use stored difficulty (updated in SubmitAnswer); first question starts at 1
	currentDifficulty := user.CurrentDifficulty
	if currentDifficulty == 0 {
		currentDifficulty = 1
	}

	// Get the next question
	question := GetNextQuestionForUser(currentDifficulty)

	return &question, currentDifficulty, nil
}

// SubmitAnswer processes an answer submission and updates user stats
func (s *QuizService) SubmitAnswer(username string, questionID int, answer string) (bool, *User, error) {
	// Get user
	user, err := s.GetOrCreateUser(username)
	if err != nil {
		return false, nil, err
	}

	// Duplicate: same question as last answered — ignore (no processing, handler returns 204)
	if lastQ, found, err := s.lastAnswerRepo.GetLastAnsweredQuestionID(username); err == nil && found && lastQ == questionID {
		return false, nil, ErrDuplicateAnswer
	}
	// On Redis error we continue and process

	// Get the question to check the answer
	question, exists := QuestionByID[questionID]
	if !exists {
		return false, nil, ErrQuestionNotFound
	}
	correctQuestion := &question

	// Check if answer is correct
	isCorrect := correctQuestion.Answer == answer

	// Update user stats
	user.TotalAnswered++
	if isCorrect {
		user.TotalCorrect++
		user.Streak++
		if user.Streak > user.MaxStreak {
			user.MaxStreak = user.Streak
		}
	} else {
		user.Streak = 0
	}

	scoreDelta := int64(0)
	if isCorrect {
		// Base score = difficulty * 10, with streak multiplier (capped at 2x)
		baseScore := int64(correctQuestion.Difficulty * 10)
		streakMultiplier := 1.0
		if user.Streak > 0 {
			streakMultiplier = 1.0 + float64(user.Streak)*0.1
			if streakMultiplier > 2.0 {
				streakMultiplier = 2.0
			}
		}
		scoreDelta = int64(float64(baseScore) * streakMultiplier)
		user.Score += scoreDelta
	}

	// Adjust difficulty for next question
	newDifficulty := s.AdjustDifficulty(user.CurrentDifficulty, isCorrect)
	user.CurrentDifficulty = newDifficulty

	// Store last answer result
	user.LastAnswerCorrect = &isCorrect

	// Update user in database
	if err := s.userRepo.UpdateUserAfterAnswer(username, user); err != nil {
		return false, nil, err
	}

	// Store last answered question ID in Redis (duplicate submit of same question will be ignored)
	if err := s.lastAnswerRepo.SetLastAnsweredQuestionID(username, questionID); err != nil {
		log.Printf("Failed to set last answered question in Redis for %s: %v", username, err)
	}

	// Update Redis ZSets (async/non-blocking - don't fail if Redis is down)
	go func() {
		if err := s.leaderboardRepo.UpdateScore(username, user.Score); err != nil {
			// Log error but don't fail the request
			// In production, you might want to use a proper logger
		}
		if err := s.leaderboardRepo.UpdateStreak(username, user.MaxStreak); err != nil {
			// Log error but don't fail the request
		}
	}()

	return isCorrect, user, nil
}

// GetUserMetrics retrieves user metrics
func (s *QuizService) GetUserMetrics(username string) (*User, error) {
	return s.GetOrCreateUser(username)
}

// GetLeaderboardByScore gets the leaderboard by score
func (s *QuizService) GetLeaderboardByScore(limit int) ([]User, error) {
	entries, err := s.leaderboardRepo.GetTopByScore(int64(limit))
	if err != nil {
		// Fallback to database if Redis fails
		return s.userRepo.GetLeaderboardByScore(limit)
	}

	// Fetch full user data from DB for each entry
	users := make([]User, 0, len(entries))
	for _, entry := range entries {
		user, err := s.userRepo.GetUser(entry.Username)
		if err != nil {
			continue // Skip if user not found
		}
		users = append(users, *user)
	}

	return users, nil
}

// GetLeaderboardByStreak gets the leaderboard by streak
func (s *QuizService) GetLeaderboardByStreak(limit int) ([]User, error) {
	entries, err := s.leaderboardRepo.GetTopByStreak(int64(limit))
	if err != nil {
		// Fallback to database if Redis fails
		return s.userRepo.GetLeaderboardByStreak(limit)
	}

	users := make([]User, 0, len(entries))
	for _, entry := range entries {
		user, err := s.userRepo.GetUser(entry.Username)
		if err != nil {
			continue
		}
		users = append(users, *user)
	}

	return users, nil
}

// GetUserRankByScore gets user's rank by score
func (s *QuizService) GetUserRankByScore(username string) (int, error) {
	rank, err := s.leaderboardRepo.GetUserRankByScore(username)
	if err != nil {
		// Fallback to database if Redis fails
		return s.userRepo.GetUserRankByScore(username)
	}
	if rank == 0 {
		// User not in Redis, try database
		return s.userRepo.GetUserRankByScore(username)
	}
	return int(rank), nil
}

// GetUserRankByStreak gets user's rank by streak
func (s *QuizService) GetUserRankByStreak(username string) (int, error) {
	rank, err := s.leaderboardRepo.GetUserRankByStreak(username)
	if err != nil {
		// Fallback to database if Redis fails
		return s.userRepo.GetUserRankByStreak(username)
	}
	if rank == 0 {
		// User not in Redis, try database
		return s.userRepo.GetUserRankByStreak(username)
	}
	return int(rank), nil
}

// Custom errors
var (
	ErrQuestionNotFound = &Error{Message: "question not found"}
	ErrDuplicateAnswer  = &Error{Message: "duplicate answer"}
)

type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
