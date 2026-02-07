package main

import (
	"log"
	"strconv"

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
// Query params: username (required)
func (h *QuizHandlers) HandleNextQuestion(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username query parameter is required",
		})
	}

	question, currentDifficulty, err := h.service.GetNextQuestionForUser(username)
	if err != nil {
		log.Printf("Error getting next question: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get next question",
		})
	}

	return c.JSON(fiber.Map{
		"questionId":        question.ID,
		"difficulty":        question.Difficulty,
		"question":          question.Question,
		"options":           question.Options,
		"currentDifficulty": currentDifficulty,
		"username":          username,
	})
}

// HandleSubmitAnswer handles POST /v1/quiz/answer
func (h *QuizHandlers) HandleSubmitAnswer(c *fiber.Ctx) error {
	var req struct {
		Username   string `json:"username"`
		QuestionID int    `json:"questionId"`
		Answer     string `json:"answer"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Username == "" || req.QuestionID == 0 || req.Answer == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username, questionId, and answer are required",
		})
	}

	isCorrect, user, err := h.service.SubmitAnswer(req.Username, req.QuestionID, req.Answer)
	if err != nil {
		if err == ErrDuplicateAnswer {
			return c.SendStatus(fiber.StatusNoContent) // duplicate — ignore, no body
		}
		log.Printf("Error submitting answer: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to submit answer",
		})
	}

	// Get user ranks
	scoreRank, _ := h.service.GetUserRankByScore(req.Username)
	streakRank, _ := h.service.GetUserRankByStreak(req.Username)

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
	username := c.Query("username")
	if username == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "username query parameter is required",
		})
	}

	user, err := h.service.GetUserMetrics(username)
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

	users, err := h.service.GetLeaderboardByScore(limit)
	if err != nil {
		log.Printf("Error getting score leaderboard: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get leaderboard",
		})
	}

	return c.JSON(fiber.Map{
		"leaderboard": users,
	})
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

	users, err := h.service.GetLeaderboardByStreak(limit)
	if err != nil {
		log.Printf("Error getting streak leaderboard: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get leaderboard",
		})
	}

	return c.JSON(fiber.Map{
		"leaderboard": users,
	})
}
