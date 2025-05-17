package helpers

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type CPUs struct {
	Period       int
	Quota        int
	AcctUsage    float64
	MaxCPUTime   float64
	TotalCPUTime float64
	Elapsed      int64
}

func CPUTimeUsage(elapsedTime time.Duration, elapsedCpuTime float64) (CPUs, error) {
	const microsecondToMillisecond int = 1000

	cpuPeriod, errCpuPeriod := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
	if errCpuPeriod != nil {
		log.Printf("err opening cfs_period_us %s", errCpuPeriod.Error())
		return CPUs{}, errCpuPeriod
	}
	cpuPeriodValue, errCpuPeriodValue := strconv.Atoi(strings.TrimSpace(string(cpuPeriod)))
	if errCpuPeriodValue != nil {
		log.Printf("err converting cpu period to int %s", errCpuPeriodValue.Error())
		return CPUs{}, errCpuPeriodValue
	}

	cpuQuota, errCpuQuota := os.ReadFile("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
	if errCpuQuota != nil {
		log.Printf("err opening cfs_quota_us %s", errCpuQuota.Error())
		return CPUs{}, errCpuQuota
	}
	cpuQuotaValue, errCpuQuotaValue := strconv.Atoi(strings.TrimSpace(string(cpuQuota)))
	if errCpuQuotaValue != nil {
		log.Printf("err converting cpu quota to int %s", errCpuQuotaValue.Error())
		return CPUs{}, errCpuQuotaValue
	}

	cpuPeriodInMs := cpuPeriodValue / microsecondToMillisecond
	cpuQuotaInMs := cpuQuotaValue / microsecondToMillisecond

	maxCpuTime := (float64(elapsedTime.Milliseconds()) / float64(cpuPeriodInMs)) * float64(cpuQuotaInMs)
	log.Printf("elapsedTime: %v | period ms: %v | quota ms: %v | elapsedCpuTime: %v | MAX CPU TIME: %v", elapsedTime.Milliseconds(), cpuPeriodInMs, cpuQuotaInMs, elapsedCpuTime, maxCpuTime)

	totalCpuTimeUsageInPercent := (elapsedCpuTime / maxCpuTime) * 100

	/* menggunakan rasio perbandingan
	cpuLimit := float64(cpuQuotaValue) / float64(cpuPeriodValue)
	cpuUsageRatio := (elapsedCpuTime * 1000_000) / float64(elapsedTime)

	cpuTotalInUsage := (cpuUsageRatio / cpuLimit) * 100
	log.Printf("cpu usage: %v | cpu limit: %v | cpu total usage: %v", cpuUsageRatio, cpuLimit, cpuTotalInUsage)
	*/

	return CPUs{
		Period:       cpuPeriodInMs,
		Quota:        cpuQuotaInMs,
		AcctUsage:    elapsedCpuTime,
		MaxCPUTime:   maxCpuTime,
		TotalCPUTime: totalCpuTimeUsageInPercent,
		Elapsed:      elapsedTime.Milliseconds(),
	}, nil
}

func GetCPUAcctUsage() (float64, error) {
	const nanoToMilli float64 = 1000_000
	cpuAcctUsage, errCpuAcctUsage := os.ReadFile("/sys/fs/cgroup/cpuacct/cpuacct.usage")
	if errCpuAcctUsage != nil {
		log.Printf("err opening cpuacct.usage %s", errCpuAcctUsage.Error())
		return 0, errCpuAcctUsage
	}
	cpuAcctUsageValue, errCpuAcctUsageValue := strconv.Atoi(strings.TrimSpace(string(cpuAcctUsage)))
	if errCpuAcctUsageValue != nil {
		log.Printf("err converting cpuacct.usage to int %s", errCpuAcctUsageValue.Error())
		return 0, errCpuAcctUsageValue
	}

	return float64(cpuAcctUsageValue) / nanoToMilli, nil
}
