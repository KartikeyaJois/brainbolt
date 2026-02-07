package main

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const (
	LeaderboardScoreKey  = "leaderboard:score"
	LeaderboardStreakKey = "leaderboard:streak"
)

// LeaderboardEntry represents a score leaderboard entry
type LeaderboardEntry struct {
	UserID int   `json:"userId"`
	Score  int64 `json:"score"`
	Rank   int64 `json:"rank"`
}

// StreakLeaderboardEntry represents a streak leaderboard entry (max_streak, not score)
type StreakLeaderboardEntry struct {
	UserID int   `json:"userId"`
	Streak int   `json:"streak"`
	Rank   int64 `json:"rank"`
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
func (r *LeaderboardRepository) UpdateScore(userID int, score int64) error {
	return r.client.ZAdd(r.ctx, LeaderboardScoreKey, redis.Z{
		Score:  float64(score),       // ZSet uses float64 for scores
		Member: strconv.Itoa(userID), // Store userID as stringified integer
	}).Err()
}

// UpdateStreak updates user's max streak in the streak leaderboard ZSet
func (r *LeaderboardRepository) UpdateStreak(userID int, maxStreak int) error {
	return r.client.ZAdd(r.ctx, LeaderboardStreakKey, redis.Z{
		Score:  float64(maxStreak),
		Member: strconv.Itoa(userID), // Store userID as stringified integer
	}).Err()
}

// Pipeline returns a new pipeline for batching Redis commands (one round-trip).
func (r *LeaderboardRepository) Pipeline() *redis.Pipeline {
	return r.client.Pipeline().(*redis.Pipeline)
}

// QueueUpdateScore queues ZADD for the score leaderboard; call Exec on the pipeline to run.
func (r *LeaderboardRepository) QueueUpdateScore(pipe *redis.Pipeline, userID int, score int64) {
	pipe.ZAdd(r.ctx, LeaderboardScoreKey, redis.Z{
		Score:  float64(score),
		Member: strconv.Itoa(userID),
	})
}

// QueueUpdateStreak queues ZADD for the streak leaderboard; call Exec on the pipeline to run.
func (r *LeaderboardRepository) QueueUpdateStreak(pipe *redis.Pipeline, userID int, maxStreak int) {
	pipe.ZAdd(r.ctx, LeaderboardStreakKey, redis.Z{
		Score:  float64(maxStreak),
		Member: strconv.Itoa(userID),
	})
}

// GetTopByScore returns top N users by score
func (r *LeaderboardRepository) GetTopByScore(limit int64) ([]LeaderboardEntry, error) {
	// ZREVRANGE returns highest to lowest (descending order)
	results, err := r.client.ZRevRangeWithScores(r.ctx, LeaderboardScoreKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	var entries []LeaderboardEntry
	for _, result := range results {
		userIDStr, ok := result.Member.(string)
		if !ok {
			continue
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			continue
		}
		entries = append(entries, LeaderboardEntry{
			UserID: userID,
			Score:  int64(result.Score),
			Rank:   0, // set below
		})
	}
	for i := range entries {
		entries[i].Rank = int64(i) + 1
	}
	return entries, nil
}

// GetTopByStreak returns top N users by max streak
func (r *LeaderboardRepository) GetTopByStreak(limit int64) ([]LeaderboardEntry, error) {
	results, err := r.client.ZRevRangeWithScores(r.ctx, LeaderboardStreakKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	var entries []LeaderboardEntry
	for _, result := range results {
		userIDStr, ok := result.Member.(string)
		if !ok {
			continue
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			continue
		}
		entries = append(entries, LeaderboardEntry{
			UserID: userID,
			Score:  int64(result.Score), // This is actually max_streak
			Rank:   0,                   // set below
		})
	}
	for i := range entries {
		entries[i].Rank = int64(i) + 1
	}
	return entries, nil
}

// GetUserRankByScore returns user's rank by score (1-indexed, 0 if not found)
func (r *LeaderboardRepository) GetUserRankByScore(userID int) (int64, error) {
	// ZREVRANK returns 0-based rank (highest score = rank 0)
	rank, err := r.client.ZRevRank(r.ctx, LeaderboardScoreKey, strconv.Itoa(userID)).Result()
	if err == redis.Nil {
		return 0, nil // User not in leaderboard
	}
	if err != nil {
		return 0, err
	}
	return rank + 1, nil // Convert to 1-indexed
}

// GetUserRankByStreak returns user's rank by streak (1-indexed, 0 if not found)
func (r *LeaderboardRepository) GetUserRankByStreak(userID int) (int64, error) {
	rank, err := r.client.ZRevRank(r.ctx, LeaderboardStreakKey, strconv.Itoa(userID)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return rank + 1, nil
}

// GetUserScore gets user's current score from ZSet
func (r *LeaderboardRepository) GetUserScore(userID int) (int64, error) {
	score, err := r.client.ZScore(r.ctx, LeaderboardScoreKey, strconv.Itoa(userID)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int64(score), nil
}

// GetUserStreak gets user's max streak from ZSet
func (r *LeaderboardRepository) GetUserStreak(userID int) (int, error) {
	score, err := r.client.ZScore(r.ctx, LeaderboardStreakKey, strconv.Itoa(userID)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return int(score), nil
}
