package main

import (
	"fmt"
	"log"
	"time"

	"brainbolt/internal/database"
	"brainbolt/internal/handlers"
	"brainbolt/internal/repository"
	"brainbolt/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 1. Initialize our external connections
	database.InitDatabases()

	// 2. Initialize repos, services, and handlers
	userRepo := repository.NewUserRepository(database.DB)
	questionRepo := repository.NewQuestionRepository(database.DB)
	leaderboardRepo := repository.NewLeaderboardRepository(database.RedisClient)
	lastAnswerRepo := repository.NewLastAnswerRepository(database.RedisClient)
	userCacheRepo := repository.NewUserCacheRepository(database.RedisClient)

	userService := service.NewUserService(userRepo, userCacheRepo)
	questionService := service.NewQuestionService(questionRepo, userRepo, userService)
	answerService := service.NewAnswerService(userService, questionRepo, lastAnswerRepo, userRepo, leaderboardRepo, userCacheRepo)
	leaderboardService := service.NewLeaderboardService(userRepo, leaderboardRepo)

	quizHandlers := handlers.NewQuizHandlers(userService, questionService, answerService, leaderboardService)

	// 4. Create a new Fiber instance
	app := fiber.New(fiber.Config{
		AppName: "BrainBolt_v1",
	})

	// 5. Middleware for better observability
	app.Use(logger.New())  // Logs every request to console
	app.Use(recover.New()) // Prevents the app from crashing on panics

	// 5.1 Simple middleware to track app-side latency
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Microseconds()
		c.Response().Header.Set("X-App-Latency-US", fmt.Sprintf("%d", duration))
		return err
	})

	// 6. Route Definitions
	api := app.Group("/v1/quiz")
	// Per-user rate limiting: extract userId from body for POST /answer, then limit by user (or IP fallback)
	api.Use(handlers.BodyUserIDMiddleware)
	api.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return handlers.RateLimitKeyByUser(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many requests. Please try again later.",
			})
		},
	}))
	api.Get("/next", quizHandlers.HandleNextQuestion)
	api.Post("/answer", quizHandlers.HandleSubmitAnswer)
	api.Get("/metrics", quizHandlers.HandleGetMetrics)

	leaderboard := app.Group("/v1/leaderboard")
	leaderboard.Get("/score", quizHandlers.HandleGetScoreBoard)
	leaderboard.Get("/streak", quizHandlers.HandleGetStreakBoard)

	// 6. Start the server
	log.Fatal(app.Listen(":3001"))
}
