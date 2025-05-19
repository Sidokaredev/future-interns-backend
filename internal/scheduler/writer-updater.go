package scheduler

import (
	"context"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/models"
	"log"
	"reflect"
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
	elements, errWriter := rdb.RPopCount(ctxRdb, "job:writer", 25000).Result()
	if errWriter != nil {
		log.Println("no job:writer are waiting...")
		return nil
	}
	if len(elements) == 0 {
		log.Println("job:writer is empty!")
		return nil
	}

	m_vacancies := []models.Vacancy{}

	ttl := 30 * time.Minute
	for _, key := range elements {
		var vacancy WriteVacancyProps
		cmd := rdb.HGetAll(ctxRdb, key)
		if errScanH := cmd.Scan(&vacancy); errScanH != nil {
			return errScanH
		}

		m_vacancies = append(m_vacancies, models.Vacancy{
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
		for idx := 0; idx < props.NumField(); idx++ {
			structTag := props.Field(idx).Tag.Get("json")
			hfields = append(hfields, structTag)
		}

		_, errExp := rdb.HExpire(ctxRdb, key, ttl, hfields...).Result()
		if errExp != nil {
			return errExp
		}
	}

	store := gormDB.CreateInBatches(&m_vacancies, 100)
	if store.Error != nil {
		return store.Error
	}
	if store.RowsAffected == 0 {
		log.Println("data writer: empty stored")
	}

	return nil
}

func DataUpdater(rdb *redis.Client, gormDB *gorm.DB) error {
	ctxRdb := context.Background()
	elements, errUpdater := rdb.RPopCount(ctxRdb, "job:updater", 25000).Result()
	if errUpdater != nil {
		log.Println("no job:updater are waiting...")
		return nil
	}
	if len(elements) == 0 {
		log.Println("job:updater is empty!")
		return nil
	}

	for idx, key := range elements {
		hfields, errHfields := rdb.HKeys(ctxRdb, key).Result()
		if errHfields != nil {
			return errHfields
		}
		// get fields of keys

		fieldsTTL, errTTL := rdb.HTTL(ctxRdb, key, hfields...).Result()
		if errTTL != nil {
			return errTTL
		}
		// get ttl of each fields

		fieldHasNoTTL := 0
		fieldsToUpdate := []string{}
		for idx, seconds := range fieldsTTL {
			if seconds == -1 {
				fieldHasNoTTL += 1
				fieldsToUpdate = append(fieldsToUpdate, hfields[idx])

				continue
			}
		}

		if fieldHasNoTTL > 0 && fieldHasNoTTL < len(hfields) {
			hvalues, errhv := rdb.HMGet(ctxRdb, key, fieldsToUpdate...).Result()
			if errhv != nil {
				return errhv
			}

			mappedColumns := map[string]interface{}{}
			for idx, value := range hvalues {
				mappedColumns[fieldsToUpdate[idx]] = value
			}

			if idx == 0 {
				log.Printf("hash keys: %v", hfields)
				log.Printf("ttl fields: %v", fieldsTTL)
				log.Printf("field being updated: %v", fieldsToUpdate)
				log.Printf("value beign updated: %v", mappedColumns)
			}

			ID := strings.TrimPrefix(key, "WB:")
			update := gormDB.Model(&models.Vacancy{Id: ID}).Updates(mappedColumns)
			if update.Error != nil {
				return update.Error
			}

			if idx == 0 {
				log.Printf("updated status: %v", update.RowsAffected)
			}
			// if update.RowsAffected == 0 {
			// 	log.Println("no columns were updated")
			// }

			_, errhexp := rdb.HExpire(ctxRdb, key, 30*time.Minute, fieldsToUpdate...).Result()
			if errhexp != nil {
				return errhexp
			}

			continue
		}
	}

	return nil
}
