package scheduler

import (
	"context"
	"errors"
	"fmt"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/models"
	"log"
	"strings"
	"sync"
	"time"
)

type VacancyProps struct {
	ID              string    `json:"id" redis:"id"`
	Position        string    `json:"position" redis:"position"`
	Description     string    `json:"description" redis:"description"`
	Qualification   string    `json:"qualification" redis:"qualification"`
	Responsibility  string    `json:"responsibility" redis:"responsibility"`
	LineIndustry    string    `json:"line_industry" redis:"line_industry"`
	EmployeeType    string    `json:"employee_type" redis:"employee_type"`
	MinExperience   string    `json:"min_experience" redis:"min_experience"`
	Salary          uint      `json:"salary" redis:"salary"`
	WorkArrangement string    `json:"work_arrangement" redis:"work_arrangement"`
	SLA             int32     `json:"sla" redis:"sla"`
	IsInactive      int       `json:"is_inactive" redis:"is_inactive"`
	EmployerID      string    `json:"employer_id" redis:"employer_id"`
	CreatedAt       time.Time `json:"created_at" redis:"created_at"`
}

var (
	Status string = "start"
	mu     sync.Mutex
)

func CacheDataFlusher() error {
	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		return errRdb
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		return errGorm
	}

	ctxRdb := context.Background()
	var cursor uint64
	for {
		var keys []string
		var errScan error
		keys, cursor, errScan = rdb.Scan(ctxRdb, cursor, "WB:*", 100).Result()
		if errScan != nil {
			return errScan
		}

		if len(keys) == 0 {
			log.Println("no action needed, 0 keys")
			return errors.New("nokeys")
		}

		// iterate over keys WB:*
		for _, key := range keys {

			// get all fields from {key}
			fields, errFields := rdb.HKeys(ctxRdb, key).Result()
			if errFields != nil {
				return errFields
			}

			// get ttl all fields from {key}
			ttl, errTTL := rdb.HTTL(ctxRdb, key, fields...).Result()
			if errTTL != nil {
				return errTTL
			}

			ttlCounter := 0
			updatedFields := []string{}
			for idxttl, seconds := range ttl {
				// if ttl (-1) no expiration, mark as new data
				if seconds == -1 {
					ttlCounter += 1
					updatedFields = append(updatedFields, fields[idxttl])
					continue
				}
			}

			if ttlCounter == len(fields) {
				var vacancy VacancyProps
				cmd := rdb.HGetAll(ctxRdb, key)
				if errScan := cmd.Scan(&vacancy); errScan != nil {
					return errScan
				}

				store := gormDB.Create(&models.Vacancy{
					Id:              vacancy.ID,
					Position:        vacancy.Position,
					Description:     vacancy.Description,
					Qualification:   vacancy.Qualification,
					Responsibility:  vacancy.Responsibility,
					LineIndustry:    vacancy.LineIndustry,
					EmployeeType:    vacancy.EmployeeType,
					MinExperience:   vacancy.MinExperience,
					Salary:          vacancy.Salary,
					WorkArrangement: vacancy.WorkArrangement,
					SLA:             vacancy.SLA,
					IsInactive:      vacancy.IsInactive != 0,
					EmployerId:      vacancy.EmployerID,
					CreatedAt:       vacancy.CreatedAt,
				})
				if store.Error != nil {
					return store.Error
				}

				_, errExp := rdb.HExpire(ctxRdb, key, 30*time.Minute, fields...).Result()
				if errExp != nil {
					return errExp
				}

				continue
			}

			if len(updatedFields) != 0 {
				values, errHMGet := rdb.HMGet(ctxRdb, key, updatedFields...).Result()
				if errHMGet != nil {
					return errHMGet
				}
				mappedColumns := map[string]interface{}{}
				for idxf, field := range updatedFields {
					mappedColumns[field] = values[idxf]
				}

				vacancyID := strings.TrimPrefix(key, "WB:")
				update := gormDB.Model(&models.Vacancy{Id: vacancyID}).Updates(mappedColumns)
				if update.Error != nil {
					return update.Error
				}
				if update.RowsAffected == 0 {
					log.Println("no columns were updated")
				}

				_, errHExp := rdb.HExpire(ctxRdb, key, 30*time.Minute, fields...).Result()
				if errHExp != nil {
					return errHExp
				}

				log.Println("columns updated \t:", len(values))
				continue
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

func StartSchedulerWriteBehind(every time.Duration) {
	for range time.Tick(every) {

		mu.Lock()
		errFlusher := CacheDataFlusher()
		if errFlusher != nil {
			Status = fmt.Sprintf("%v", errFlusher)
			mu.Unlock()
			continue
		}

		Status = "done"
		mu.Unlock()
	}
}

func GetFlusherStatus() string {
	mu.Lock()
	defer mu.Unlock()

	return Status
}
