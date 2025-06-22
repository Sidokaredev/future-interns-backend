package handlers

import (
	"context"
	"fmt"
	initializer "go-write-through-service/init"
	"go-write-through-service/internal/models"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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

type RequestBodyBind struct {
	Intersection string               `json:"intersection" binding:"required"`
	Data         []UpdateVacancyProps `json:"data" binding:"required"`
}

func (h *VacancyHandler) WriteThroughService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []WriteVacancyProps

	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil { // binding request body
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "failed to bind collection of JSON fields",
		})

		return
	}

	rdb, errRedis := initializer.GetRedisDB() // redis conn
	if errRedis != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRedis.Error(),
			"message": "fail while gettting RDB instance",
		})

		return
	}
	gormDB, errGorm := initializer.GetMssqlDB() // mssql conn
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

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

	indexes := []string{}
	indexesMap := map[string]bool{}

	ctxRdb := context.Background()
	ttl := 30 * time.Minute
	members := []redis.Z{}
	pipe := rdb.Pipeline()

	for _, vacancy := range RequestBody { // loop over request body
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

		hashFields := []string{}

		props := reflect.TypeOf(vacancy)
		for idx := 0; idx < props.NumField(); idx++ {
			structTag := props.Field(idx).Tag.Get("json")
			hashFields = append(hashFields, structTag)
		}

		key := fmt.Sprintf("WT:%s", vacancy.ID)
		members = append(members, redis.Z{
			Score:  float64(vacancy.CreatedAt.UnixNano()),
			Member: key,
		})
		pipe.HSet(ctxRdb, key, vacancy)
		pipe.HExpire(ctxRdb, key, ttl, hashFields...)

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

	indexConcat := ""
	for i, ZKey := range indexes {
		if i == (len(indexes) - 1) {
			indexConcat += strings.TrimPrefix(ZKey, "index:")
		} else {
			indexConcat += fmt.Sprintf("%s,", strings.TrimPrefix(ZKey, "index:"))
		}

		pipe.ZAddArgs(ctxRdb, ZKey, redis.ZAddArgs{
			GT:      true,
			Members: members,
		})
	}

	pipe.ZInterStore(ctxRdb, indexConcat, &redis.ZStore{
		Keys:      indexes[:],
		Aggregate: "MAX",
	})
	pipe.Expire(ctxRdb, indexConcat, ttl)

	if _, errExecute := pipe.Exec(ctxRdb); errExecute != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errExecute.Error(),
			"message": "fail when executing SADD operation",
		})
		return
	}

	errStoreVacancies := gormDB.CreateInBatches(&m_vacancies, 100).Error
	if errStoreVacancies != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreVacancies.Error(),
			"message": fmt.Sprintf("failed storing %v vacancies", len(m_vacancies)),
		})

		return
	}

	ctx.Set("CACHE_HIT", 1)
	ctx.Set("CACHE_TYPE", "write-through")

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    vacanciesID,
	})
} // in use

func (h *VacancyHandler) UpdateWriteThroughService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody RequestBodyBind
	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check JSON fields",
		})
		return
	}

	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	ctxRdb := context.Background()
	ttl := 30 * time.Minute
	pipe := rdb.Pipeline()

	tx := gormDB.Begin()

	for _, vacancy := range RequestBody.Data { // data: []UpdateVacancyProps
		props := reflect.TypeOf(vacancy)
		values := reflect.ValueOf(vacancy)

		mappedColumns := map[string]interface{}{}

		hashFields := []interface{}{}
		fields := []string{}
		for idx := 0; idx < props.NumField(); idx++ {
			structTag := props.Field(idx).Tag.Get("json")
			if values.Field(idx).Kind() == reflect.Ptr && values.Field(idx).IsNil() {
				continue
			}
			if structTag == "id" && values.Field(idx).Kind() == reflect.String {
				continue
			}
			mappedColumns[structTag] = values.Field(idx).Elem().Interface()

			hashFields = append(hashFields, structTag, values.Field(idx).Elem().Interface())
			fields = append(fields, structTag)
		}

		key := fmt.Sprintf("WT:%s", vacancy.ID)
		pipe.HSet(ctxRdb, key, hashFields...)
		pipe.HExpire(ctxRdb, key, ttl, fields...)

		update := tx.Model(&models.Vacancy{Id: vacancy.ID}).Updates(mappedColumns)
		if update.RowsAffected == 0 {
			tx.Rollback()
		}
	}

	pipe.Expire(ctxRdb, RequestBody.Intersection, ttl)

	_, errExecute := pipe.Exec(ctxRdb)
	if errExecute != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errExecute.Error(),
			"message": "fail updating hashes",
		})
		return
	}

	if errCommit := tx.Commit().Error; errCommit != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCommit.Error(),
			"message": "fail updating data at database",
		})
		return
	}

	ctx.Set("CACHE_HIT", 1)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "write-through")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    "successfully update all data",
	})
}

func (handler *VacancyHandler) ReadWriteThroughService(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
	WorkArrangement, _ := ctx.GetQuery("workArrangement")

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")
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
	queryParams := []interface{}{
		fmt.Sprintf("%%%s%%", lineIndustryQuery),
		fmt.Sprintf("%%%s%%", employeeTypeQuery),
		fmt.Sprintf("%%%s%%", WorkArrangement),
	}
	vacancies := []map[string]interface{}{}
	getVacancies := gormDB.Raw(sqlQuery, queryParams...).Scan(&vacancies)
	if getVacancies.Error != nil {
		ctx.Set("CACHE_TYPE", "invalid-write-through")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getVacancies.Error.Error(),
			"message": "there something went wrong with sql query",
		})
		return
	}

	// employerKeys := []string{
	// 	"employer_id",
	// 	"name",
	// 	"legal_name",
	// 	"location",
	// 	"profile_image_id",
	// }
	// for _, vacancy := range vacancies {
	// 	employer := map[string]interface{}{}
	// 	for _, key := range employerKeys {
	// 		if key == "employer_id" {
	// 			employer["id"] = vacancy[key]
	// 			delete(vacancy, key)
	// 			continue
	// 		}
	// 		employer[key] = vacancy[key]
	// 		delete(vacancy, key)
	// 	}
	// 	helpers.TransformsIdToPath([]string{"profile_image_id"}, employer)
	// 	vacancy["employer"] = employer
	// }

	ctx.Set("CACHE_TYPE", "write-through")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    vacancies,
	})
} // in use

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
