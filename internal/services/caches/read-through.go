package caches

import (
	"context"
	"fmt"
	initializer "go-read-through-service/init"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type ReadThrough struct {
	Fallback     FallbackCall
	FallbackArgs FallbackArgs
}

// -> membuat cache-aside instance baru
func NewReadThrough(fb FallbackCall, fbArgs FallbackArgs) *ReadThrough {
	return &ReadThrough{
		Fallback:     fb,
		FallbackArgs: fbArgs,
	}
}

// -> mengambil data dengan menerapkan pola 'cache-aside'
func (rt *ReadThrough) GetCache(args CacheArgs, dest *[]map[string]any) error {
	rdb, err := initializer.GetRedisDB()
	if err != nil {
		return err
	}

	log.Printf("Looking for [%v] ... \n", args.Intersection)
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
		fallback, errFallback := rt.Fallback(rt.FallbackArgs)
		if errFallback != nil {
			return fmt.Errorf("fallback error: %v", errFallback.Error())
		}

		if len(fallback.Data) == 0 {
			log.Println(":fallback return empty data")
			*dest = []map[string]any{}
			return nil
		}

		hash := ExtractToHash(args.CacheProps.KeyPropName, fallback.Data)
		sortedSet := NewSortedSetCollection(hash, SortedSetArgs{
			ScorePropName:  args.CacheProps.ScorePropName,
			ScoreType:      args.CacheProps.ScoreType,
			MemberPropName: args.CacheProps.MemberPropName,
		})
		sortedSet.Keys = fallback.Indexes

		errZAdd := sortedSet.Add(1 * time.Hour)
		if errZAdd != nil {
			return errZAdd
		}

		pipe := rdb.Pipeline()
		pipe.ZAddArgs(ctx, args.Intersection, redis.ZAddArgs{
			GT:      true,
			Members: sortedSet.Collection,
		})
		pipe.Expire(ctx, args.Intersection, 1*time.Hour)
		if cmds, errExec := pipe.Exec(ctx); errExec != nil {
			for _, cmd := range cmds {
				log.Printf("cmd: %v | args: %v | err: %v", cmd.FullName(), cmd.Args(), cmd.Err())
			}
			return errExec
		}

		errHSet := hash.Add(1 * time.Hour)
		if errHSet != nil {
			return errHSet
		}
		log.Printf("Indexes return from fallback [%v] ... \n", fallback.Indexes)
		log.Printf("Fallback looking for [%v] ... \n", args.Intersection)
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

		return nil
	}

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
