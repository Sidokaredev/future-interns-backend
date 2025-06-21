package helpers

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type Memories struct {
	Limit    int64
	MaxUsage int64
	Usage    int64
	Percent  float64
}

func GetMemoryUsage() (Memories, error) {
	memLimit, errMemLimit := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes")
	if errMemLimit != nil {
		log.Printf("err mem limit: %s", errMemLimit.Error())
		return Memories{}, errMemLimit
	}
	memLimitValue, errMemLimitValue := strconv.Atoi(strings.TrimSpace(string(memLimit)))
	if errMemLimitValue != nil {
		log.Printf("err convert mem limit: %s", errMemLimitValue.Error())
		return Memories{}, errMemLimitValue
	}

	maxMemUsage, errMaxMemUsage := os.ReadFile("/sys/fs/cgroup/memory/memory.max_usage_in_bytes")
	if errMaxMemUsage != nil {
		log.Printf("err max mem usage: %s", errMaxMemUsage.Error())
		return Memories{}, errMaxMemUsage
	}
	maxMemUsageValue, errMaxMemUsageValue := strconv.Atoi(strings.TrimSpace(string(maxMemUsage)))
	if errMaxMemUsageValue != nil {
		log.Printf("err convert max mem usage: %s", errMaxMemUsageValue.Error())
		return Memories{}, errMaxMemUsageValue
	}

	memUsage, errMemUsage := os.ReadFile("/sys/fs/cgroup/memory/memory.usage_in_bytes")
	if errMemUsage != nil {
		log.Printf("err mem usage: %s", errMemUsage)
		return Memories{}, errMemUsage
	}
	memUsageValue, errMemUsageValue := strconv.Atoi(strings.TrimSpace(string(memUsage)))
	if errMemUsageValue != nil {
		log.Printf("err convert mem usage: %s", errMemUsageValue.Error())
		return Memories{}, nil
	}

	return Memories{
		Limit:    int64(memLimitValue),
		MaxUsage: int64(maxMemUsageValue),
		Usage:    int64(memUsageValue),
		Percent:  (float64(memUsageValue) / float64(memLimitValue)) * 100,
	}, nil
} // cgroup v1

func CalcMemUsage() (Memories, error) {
	memLimit, errMemLimit := os.ReadFile("/sys/fs/cgroup/memory.max")
	if errMemLimit != nil {
		return Memories{}, errMemLimit
	}

	memLimitValue, errMemLimitValue := strconv.ParseFloat(strings.TrimSpace(string(memLimit)), 64)
	if errMemLimitValue != nil {
		log.Printf("err parse mem limit: %v", string(memLimit))
		return Memories{}, errMemLimitValue
	}

	memCurrent, errMemCurrent := os.ReadFile("/sys/fs/cgroup/memory.current")
	if errMemCurrent != nil {
		return Memories{}, errMemCurrent
	}

	memCurrentValue, errMemCurrentValue := strconv.ParseFloat(strings.TrimSpace(string(memCurrent)), 64)
	if errMemCurrentValue != nil {
		log.Printf("err parse mem current: %v", string(memCurrent))
		return Memories{}, errMemCurrentValue
	}

	memUsageInPercent := (memCurrentValue / memLimitValue) * 100
	// log.Printf("mem limit: %vB | mem usage: %vB | percent: %v", memLimitValue, memCurrentValue, memUsageInPercent)

	return Memories{
		Limit:    int64(memLimitValue),
		MaxUsage: 0,
		Usage:    int64(memCurrentValue),
		Percent:  memUsageInPercent,
	}, nil
} // cgroup v2
