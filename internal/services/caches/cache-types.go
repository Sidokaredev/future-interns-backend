package caches

type ExecutorCall func(data any, hash HashCollection, sortedset *SortedSetCollection) error
type ExecutorArgs struct {
}

type FallbackCall func(args FallbackArgs) (*FallbackReturn, error)
type FallbackReturn struct {
	Data    []map[string]any
	Indexes []string
}
type FallbackArgs struct {
	// vacancies.VacanciesArgs
	Offset    int
	FetchNext int
}

type CacheArgs struct {
	Indexes    []string
	CacheProps CacheProps
}
type CacheProps struct {
	KeyPropName    string
	ScorePropName  string
	ScoreType      string
	MemberPropName string
}
