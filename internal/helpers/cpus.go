package helpers

import (
	"bufio"
	"errors"
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

func GetCPUStatUsage() (float64, error) {
	const microToMilli = 1_000
	cpuStatFile, errCpuStatFile := os.Open("/sys/fs/cgroup/cpu.stat")
	if errCpuStatFile != nil {
		return 0, errCpuStatFile
	}

	defer cpuStatFile.Close()

	scanner := bufio.NewScanner(cpuStatFile)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		log.Printf("line : %s", line)

		parts := strings.Fields(line)
		if len(parts) != 2 {
			log.Println("string has less than 2 parts!")
			continue
		}

		if parts[0] == "usage_usec" {
			cpuUsageInMicroseconds, errParseFloat := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
			if errParseFloat != nil {
				return 0, errParseFloat
			}

			return (cpuUsageInMicroseconds / microToMilli), nil
		}
	}

	if errScan := scanner.Err(); errScan != nil {
		return 0, nil
	}

	return 0, errors.New("cannot find 'usage_usec' value")
}

func CalcCPUTimeUsage(elapsedTime time.Duration, elapsedCpuTime float64) (CPUs, error) {
	const microToMilli = 1_000

	cpuMax, errCpuMax := os.ReadFile("/sys/fs/cgroup/cpu.max")
	if errCpuMax != nil {
		return CPUs{}, errCpuMax
	}

	parts := strings.Fields(string(cpuMax))
	if len(parts) != 2 {
		return CPUs{}, errors.New("invalid value format for cpu.stat file")
	}

	if strings.TrimSpace(parts[0]) == "max" {
		return CPUs{}, errors.New("machine has no cpu quota")
	}

	cpuQuotaValue, errQuotaValue := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if errQuotaValue != nil {
		log.Printf("err parse cpu quota: %s", parts[0])
		return CPUs{}, errQuotaValue
	}

	cpuPeriodValue, errPeriodValue := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if errPeriodValue != nil {
		log.Printf("err parse cpu period: %s", parts[1])
		return CPUs{}, errPeriodValue
	}

	cpuQuotaInMs := cpuQuotaValue / microToMilli
	cpuPeriodInMs := cpuPeriodValue / microToMilli

	maxCpuTime := (float64(elapsedTime.Milliseconds()) / float64(cpuPeriodInMs)) * float64(cpuQuotaInMs)

	totalCpuTimeInPercent := (elapsedCpuTime / maxCpuTime) * 100
	log.Printf("time: %vms | cpu time usage: %.2f | cpu quota ms: %.2f | cpu period ms: %.2f | max cpu ms: %.2f | percent: %v", elapsedTime.Milliseconds(), elapsedCpuTime, cpuQuotaInMs, cpuPeriodInMs, maxCpuTime, ((totalCpuTimeInPercent * 100) / 100))

	return CPUs{
		Period:       int(cpuPeriodInMs),
		Quota:        int(cpuQuotaInMs),
		AcctUsage:    elapsedCpuTime,
		MaxCPUTime:   maxCpuTime,
		TotalCPUTime: totalCpuTimeInPercent,
		Elapsed:      elapsedTime.Milliseconds(),
	}, nil
}
