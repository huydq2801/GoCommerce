package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"

	"huydq2801.com/golearn/internal/utils"
)

type client struct {
	limiter           *rate.Limiter
	lastSeen, lastLog time.Time
}
type limiterStore struct {
	mu               sync.Mutex
	clients          map[string]*client
	nextCleanup      time.Time
	perSecond, burst int
	now              func() time.Time
}

func RateLimiterMiddleware(logger *zerolog.Logger, perSecond, burst int) gin.HandlerFunc {
	store := &limiterStore{clients: make(map[string]*client), perSecond: perSecond, burst: burst, now: time.Now}
	return func(c *gin.Context) {
		allowed, logRejection := store.allow(c.ClientIP())
		if !allowed {
			if logRejection {
				logger.Warn().Str("request_id", c.GetString("request_id")).Msg("rate_limit_exceeded")
			}
			c.Header("Retry-After", "1")
			utils.ResponseError(c, utils.NewError("Too many requests", utils.ErrCodeTooManyRequests))
			c.Abort()
			return
		}
		c.Next()
	}
}
func (s *limiterStore) allow(ip string) (bool, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if !now.Before(s.nextCleanup) {
		for key, c := range s.clients {
			if now.Sub(c.lastSeen) > 3*time.Minute {
				delete(s.clients, key)
			}
		}
		s.nextCleanup = now.Add(time.Minute)
	}
	c := s.clients[ip]
	if c == nil {
		// Bound memory even when many distinct source addresses appear.
		if len(s.clients) >= 10000 {
			return false, false
		}
		c = &client{limiter: rate.NewLimiter(rate.Limit(s.perSecond), s.burst)}
		s.clients[ip] = c
	}
	c.lastSeen = now
	if c.limiter.AllowN(now, 1) {
		return true, false
	}
	logIt := c.lastLog.IsZero() || now.Sub(c.lastLog) >= 10*time.Second
	if logIt {
		c.lastLog = now
	}
	return false, logIt
}
