package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

// QuizService handles business logic for the quiz system
type QuizService struct {
	userRepo        *UserRepository
	leaderboardRepo *LeaderboardRepository
	lastAnswerRepo  *LastAnswerRepository
	userCacheRepo   *UserCacheRepository
}

// NewQuizService creates a new quiz service
func NewQuizService(userRepo *UserRepository, leaderboardRepo *LeaderboardRepository, lastAnswerRepo *LastAnswerRepository, userCacheRepo *UserCacheRepository) *QuizService {
	return &QuizService{
		userRepo:        userRepo,
		leaderboardRepo: leaderboardRepo,
		lastAnswerRepo:  lastAnswerRepo,
		userCacheRepo:   userCacheRepo,
	}
}

// GetUserFromCacheOrDB returns the user from cache first, then DB; on DB hit populates cache. Use this wherever user data is needed.
func (s *QuizService) GetUserFromCacheOrDB(userID int) (*User, error) {
	if s.userCacheRepo != nil {
		if user, err := s.userCacheRepo.Get(userID); err == nil && user != nil {
			return user, nil
		}
	}
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if s.userCacheRepo != nil {
		_ = s.userCacheRepo.Set(userID, user)
	}
	return user, nil
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

// CalculateScore computes the score delta for a correct answer based on difficulty,
// current streak, and overall accuracy. Pass the already-updated user stats (after
// TotalAnswered, TotalCorrect, and Streak have been updated for this answer).
func (s *QuizService) CalculateScore(difficulty int, streak, totalCorrect, totalAnswered int) int64 {
	baseScore := int64(difficulty * 10)

	streakMultiplier := 1.0
	if streak > 0 {
		streakMultiplier = 1.0 + float64(streak)*0.1
		if streakMultiplier > 2.0 {
			streakMultiplier = 2.0
		}
	}

	accuracy := 0.0
	if totalAnswered > 0 {
		accuracy = float64(totalCorrect) / float64(totalAnswered)
	}
	accuracyMultiplier := 0.5 + (accuracy * 1.0)

	return int64(float64(baseScore) * streakMultiplier * accuracyMultiplier)
}

// GetUserByID gets a user by ID (cache first, then DB)
func (s *QuizService) GetUserByID(userID int) (*User, error) {
	return s.GetUserFromCacheOrDB(userID)
}

// GetNextQuestionForUser gets the next question for a user.
// Difficulty is already updated when the user submits an answer (SubmitAnswer).
// Ensures the question hasn't been asked to this user before.
func (s *QuizService) GetNextQuestionForUser(userID int) (*Question, int, error) {
	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, 0, err
	}

	// Use stored difficulty (updated in SubmitAnswer); first question starts at 1
	currentDifficulty := user.CurrentDifficulty
	if currentDifficulty == 0 {
		currentDifficulty = 1
	}

	// Get questions already asked to this user
	askedQuestions, err := s.userRepo.GetAskedQuestionIDs(userID)
	if err != nil {
		return nil, 0, err
	}

	// Get all questions for this difficulty level
	questions, exists := QuestionPool[currentDifficulty]
	if !exists || len(questions) == 0 {
		// Fallback to difficulty 1 if no questions at current difficulty
		questions = QuestionPool[1]
		if len(questions) == 0 {
			return nil, 0, ErrQuestionNotFound
		}
		currentDifficulty = 1
	}

	// Filter out questions already asked
	availableQuestions := make([]Question, 0)
	for _, q := range questions {
		if !askedQuestions[q.ID] {
			availableQuestions = append(availableQuestions, q)
		}
	}

	// If all questions at this difficulty have been asked, use all questions (allow repeats)
	if len(availableQuestions) == 0 {
		availableQuestions = questions
	}

	// Pick a random question from available ones
	selectedQuestion := availableQuestions[rand.Intn(len(availableQuestions))]

	// Record that this question was asked to this user
	if err := s.userRepo.RecordQuestionAsked(userID, selectedQuestion.ID); err != nil {
		log.Printf("Failed to record question asked for userID %d, questionID %d: %v", userID, selectedQuestion.ID, err)
		// Continue anyway - don't fail the request
	}

	return &selectedQuestion, currentDifficulty, nil
}

