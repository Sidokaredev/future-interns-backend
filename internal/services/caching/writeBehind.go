package caching

import (
	"context"
	"fmt"
	initializer "future-interns-backend/init"
	"time"
)

type WriteBehind struct {
	Fallback     FallbackCall
	FallbackArgs []any
}

func NewWriteBehind(fb FallbackCall, fbArgs []any) *WriteBehind {
	return &WriteBehind{
		Fallback:     fb,
		FallbackArgs: fbArgs,
	}
}

func (wb *WriteBehind) SetCache(args SetCacheArgs, data any) error {
	hash := ExtractToHash(args.KeyPropName, data)
	sortedSet := NewSortedSetCollection(hash, SortedSetArgs{
		ScorePropName:  args.ScorePropName,
		ScoreType:      args.ScoreType,
		MemberPropName: args.MemberPropName,
	})
	sortedSet.Keys = args.Indexes

	rdb, err := initializer.GetRedisDB()
	if err != nil {
		return err
	}

	rdbCtx := context.Background()
	pipe := rdb.Pipeline()
	for _, h := range hash.Collection {
		pipe.LPush(rdbCtx, args.JobName, h.Key)
	}

	if cmds, errExec := pipe.Exec(rdbCtx); errExec != nil {
		for _, cmd := range cmds {
			fmt.Printf("redis query: %v | err: %v \n", cmd.Args(), cmd.Err())
		}
		return errExec
	}

	errZAdd := sortedSet.Add(0 * time.Nanosecond)
	if errZAdd != nil {
		return errZAdd
	}

	errHSet := hash.Add((0 * time.Nanosecond))
	if errHSet != nil {
		return errHSet
	}

	return nil
}

func (wb *WriteBehind) GetCache(args GetCacheArgs, dest *[]map[string]any) error {
	return nil
}
