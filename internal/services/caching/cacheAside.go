package caching

import (
	"context"
	"fmt"
	initializer "future-interns-backend/init"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheAside struct {
	Fallback     FallbackCall
	FallbackArgs []any
}

func NewCacheAside(fb FallbackCall, fbArgs []any) *CacheAside {
	return &CacheAside{
		Fallback:     fb,
		FallbackArgs: fbArgs,
	}
}

func (ca *CacheAside) SetCache(args SetCacheArgs, data any) error {
	return nil
}

func (ca *CacheAside) GetCache(args GetCacheArgs, dest *[]map[string]any) error {
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

		fallback, errFallback := ca.Fallback(ca.FallbackArgs...)
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

		*dest = fallback.Data

		return nil
	}

	// should have a fallback here

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

func TransformNestedMap(flatMap map[string]string) map[string]any {
	nestedMap := make(map[string]any)

	for key, val := range flatMap {
		parts := strings.Split(key, ".")
		currentMap := nestedMap

		for i, part := range parts {
			if i == (len(parts) - 1) {
				num, errNum := strconv.Atoi(val)
				if errNum != nil {
					currentMap[part] = val
					continue
				}

				currentMap[part] = num
			} else {
				if _, ok := currentMap[part]; !ok {
					currentMap[part] = make(map[string]any)
				}

				currentMap = currentMap[part].(map[string]any)
			}
		}
	}

	return nestedMap
}
