package caches

type ExecutorCall func(data any) error

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

const JobPipelinesKey string = "job:writer@pipelines"
