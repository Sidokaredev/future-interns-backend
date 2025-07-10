package caches

import "log"

type WriteBehind struct {
	ExecutorFunc ExecutorCall
}

func NewWriteBehind(exec ExecutorCall) *WriteBehind {
	return &WriteBehind{
		ExecutorFunc: exec,
	}
}

func (wb *WriteBehind) SetCache(data any, args CacheArgs) error {
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
		return nil
	}

	log.Println("Cache: executing Executor Func ...")
	errExecutorCall := wb.ExecutorFunc(data)
	if errExecutorCall != nil {
		return errExecutorCall
	}

	return nil
}
