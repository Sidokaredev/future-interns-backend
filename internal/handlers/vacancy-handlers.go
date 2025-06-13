package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	initializer "go-read-through-service/init"
	"go-read-through-service/internal/models"
	"log"
	"net/http"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type VacancyHandler struct {
}
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
	IsInactive      bool      `json:"is_inactive" redis:"is_inactive"`
	EmployerID      string    `json:"employer_id" redis:"employer_id"`
	CreatedAt       time.Time `json:"created_at" redis:"created_at"`
}

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

type UpdateVacancyProps struct {
	ID              string     `json:"id" binding:"required"`
	Position        *string    `json:"position"`
	Description     *string    `json:"description"`
	Qualification   *string    `json:"qualification"`
	Responsibility  *string    `json:"responsibility"`
	LineIndustry    *string    `json:"line_industry"`
	EmployeeType    *string    `json:"employee_type"`
	MinExperience   *string    `json:"min_experience"`
	Salary          *uint      `json:"salary"`
	WorkArrangement *string    `json:"work_arrangement"`
	SLA             *int32     `json:"sla"`
	IsInactive      *bool      `json:"is_inactive"`
	EmployerID      *string    `json:"employer_id"`
	CreatedAt       *time.Time `json:"created_at"`
}

func (handler *VacancyHandler) WriteVacanciesReadThrough(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []WriteVacancyProps

	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "failed to bind collection of JSON fields",
		})

		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "failed getting GORM instance",
		})

		ctx.Abort()
		return
	}

	m_vacancies := []models.Vacancy{}
	vacanciesID := []string{}

	for _, vacancy := range RequestBody {
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
		vacanciesID = append(vacanciesID, vacancy.ID)
	}

	errStoreVacancies := gormDB.CreateInBatches(&m_vacancies, 100).Error
	if errStoreVacancies != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreVacancies.Error(),
			"message": fmt.Sprintf("failed storing %v vacancies", len(m_vacancies)),
		})

		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "read-through")

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    vacanciesID,
	})
}

func (handler *VacancyHandler) GetVacanciesReadThrough(ctx *gin.Context) {
	ctx.Set("CACHE_TYPE", "read-through")

	claim, exists := ctx.Get("claim")
	if !exists {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"error":   "key 'claim' doens't exist, probably authentication middleware",
			"message": "user has no credentials",
		})

		return
	}

	rdb, errRedis := initializer.GetRedisDB()
	if errRedis != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRedis.Error(),
			"message": "fail while gettting RDB instance",
		})

		return
	}

	var vacancies []map[string]interface{}
	var applied []string

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail while getting MSSQL instance",
		})

		return
	}

	errGetApplied := gormDB.Transaction(func(tx *gorm.DB) error {
		var candidateID string
		errGetCandidateID := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", claim).First(&candidateID).Error
		if errGetCandidateID != nil {
			return errGetCandidateID
		}

		getPipelines := tx.Model(&models.Pipeline{}).Select("vacancy_id").Where("candidate_id = ?", candidateID).Find(&applied)
		if getPipelines.RowsAffected == 0 {
			log.Println("candidate has zero apply")
			applied = []string{}
		}
		return nil
	})

	if errGetApplied != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetApplied.Error(),
			"message": "error getting applied vacancies",
		})

		return
	}

	var cachedRequest atomic.Int32

	c := context.Background()
checkcached_label:
	cachedJSON, errCached := rdb.Get(c, claim.(string)).Result()
	if errCached != nil {
		if cachedRequest.Load() > 3 {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "uncontrolled loop cache rerquest",
				"message": "cached request has been trying for more than 3 times",
			})

			ctx.Abort()
			return
		}

		ctx.Set("CACHE_MISS", 1) // set cache miss
		ctx.Set("CACHE_HIT", 0)  // set cache hit

		cachedRequest.Add(1)

		reqBody := map[string]int{"redis_database": viper.GetInt("redis.dev.database")}
		reqBodyJSON, errEncode := json.Marshal(&reqBody)
		if errEncode != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errEncode.Error(),
				"message": "fail encoding request body",
			})

			return
		}

		req, errReq := http.NewRequest("POST", "http://localhost:8004/api/v1/no-cache/vacancies/set-cache", bytes.NewBuffer(reqBodyJSON))
		if errReq != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errReq.Error(),
				"message": "fail to make cache request",
			})

			return
		}
		req.Header.Add("Authorization", ctx.GetHeader("Authorization"))

		client := &http.Client{}
		resp, errResp := client.Do(req)
		if errResp != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errResp.Error(),
				"message": "fail getting response",
			})

			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "cache request fail",
				"message": "cache service operation failed",
			})

			return
		}

		goto checkcached_label
	}

	if errDecode := json.Unmarshal([]byte(cachedJSON), &vacancies); errDecode != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDecode.Error(),
			"message": "error decoding cached JSON data",
		})

		return
	}

	ctx.Set("CACHE_MISS", 0) // set cache miss
	ctx.Set("CACHE_HIT", 1)  // set cache hit

	log.Println("go for cached vacancies...")
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"applied":   applied,
			"vacancies": vacancies,
		},
	})
}

