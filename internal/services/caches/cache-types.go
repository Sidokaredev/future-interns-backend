package caches

import "go-cache-aside-service/internal/services/vacancies"

type FallbackCall func(args FallbackArgs) (*FallbackReturn, error)
type FallbackReturn struct {
	Data    []map[string]any
	Indexes []string
}
type FallbackArgs struct {
	vacancies.VacanciesArgs
	Offset    int
	FetchNext int
}

type CacheArgs struct {
	Indexes      []string
	CacheProps   CacheProps
	Intersection string
	Min, Max     string
	Offset       int64
	Count        int64
}
type CacheProps struct {
	KeyPropName    string
	ScorePropName  string
	ScoreType      string
	MemberPropName string
}
