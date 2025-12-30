package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisAdapter struct.
// This struct is used to represent a Redis adapter.
//
// Attributes:
//   - Client (*redis.Client): The Redis client.
//   - logger (*logrus.Logger): The logger.
type RedisAdapter struct {
	Client *redis.Client
	logger *logrus.Logger
}

// NewRedisAdapter function.
// This function is used to create a new Redis adapter.
//
// Parameters:
//   - redisURI (string): The Redis URI.
//   - logger (*logrus.Logger): The logger.
//
// Returns:
//   - *RedisAdapter: The Redis adapter.
//   - error: The error.
func NewRedisAdapter(redisURI string, logger *logrus.Logger) (*RedisAdapter, error) {
	opt, err := redis.ParseURL(redisURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URI: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis!")

	return &RedisAdapter{
		Client: client,
		logger: logger,
	}, nil
}

// Set function.
// This function is used to set a key-value pair in Redis.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - key (string): The key.
//   - value (interface{}): The value.
//   - expiration (time.Duration): The expiration time.
//
// Returns:
//   - error: The error.
func (r *RedisAdapter) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	_, err := r.Client.Set(ctx, key, value, expiration).Result()
	if err != nil {
		return fmt.Errorf("failed to set key '%s' in Redis: %w", key, err)
	}

	r.logger.WithFields(logrus.Fields{"key": key, "value": value}).Info("Set Key in Redis!")
	return nil
}

// Get function.
// This function is used to get a value from Redis.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - key (string): The key.
//
// Returns:
//   - string: The value.
//   - error: The error.
func (r *RedisAdapter) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", errors.New("key not found in redis")
		}
		return "", fmt.Errorf("failed to get key '%s' from Redis: %w", key, err)
	}

	r.logger.WithFields(logrus.Fields{"key": key, "value": val}).Info("Get Key from Redis!")
	return val, nil
}

// Delete function.
// This function is used to delete a key from Redis.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - key (string): The key.
//
// Returns:
//   - error: The error.
func (r *RedisAdapter) Delete(ctx context.Context, key string) error {
	_, err := r.Client.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to delete key '%s' from Redis: %w", key, err)
	}

	r.logger.WithFields(logrus.Fields{"key": key}).Info("Delete Key from Redis!")
	return nil
}

// Close function.
// This function is used to close the Redis connection.
//
// Returns:
//   - error: The error.
func (r *RedisAdapter) Close() error {
	err := r.Client.Close()
	if err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}
	return nil
}
