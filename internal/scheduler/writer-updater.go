package scheduler

import (
	"context"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/models"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type WriteVacancyProps struct {
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

func StartWriterJob(every time.Duration) {
	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		panic(errRdb)
	}
	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		panic(errGorm)
	}

	for range time.Tick(every) {
		errWriterJob := DataWriter(rdb, gormDB)
		if errWriterJob != nil {
			log.Println("err writer \t:", errWriterJob.Error())
			continue
		}
		log.Println("job:writer executed!")
	}
}
func StartUpdaterJob(every time.Duration) {
	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		panic(errRdb)
	}
	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		panic(errGorm)
	}

	for range time.Tick(every) {
		errUpdaterJob := DataUpdater(rdb, gormDB)
		if errUpdaterJob != nil {
			log.Println("err data updater \t:", errUpdaterJob.Error())
			continue
		}
		log.Println("job:updater executed!")
	}
}

func DataWriter(rdb *redis.Client, gormDB *gorm.DB) error {
	ctxRdb := context.Background()
	elements, errWriter := rdb.RPopCount(ctxRdb, "job:writer", 20).Result() // default to 25000
	if errWriter != nil {
		log.Printf("Writer Err: %s", errWriter.Error())
		log.Println("no job:writer are waiting...")
		return nil
	}
	if len(elements) == 0 {
		log.Println("job:writer is empty!")
		return nil
	}

	m_vacancies := []models.Vacancy{}

	ttl := 30 * time.Minute
	pipe := rdb.Pipeline()
	for _, intersectionKey := range elements { // intersection key
		intersecMembers, errIntersec := rdb.ZRevRangeByScore(ctxRdb, intersectionKey, &redis.ZRangeBy{ // get members by intersection key DESC
			Min:    "-inf",
			Max:    strconv.FormatInt(time.Now().UnixNano(), 10),
			Offset: 0,
			Count:  500,
		}).Result()
		if errIntersec != nil {
			return errIntersec
		}

		for _, HKey := range intersecMembers { // intersection members
			var vacancy WriteVacancyProps
			cmd := rdb.HGetAll(ctxRdb, HKey) // Hash field-value
			if errScanH := cmd.Scan(&vacancy); errScanH != nil {
				return errScanH
			}

			m_vacancies = append(m_vacancies, models.Vacancy{ // collect Hash data
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

			props := reflect.TypeOf(vacancy)
			hfields := []string{}
			for idx := 0; idx < props.NumField(); idx++ { // get Hash fields
				structTag := props.Field(idx).Tag.Get("json")
				hfields = append(hfields, structTag)
			}

			pipe.HExpire(ctxRdb, HKey, ttl, hfields...).Result() // set Hash fields TTL using pipe
		}

		pipe.Expire(ctxRdb, intersectionKey, ttl) // set ZIntersection TTL using pipe
	}

	store := gormDB.CreateInBatches(&m_vacancies, 100)
	if store.Error != nil {
		return store.Error
	}
	if store.RowsAffected == 0 {
		log.Println("data writer: empty stored")
	}

	if _, errExec := pipe.Exec(ctxRdb); errExec != nil {
		return errExec
	}

	return nil
}

func DataUpdater(rdb *redis.Client, gormDB *gorm.DB) error {
	ctxRdb := context.Background()
	elements, errUpdater := rdb.RPopCount(ctxRdb, "job:updater", 20).Result()
	if errUpdater != nil {
		log.Printf("Updater Err: %v", errUpdater.Error())
		log.Println("no job:updater are waiting...")
		return nil
	}
	if len(elements) == 0 {
		log.Println("job:updater is empty!")
		return nil
	}

	ttl := 30 * time.Minute
	pipe := rdb.Pipeline()
	for _, intersectionKey := range elements {
		intersecMembers, errIntersec := rdb.ZRevRangeByScore(ctxRdb, intersectionKey, &redis.ZRangeBy{ // get members by intersection key DESC
			Min:    "-inf",
			Max:    strconv.FormatInt(time.Now().UnixNano(), 10),
			Offset: 0,
			Count:  500,
		}).Result()
		if errIntersec != nil {
			return errIntersec
		}

		for IHash, HKey := range intersecMembers {
			hfields, errHfields := rdb.HKeys(ctxRdb, HKey).Result() // get Hash fields (name)
			if errHfields != nil {
				return errHfields
			}

			fieldsTTL, errTTL := rdb.HTTL(ctxRdb, HKey, hfields...).Result() // get Hash fields TTL
			if errTTL != nil {
				return errTTL
			}

			fieldHasNoTTL := 0
			fieldsToUpdate := []string{}
			for idx, seconds := range fieldsTTL { // check number of TTL
				if seconds == -1 {
					fieldHasNoTTL += 1
					fieldsToUpdate = append(fieldsToUpdate, hfields[idx])

					continue
				}
			}

			if fieldHasNoTTL > 0 && fieldHasNoTTL < len(hfields) {
				hvalues, errhv := rdb.HMGet(ctxRdb, HKey, fieldsToUpdate...).Result() // get Hash values of fields with no TTL
				if errhv != nil {
					return errhv
				}

				mappedColumns := map[string]interface{}{}
				for idx, value := range hvalues {
					mappedColumns[fieldsToUpdate[idx]] = value
				}

				if IHash == 0 {
					log.Printf("hash keys: %v", hfields)
					log.Printf("ttl fields: %v", fieldsTTL)
					log.Printf("field being updated: %v", fieldsToUpdate)
					log.Printf("value being updated: %v", mappedColumns)
				}

				ID := strings.TrimPrefix(HKey, "WB:")
				update := gormDB.Model(&models.Vacancy{Id: ID}).Updates(mappedColumns)
				if update.Error != nil {
					return update.Error
				}

				if IHash == 0 {
					log.Printf("updated status: %v", update.RowsAffected)
				}

				// _, errhexp := rdb.HExpire(ctxRdb, HKey, 30*time.Minute, fieldsToUpdate...).Result()
				/*
					logically if there any data updated, all Hash fields should re-assign TTL
				*/
				pipe.HExpire(ctxRdb, HKey, ttl, hfields...)
			}
		}
		pipe.Expire(ctxRdb, intersectionKey, ttl) // re-assign TTL
	}

	if _, errExec := pipe.Exec(ctxRdb); errExec != nil {
		log.Println("Redis: pipeline execution error!")
		return errExec
	}

	return nil
}