// SubmitAnswer processes an answer submission and updates user stats
func (s *QuizService) SubmitAnswer(userID int, questionID int, answer string) (bool, *User, error) {
	// Get user
	user, err := s.GetUserByID(userID)
	if err != nil {
		return false, nil, err
	}

	// Duplicate: same question as last answered — ignore (no processing, handler returns 204)
	if lastQ, found, err := s.lastAnswerRepo.GetLastAnsweredQuestionID(userID); err == nil && found && lastQ == questionID {
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
		scoreDelta = s.CalculateScore(correctQuestion.Difficulty, user.Streak, user.TotalCorrect, user.TotalAnswered)
		user.Score += scoreDelta
	}

	// Adjust difficulty for next question
	newDifficulty := s.AdjustDifficulty(user.CurrentDifficulty, isCorrect)
	user.CurrentDifficulty = newDifficulty

	// Store last answer result
	user.LastAnswerCorrect = &isCorrect
	now := time.Now()
	user.LastAnsweredAt = &now

	// Update user in database
	if err := s.userRepo.UpdateUserAfterAnswer(userID, user); err != nil {
		return false, nil, err
	}

	// Cache user after answer (TTL 1 day)
	if s.userCacheRepo != nil {
		_ = s.userCacheRepo.Set(userID, user)
	}

	// Store last answered question ID in Redis (duplicate submit of same question will be ignored)
	if err := s.lastAnswerRepo.SetLastAnsweredQuestionID(userID, questionID); err != nil {
		log.Printf("Failed to set last answered question in Redis for userID %d: %v", userID, err)
	}

	// Update Redis ZSets in a goroutine; wait only for this request's update (no global lock)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := s.leaderboardRepo.UpdateScore(userID, user.Score); err != nil {
			log.Printf("Leaderboard UpdateScore failed for userID %d: %v", userID, err)
		}
		if err := s.leaderboardRepo.UpdateStreak(userID, user.MaxStreak); err != nil {
			log.Printf("Leaderboard UpdateStreak failed for userID %d: %v", userID, err)
		}
	}()
	wg.Wait()

	return isCorrect, user, nil
}

// GetUserMetrics retrieves user metrics
func (s *QuizService) GetUserMetrics(userID int) (*User, error) {
	return s.GetUserByID(userID)
}

// GetLeaderboardByScore gets the leaderboard by score
func (s *QuizService) GetLeaderboardByScore(limit int) ([]User, error) {
	entries, err := s.leaderboardRepo.GetTopByScore(int64(limit))
	if err != nil {
		// Fallback to database if Redis fails
		return s.userRepo.GetLeaderboardByScore(limit)
	}

	// Batch fetch user IDs from entries
	userIDs := make([]int, len(entries))
	for i, entry := range entries {
		userIDs[i] = entry.UserID
	}

	// Fetch users from DB in batch
	users, err := s.userRepo.GetUsersByIDs(userIDs)
	if err != nil {
		return nil, err
	}

	// Cache the fetched users
	if s.userCacheRepo != nil {
		for i := range users {
			_ = s.userCacheRepo.Set(users[i].ID, &users[i])
		}
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

	// Batch fetch user IDs from entries
	userIDs := make([]int, len(entries))
	for i, entry := range entries {
		userIDs[i] = entry.UserID
	}

	// Fetch users from DB in batch
	users, err := s.userRepo.GetUsersByIDs(userIDs)
	if err != nil {
		return nil, err
	}

	// Cache the fetched users
	if s.userCacheRepo != nil {
		for i := range users {
			_ = s.userCacheRepo.Set(users[i].ID, &users[i])
		}
	}

	return users, nil
}

// GetUserRankByScore gets user's rank by score
func (s *QuizService) GetUserRankByScore(userID int) (int, error) {
	rank, err := s.leaderboardRepo.GetUserRankByScore(userID)
	if err != nil {
		// Fallback to database if Redis fails
		return s.userRepo.GetUserRankByScore(userID)
	}
	if rank == 0 {
		// User not in Redis, try database
		return s.userRepo.GetUserRankByScore(userID)
	}
	return int(rank), nil
}

// GetUserRankByStreak gets user's rank by streak
func (s *QuizService) GetUserRankByStreak(userID int) (int, error) {
	rank, err := s.leaderboardRepo.GetUserRankByStreak(userID)
	if err != nil {
		// Fallback to database if Redis fails
		return s.userRepo.GetUserRankByStreak(userID)
	}
	if rank == 0 {
		// User not in Redis, try database
		return s.userRepo.GetUserRankByStreak(userID)
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
