package initializer

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var (
	redisDB *redis.Client
)

func RedisServerInit() error {
	var redisconfig struct {
		Address  string
		Password string
		Database int
		Protocol int
	}

	if err := viper.UnmarshalKey("redis.dev", &redisconfig); err != nil {
		return err
	}

	redisDB = redis.NewClient(&redis.Options{
		Addr:     redisconfig.Address,
		Password: redisconfig.Password,
		DB:       redisconfig.Database,
		Protocol: redisconfig.Protocol,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := redisDB.Ping(ctx).Err(); err != nil {
		return err
	}

	log.Println("redis server connection established")
	return nil
}

func GetRedisDB() (*redis.Client, error) {
	if redisDB == nil {
		return nil, errors.New("redis Client instance empty")
	}

	return redisDB, nil
}
