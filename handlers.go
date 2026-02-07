package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// QuizHandlers contains HTTP handlers for quiz endpoints
type QuizHandlers struct {
	service *QuizService
}

// NewQuizHandlers creates a new quiz handlers instance
func NewQuizHandlers(service *QuizService) *QuizHandlers {
	return &QuizHandlers{service: service}
}

// HandleNextQuestion handles GET /v1/quiz/next
// Query params: userId (required)
func (h *QuizHandlers) HandleNextQuestion(c *fiber.Ctx) error {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId query parameter is required",
		})
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId must be a valid integer",
		})
	}

	question, currentDifficulty, err := h.service.GetNextQuestionForUser(userID)
	if err != nil {
		log.Printf("Error getting next question for userID %d: %v", userID, err)
		// Check if it's a user not found error
		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("User with ID %d not found", userID),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to get next question",
			"details": errMsg,
		})
	}

	return c.JSON(fiber.Map{
		"questionId":        question.ID,
		"difficulty":        question.Difficulty,
		"question":          question.Question,
		"options":           question.Options,
		"currentDifficulty": currentDifficulty,
		"userId":            userID,
	})
}

// HandleSubmitAnswer handles POST /v1/quiz/answer
func (h *QuizHandlers) HandleSubmitAnswer(c *fiber.Ctx) error {
	var req struct {
		UserID     int    `json:"userId"`
		QuestionID int    `json:"questionId"`
		Answer     string `json:"answer"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == 0 || req.QuestionID == 0 || req.Answer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId, questionId, and answer are required",
		})
	}

	isCorrect, user, err := h.service.SubmitAnswer(req.UserID, req.QuestionID, req.Answer)
	if err != nil {
		if err == ErrDuplicateAnswer {
			return c.SendStatus(fiber.StatusNoContent) // duplicate — ignore, no body
		}
		log.Printf("Error submitting answer: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to submit answer",
		})
	}

	// Get user ranks in parallel
	var scoreRank, streakRank int
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		scoreRank, _ = h.service.GetUserRankByScore(req.UserID)
	}()
	go func() {
		defer wg.Done()
		streakRank, _ = h.service.GetUserRankByStreak(req.UserID)
	}()
	wg.Wait()

	return c.JSON(fiber.Map{
		"correct":               isCorrect,
		"newDifficulty":         user.CurrentDifficulty,
		"newStreak":             user.Streak,
		"totalScore":            user.Score,
		"leaderboardRankScore":  scoreRank,
		"leaderboardRankStreak": streakRank,
	})
}

// HandleGetMetrics handles GET /v1/quiz/metrics
func (h *QuizHandlers) HandleGetMetrics(c *fiber.Ctx) error {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId query parameter is required",
		})
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "userId must be a valid integer",
		})
	}

	user, err := h.service.GetUserMetrics(userID)
	if err != nil {
		log.Printf("Error getting user metrics: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user metrics",
		})
	}

	// Calculate accuracy
	accuracy := 0.0
	if user.TotalAnswered > 0 {
		accuracy = float64(user.TotalCorrect) / float64(user.TotalAnswered) * 100
	}

	return c.JSON(fiber.Map{
		"currentDifficulty": user.CurrentDifficulty,
		"streak":            user.Streak,
		"maxStreak":         user.MaxStreak,
		"totalScore":        user.Score,
		"accuracy":          accuracy,
		"totalCorrect":      user.TotalCorrect,
		"totalAnswered":     user.TotalAnswered,
	})
}

// HandleGetScoreBoard handles GET /v1/leaderboard/score
func (h *QuizHandlers) HandleGetScoreBoard(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Cap at 100
	}

	entries, err := h.service.GetLeaderboardEntriesByScore(limit)
	if err != nil {
		log.Printf("Error getting score leaderboard: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get leaderboard",
		})
	}

	return c.JSON(entries)
}

// HandleGetStreakBoard handles GET /v1/leaderboard/streak
func (h *QuizHandlers) HandleGetStreakBoard(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Cap at 100
	}

	entries, err := h.service.GetLeaderboardEntriesByStreak(limit)
	if err != nil {
		log.Printf("Error getting streak leaderboard: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get leaderboard",
		})
	}

	return c.JSON(entries)
}
