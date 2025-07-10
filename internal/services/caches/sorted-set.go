package caches

import (
	"context"
	initializer "go-write-behind-service/init"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type SortedSetCollection struct {
	Collection []redis.Z `json:"collection"`
	Min        string    `json:"min"`
	Max        string    `json:"max"`
	Offset     int64     `json:"offset"`
	Count      int64     `json:"count"`
	Keys       []string  `json:"keys"`
	// Intersection string    `json:"intersection"`
}

type SortedSetArgs struct {
	ScorePropName  string
	ScoreType      string // [int32, int64, string-time, time.Time]
	MemberPropName string
}

func NewSortedSetCollection(Hash HashCollection, args SortedSetArgs) *SortedSetCollection {
	sortedSet := &SortedSetCollection{}

	for _, hash := range Hash.Collection {
		member := redis.Z{Member: hash.Key}
		for index, field := range hash.FieldValues {
			if field == args.ScorePropName {
				switch args.ScoreType {
				case "int32", "int64":
					scoreAsInt := hash.FieldValues[index+1].(int)
					member.Score = float64(scoreAsInt)

				case "string-time":
					tm, err := time.Parse(time.RFC3339, hash.FieldValues[index+1].(string))
					if err != nil {
						panic(err.Error())
					}
					scoreAsString := tm.UnixNano()
					member.Score = float64(scoreAsString)

				case "time.Time":
					scoreAsTime := hash.FieldValues[index+1].(time.Time).UnixNano()
					member.Score = float64(scoreAsTime)
				}
			}
		}

		sortedSet.Collection = append(sortedSet.Collection, member)
	}

	return sortedSet
}

/*
set Collection(redis.Z) as members on every Keys
*/
func (ss *SortedSetCollection) Add() error {
	rdb, err := initializer.GetRedisDB()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pipe := rdb.Pipeline()
	for _, key := range ss.Keys {
		pipe.ZAddArgs(ctx, key, redis.ZAddArgs{
			GT:      true,
			Members: ss.Collection,
		})
	}

	if cmds, errExec := pipe.Exec(ctx); errExec != nil {
		for _, cmd := range cmds {
			log.Printf("cmd: %v | args: %v | err: %v", cmd.FullName(), cmd.Args(), cmd.Err())
		}
		return errExec
	}

	return nil
}

func (ss *SortedSetCollection) SetExpireByKeys(ttl time.Duration) error {
	rdb, err := initializer.GetRedisDB()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pipe := rdb.Pipeline()
	for _, key := range ss.Keys {
		if ttl.Nanoseconds() != 0 {
			pipe.Expire(ctx, key, ttl)
		}
	}

	if cmds, errExec := pipe.Exec(ctx); errExec != nil {
		for _, cmd := range cmds {
			log.Printf("cmd: %v | args: %v | err: %v", cmd.FullName(), cmd.Args(), cmd.Err())
		}
		return errExec
	}
	return nil
}
