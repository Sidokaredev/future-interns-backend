package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	initializer "go-read-through-service/init"
	"go-read-through-service/internal/helpers"
	"go-read-through-service/internal/models"
	"go-read-through-service/internal/services/caches"
	"go-read-through-service/internal/services/vacancies"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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
} // in use

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
} // in use

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

	rdb, errRdb := initializer.GetRedisDB() // redis conn
	if errRdb != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB() // mssql conn
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	indexes := [3]string{ // wilcard mssql query
		fmt.Sprintf("index:%s", lineIndustryQuery),
		fmt.Sprintf("index:%s", employeeTypeQuery),
		fmt.Sprintf("index:%s", workArrangement),
	}
	indexConcat := fmt.Sprintf("%s,%s,%s", lineIndustryQuery, employeeTypeQuery, workArrangement) // intersection key

	rdbCtx := context.Background()

	zInter, errZInter := rdb.ZRevRangeByScore(rdbCtx, indexConcat, &redis.ZRangeBy{ // get intersection members DESC
		Min:    "-inf",
		Max:    strconv.FormatInt(time.Now().UnixNano(), 10),
		Offset: 0,
		Count:  500,
	}).Result()
	if errZInter != nil {
		ctx.Set("CACHE_TYPE", "invalid-read-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errZInter.Error(),
			"message": fmt.Sprintf("gagal mendapatkan interseksi | key: %s", indexConcat),
		})
		return
	}

	if len(zInter) == 0 { // if empty intersection
		// log.Printf("EMPTY INTERSECTION KEY: %s", indexConcat)

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
		members := []redis.Z{}
		for _, vacancy := range vacancies { // get Hash by key
			hfields := []string{}

			props := reflect.TypeOf(vacancy)
			for i := 0; i < props.NumField(); i++ {
				structTag := props.Field(i).Tag.Get("json")
				hfields = append(hfields, structTag)
			}

			key := fmt.Sprintf("RT:%s", vacancy.ID)
			pipe.HSet(rdbCtx, key, vacancy)
			pipe.HExpire(rdbCtx, key, 30*time.Minute, hfields...)

			members = append(members, redis.Z{
				Score:  float64(vacancy.CreatedAt.UnixNano()),
				Member: key,
			})
		}

		for _, index := range indexes { // add sorted-set members
			pipe.ZAddArgs(rdbCtx, index, redis.ZAddArgs{
				GT:      true,
				Members: members,
			})
		}

		pipe.ZInterStore(rdbCtx, indexConcat, &redis.ZStore{ // store intersection members
			Keys:      indexes[:],
			Aggregate: "MAX",
		})
		pipe.Expire(rdbCtx, indexConcat, 30*time.Minute)

		if _, errExec := pipe.Exec(rdbCtx); errExec != nil {
			ctx.Set("CACHE_TYPE", "invalid-read-through")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": true,
				"error":   errExec.Error(),
				"message": "there was an err query from the pipeline",
			})
			return
		}

		zInter, errZInter := rdb.ZRevRangeByScore(rdbCtx, indexConcat, &redis.ZRangeBy{ // get intersection members DESC
			Min:    "-inf",
			Max:    strconv.FormatInt(time.Now().UnixNano(), 10),
			Offset: 0,
			Count:  500,
		}).Result()
		if errZInter != nil {
			ctx.Set("CACHE_TYPE", "invalid-read-through")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errZInter.Error(),
				"message": fmt.Sprintf("gagal mendapatkan interseksi | key: %s", indexConcat),
			})
			return
		}

		uncachedCount := 0
		cachedVacancies := []VacancyProps{}
		for _, key := range zInter { // get Hash by key
			var vacancy VacancyProps

			cmdHash := rdb.HGetAll(rdbCtx, key)
			if len(cmdHash.Val()) == 0 {
				uncachedCount += 1
			}
			if errScanHash := cmdHash.Scan(&vacancy); errScanHash != nil { // scan Hash into struct
				ctx.Set("CACHE_TYPE", "invalid-read-through")

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errScanHash.Error(),
					"message": "gagal melakukan scan ke struct VacancyProps",
				})
				return
			}

			cachedVacancies = append(cachedVacancies, vacancy)
		}
		// log.Printf("ZINTER - UNCACHED IN TOTAL: %v", uncachedCount)

		ctx.Set("CACHE_HIT", 1)
		ctx.Set("CACHE_MISS", 0)
		ctx.Set("CACHE_TYPE", "read-through")

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    cachedVacancies,
		})
		return
	}

	vacancies := []VacancyProps{}
	uncachedVacancyKeys := []string{}
	for _, key := range zInter { // get Hash by key
		var vacancy VacancyProps

		cmdHash := rdb.HGetAll(rdbCtx, key)
		if len(cmdHash.Val()) == 0 { // if Hash doesn't exist
			uncachedVacancyKeys = append(uncachedVacancyKeys, strings.TrimPrefix(key, "RT:"))
		} else {
			if errScanHash := cmdHash.Scan(&vacancy); errScanHash != nil { // scan Hash into struct
				ctx.Set("CACHE_TYPE", "invalid-read-through")

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errScanHash.Error(),
					"message": "gagal melakukan scan ke struct VacancyProps",
				})
				return
			}

			vacancies = append(vacancies, vacancy) // collect data
		}
	}

	log.Printf("UNCACHED IN TOTAL: %v of %v", len(uncachedVacancyKeys), len(zInter))

	if len(uncachedVacancyKeys) > 0 { // keys of Hash doesn't exist
		log.Println("PARTIALLY UNCACHED ... ...")
		var uncachedVacancies []VacancyProps
		sql := `
			SELECT
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
				AND
				id IN ?
		`
		unCachedQueryParams := []interface{}{
			fmt.Sprintf("%%%s%%", lineIndustryQuery),
			fmt.Sprintf("%%%s%%", employeeTypeQuery),
			fmt.Sprintf("%%%s%%", workArrangement),
			uncachedVacancyKeys,
		}
		read := gormDB.Raw(sql, unCachedQueryParams...).Scan(&uncachedVacancies)
		if read.Error != nil {
			ctx.Set("CACHE_TYPE", "invalid-read-through")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   read.Error.Error(),
				"message": "query[SELECT] gagal",
			})
			return
		}

		pipe := rdb.Pipeline()
		members := []redis.Z{}
		for _, uncachedVacancy := range uncachedVacancies { // store uncached-vacancies
			hfields := []string{}

			props := reflect.TypeOf(uncachedVacancy)
			for idx := 0; idx < props.NumField(); idx++ {
				structTag := props.Field(idx).Tag.Get("json")
				hfields = append(hfields, structTag)
			}

			key := fmt.Sprintf("RT:%s", uncachedVacancy.ID)
			pipe.HSet(rdbCtx, key, uncachedVacancy)
			pipe.HExpire(rdbCtx, key, 30*time.Minute, hfields...)

			members = append(members, redis.Z{
				Score:  float64(uncachedVacancy.CreatedAt.UnixNano()),
				Member: key,
			})
		}

		for _, index := range indexes { // store sorted-set members
			pipe.ZAddArgs(rdbCtx, index, redis.ZAddArgs{
				GT:      true,
				Members: members,
			})
		}

		pipe.ZInterStore(rdbCtx, indexConcat, &redis.ZStore{ // store intersection members
			Keys:      indexes[:],
			Aggregate: "MAX",
		})
		pipe.Expire(rdbCtx, indexConcat, 30*time.Minute)

		if _, errExec := pipe.Exec(rdbCtx); errExec != nil {
			ctx.Set("CACHE_TYPE", "invalid-read-through")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": true,
				"error":   errExec.Error(),
				"message": "there was an err query from the pipeline",
			})
			return
		}

		cachedVacancies := []VacancyProps{}
		for _, key := range uncachedVacancyKeys { // get Hash by key
			var vacancy VacancyProps

			cmdHash := rdb.HGetAll(rdbCtx, fmt.Sprintf("RT:%s", key))
			if errScanHash := cmdHash.Scan(&vacancy); errScanHash != nil { // scan Hash into struct
				ctx.Set("CACHE_TYPE", "invalid-read-through")

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errScanHash.Error(),
					"message": "gagal melakukan scan ke struct VacancyProps",
				})
				return
			}

			cachedVacancies = append(cachedVacancies, vacancy)
		}

		vacancies = append(vacancies, cachedVacancies...) // collect vacancy

		ctx.Set("CACHE_HIT", 0)
		ctx.Set("CACHE_MISS", 1)
		ctx.Set("CACHE_TYPE", "read-through")

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    vacancies,
		})
		return
	}

	ctx.Set("CACHE_HIT", 1)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "read-through")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    vacancies,
	})
} // in use

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
} // in use

