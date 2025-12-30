package middlewares

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RateLimiterMiddleware is a middleware that limits the number of requests from a single IP address
// using a sliding window approach with Redis Sorted Sets. It enforces a rate limit of 'limit'
// requests per 'duration'.
//
// Parameters:
//   - redisClient (*redis.Client): Redis client to use for rate limiting
//   - limit int: Maximum number of requests allowed within the duration
//   - duration time.Duration: Time window for the rate limit
//   - logger (*logrus.Logger): Logger for logging rate limiting events
//
// Returns:
//   - gin.HandlerFunc: Gin middleware handler for rate limiting
func RateLimiterMiddleware(redisClient *redis.Client, limit int, duration time.Duration, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		realIP, exists := c.Get(RealIPKey)
		if !exists {
			logger.Warn("Real IP Middleware not Configured Correctly! Falling Back to RemoteAddr for Rate Limiting!")
			realIP = c.ClientIP()
		}
		ipAddress := realIP.(string)

		key := "rl:ip:" + ipAddress
		now := time.Now()
		windowStart := now.Add(-duration).UnixNano() / 1e6

		keyType, err := redisClient.Type(c.Request.Context(), key).Result()
		if err == nil && keyType != "zset" && keyType != "none" {
			err = redisClient.Del(c.Request.Context(), key).Err()
			if err != nil {
				logger.WithFields(logrus.Fields{
					"error": err,
					"ip":    ipAddress,
				}).Error("Failed to Delete Old Key Format!")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
				return
			}
		}

		member := strconv.FormatInt(now.UnixNano()/1e6, 10)

		err = redisClient.ZRemRangeByScore(c.Request.Context(), key, "0", strconv.FormatInt(windowStart, 10)).Err()
		if err != nil && err != redis.Nil {
			logger.WithFields(logrus.Fields{
				"error": err,
				"ip":    ipAddress,
			}).Error("Failed to Remove Old Entries!")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
			return
		}

		err = redisClient.ZAdd(c.Request.Context(), key, redis.Z{
			Score:  float64(now.UnixNano() / 1e6),
			Member: member,
		}).Err()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
				"ip":    ipAddress,
			}).Error("Failed to Add New Request!")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
			return
		}

		err = redisClient.Expire(c.Request.Context(), key, duration).Err()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
				"ip":    ipAddress,
			}).Error("Failed to Set Expiration!")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
			return
		}

		count, err := redisClient.ZCard(c.Request.Context(), key).Result()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
				"ip":    ipAddress,
			}).Error("Failed to Get Request Count!")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
			return
		}

		if count > int64(limit) {
			oldestRequest, err := redisClient.ZRange(c.Request.Context(), key, 0, 0).Result()
			if err != nil || len(oldestRequest) == 0 {
				logger.WithFields(logrus.Fields{
					"error": err,
					"ip":    ipAddress,
				}).Error("Failed to Get Oldest Request Timestamp!")
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error!"})
				return
			}

			oldestTimestamp, _ := strconv.ParseInt(oldestRequest[0], 10, 64)
			retryAfter := time.Duration(oldestTimestamp+duration.Milliseconds()-now.UnixNano()/1e6) * time.Millisecond

			c.Header("Retry-After", strconv.Itoa(int(retryAfter.Seconds())))
			logger.WithFields(logrus.Fields{
				"ip":       ipAddress,
				"limit":    limit,
				"duration": duration,
				"count":    count,
			}).Warn("Rate Limit Exceeded!")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too Many Requests!",
				"message": "Rate Limit Exceeded! Please Try Again Later!",
			})
			return
		}

		c.Next()
	}
}
