package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VacancyHandlers struct {
}

func (h *VacancyHandlers) GetVacancies(ctx *gin.Context) {
	/*
		API Requirements:
		1. Only retrieve 10 vacancies if it not logged in.
		2. Vacancies only can be applied to candidates.
		3. Duplicate applied is not allowed.
		4.
	*/
	authenticated, _ := ctx.Get("authenticated")
	identity, _ := ctx.Get("identity-access")
	// permissions, _ := ctx.Get("permissions")

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")

	pageQuery, _ := ctx.GetQuery("page")
	page, errConvPage := strconv.Atoi(pageQuery)
	if errConvPage != nil {
		page = 1
	}

	limitQuery, _ := ctx.GetQuery("limit")
	limit, errConvLimit := strconv.Atoi(limitQuery)
	if errConvLimit != nil {
		limit = 10
	}

	if (page > 1 || limit > 10) && !authenticated.(bool) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid access token",
			"message": "Log in to gain exclusive access to all job openings and discover your next opportunity.",
		})

		ctx.Abort()
		return
	}
	if limitQuery == "none" && authenticated.(bool) {
		limit = -1
	}

	offset := (limit * page) - limit

	keywordQuery, _ := ctx.GetQuery("keyword")
	keyword := fmt.Sprintf("%%%s%%", strings.TrimSpace(keywordQuery))
	locationQuery, _ := ctx.GetQuery("location")
	location := fmt.Sprintf("%%%s%%", strings.TrimSpace(locationQuery))

	var vacancies []map[string]interface{}
	var applied []string

	gormDB, _ := initializer.GetGorm()
	errGetVacancies := gormDB.Transaction(func(tx *gorm.DB) error {
		errGetVacanciesList := tx.Model(&models.Vacancy{}).Select([]string{
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
			Order("vacancies.created_at DESC").
			Where("vacancies.is_inactive = ? AND employers.location LIKE ? AND (vacancies.position LIKE ? OR vacancies.description LIKE ? OR vacancies.qualification LIKE ? OR vacancies.responsibility LIKE ?)", false, location, keyword, keyword, keyword, keyword).
			Limit(limit).Offset(offset).Find(&vacancies)

		if errGetVacanciesList.Error != nil {
			return errGetVacanciesList.Error
		} else if errGetVacanciesList.RowsAffected == 0 {
			return fmt.Errorf("%v rows, no data vacancies found", errGetVacanciesList.RowsAffected)
		}

		if authenticated.(bool) && identity.(string) == "candidate" {
			claims := ParseJWT(bearerToken)
			var candidateID string
			errCandidateID := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", claims.Id).First(&candidateID).Error
			if errCandidateID != nil {
				log.Println("candidate has not completed their profile as a candidate")
				applied = []string{}
				return nil
				// return errCandidateID
			}

			getPipelines := tx.Model(&models.Pipeline{}).Select([]string{"vacancy_id"}).Where("candidate_id = ?", candidateID).Find(&applied)
			if getPipelines.Error != nil {
				return getPipelines.Error
			}
		}
		return nil
	})

	if errGetVacancies != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetVacancies.Error(),
			"message": "error querying or record not found",
		})

		ctx.Abort()
		return
	}

	employerKeys := []string{
		"name",
		"legal_name",
		"location",
		"profile_image_id",
	}

	for _, vacancy := range vacancies {
		employer := map[string]interface{}{}
		for _, key := range employerKeys {
			employer[key] = vacancy[key]
		}
		TransformsIdToPath([]string{"profile_image_id"}, employer)
		vacancy["employer"] = employer
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"vacancies": vacancies,
			"applied":   applied,
		},
	})
}

func (h *VacancyHandlers) GetVacancyDetail(ctx *gin.Context) {
	authenticated, _ := ctx.Get("authenticated")
	identity, _ := ctx.Get("identity-access")
	// permissions, _ := ctx.Get("permissions")

	vacancyID := ctx.Param("id")
	if vacancyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing id param for Vacancy ID",
			"message": "please specify :id param for Vacancy ID",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")

	gormDB, _ := initializer.GetGorm()
	vacancy := map[string]interface{}{}
	applied := false
	errGetVacancy := gormDB.Transaction(func(tx *gorm.DB) error {
		errGet := tx.Model(&models.Vacancy{}).Select([]string{
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
		}).Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
			Where("vacancies.is_inactive = ? AND vacancies.id = ?", false, vacancyID).
			First(&vacancy).Error

		if errGet != nil {
			return errGet
		}

		if authenticated.(bool) && identity.(string) == "candidate" {
			claims := ParseJWT(bearerToken)

			var candidateID string
			errGetCandidateID := tx.Model(&models.Candidate{}).
				Select("id").
				Where("user_id = ?", claims.Id).
				First(&candidateID).Error
			if errGetCandidateID != nil {
				return errGetCandidateID
			}

			errAppliedCheck := tx.Model(&models.Pipeline{}).Select("1").Where("candidate_id = ? AND vacancy_id = ?", candidateID, vacancyID).First(&applied).Error
			if errAppliedCheck != nil {
				log.Println(errAppliedCheck.Error())
			}
		}

		return nil
	})

	if errGetVacancy != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetVacancy.Error(),
			"message": "this could be data not found in database",
		})

		ctx.Abort()
		return
	}

	employerKeys := []string{
		"name",
		"legal_name",
		"location",
		"profile_image_id",
	}

	employer := map[string]interface{}{}
	for _, key := range employerKeys {
		employer[key] = vacancy[key]
		delete(vacancy, key)
	}
	TransformsIdToPath([]string{"profile_image_id"}, employer)
	vacancy["employer"] = employer

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"vacancy": vacancy,
			"applied": applied,
		},
	})
}
