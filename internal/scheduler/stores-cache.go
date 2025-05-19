package scheduler

import (
	"context"
	"encoding/json"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/models"
	"log"
	"time"
)

func StoreCaches() error {
	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		log.Printf("[%v] \t: err redis -> %v", time.Now().Format("02/01/2006 15:04:05"), errRdb.Error())
		return errRdb
	}

	rdbContext := context.Background()

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		log.Printf("[%v] \t: err gorm -> %v", time.Now().Format("02/01/2006 15:04:05"), errGorm.Error())
		return errGorm
	}

	var cursor uint64

	for {
		var err error
		var slicedKeys []string
		slicedKeys, cursor, err = rdb.Scan(rdbContext, cursor, "*", 100).Result()
		if err != nil {
			log.Printf("[%v] \t: fail scanning keys -> %v", time.Now().Format("02/01/2006 15:04:05"), err.Error())
			return err
		}

		// check the length of sliced keys
		if len(slicedKeys) == 0 {
			log.Printf("[%v] \t: empty keys on redis", time.Now().Format("02/01/2006 15:04:05"))
			return nil
		}

		values, errGet := rdb.MGet(rdbContext, slicedKeys...).Result()
		if errGet != nil {
			log.Printf("[%v] \t: fail mget values -> %v", time.Now().Format("02/01/2006 15:04:05"), errGet.Error())
			return errGet
		}

		m_vacancies := []models.Vacancy{}
		for _, value := range values {
			if value == nil {
				continue
			}

			var vacancy struct {
				ID              string    `json:"id"`
				Position        string    `json:"position"`
				Description     string    `json:"description"`
				Qualification   string    `json:"qualification"`
				Responsibility  string    `json:"responsibility"`
				LineIndustry    string    `json:"line_industry"`
				EmployeeType    string    `json:"employee_type"`
				MinExperience   string    `json:"min_experience"`
				Salary          int64     `json:"salary"`
				WorkArrangement string    `json:"work_arrangement"`
				SLA             int       `json:"sla"`
				IsInactive      int       `json:"is_inactive"`
				EmployerID      string    `json:"employer_id"`
				CreatedAt       time.Time `json:"created_at"`
			}

			if jsonVal, ok := value.(string); ok {
				if errDecode := json.Unmarshal([]byte(jsonVal), &vacancy); errDecode != nil {
					log.Printf("[%v] \t: unmarshal -> %v", time.Now().Format("02/01/2006 15:04:05"), errDecode.Error())
					continue
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
					Salary:          uint(vacancy.Salary),
					WorkArrangement: vacancy.WorkArrangement,
					SLA:             int32(vacancy.SLA),
					IsInactive:      vacancy.IsInactive != 0,
					EmployerId:      vacancy.EmployerID,
					CreatedAt:       vacancy.CreatedAt,
				})
			} else {
				log.Printf("[%v] \t: string -> value is not string", time.Now().Format("02/01/2006 15:04:05"))
				continue
			}
		}

		errStoreVacancies := gormDB.CreateInBatches(&m_vacancies, 25).Error
		if errStoreVacancies != nil {
			log.Printf("[%v] \t: gorm -> %v", time.Now().Format("02/01/2006 15:04:05"), errStoreVacancies.Error())
			return errStoreVacancies
		}
		log.Printf("[%v] \t: gorm -> data stored", time.Now().Format("02/01/2006 15:04:05"))

		rdbPipeline := rdb.Pipeline()
		for _, key := range slicedKeys {
			rdbPipeline.Del(rdbContext, key)
		}

		_, errPipe := rdbPipeline.Exec(rdbContext)
		if errPipe != nil {
			log.Printf("[%v] \t: err pipe -> %v", time.Now().Format("02/01/2006 15:04:05"), errPipe.Error())
			return errPipe
		}
		log.Printf("[%v] \t: redis -> data deleted", time.Now().Format("02/01/2006 15:04:05"))

		if cursor == 0 {
			break
		}
	}

	return nil
}
