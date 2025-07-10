package service_schedulers

import "time"

type IJobScheduler interface {
	Use() error
}

type SchedulerFunc func() (<-chan string, error)

type SchedulerMember struct {
	SchedulerCall SchedulerFunc
	Every         time.Time
}
