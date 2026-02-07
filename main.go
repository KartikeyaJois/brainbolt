package main

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 1. Initialize our external connections
	InitDatabases()

	// 2. Load our questions into memory for fast access
	SeedQuestions()

	// 3. Initialize handlers (service and repository layers)
	userRepo := NewUserRepository(DB)
	leaderboardRepo := NewLeaderboardRepository(RedisClient)
	lastAnswerRepo := NewLastAnswerRepository(RedisClient)
	userCacheRepo := NewUserCacheRepository(RedisClient)
	quizService := NewQuizService(userRepo, leaderboardRepo, lastAnswerRepo, userCacheRepo)
	quizHandlers := NewQuizHandlers(quizService)

	// 4. Create a new Fiber instance
	app := fiber.New(fiber.Config{
		AppName: "BrainBolt_v1",
	})

	// 5. Middleware for better observability
	app.Use(logger.New())  // Logs every request to console
	app.Use(recover.New()) // Prevents the app from crashing on panics

	// 6. Route Definitions
	api := app.Group("/v1/quiz")
	// Per-user rate limiting: extract userId from body for POST /answer, then limit by user (or IP fallback)
	api.Use(BodyUserIDMiddleware)
	api.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return RateLimitKeyByUser(c)
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
