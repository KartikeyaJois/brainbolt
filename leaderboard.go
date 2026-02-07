package main

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const (
	LeaderboardScoreKey  = "leaderboard:score"
	LeaderboardStreakKey = "leaderboard:streak"
)

// LeaderboardEntry represents a leaderboard entry
type LeaderboardEntry struct {
	Username string `json:"username"`
	Score    int64  `json:"score"` // For streak leaderboard, this is max_streak
	Rank     int64  `json:"rank"`
}

// LeaderboardRepository handles Redis ZSet operations for leaderboards
type LeaderboardRepository struct {
	client *redis.Client
	ctx    context.Context
}

// NewLeaderboardRepository creates a new leaderboard repository
func NewLeaderboardRepository(client *redis.Client) *LeaderboardRepository {
	return &LeaderboardRepository{
		client: client,
		ctx:    context.Background(),
	}
}

// UpdateScore updates user's score in the score leaderboard ZSet
func (r *LeaderboardRepository) UpdateScore(username string, score int64) error {
	return r.client.ZAdd(r.ctx, LeaderboardScoreKey, redis.Z{
		Score:  float64(score), // ZSet uses float64 for scores
		Member: username,
	}).Err()
}

// UpdateStreak updates user's max streak in the streak leaderboard ZSet
func (r *LeaderboardRepository) UpdateStreak(username string, maxStreak int) error {
	return r.client.ZAdd(r.ctx, LeaderboardStreakKey, redis.Z{
		Score:  float64(maxStreak),
		Member: username,
	}).Err()
}

// GetTopByScore returns top N users by score
func (r *LeaderboardRepository) GetTopByScore(limit int64) ([]LeaderboardEntry, error) {
	// ZREVRANGE returns highest to lowest (descending order)
	results, err := r.client.ZRevRangeWithScores(r.ctx, LeaderboardScoreKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, len(results))
	for i, result := range results {
		entries[i] = LeaderboardEntry{
			Username: result.Member.(string),
			Score:    int64(result.Score),
			Rank:     int64(i) + 1, // 1-indexed rank
		}
	}

	return entries, nil
}

// GetTopByStreak returns top N users by max streak
func (r *LeaderboardRepository) GetTopByStreak(limit int64) ([]LeaderboardEntry, error) {
	results, err := r.client.ZRevRangeWithScores(r.ctx, LeaderboardStreakKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]LeaderboardEntry, len(results))
	for i, result := range results {
		entries[i] = LeaderboardEntry{
			Username: result.Member.(string),
			Score:    int64(result.Score), // This is actually max_streak
			Rank:     int64(i) + 1,
		}
	}

	return entries, nil
}

// GetUserRankByScore returns user's rank by score (1-indexed, 0 if not found)
func (r *LeaderboardRepository) GetUserRankByScore(username string) (int64, error) {
	// ZREVRANK returns 0-based rank (highest score = rank 0)
	rank, err := r.client.ZRevRank(r.ctx, LeaderboardScoreKey, username).Result()
	if err == redis.Nil {
		return 0, nil // User not in leaderboard
	}
	if err != nil {
		return 0, err
	}
	return rank + 1, nil // Convert to 1-indexed
}

// GetUserRankByStreak returns user's rank by streak (1-indexed, 0 if not found)
func (r *LeaderboardRepository) GetUserRankByStreak(username string) (int64, error) {
	rank, err := r.client.ZRevRank(r.ctx, LeaderboardStreakKey, username).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return rank + 1, nil
}

// GetUserScore gets user's current score from ZSet
func (r *LeaderboardRepository) GetUserScore(username string) (int64, error) {
	score, err := r.client.ZScore(r.ctx, LeaderboardScoreKey, username).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int64(score), nil
}

// GetUserStreak gets user's max streak from ZSet
func (r *LeaderboardRepository) GetUserStreak(username string) (int, error) {
	score, err := r.client.ZScore(r.ctx, LeaderboardStreakKey, username).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(score), nil
}
