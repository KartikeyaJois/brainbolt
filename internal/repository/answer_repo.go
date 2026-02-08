package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	lastAnswerKeyPrefix = "user:last_answer:"
	lastAnswerTTL       = 5 * time.Second
)

// LastAnswerRepository persists "last answered question" per user in Redis (duplicate submit = ignore).
type LastAnswerRepository struct {
	client *redis.Client
	ctx    context.Context
}

// NewLastAnswerRepository creates a new last-answer repository.
func NewLastAnswerRepository(client *redis.Client) *LastAnswerRepository {
	return &LastAnswerRepository{
		client: client,
		ctx:    context.Background(),
	}
}

// GetLastAnsweredQuestionID returns the last answered question ID for the user, or (0, false, nil) if not set.
func (r *LastAnswerRepository) GetLastAnsweredQuestionID(userID int) (questionID int, found bool, err error) {
	key := lastAnswerKeyPrefix + strconv.Itoa(userID)
	q, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	id, err := strconv.Atoi(q)
	if err != nil {
		return 0, false, nil // invalid value, treat as not set
	}
	return id, true, nil
}

// SetLastAnsweredQuestionID stores the last answered question ID for the user (single key with TTL).
func (r *LastAnswerRepository) SetLastAnsweredQuestionID(userID int, questionID int) error {
	key := lastAnswerKeyPrefix + strconv.Itoa(userID)
	return r.client.Set(r.ctx, key, strconv.Itoa(questionID), lastAnswerTTL).Err()
}

// QueueSetLastAnswered queues SET for the last-answered question; call Exec on the pipeline to run.
func (r *LastAnswerRepository) QueueSetLastAnswered(pipe *redis.Pipeline, userID int, questionID int) {
	key := lastAnswerKeyPrefix + strconv.Itoa(userID)
	pipe.Set(r.ctx, key, strconv.Itoa(questionID), lastAnswerTTL)
}
