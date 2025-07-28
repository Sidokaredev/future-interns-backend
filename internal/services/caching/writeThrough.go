package caching

import "time"

type WriteThrough struct {
	Fallback     FallbackCall
	FallbackArgs []any
}

func NewWriteThrough(fb FallbackCall, fbArgs []any) *WriteThrough {
	return &WriteThrough{
		Fallback:     fb,
		FallbackArgs: fbArgs,
	}
}

func (wt *WriteThrough) SetCache(args SetCacheArgs, data any) error {
	hash := ExtractToHash(args.KeyPropName, data)
	sortedSet := NewSortedSetCollection(hash, SortedSetArgs{
		ScorePropName:  args.ScorePropName,
		ScoreType:      args.ScoreType,
		MemberPropName: args.MemberPropName,
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

	_, errFallback := wt.Fallback(data)
	if errFallback != nil {
		return errFallback
	}

	return nil
}

func (wt *WriteThrough) GetCache(args GetCacheArgs, dest *[]map[string]any) error {
	return nil
}
