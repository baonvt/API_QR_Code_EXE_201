package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

// ConnectRedis kết nối đến Redis server
func ConnectRedis() {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")

	// Set default values
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "6379"
	}

	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	// Tạo Redis client
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	// Test connection
	_, err = RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Redis features will be disabled")
		RedisClient = nil
		return
	}

	log.Println("Redis connected successfully!")
}

// GetRedis trả về instance Redis client
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis đóng kết nối Redis
func CloseRedis() {
	if RedisClient != nil {
		err := RedisClient.Close()
		if err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		}
	}
}

// SetCache lưu dữ liệu vào cache với TTL
func SetCache(key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	return RedisClient.Set(Ctx, key, value, expiration).Err()
}

// GetCache lấy dữ liệu từ cache
func GetCache(key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("redis client is not initialized")
	}
	return RedisClient.Get(Ctx, key).Result()
}

// DeleteCache xóa dữ liệu từ cache
func DeleteCache(key string) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	return RedisClient.Del(Ctx, key).Err()
}

// ExistsCache kiểm tra key có tồn tại trong cache không
func ExistsCache(key string) (bool, error) {
	if RedisClient == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}
	result, err := RedisClient.Exists(Ctx, key).Result()
	return result > 0, err
}

// SetCacheJSON lưu dữ liệu JSON vào cache
func SetCacheJSON(key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	return RedisClient.Set(Ctx, key, value, expiration).Err()
}

// FlushDB xóa tất cả keys trong database hiện tại
func FlushDB() error {
	if RedisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	return RedisClient.FlushDB(Ctx).Err()
}