// public
func VacanciesWithReadThrough(gctx *gin.Context) {
	// middleware -> public-identity-check
	identity := gctx.GetString("identity")
	authenticated := gctx.GetBool("authenticated")
	userID := gctx.GetString("user-id")

	page, errConvPage := strconv.Atoi(gctx.Query("page"))
	if errConvPage != nil {
		page = 1
	}
	limit, errConvLimit := strconv.Atoi(gctx.Query("limit"))
	if errConvLimit != nil {
		limit = 10
	}

	if !authenticated && (page > 1 || limit > 10) { // pengguna yang belum terauthentikasi hanya tidak bisa melihat lebih dari 10 data
		gctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "pengguna tidak ter-autentikasi",
			"message": "Masuk terlebih dahulu untuk melihat lebih banyak lowongan pekerjaan",
		})
		return
	}

	offset := (limit * page) - limit

	searchQueriesWilcards := []string{
		fmt.Sprintf("%%%s%%", strings.TrimSpace(strings.ToLower(gctx.Query("keyword")))),  // -> keyword harus berupa string LowerCase
		fmt.Sprintf("%%%s%%", strings.TrimSpace(strings.ToLower(gctx.Query("location")))), // -> location harus berupa string LowerCase
		fmt.Sprintf("%%%s%%", strings.TrimSpace(gctx.Query("lineIndustry"))),
		fmt.Sprintf("%%%s%%", strings.TrimSpace(gctx.Query("employeeType"))),
	}

	queryValues := []string{}
	re := regexp.MustCompile("%(.+?)%") // mengambil search query parameter yang hanya memiliki nilai [%SEARCH%]
	for i, val := range searchQueriesWilcards {
		if i == 0 && val == "%%" {
			queryValues = append(queryValues, "no-keyword") // -> jika keyword kosong, maka 'no-keyword'
		}

		if val != "%%" {
			captured := re.FindAllStringSubmatch(val, -1)
			queryValues = append(queryValues, captured[0][(len(captured[0])-1)])
		}
	}

	intersectionKey := "" // menyusun intersection untuk setiap search query parameter [Keyword:Location:LineIndustry:EmployeeType]
	for index, val := range queryValues {
		if index == (len(queryValues) - 1) {
			intersectionKey += fmt.Sprintf("%v", val)
			continue
		}
		intersectionKey += fmt.Sprintf("%v:", val)
	}

	var scoreMax string
	tmScore, errParse := time.Parse(time.RFC3339, gctx.Query("time")) // mengambil nilai [time] sebagai UnixNano untuk Redis dan RFC3339 untuk MS SQL Server
	if errParse != nil {
		tmScore = time.Now()
		scoreMax = strconv.Itoa(int(tmScore.UnixNano()))
	} else {
		scoreMax = strconv.Itoa(int(tmScore.UnixNano()))
	}

	// -> Fallback Args for Cache
	var fallbackArgs caches.FallbackArgs = caches.FallbackArgs{
		VacanciesArgs: vacancies.VacanciesArgs{
			Time:         tmScore.Format(time.RFC3339),
			Keyword:      searchQueriesWilcards[0],
			Location:     searchQueriesWilcards[1],
			LineIndustry: searchQueriesWilcards[2],
			EmployeeType: searchQueriesWilcards[3],
		},
		Offset:    offset,
		FetchNext: limit,
	}

	// -> mengambil total vacancies dan candidateID
	var count int64
	var candidateID string
	errCount := vacancies.CountAndCandidateID(userID, identity, fallbackArgs.VacanciesArgs, &count, &candidateID)
	if errCount != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCount.Error(),
			"message": "Gagal mendapatkan data candidate dan total lowongan pekerjaan",
		})
		return
	}

	// -> fallback func untuk data pipelines/applied
	var appliedFallbackFunc caches.FallbackCall = func(args caches.FallbackArgs) (*caches.FallbackReturn, error) {
		DB, errDB := initializer.GetMssqlDB()
		if errDB != nil {
			return nil, errDB
		}

		var candidateID string
		var applied []map[string]any
		var indexes []string

		errPipelines := DB.Transaction(func(tx *gorm.DB) error {
			log.Printf("user id: %v | identity: %v \n", userID, identity)
			if userID == "" || identity != "candidate" {
				applied = []map[string]any{}
				indexes = []string{}
				return nil
			}

			fetchCandidate := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", userID).First(&candidateID)
			if fetchCandidate.Error != nil {
				return fetchCandidate.Error
			}
			indexes = append(indexes, fmt.Sprintf("pipelines:%v", candidateID))

			errApplied := tx.Model(&models.Pipeline{}).Select([]string{
				"id",
				"stage",
				"status",
				"candidate_id",
				"vacancy_id",
				"created_at",
			}).Where("candidate_id = ?", candidateID).Find(&applied).Error
			if errApplied != nil {
				return errApplied
			}
			return nil
		})
		if errPipelines != nil {
			return nil, errPipelines
		}

		return &caches.FallbackReturn{
			Data:    applied,
			Indexes: indexes,
		}, nil
	}
	rt_applied := caches.NewReadThrough(appliedFallbackFunc, fallbackArgs)

	var applied []map[string]any // -> harus diubah/transform menjadi []string yang hanya berisi 'vacancy_id'
	errApplied_rt := rt_applied.GetCache(caches.CacheArgs{
		Intersection: fmt.Sprintf("pipelines:%v", candidateID),
		Min:          "-inf",
		Max:          scoreMax,
		Indexes: []string{
			fmt.Sprintf("pipelines:%v", candidateID),
		},
		CacheProps: caches.CacheProps{
			KeyPropName:    "id",
			ScorePropName:  "created_at",
			ScoreType:      "time.Time",
			MemberPropName: "id",
		},
	}, &applied)
	if errApplied_rt != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errApplied_rt.Error(),
			"message": "Gagal melakukan cache data [pipelines] dengan pola [read-through]",
		})
		return
	}
	sliceApplied := []string{}
	for _, element := range applied {
		stringID, ok := element["vacancy_id"].(string)
		if ok {
			sliceApplied = append(sliceApplied, stringID)
		}
	}

	// -> fallback func untuk data vacancies
	var fallbackFunc caches.FallbackCall = func(args caches.FallbackArgs) (*caches.FallbackReturn, error) {
		var vacancies []map[string]any
		DB, _ := initializer.GetMssqlDB()
		getVacancies := DB.Model(&models.Vacancy{}).
			Select([]string{
				"vacancies.id",
				"vacancies.position",
				"vacancies.description",
				"vacancies.qualification",
				"vacancies.responsibility",
				"vacancies.line_industry",
				"vacancies.employee_type",
				"vacancies.min_experience",
				"vacancies.salary",
				"vacancies.work_arrangement",
				"vacancies.sla",
				"vacancies.is_inactive",
				"vacancies.created_at",
				"employers.name",
				"employers.legal_name",
				"employers.location",
				"employers.profile_image_id",
			}).
			Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
			Where(`vacancies.is_inactive = ? AND
				vacancies.created_at <= ? AND
				employers.location LIKE ? AND
				(vacancies.position LIKE ? OR
				vacancies.description LIKE ? OR
				vacancies.qualification LIKE ? OR
				vacancies.responsibility LIKE ?) AND
				vacancies.line_industry LIKE ? AND
				vacancies.employee_type LIKE ?`,
				false,
				args.Time,
				args.Location,
				args.Keyword,
				args.Keyword,
				args.Keyword,
				args.Keyword,
				args.LineIndustry,
				args.EmployeeType).
			Order("vacancies.created_at DESC").
			Offset(args.Offset).
			Limit(args.FetchNext).
			Find(&vacancies)

		if getVacancies.Error != nil {
			return nil, getVacancies.Error
		}

		if getVacancies.RowsAffected == 0 {
			return &caches.FallbackReturn{
				Data:    []map[string]any{},
				Indexes: []string{},
			}, nil
		}

		employerKeys := []string{
			"name",
			"legal_name",
			"location",
			"profile_image_id",
		}

		// -> harus mendata indexes nya [line_industry, employee_type]
		indexesCheck := map[string]bool{}
		indexes := []string{}

		if args.Keyword == "%%" { // -> cek nilai keyword
			indexes = append(indexes, "no-keyword")
		}

		for _, vacancy := range vacancies {
			if !indexesCheck[vacancy["line_industry"].(string)] {
				indexesCheck[vacancy["line_industry"].(string)] = true
				indexes = append(indexes, vacancy["line_industry"].(string))
			}
			if !indexesCheck[vacancy["employee_type"].(string)] {
				indexesCheck[vacancy["employee_type"].(string)] = true
				indexes = append(indexes, vacancy["employee_type"].(string))
			}
			if !indexesCheck[vacancy["location"].(string)] {
				indexesCheck[vacancy["location"].(string)] = true
				indexes = append(indexes, vacancy["location"].(string))
			}

			employer := map[string]any{}
			for _, key := range employerKeys {
				employer[key] = vacancy[key]
			}
			helpers.TransformsIdToPath([]string{"profile_image_id"}, employer)
			vacancy["employer"] = employer
		}

		return &caches.FallbackReturn{
			Data:    vacancies,
			Indexes: indexes,
		}, nil
	}
	rt_vacancies := caches.NewReadThrough(fallbackFunc, fallbackArgs)

	var vacancies []map[string]any
	errCache := rt_vacancies.GetCache(caches.CacheArgs{
		Intersection: intersectionKey,
		Min:          "-inf",
		Max:          scoreMax,
		Count:        int64(limit),
		Offset:       int64(offset),
		Indexes:      queryValues,
		CacheProps: caches.CacheProps{
			KeyPropName:    "id",
			ScorePropName:  "created_at",
			ScoreType:      "time.Time",
			MemberPropName: "id",
		},
	}, &vacancies)
	if errCache != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCache.Error(),
			"message": "Gagal melakukan cache dengan pola [:read-through]",
		})

		return
	}

	var last_time string
	if len(vacancies) == 0 {
		last_time = "Data pencarian telah ditampilkan secara kesluruhan"
	} else {
		timeParse, ok := vacancies[len(vacancies)-1]["created_at"].(time.Time)
		if ok {
			last_time = timeParse.Format(time.RFC3339)
		} else {
			last_time = vacancies[len(vacancies)-1]["created_at"].(string)
		}
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count":     count,
			"applied":   sliceApplied,
			"vacancies": vacancies,
			"last_time": last_time, // validate if len == 0
		},
	})
}

