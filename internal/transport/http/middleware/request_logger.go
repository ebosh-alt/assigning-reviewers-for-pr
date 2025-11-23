// Package middleware contains HTTP middlewares for delivery.
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// RequestLogger logs HTTP requests with method, path, status and duration.
func RequestLogger(log *zap.SugaredLogger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		dur := time.Since(start)
		reqID, _ := c.Locals("requestid").(string)
		if reqID == "" {
			reqID = c.Get(fiber.HeaderXRequestID)
		}
		log.Infow("http",
			"method", c.Method(),
			"path", c.OriginalURL(),
			"status", c.Response().StatusCode(),
			"duration_ms", float64(dur.Microseconds())/1000.0,
			"request_id", reqID,
		)
		return err
	}
}
