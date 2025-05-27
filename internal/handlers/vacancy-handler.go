package handlers

import (
	"context"
	"fmt"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/helpers"
	"go-write-behind-service/internal/scheduler"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VacancyHandler struct {
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
type ReadVacancyProps struct {
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

func (h *VacancyHandler) StatusScheduler(ctx *gin.Context) {
	status := scheduler.GetFlusherStatus()

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

func (handler *VacancyHandler) WriteBehindService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []WriteVacancyProps

	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "failed to bind collection of JSON fields",
		})
		return
	}

	rdb, errRedis := initializer.GetRedisDB()
	if errRedis != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRedis.Error(),
			"message": "fail while gettting RDB instance",
		})
		return
	}

	vacanciesID := []string{}

	indexes := []string{}
	indexesMap := map[string]bool{}

	ctxRdb := context.Background()
	keysCollection := []string{}

	pipe := rdb.Pipeline()

	for _, vacancy := range RequestBody {
		if !indexesMap[vacancy.LineIndustry] {
			indexesMap[vacancy.LineIndustry] = true
			indexes = append(indexes, fmt.Sprintf("index:%s", vacancy.LineIndustry))
		}
		if !indexesMap[vacancy.EmployeeType] {
			indexesMap[vacancy.EmployeeType] = true
			indexes = append(indexes, fmt.Sprintf("index:%s", vacancy.EmployeeType))
		}
		if !indexesMap[vacancy.WorkArrangement] {
			indexesMap[vacancy.WorkArrangement] = true
			indexes = append(indexes, fmt.Sprintf("index:%s", vacancy.WorkArrangement))
		}

		key := fmt.Sprintf("WB:%s", vacancy.ID)
		keysCollection = append(keysCollection, key)
		pipe.HSet(ctxRdb, key, vacancy)

		vacanciesID = append(vacanciesID, vacancy.ID)
	}

	for _, index := range indexes {
		pipe.SAdd(ctxRdb, index, keysCollection)
	}

	pipe.LPush(ctxRdb, "job:writer", keysCollection)

	_, errExecute := pipe.Exec(ctxRdb)
	if errExecute != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errExecute.Error(),
			"message": "fail when executing SADD operation",
		})
		return
	}

	ctx.Set("CACHE_HIT", 1)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "write-behind")

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    vacanciesID,
	})
}

func (handler *VacancyHandler) UpdateWriteBehindService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []UpdateVacancyProps
	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check JSON fields",
		})
		return
	}

	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	}

	ctxRdb := context.Background()
	pipe := rdb.Pipeline()
	keysCollection := []string{}

	counter := 0
	for _, vacancy := range RequestBody {
		values := reflect.ValueOf(vacancy)
		props := reflect.TypeOf(vacancy)
		hfvalues := []interface{}{}
		for idx := 0; idx < props.NumField(); idx++ {
			structTag := props.Field(idx).Tag.Get("json")
			if values.Field(idx).Kind() == reflect.Ptr && values.Field(idx).IsNil() {
				continue
			}
			if structTag == "id" && values.Field(idx).Kind() == reflect.String {
				continue
			}

			value := values.Field(idx).Elem().Interface()
			hfvalues = append(hfvalues, structTag, value)
		}

		key := fmt.Sprintf("WB:%s", vacancy.ID)
		keysCollection = append(keysCollection, key)
		pipe.HSet(ctxRdb, key, hfvalues...)

		counter += 1
	}

	pipe.LPush(ctxRdb, "job:updater", keysCollection)
	_, errExec := pipe.Exec(ctxRdb)
	if errExec != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errExec.Error(),
			"message": "faile executing pipeline",
		})
		return
	}

	ctx.Set("CACHE_HIT", 1)
	ctx.Set("CACHE_TYPE", "write-behind")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    counter,
	})
}

// func (handler *VacancyHandler) ReadWriteBehindService(ctx *gin.Context) {
// 	ctx.Set("CACHE_HIT", 0)
// 	ctx.Set("CACHE_MISS", 0)

// 	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
// 	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
// 	workArrangement, _ := ctx.GetQuery("workArrangement")

// 	rdb, errRdb := initializer.GetRedisDB()
// 	if errRdb != nil {
// 		ctx.Set("CACHE_TYPE", "invalid-write-behind")

// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"success": false,
// 			"error":   errRdb.Error(),
// 			"message": "fail getting Redis instance connection",
// 		})
// 		return
// 	}

// 	gormDB, errGorm := initializer.GetMssqlDB()
// 	if errGorm != nil {
// 		ctx.Set("CACHE_TYPE", "invalid-write-behind")

// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"success": false,
// 			"error":   errGorm.Error(),
// 			"message": "fail getting GORM instance connection",
// 		})
// 		return
// 	}

// 	indexes := [3]string{
// 		fmt.Sprintf("index:%s", lineIndustryQuery),
// 		fmt.Sprintf("index:%s", employeeTypeQuery),
// 		fmt.Sprintf("index:%s", workArrangement),
// 	}

// 	rdbCtx := context.Background()
// 	sizeInterCard, errInterCard := rdb.SInterCard(rdbCtx, 500, indexes[0], indexes[1], indexes[2]).Result()
// 	if errInterCard != nil {
// 		ctx.Set("CACHE_TYPE", "invalid-write-behind")

// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"success": false,
// 			"error":   errInterCard.Error(),
// 			"message": "fail counting intersection size",
// 		})
// 		return
// 	}

// 	log.Printf("%s&%s&%s", indexes[0], indexes[1], indexes[2])
// 	if sizeInterCard < 500 {
// 		log.Println("reading from database...")
// 		ctx.Set("CACHE_HIT", 0)
// 		ctx.Set("CACHE_MISS", 1)

// 		queryParams := []interface{}{
// 			fmt.Sprintf("%%%s%%", lineIndustryQuery),
// 			fmt.Sprintf("%%%s%%", employeeTypeQuery),
// 			fmt.Sprintf("%%%s%%", workArrangement),
// 		}
// 		vacancies, errRead := ReadFromDatabase(gormDB, queryParams...)
// 		if errRead != nil {
// 			ctx.Set("CACHE_TYPE", "invalid-write-behind")

// 			ctx.JSON(http.StatusInternalServerError, gin.H{
// 				"success": false,
// 				"error":   errRead.Error(),
// 				"message": "fail reading from database source",
// 			})
// 			return
// 		}

// 		ctx.Set("CACHE_TYPE", "write-behind")

// 		ctx.JSON(http.StatusOK, gin.H{
// 			"success": true,
// 			"data":    vacancies,
// 		})
// 		return
// 	}

// 	sInter, errSInter := rdb.SInter(rdbCtx, indexes[0], indexes[1], indexes[2]).Result()
// 	if errSInter != nil {
// 		ctx.Set("CACHE_TYPE", "invalid-write-behind")

// 		ctx.JSON(http.StatusInternalServerError, gin.H{
// 			"success": false,
// 			"error":   errSInter.Error(),
// 			"message": "fail getting intersection values",
// 		})
// 		return
// 	}

// 	vacancies := []ReadVacancyProps{}
// 	for _, key := range sInter {
// 		var vacancy ReadVacancyProps

// 		cmd := rdb.HGetAll(rdbCtx, key)
// 		if errScanHash := cmd.Scan(&vacancy); errScanHash != nil {
// 			ctx.Set("CACHE_TYPE", "invalid-write-behind")

// 			ctx.JSON(http.StatusInternalServerError, gin.H{
// 				"success": false,
// 				"error":   errScanHash.Error(),
// 				"message": "fail scanning hash field-value",
// 			})
// 			return
// 		}

// 		vacancies = append(vacancies, vacancy)
// 	}

// 	ctx.Set("CACHE_HIT", 1)
// 	ctx.Set("CACHE_MISS", 0)
// 	ctx.Set("CACHE_TYPE", "write-behind")

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"success": true,
// 		"data":    vacancies,
// 	})
// }

func (handler *VacancyHandler) ReadWriteBehindService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
	WorkArrangement, _ := ctx.GetQuery("workArrangement")

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	sqlQuery := `
		SELECT TOP(500)
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
		ORDER BY
			created_at DESC
	`
	vacancies := []map[string]interface{}{}
	getVacancies := gormDB.Raw(sqlQuery,
		fmt.Sprintf("%%%s%%", lineIndustryQuery),
		fmt.Sprintf("%%%s%%", employeeTypeQuery),
		fmt.Sprintf("%%%s%%", WorkArrangement),
	).Scan(&vacancies)
	if getVacancies.Error != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-behind")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getVacancies.Error.Error(),
			"message": "there something went wrong with sql query",
		})
		return
	}

	ctx.Set("CACHE_TYPE", "write-behind")

	employerKeys := []string{
		"employer_id",
		"name",
		"legal_name",
		"location",
		"profile_image_id",
	}
	for _, vacancy := range vacancies {
		employer := map[string]interface{}{}
		for _, key := range employerKeys {
			if key == "employer_id" {
				employer["id"] = vacancy[key]
				delete(vacancy, key)
				continue
			}
			employer[key] = vacancy[key]
			delete(vacancy, key)
		}
		helpers.TransformsIdToPath([]string{"profile_image_id"}, employer)
		vacancy["employer"] = employer
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    vacancies,
	})
}

func ReadFromDatabase(gormDB *gorm.DB, queryParams ...interface{}) ([]ReadVacancyProps, error) {
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
	var vacancies []ReadVacancyProps
	read := gormDB.Raw(sql, queryParams...).Scan(&vacancies)
	if read.Error != nil {
		return nil, read.Error
	}

	return vacancies, nil
}

func (handler *VacancyHandler) BackgroundJobStatus(ctx *gin.Context) {
	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	}

	ctxRdb := context.Background()
	writerJobSize, errWriter := rdb.LLen(ctxRdb, "job:writer").Result()
	if errWriter != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errWriter.Error(),
			"message": "fail getting size of job:writer",
		})
		return
	}
	updaterJobSize, errUpdater := rdb.LLen(ctxRdb, "job:updater").Result()
	if errUpdater != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errUpdater.Error(),
			"message": "fail getting size of job:updater",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    writerJobSize + updaterJobSize,
	})
}
