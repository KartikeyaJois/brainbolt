package main

import (
	"time"
)

// StreakDecayWindow is the time after which streak starts degrading (e.g. lose 1 per full window since last answer).
const StreakDecayWindow = 24 * time.Hour

// UserService handles user-related business logic (cache, streak decay, metrics).
type UserService struct {
	userRepo      *UserRepository
	userCacheRepo *UserCacheRepository
}

// NewUserService creates a new user service.
func NewUserService(userRepo *UserRepository, userCacheRepo *UserCacheRepository) *UserService {
	return &UserService{
		userRepo:      userRepo,
		userCacheRepo: userCacheRepo,
	}
}

// applyStreakDecay reduces user.Streak based on time since LastAnsweredAt (1 per full StreakDecayWindow). Returns true if streak was changed.
func (s *UserService) applyStreakDecay(user *User) bool {
	if user.LastAnsweredAt == nil {
		return false
	}
	elapsed := time.Since(*user.LastAnsweredAt)
	if elapsed < StreakDecayWindow {
		return false
	}
	periods := int(elapsed / StreakDecayWindow)
	newStreak := user.Streak - periods
	if newStreak < 0 {
		newStreak = 0
	}
	if newStreak == user.Streak {
		return false
	}
	user.Streak = newStreak
	return true
}

// GetUserFromCacheOrDB returns the user from cache first, then DB; on DB hit populates cache.
// Applies streak decay based on time since last answer; persists and updates cache if streak was decayed.
func (s *UserService) GetUserFromCacheOrDB(userID int) (*User, error) {
	if s.userCacheRepo != nil {
		if user, err := s.userCacheRepo.Get(userID); err == nil && user != nil {
			if s.applyStreakDecay(user) {
				_ = s.userRepo.UpdateUserStreak(userID, user.Streak)
				_ = s.userCacheRepo.Set(userID, user)
			}
			return user, nil
		}
	}
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if s.applyStreakDecay(user) {
		_ = s.userRepo.UpdateUserStreak(userID, user.Streak)
	}
	if s.userCacheRepo != nil {
		_ = s.userCacheRepo.Set(userID, user)
	}
	return user, nil
}

// GetUserByID gets a user by ID (cache first, then DB).
func (s *UserService) GetUserByID(userID int) (*User, error) {
	return s.GetUserFromCacheOrDB(userID)
}

// GetUserMetrics retrieves user metrics (same as GetUserByID).
func (s *UserService) GetUserMetrics(userID int) (*User, error) {
	return s.GetUserByID(userID)
}
