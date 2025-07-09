package caches

import "log"

type WriteThrough struct {
	ExecutorFunc ExecutorCall
}

func NewWriteThrough(exec ExecutorCall) *WriteThrough {
	return &WriteThrough{
		ExecutorFunc: exec,
	}
}

func (wt *WriteThrough) SetCache(data any, args CacheArgs) error {
	hash := ExtractToHash(args.CacheProps.KeyPropName, data)
	sortedSet := NewSortedSetCollection(hash, SortedSetArgs{
		ScorePropName:  args.CacheProps.ScorePropName,
		ScoreType:      args.CacheProps.ScoreType,
		MemberPropName: args.CacheProps.MemberPropName,
	})
	sortedSet.Keys = args.Indexes

	log.Printf("Cache: adding sortedset into indexes -> %v", args.Indexes)
	errZAdd := sortedSet.Add()
	if errZAdd != nil {
		return errZAdd
	}

	errHSet := hash.Add()
	if errHSet != nil {
		return errHSet
	}

	log.Println("Cache: executing Executor Func ...")
	errExecutor := wt.ExecutorFunc(data, hash, sortedSet)
	if errExecutor != nil {
		return errExecutor
	}

	return nil
}