func VacanciesByIdWithReadThrough(gctx *gin.Context) {
	authenticated := gctx.GetBool("authenticated")
	identity := gctx.GetString("identity")
	userID := gctx.GetString("user-id")

	vacancyID := gctx.Param("id")
	if vacancyID == "" {
		gctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "parameter :id untuk vacancy_id tidak tersedia",
			"message": "Periksa kembali dan pastikan nilai parameter :id untuk vacancy_id",
		})
		return
	}

	DB, errDB := initializer.GetMssqlDB()
	if errDB != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDB.Error(),
			"message": "Gagal memanggil GORM Instance connection",
		})
		return
	}

	ctx := context.Background()
	rdb, err := initializer.GetRedisDB()
	if err != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"message": "Gagal memanggil koneksi Redis Server",
		})
		return
	}

	applied := []string{}
	var vacancy map[string]any

	if authenticated && identity == "candidate" {
		var candidateID string
		errCandidate := DB.Model(&models.Candidate{}).Select("id").Where("user_id = ?", userID).First(&candidateID).Error
		if errCandidate != nil {
			gctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errCandidate.Error(),
				"message": "Data pengguna sebagai kandidat tidak ditemukan",
			})
			return
		}

		log.Println("Looking for cached pipelines ...")
		key := fmt.Sprintf("pipelines:%v", candidateID)
		find_pipelines, errFind := rdb.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
			Min: "-inf",
			Max: "+inf",
		}).Result()
		if errFind != nil {
			gctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errFind.Error(),
				"message": fmt.Sprintf("Gagal menjalankan Redis ZREVRANGEBYSCORE untuk key [%v]", key),
			})
			return
		}

		if len(find_pipelines) == 0 {
			log.Println("Falling back pipelines from database ...")

			pipelinesFromDB := []map[string]any{}
			errPipelineFromDB := DB.Model(&models.Pipeline{}).Select([]string{
				"id",
				"stage",
				"status",
				"candidate_id",
				"vacancy_id",
				"created_at",
			}).Where("candidate_id = ?", candidateID).Find(&pipelinesFromDB).Error
			if errPipelineFromDB != nil {
				gctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errPipelineFromDB.Error(),
					"message": "Gagal menjalankan query data Pipelines",
				})
				return
			}

			if len(pipelinesFromDB) > 0 {
				hash := caches.ExtractToHash("id", pipelinesFromDB)
				sortedSet := caches.NewSortedSetCollection(hash, caches.SortedSetArgs{
					ScorePropName:  "created_at",
					ScoreType:      "time.Time",
					MemberPropName: "id",
				})
				sortedSet.Keys = []string{
					key,
				}
				errZAdd := sortedSet.Add(1 * time.Hour)
				if errZAdd != nil {
					gctx.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"error":   errZAdd.Error(),
						"message": "Gagal menjalankan Redis ZADD data pipelines",
					})
					return
				}
				errHSet := hash.Add(1 * time.Hour)
				if errHSet != nil {
					gctx.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"error":   errHSet.Error(),
						"message": "Gagal menjalankan Redis HSET data Pipelines",
					})
					return
				}

				log.Println("Looking for new cached pipelines ...")
				new_cache_pipelines, errNewPipelines := rdb.ZRevRangeByScore(ctx, key, &redis.ZRangeBy{
					Min: "-inf",
					Max: "+inf",
				}).Result()
				if errNewPipelines != nil {
					gctx.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"error":   errNewPipelines.Error(),
						"message": fmt.Sprintf("Gagal menjalankan Redis ZREVRANGEBYSCORE untuk key [%v]", key),
					})
					return
				}

				cached_pipelines := []map[string]string{}
				for _, id := range new_cache_pipelines {
					hash, errHGetAll := rdb.HGetAll(ctx, id).Result()
					if errHGetAll != nil {
						gctx.JSON(http.StatusInternalServerError, gin.H{
							"success": false,
							"error":   errHGetAll.Error(),
							"message": fmt.Sprintf("Gagal menjalankan Redis HGETALL untuk key [%v]", id),
						})
						return
					}
					cached_pipelines = append(cached_pipelines, hash)
				}

				for _, element := range cached_pipelines {
					applied = append(applied, element["vacancy_id"])
				}
			}
		} else {
			cached_pipelines := []map[string]string{}
			for _, id := range find_pipelines {
				hash, errHGetAll := rdb.HGetAll(ctx, id).Result()
				if errHGetAll != nil {
					gctx.JSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"error":   errHGetAll.Error(),
						"message": fmt.Sprintf("Gagal menjalankan Redis HGETALL untuk key [%v]", id),
					})
					return
				}
				cached_pipelines = append(cached_pipelines, hash)
			}

			for _, element := range cached_pipelines {
				applied = append(applied, element["vacancy_id"])
			}
		}
	}

	log.Println("Looking for cached vacancy ...")
	cache_vacancy, errHash := rdb.HGetAll(ctx, vacancyID).Result()
	if errHash != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errHash.Error(),
			"message": fmt.Sprintf("Gagal mendapatkan cache Hash [%v]", vacancyID),
		})
		return
	}

	if len(cache_vacancy) == 0 {
		log.Println("Falling back vacancy from database ...")

		var new_vacancy map[string]any
		errVacancy := vacancies.ById(vacancyID, &new_vacancy)
		if errVacancy != nil {
			gctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errVacancy.Error(),
				"message": "Gagal mendapatkan data vacancy berdasarkan ID",
			})
			return
		}

		employerKeys := []string{
			"name",
			"legal_name",
			"location",
			"profile_image_id",
		}
		employer := map[string]any{}
		for _, key := range employerKeys {
			employer[key] = new_vacancy[key]
		}
		helpers.TransformsIdToPath([]string{"profile_image_id"}, employer)
		new_vacancy["employer"] = employer

		hash := caches.ExtractToHash("id", new_vacancy)
		errHSet := hash.Add(1 * time.Hour)
		if errHSet != nil {
			gctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errHSet.Error(),
				"message": "Gagal melakukan cache data pekerjaan",
			})
			return
		}

		new_cache_vacancy, errNewHash := rdb.HGetAll(ctx, vacancyID).Result()
		if errNewHash != nil {
			gctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errNewHash.Error(),
				"message": fmt.Sprintf("Gagal mendapatkan cache Hash [%v]", vacancyID),
			})
			return
		}

		vacancy = caches.TransformNestedMap(new_cache_vacancy)
	} else {
		vacancy = caches.TransformNestedMap(cache_vacancy)
	}

	gctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"applied": applied,
			"vacancy": vacancy,
		},
	})
}
