package audit

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a Gin middleware that logs every API request as an
// audit event. Health-check requests (GET /api/health) are skipped to
// reduce noise. Logging is performed asynchronously so it does not add
// latency to the response.
func Middleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip health checks
		if c.Request.Method == "GET" && c.FullPath() == "/api/health" {
			c.Next()
			return
		}

		start := time.Now()

		// Process the request
		c.Next()

		// Capture values for the goroutine (avoid races on gin.Context)
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		status := c.Writer.Status()
		ip := c.ClientIP()
		duration := time.Since(start)

		success := status >= 200 && status < 400
		var errMsg string
		if !success {
			// Capture first error from Gin context if available
			if len(c.Errors) > 0 {
				errMsg = c.Errors.String()
			}
		}

		// Determine action from HTTP method
		action := methodToAction(method)

		event := Event{
			Timestamp: start,
			EventType: "api.request",
			Action:    action,
			Resource:  path,
			IPAddress: ip,
			Success:   success,
			ErrorMsg:  errMsg,
			Details: map[string]interface{}{
				"method":      method,
				"status_code": status,
				"duration_ms": duration.Milliseconds(),
			},
		}

		// Log asynchronously to avoid slowing down the response
		go func() {
			if err := logger.Log(context.Background(), event); err != nil {
				log.Printf("audit middleware: failed to log event: %v", err)
			}
		}()
	}
}

func methodToAction(method string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return method
	}
}
