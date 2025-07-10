package service_schedulers

import (
	writer_pipelines "go-write-behind-service/internal/services/scheduler/writer"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type JobWriter struct {
	Schedulers []SchedulerFunc
}

func (job *JobWriter) Use(schedulers ...SchedulerFunc) {
	job.Schedulers = append(job.Schedulers, schedulers...)
}

func (job *JobWriter) Run(every time.Duration) {
	if every.Nanoseconds() <= 0 {
		panic("job:writer -> harus terdapat rentang waktu eksekusi")
	}

	if len(job.Schedulers) == 0 {
		panic("job:writer -> tidak ada tugas yang dijadwalkan")
	}

	// status := make(chan string)
	// for range time.Tick(every) {
	//   for _, job := range job.Schedulers {
	//     go job()
	//   }
	// }
}

func RunPipelinesScheduler(every time.Duration) {
	for range time.Tick(every) {
		log.Println("Job: executing pipelines scheduler ...")
		errExecute := writer_pipelines.PipelinesWriteScheduler()
		if errExecute != nil {
			if errExecute == redis.Nil {
				log.Println("Tidak terdapat antrean job untuk data Pipelines")
				continue
			}
			log.Printf("Job Error: %v \n", errExecute.Error())
			continue
		}

		log.Println("job:writer@pipelines done!")
	}
}
