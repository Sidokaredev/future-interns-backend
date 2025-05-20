package handlers

import (
	"fmt"
	initializer "go-no-cache-service/init"
	"go-no-cache-service/internal/helpers"
	"go-no-cache-service/internal/models"
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

type VacancyHandler struct {
}

type VacancyProps struct {
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

/* Read Ops */
func (handler *VacancyHandler) GetVacanciesNoCache(ctx *gin.Context) {
	ctx.Set("CACHE_TYPE", "no-cache")

	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
	WorkArrangement, _ := ctx.GetQuery("workArrangement")

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	sqlQuery := `
		SELECT TOP(500)
			*
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getVacancies.Error.Error(),
			"message": "there something went wrong with sql query",
		})
		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

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

/* Write Ops */
func (handler *VacancyHandler) WriteVacanciesNoCache(ctx *gin.Context) {
	ctx.Set("CACHE_TYPE", "no-cache")

	var RequestBody []struct {
		ID              string    `json:"id"`
		Position        string    `json:"position"`
		Description     string    `json:"description"`
		Qualification   string    `json:"qualification"`
		Responsibility  string    `json:"responsibility"`
		LineIndustry    string    `json:"line_industry"`
		EmployeeType    string    `json:"employee_type"`
		MinExperience   string    `json:"min_experience"`
		Salary          uint      `json:"salary"`
		WorkArrangement string    `json:"work_arrangement"`
		SLA             int32     `json:"sla"`
		IsInactive      int       `json:"is_inactive"`
		EmployerID      string    `json:"employer_id"`
		CreatedAt       time.Time `json:"created_at"`
	}

	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-no-cache")
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "failed to bind collection of JSON fields",
		})

		ctx.Abort()
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-no-cache")
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
		vacanciesID = append(vacanciesID, vacancy.ID)
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
	}

	errStoreVacancies := gormDB.CreateInBatches(&m_vacancies, 100).Error
	if errStoreVacancies != nil {
		ctx.Set("CACHE_TYPE", "invalid-no-cache")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreVacancies.Error(),
			"message": fmt.Sprintf("failed storing %v vacancies", len(m_vacancies)),
		})

		ctx.Abort()
		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0) // write ops doesnt relate to cache hit or miss

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    vacanciesID,
	})
}

func (handler *VacancyHandler) ReadVacanciesNoCache(ctx *gin.Context) {
	ctx.Set("CACHE_TYPE", "no-cache-write")

	lineIndustryQuery, _ := ctx.GetQuery("lineIndustry")
	employeeTypeQuery, _ := ctx.GetQuery("employeeType")
	WorkArrangement, _ := ctx.GetQuery("workArrangement")

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	sqlQuery := `
		SELECT TOP(500)
			*
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
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getVacancies.Error.Error(),
			"message": "there something went wrong with sql query",
		})
		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 1)

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

func (handler *VacancyHandler) UpdateVacanciesNoCache(ctx *gin.Context) {
	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)

	var RequestBody []UpdateVacancyProps
	if errBind := ctx.ShouldBindJSON(&RequestBody); errBind != nil {
		ctx.Set("CACHE_TYPE", "invalid-no-cache")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check JSON fields",
		})
		return
	}

	gormDB, errGorm := initializer.GetMssqlDB()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-no-cache")

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
		ctx.Set("CACHE_TYPE", "invalid-no-cache")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCommit.Error(),
			"message": "fail updating data at database",
		})
		return
	}

	ctx.Set("CACHE_HIT", 0)
	ctx.Set("CACHE_MISS", 0)
	ctx.Set("CACHE_TYPE", "no-cache")

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    "successfully update all data",
	})
}