func (handler *VacancyHandler) ReadCacheAsideVacancies(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
	WorkArrangement, _ := ctx.GetQuery("workArrangement")

	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		ctx.Set("CACHE_TYPE", "read-through-INVALID")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "read-through-INVALID")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	vacancyQueryFunc := func(values ...string) bool {
		compareValues := []string{
			lineIndustryQuery,
			employeeTypeQuery,
			WorkArrangement,
		}
		for idxv, v := range values {
			if compareValues[idxv] != v {
				return false
			}
		}

		return true
	}

	vacancies := []VacancyProps{}

	ctxRdb := context.Background()
	var cursor uint64
	for {
		var keys []string
		var errScan error
		keys, cursor, errScan = rdb.Scan(ctxRdb, cursor, "RT:*", 50).Result()
		if errScan != nil {
			ctx.Set("CACHE_TYPE", "read-through-INVALID")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errScan.Error(),
				"message": "fail scanning matching keys",
			})
			return
		}

		for idxk, key := range keys {
			var vacancy VacancyProps

			cmd := rdb.HGetAll(ctxRdb, key)
			if errScan := cmd.Scan(&vacancy); errScan != nil {
				ctx.Set("CACHE_TYPE", "read-through-INVALID")

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errScan.Error(),
					"message": fmt.Sprintf("fail scanning field-value at index:%d", idxk),
				})
				return
			}

			if vacancyQueryFunc(
				vacancy.LineIndustry,
				vacancy.EmployeeType,
				vacancy.WorkArrangement,
			) {
				vacancies = append(vacancies, vacancy)
			}
		}

		if cursor == 0 {
			break
		}
	}

	if len(vacancies) < 500 {
		ctx.Set("CACHE_HIT", 0)
		ctx.Set("CACHE_MISS", 1)

		sqlQuery := `
		SELECT TOP (500)
			id,
			position,
			description,
			qualification,
			responsibility,
			line_industry,
			employee_type,
			min_experience,
			salary,
			work_arrangement,
			sla,
			is_inactive,
			employer_id,
			created_at
		FROM
			vacancies
		WHERE
			line_industry LIKE ?
			AND
			employee_type LIKE ?
			AND
			work_arrangement LIKE ?
	`

		vacanciesUpdate := []map[string]interface{}{}
		getVacancies := gormDB.Raw(sqlQuery,
			fmt.Sprintf("%%%s%%", lineIndustryQuery),
			fmt.Sprintf("%%%s%%", employeeTypeQuery),
			fmt.Sprintf("%%%s%%", WorkArrangement),
		).Scan(&vacanciesUpdate)
		if getVacancies.Error != nil {
			ctx.Set("CACHE_TYPE", "read-through-INVALID")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   getVacancies.Error.Error(),
				"message": "there was an issue with sql query",
			})
			return
		}

		if getVacancies.RowsAffected == 0 {
			log.Println("empty vacancies!")
		} else {
			errUpdate := UpdateCaches(rdb, vacanciesUpdate)
			if errUpdate != nil {
				log.Println("fail updating cache from database source!")
			}
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    vacanciesUpdate,
		})
		return
	}

	ctx.Set("CACHE_TYPE", "read-through")
	ctx.Set("CACHE_HIT", 1)
	ctx.Set("CACHE_MISS", 0)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    vacancies,
	})
}

func (handler *VacancyHandler) UpdateVacanciesReadThrough(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []UpdateVacancyProps
	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check JSON fields",
		})
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	tx := gormDB.Begin()
	for _, vacancy := range RequestBody {
		props := reflect.TypeOf(vacancy)
		values := reflect.ValueOf(vacancy)
		mappedColumns := map[string]interface{}{}
		for idx := 0; idx < props.NumField(); idx++ {
			structTag := props.Field(idx).Tag.Get("json")
			if values.Field(idx).Kind() == reflect.Ptr && values.Field(idx).IsNil() {
				continue
			}
			if structTag == "id" && values.Field(idx).Kind() == reflect.String {
				continue
			}
			mappedColumns[structTag] = values.Field(idx).Elem().Interface()
		}

		update := tx.Model(&models.Vacancy{Id: vacancy.ID}).Updates(mappedColumns)
		if update.RowsAffected == 0 {
			tx.Rollback()
		}
	}

	if errCommit := tx.Commit().Error; errCommit != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCommit.Error(),
			"message": "fail updating data at database",
		})
		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "read-through")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    "successfully update all data",
	})
}

