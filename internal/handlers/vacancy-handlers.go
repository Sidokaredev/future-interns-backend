package handlers

import (
	"context"
	"fmt"
	initializer "go-cache-aside-service/init"
	"go-cache-aside-service/internal/models"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

func (handler *VacancyHandler) WriteVacanciesCacheAside(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []WriteVacancyProps

	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "failed to bind collection of JSON fields",
		})

		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

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
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreVacancies.Error(),
			"message": fmt.Sprintf("failed storing %v vacancies", len(m_vacancies)),
		})

		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "cache-aside")

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    vacanciesID,
	})
} // in use

func (handler *VacancyHandler) UpdateVacanciesCacheAside(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []UpdateVacancyProps
	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check JSON fields",
		})
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

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
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCommit.Error(),
			"message": "fail updating data at database",
		})
		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "cache-aside")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    "successfully update all data",
	})
} // in use

func UpdateCaches(rdb *redis.Client, vacancies []map[string]interface{}) error {
	c := context.Background()
	for _, vacancy := range vacancies {
		key := fmt.Sprintf("CA:%s", vacancy["id"])
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

func (handler *VacancyHandler) ReadCacheAsideService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
	workArrangement, _ := ctx.GetQuery("workArrangement")

	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	} // redis conn

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	} // mssql conn

	indexes := [3]string{ // key of SET
		fmt.Sprintf("index:%s", lineIndustryQuery),
		fmt.Sprintf("index:%s", employeeTypeQuery),
		fmt.Sprintf("index:%s", workArrangement),
	}
	indexConcat := fmt.Sprintf("%s,%s,%s", lineIndustryQuery, employeeTypeQuery, workArrangement) // intersection key

	rdbCtx := context.Background()

	/*
		sInter, errSInter := rdb.SInter(rdbCtx, indexes[0], indexes[1], indexes[2]).Result() // members INTERSECTION
		if errSInter != nil {
			ctx.Set("CACHE_TYPE", "invalid-cache-aside")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errSInter.Error(),
				"message": "fail getting intersection values",
			})
			return
		}
	*/
	zInter, errZInter := rdb.ZRevRangeByScore(rdbCtx, indexConcat, &redis.ZRangeBy{ // get intersection DESC by unixnano as maximum score
		Min:    "-inf",
		Max:    strconv.FormatInt(time.Now().UnixNano(), 10),
		Offset: 0,
		Count:  500,
	}).Result()
	if errZInter != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errZInter.Error(),
			"message": fmt.Sprintf("gagal mendapatkan interseksi | key: %s", indexConcat),
		})
		return
	}

	if len(zInter) == 0 { // length members
		log.Printf("EMPTY INTERSECTION KEY: %s", indexConcat)
		ctx.Set("CACHE_HIT", 0)
		ctx.Set("CACHE_MISS", 1)

		queryParams := []interface{}{
			fmt.Sprintf("%%%s%%", lineIndustryQuery),
			fmt.Sprintf("%%%s%%", employeeTypeQuery),
			fmt.Sprintf("%%%s%%", workArrangement),
		}
		vacancies, errRead := ReadFromDatabase(gormDB, queryParams...)
		if errRead != nil {
			ctx.Set("CACHE_TYPE", "invalid-cache-aside")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errRead.Error(),
				"message": "fail reading from database source",
			})
			return
		}

		pipe := rdb.Pipeline()
		// keysCollection := []interface{}{}
		members := []redis.Z{}
		for _, vacancy := range vacancies {
			hfields := []string{}

			props := reflect.TypeOf(vacancy)
			for idx := 0; idx < props.NumField(); idx++ {
				structTag := props.Field(idx).Tag.Get("json")
				hfields = append(hfields, structTag)
			}

			key := fmt.Sprintf("CA:%s", vacancy.ID)
			pipe.HSet(rdbCtx, key, vacancy)
			pipe.HExpire(rdbCtx, key, 30*time.Minute, hfields...)

			// keysCollection = append(keysCollection, key)
			members = append(members, redis.Z{
				Score:  float64(vacancy.CreatedAt.UnixNano()),
				Member: key,
			})
		}

		for _, index := range indexes {
			pipe.ZAddArgs(rdbCtx, index, redis.ZAddArgs{
				XX:      true,
				GT:      true,
				Members: members,
			})
			// pipe.SAdd(rdbCtx, index, keysCollection...) // interface
		}

		pipe.ZInterStore(rdbCtx, indexConcat, &redis.ZStore{
			Keys:      indexes[:],
			Aggregate: "MAX",
		})
		pipe.Expire(rdbCtx, indexConcat, 30*time.Minute)

		if _, errExec := pipe.Exec(rdbCtx); errExec != nil {
			ctx.Set("CACHE_TYPE", "invalid-cache-aside")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": true,
				"error":   errExec.Error(),
				"message": "there was an err query from the pipeline",
			})
			return
		}

		ctx.Set("CACHE_TYPE", "cache-aside")
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    vacancies,
		})
		return
	}

	vacancies := []VacancyProps{}
	// removedMemberCount := 0
	unCachedVacancyKeys := []string{}
	for _, key := range zInter {
		var vacancy VacancyProps

		cmd := rdb.HGetAll(rdbCtx, key)
		if len(cmd.Val()) == 0 { // empty HASH by key
			// for _, index := range indexes { // remove member over SET by key
			// 	setRemStatus, errSetRem := rdb.SRem(rdbCtx, index, key).Result()
			// 	if errSetRem != nil {
			// 		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

			// 		ctx.JSON(http.StatusInternalServerError, gin.H{
			// 			"success": false,
			// 			"error":   errSetRem.Error(),
			// 			"message": fmt.Sprintf("terjadi kegagalan ketika menghapus member[%s] dari set[%s]", key, index),
			// 		})
			// 		return
			// 	}

			// 	if setRemStatus == 1 { // counting removed members and collect uncached vacancy id
			// 		removedMemberCount += 1
			// 		unCachedVacancyKeys = append(unCachedVacancyKeys, strings.TrimPrefix(key, "CA:"))
			// 	}
			// }
			unCachedVacancyKeys = append(unCachedVacancyKeys, strings.TrimPrefix(key, "CA:"))
		} else { // scanning HASH value
			if errScanHash := cmd.Scan(&vacancy); errScanHash != nil {
				ctx.Set("CACHE_TYPE", "invalid-cache-aside")

				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errScanHash.Error(),
					"message": "fail scanning hash field-value",
				})
				return
			}

			vacancies = append(vacancies, vacancy) // collect vacancy
		}
	}

	log.Printf("UNCACHED IN TOTAL: %v of %v", len(unCachedVacancyKeys), len(zInter))

	if len(unCachedVacancyKeys) > 0 { // uncached vacancy id more than 0
		ctx.Set("CACHE_HIT", 0)
		ctx.Set("CACHE_MISS", 1)

		var unCachedVacancies []VacancyProps
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
			unCachedVacancyKeys,
		}
		read := gormDB.Raw(sql, unCachedQueryParams...).Scan(&unCachedVacancies)
		if read.Error != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   read.Error.Error(),
				"message": "query[SELECT] gagal",
			})
			return
		}

		pipe := rdb.Pipeline() // set vacancy id as members
		// keysCollection := []interface{}{}
		members := []redis.Z{}
		for _, uncachedVacancy := range unCachedVacancies { // uncached vacancies from query
			hfields := []string{}

			props := reflect.TypeOf(uncachedVacancy)
			for idx := 0; idx < props.NumField(); idx++ {
				structTag := props.Field(idx).Tag.Get("json")
				hfields = append(hfields, structTag)
			}

			key := fmt.Sprintf("CA:%s", uncachedVacancy.ID)
			pipe.HSet(rdbCtx, key, uncachedVacancy)
			pipe.HExpire(rdbCtx, key, 30*time.Minute, hfields...)

			// keysCollection = append(keysCollection, key)
			members = append(members, redis.Z{
				Score:  float64(uncachedVacancy.CreatedAt.UnixNano()),
				Member: key,
			})
		}

		for _, index := range indexes {
			pipe.ZAddArgs(rdbCtx, index, redis.ZAddArgs{
				XX:      true,
				GT:      true,
				Members: members,
			})
			// pipe.SAdd(rdbCtx, index, keysCollection...) // interface members
		}

		pipe.ZInterStore(rdbCtx, indexConcat, &redis.ZStore{
			Keys:      indexes[:],
			Aggregate: "MAX",
		})
		pipe.Expire(rdbCtx, indexConcat, 30*time.Minute)

		if _, errExec := pipe.Exec(rdbCtx); errExec != nil {
			ctx.Set("CACHE_TYPE", "invalid-cache-aside")

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": true,
				"error":   errExec.Error(),
				"message": "there was an err query from the pipeline",
			})
			return
		}

		vacancies = append(vacancies, unCachedVacancies...) // collect vacancy
	} else {
		ctx.Set("CACHE_HIT", 1)
		ctx.Set("CACHE_MISS", 0)
	}

	ctx.Set("CACHE_TYPE", "cache-aside")

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
	var vacancies []VacancyProps
	read := gormDB.Raw(sql, queryParams...).Scan(&vacancies)
	if read.Error != nil {
		return nil, read.Error
	}

	return vacancies, nil
}
