package middleware

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"
	"huydq2801.com/golearn/internal/utils"
)

type Client struct {
	limiter  *rate.Limiter
	lastseen time.Time
}

var (
	clients = make(map[string]*Client)
	mu      sync.Mutex
)

func getClientIP(ctx *gin.Context) string {
	ip := ctx.ClientIP()
	if ip == "" {
		ip = ctx.Request.RemoteAddr
	}
	return ip
}

// hey -n 20 -c 1 -H "X-API-Key: " http://localhost:8080/api/v1/users/
func CleanUpClients() {
	for {
		time.Sleep(time.Minute)
		mu.Lock()
		for ip, client := range clients {
			if time.Since(client.lastseen) > 3*time.Minute {
				delete(clients, ip)
				log.Printf("Removed client %s", ip)
			}
		}
		mu.Unlock()
	}
}
func getRateLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	client, exists := clients[ip]
	if !exists {
		requestSecStr := utils.GetEnv("RATE_LIMITER_REQUEST_SEC", "5")
		requestBurstStr := utils.GetEnv("RATE_LIMITER_REQUEST_BURST", "10")
		requestSec, err := strconv.Atoi(requestSecStr)
		if err != nil {
			panic("Failed to parse requestSec: " + err.Error())
		}
		requestBurst, err := strconv.Atoi(requestBurstStr)
		if err != nil {
			panic("Failed to parse requestBurst: " + err.Error())
		}
		limiter := rate.NewLimiter(rate.Limit(requestSec), requestBurst)
		newClient := &Client{limiter, time.Now()}
		clients[ip] = newClient
		return limiter
	}
	client.lastseen = time.Now()
	return client.limiter
}
func RateLimiterMiddleware(rateLimiterLogger *zerolog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := getClientIP(ctx)
		limiter := getRateLimiter(ip)
		if !limiter.Allow() {
			if shouldLogRateLimit(ip) {
				rateLimiterLogger.Warn().
					Str("method", ctx.Request.Method).
					Str("path", ctx.Request.URL.Path).
					Str("query", ctx.Request.URL.RawQuery).
					Str("client_ip", ctx.ClientIP()).
					Str("user_agent", ctx.Request.UserAgent()).
					Str("referer", ctx.Request.Referer()).
					Str("protocol", ctx.Request.Proto).
					Str("host", ctx.Request.Host).
					Str("remote_addr", ctx.Request.RemoteAddr).
					Str("request_uri", ctx.Request.RequestURI).
					Interface("headers", ctx.Request.Header).
					Msg("rate limiter exceeded")
					ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
						"error":   "Too many request",
						"message": "Bạn có quá nhiều request",
					})
					return
			}
		}
		ctx.Next()
	}
}

var rateLimitLogCache = sync.Map{}

const rateLimitLogTTL = 10 * time.Second

func shouldLogRateLimit(ip string) bool {
	now := time.Now()
	if val, ok := rateLimitLogCache.Load(ip); ok {
		if time.Since(val.(time.Time)) < rateLimitLogTTL {
			return false
		}
	}
	rateLimitLogCache.Store(ip, now)
	return true
}
