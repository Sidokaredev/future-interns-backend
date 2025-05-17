package initializer

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
	// once sync.Once
)

func InitRedisClient() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
		Protocol: 3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed connect to redis: %v", err)
	} else {
		log.Println("redis connection established")
	}
	// once.Do(func() {
	// })
}

func GetRedis() *redis.Client {
	if rdb == nil {
		log.Fatal("redis has not been initialized")
	}

	return rdb
}
