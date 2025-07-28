package caching

type ICacheStrategy interface {
	// SetCache(args SetCacheArgs, data []map[string]any) error
	SetCache(args SetCacheArgs, data any) error
	GetCache(args GetCacheArgs, dest *[]map[string]any) error
}

type FallbackReturnValues struct {
	Data    []map[string]any
	Indexes []string
}

type FallbackCall func(params ...any) (*FallbackReturnValues, error)

type CacheProps struct {
	KeyPropName    string
	ScorePropName  string
	ScoreType      string
	MemberPropName string
}

type GetCacheArgs struct {
	Indexes      []string
	CacheArgs    CacheProps
	Intersection string
	Min, Max     string
	Offset       int64
	Count        int64
}

type SetCacheArgs struct {
	KeyPropName    string
	ScorePropName  string
	ScoreType      string
	MemberPropName string
	Indexes        []string
	JobName        string
}

func NewCacheStrategy(cacheType string, fallback FallbackCall, fallbackArgs []any) ICacheStrategy {
	if cacheType == "cache-aside" {
		return NewCacheAside(fallback, fallbackArgs)
	}

	if cacheType == "read-through" {
		return NewReadThrough(fallback, fallbackArgs)
	}

	if cacheType == "write-through" {
		return NewWriteThrough(fallback, fallbackArgs)
	}

	if cacheType == "write-behind" {
		return NewWriteBehind(fallback, fallbackArgs)
	}

	return nil
}
