package service

import (
	"brainbolt/internal/repository"
)

// LeaderboardService handles leaderboard and rank logic (Redis with DB fallback).
type LeaderboardService struct {
	userRepo        *repository.UserRepository
	leaderboardRepo *repository.LeaderboardRepository
}

// NewLeaderboardService creates a new leaderboard service.
func NewLeaderboardService(userRepo *repository.UserRepository, leaderboardRepo *repository.LeaderboardRepository) *LeaderboardService {
	return &LeaderboardService{
		userRepo:        userRepo,
		leaderboardRepo: leaderboardRepo,
	}
}

// GetLeaderboardEntriesByScore returns leaderboard entries (userId, score, rank) from Redis; fallback to DB.
func (s *LeaderboardService) GetLeaderboardEntriesByScore(limit int) ([]repository.LeaderboardEntry, error) {
	entries, err := s.leaderboardRepo.GetTopByScore(int64(limit))
	if err != nil {
		users, err2 := s.userRepo.GetLeaderboardByScore(limit)
		if err2 != nil {
			return nil, err
		}
		entries = make([]repository.LeaderboardEntry, len(users))
		for i, u := range users {
			entries[i] = repository.LeaderboardEntry{UserID: u.ID, Score: u.Score, Rank: int64(i + 1)}
		}
		return entries, nil
	}
	return entries, nil
}

// GetLeaderboardEntriesByStreak returns streak leaderboard entries (userId, streak, rank).
func (s *LeaderboardService) GetLeaderboardEntriesByStreak(limit int) ([]repository.StreakLeaderboardEntry, error) {
	entries, err := s.leaderboardRepo.GetTopByStreak(int64(limit))
	if err != nil {
		users, err2 := s.userRepo.GetLeaderboardByStreak(limit)
		if err2 != nil {
			return nil, err
		}
		out := make([]repository.StreakLeaderboardEntry, len(users))
		for i, u := range users {
			out[i] = repository.StreakLeaderboardEntry{UserID: u.ID, Streak: u.MaxStreak, Rank: int64(i + 1)}
		}
		return out, nil
	}
	out := make([]repository.StreakLeaderboardEntry, len(entries))
	for i, e := range entries {
		out[i] = repository.StreakLeaderboardEntry{UserID: e.UserID, Streak: int(e.Score), Rank: e.Rank}
	}
	return out, nil
}

// GetUserRankByScore gets user's rank by score.
func (s *LeaderboardService) GetUserRankByScore(userID int) (int, error) {
	rank, err := s.leaderboardRepo.GetUserRankByScore(userID)
	if err != nil {
		return s.userRepo.GetUserRankByScore(userID)
	}
	if rank == 0 {
		return s.userRepo.GetUserRankByScore(userID)
	}
	return int(rank), nil
}

// GetUserRankByStreak gets user's rank by streak.
func (s *LeaderboardService) GetUserRankByStreak(userID int) (int, error) {
	rank, err := s.leaderboardRepo.GetUserRankByStreak(userID)
	if err != nil {
		return s.userRepo.GetUserRankByStreak(userID)
	}
	if rank == 0 {
		return s.userRepo.GetUserRankByStreak(userID)
	}
	return int(rank), nil
}
