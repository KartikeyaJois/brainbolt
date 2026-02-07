package main

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// answerBody is used only to extract userId from POST /answer body for rate limiting.
type answerBody struct {
	UserID int `json:"userId"`
}

// BodyUserIDMiddleware runs for POST requests and extracts userId from the JSON body,
// stores it in Locals so the rate limiter can key by user. Restores the body for the handler.
func BodyUserIDMiddleware(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Next()
	}
	body := c.Body()
	if len(body) == 0 {
		return c.Next()
	}
	var req answerBody
	if err := json.Unmarshal(body, &req); err != nil {
		return c.Next()
	}
	if req.UserID != 0 {
		c.Locals("rateLimitUserId", req.UserID)
	}
	// Restore body so the handler can parse it again
	c.Request().SetBody(body)
	return c.Next()
}

// RateLimitKeyByUser returns a key for the rate limiter: per-user when userId is available, else per-IP.
func RateLimitKeyByUser(c *fiber.Ctx) string {
	if uid := c.Locals("rateLimitUserId"); uid != nil {
		return "user:" + fmt.Sprint(uid)
	}
	if q := c.Query("userId"); q != "" {
		return "user:" + q
	}
	return c.IP()
}
