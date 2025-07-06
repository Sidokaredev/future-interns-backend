package caching

type ICacheStrategy interface {
	SetCache(args SetCacheArgs, data []map[string]any) error
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
}

func NewCacheStrategy(caceType string, fallback FallbackCall, fallbackArgs []any) ICacheStrategy {
	if caceType == "cache-aside" {
		return NewCacheAside(fallback, fallbackArgs)
	}

	if caceType == "read-through" {
		return NewReadThrough(fallback, fallbackArgs)
	}

	return nil
}
