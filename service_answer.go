package main

import (
	"context"
	"log"
	"sync"
	"time"
)

// AnswerService handles answer submission business logic.
type AnswerService struct {
	userService     *UserService
	questionRepo    *QuestionRepository
	lastAnswerRepo  *LastAnswerRepository
	userRepo        *UserRepository
	leaderboardRepo *LeaderboardRepository
	userCacheRepo   *UserCacheRepository
}

// NewAnswerService creates a new answer service.
func NewAnswerService(
	userService *UserService,
	questionRepo *QuestionRepository,
	lastAnswerRepo *LastAnswerRepository,
	userRepo *UserRepository,
	leaderboardRepo *LeaderboardRepository,
	userCacheRepo *UserCacheRepository,
) *AnswerService {
	return &AnswerService{
		userService:     userService,
		questionRepo:    questionRepo,
		lastAnswerRepo:  lastAnswerRepo,
		userRepo:        userRepo,
		leaderboardRepo: leaderboardRepo,
		userCacheRepo:   userCacheRepo,
	}
}

// AdjustDifficulty adjusts the difficulty based on whether the last answer was correct (1-10).
func (s *AnswerService) AdjustDifficulty(currentDifficulty int, lastAnswerCorrect bool) int {
	if lastAnswerCorrect {
		if currentDifficulty < 10 {
			return currentDifficulty + 1
		}
		return 10
	}
	if currentDifficulty > 1 {
		return currentDifficulty - 1
	}
	return 1
}

// CalculateScore computes the score delta for a correct answer.
func (s *AnswerService) CalculateScore(difficulty int, streak, totalCorrect, totalAnswered int) int64 {
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

// SubmitAnswer processes an answer submission and updates user stats.
func (s *AnswerService) SubmitAnswer(userID int, questionID int, answer string) (bool, *User, error) {
	var user *User
	var userErr error
	var lastQ int
	var lastFound bool
	var lastErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		user, userErr = s.userService.GetUserByID(userID)
	}()
	go func() {
		defer wg.Done()
		lastQ, lastFound, lastErr = s.lastAnswerRepo.GetLastAnsweredQuestionID(userID)
	}()
	wg.Wait()
	if userErr != nil {
		return false, nil, userErr
	}

	if s.userService.applyStreakDecay(user) {
		_ = s.userRepo.UpdateUserStreak(userID, user.Streak)
		if s.userCacheRepo != nil {
			_ = s.userCacheRepo.Set(userID, user)
		}
	}

	if lastErr == nil && lastFound && lastQ == questionID {
		return false, nil, ErrDuplicateAnswer
	}

	question, err := s.questionRepo.GetQuestionByID(questionID)
	if err != nil || question == nil {
		return false, nil, ErrQuestionNotFound
	}

	isCorrect := question.Answer == answer

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

	if isCorrect {
		scoreDelta := s.CalculateScore(question.Difficulty, user.Streak, user.TotalCorrect, user.TotalAnswered)
		user.Score += scoreDelta
	}

	user.CurrentDifficulty = s.AdjustDifficulty(user.CurrentDifficulty, isCorrect)
	now := time.Now()
	user.LastAnsweredAt = &now

	if err := s.userRepo.UpdateUserAfterAnswer(userID, user); err != nil {
		return false, nil, err
	}

	pipe := s.leaderboardRepo.Pipeline()
	if s.userCacheRepo != nil {
		_ = s.userCacheRepo.QueueSet(pipe, userID, user)
	}
	s.lastAnswerRepo.QueueSetLastAnswered(pipe, userID, questionID)
	s.leaderboardRepo.QueueUpdateScore(pipe, userID, user.Score)
	s.leaderboardRepo.QueueUpdateStreak(pipe, userID, user.MaxStreak)
	if _, err := pipe.Exec(context.Background()); err != nil {
		log.Printf("Redis pipeline Exec failed for userID %d: %v", userID, err)
	}

	return isCorrect, user, nil
}
