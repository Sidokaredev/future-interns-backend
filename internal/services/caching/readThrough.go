package caching

import (
	"context"
	"fmt"
	initializer "future-interns-backend/init"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type ReadThrough struct {
	Fallback     FallbackCall
	FallbackArgs []any
}

func NewReadThrough(fb FallbackCall, fbArgs []any) *ReadThrough {
	return &ReadThrough{
		Fallback:     fb,
		FallbackArgs: fbArgs,
	}
}

func (rt *ReadThrough) SetCache(args SetCacheArgs, data any) error {
	return nil
}

func (rt *ReadThrough) GetCache(args GetCacheArgs, dest *[]map[string]any) error {
	rdb, err := initializer.GetRedisDB()
	if err != nil {
		return err
	}

	ctx := context.Background()
	vacancyKeys, errKeys := rdb.ZRevRangeByScore(ctx, args.Intersection, &redis.ZRangeBy{
		Min:    args.Min,
		Max:    args.Max,
		Offset: args.Offset,
		Count:  args.Count,
	}).Result()
	if errKeys != nil {
		return errKeys
	}

	if len(vacancyKeys) == 0 {
		log.Println("calling fallback ...")

		fallback, errFallback := rt.Fallback(rt.FallbackArgs...)
		if errFallback != nil {
			return fmt.Errorf("fallback error: %v", errFallback.Error())
		}

		if len(fallback.Data) == 0 {
			log.Println(":fallback return empty data")
			*dest = []map[string]any{}
			return nil
		}

		hash := ExtractToHash(args.CacheArgs.KeyPropName, fallback.Data)
		sortedSet := NewSortedSetCollection(hash, SortedSetArgs{
			ScorePropName:  args.CacheArgs.ScorePropName,
			ScoreType:      args.CacheArgs.ScoreType,
			MemberPropName: args.CacheArgs.MemberPropName,
		})

		sortedSet.Keys = args.Indexes

		errZAdd := sortedSet.Add(1 * time.Hour)
		if errZAdd != nil {
			return errZAdd
		}

		errHSet := hash.Add(1 * time.Hour)
		if errHSet != nil {
			return errHSet
		}

		vacancyKeys, errKeys := rdb.ZRevRangeByScore(ctx, args.Intersection, &redis.ZRangeBy{
			Min:    args.Min,
			Max:    args.Max,
			Offset: args.Offset,
			Count:  args.Count,
		}).Result()
		if errKeys != nil {
			return errKeys
		}

		for _, key := range vacancyKeys {
			hash, errHash := rdb.HGetAll(ctx, key).Result()
			if errHash != nil {
				return errHash
			}
			nestedMap := TransformNestedMap(hash)
			*dest = append(*dest, nestedMap)
		}

	} else {
		for _, key := range vacancyKeys {
			hash, errHash := rdb.HGetAll(ctx, key).Result()
			if errHash != nil {
				return errHash
			}
			nestedMap := TransformNestedMap(hash)
			*dest = append(*dest, nestedMap)
		}

		return nil
	}

	return nil
}
