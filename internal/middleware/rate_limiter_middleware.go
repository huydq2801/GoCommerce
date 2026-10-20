package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
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
		limiter := rate.NewLimiter(5, 10)
		newClient := &Client{limiter, time.Now()}
		clients[ip] = newClient
		return limiter
	}
	client.lastseen = time.Now()
	return client.limiter
}
func RateLimiterMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ip := getClientIP(ctx)
		limiter := getRateLimiter(ip)
		if !limiter.Allow() {
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "To mant request",
				"message": "Bạn có quá nhiều request",
			})
			return
		}
		ctx.Next()
	}
}