func UpdateCaches(rdb *redis.Client, vacancies []map[string]interface{}) error {
	c := context.Background()
	for _, vacancy := range vacancies {
		key := fmt.Sprintf("RT:%s", vacancy["id"])
		_, errHSet := rdb.HSet(c, key, vacancy).Result()
		if errHSet != nil {
			return errHSet
		}

		fields := []string{}
		for key := range vacancy {
			fields = append(fields, key)
		}

		_, errHExp := rdb.HExpire(c, key, 30*time.Minute, fields...).Result()
		if errHExp != nil {
			return errHExp
		}
	}

	return nil
}

// using set indexing
func (handler *VacancyHandler) ReadThroughService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
	workArrangement, _ := ctx.GetQuery("workArrangement")

	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	indexes := [3]string{
		fmt.Sprintf("index:%s", lineIndustryQuery),
		fmt.Sprintf("index:%s", employeeTypeQuery),
		fmt.Sprintf("index:%s", workArrangement),
	}

	rdbCtx := context.Background()
	log.Printf("FINDING INDEX \t: %s | %s | %s", indexes[0], indexes[1], indexes[2])

	sizeInterCard, errInterCard := rdb.SInterCard(rdbCtx, 500, indexes[0], indexes[1], indexes[2]).Result()
	if errInterCard != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errInterCard.Error(),
			"message": "fail counting intersection size",
		})
		return
	}
	if sizeInterCard < 500 {
		log.Println("reading from database...")
		ctx.Set("CACHE_HIT", 0)
		ctx.Set("CACHE_MISS", 1)

		queryParams := []interface{}{
			fmt.Sprintf("%%%s%%", lineIndustryQuery),
			fmt.Sprintf("%%%s%%", employeeTypeQuery),
			fmt.Sprintf("%%%s%%", workArrangement),
		}
		vacancies, errRead := ReadFromDatabase(gormDB, queryParams...)
		if errRead != nil {
			ctx.Set("CACHE_TYPE", "invalid-read-through")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errRead.Error(),
				"message": "fail reading from database source",
			})
			return
		}

		pipe := rdb.Pipeline()
		keysCollection := []string{}
		for _, vacancy := range vacancies {
			hfields := []string{}
			hfieldValues := []interface{}{}

			values := reflect.ValueOf(vacancy)
			props := reflect.TypeOf(vacancy)

			for i := 0; i < props.NumField(); i++ {
				structTag := props.Field(i).Tag.Get("json")
				hfields = append(hfields, structTag)

				value := values.Field(i)
				hfieldValues = append(hfieldValues, structTag, value.Interface())
			}

			key := fmt.Sprintf("RT:%s", vacancy.ID)
			// pipe.HSet(rdbCtx, key, vacancy)
			pipe.HSet(rdbCtx, key, hfieldValues)
			pipe.HExpire(rdbCtx, key, 30*time.Minute, hfields...)

			keysCollection = append(keysCollection, key)
		}

		for _, index := range indexes {
			pipe.SAdd(rdbCtx, index, keysCollection)
		}

		if _, errExec := pipe.Exec(rdbCtx); errExec != nil {
			ctx.Set("CACHE_TYPE", "invalid-read-through")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": true,
				"error":   errExec.Error(),
				"message": "there was an err query from the pipeline",
			})
			return
		}
	}

	sInter, errSInter := rdb.SInter(rdbCtx, indexes[0], indexes[1], indexes[2]).Result()
	if errSInter != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errSInter.Error(),
			"message": "fail getting intersection values",
		})
		return
	}

	vacancies := []VacancyProps{}
	for _, key := range sInter {
		var vacancy VacancyProps

		cmd := rdb.HGetAll(rdbCtx, key)
		if errScanHash := cmd.Scan(&vacancy); errScanHash != nil {
			ctx.Set("CACHE_TYPE", "invalid-read-through")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errScanHash.Error(),
				"message": "fail scanning hash field-value",
			})
			return
		}

		vacancies = append(vacancies, vacancy)
	}

	ctx.Set("CACHE_HIT", 1)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "read-through")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    vacancies,
	})
}

func ReadFromDatabase(gormDB *gorm.DB, queryParams ...interface{}) ([]VacancyProps, error) {
	sql := `
		SELECT TOP (500)
			id,
			position,
			description,
			qualification,
			responsibility,
			line_industry,
			employee_type,
			min_experience,
			salary,
			work_arrangement,
			sla,
			is_inactive,
			employer_id,
			created_at
		FROM
			vacancies
		WHERE
			line_industry LIKE ?
			AND
			employee_type LIKE ?
			AND
			work_arrangement LIKE ?
	`
	// var vacancies []map[string]interface{}
	var vacancies []VacancyProps
	read := gormDB.Raw(sql, queryParams...).Scan(&vacancies)
	if read.Error != nil {
		return nil, read.Error
	}

	return vacancies, nil
}
