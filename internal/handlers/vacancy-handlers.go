package handlers

import (
	"context"
	"encoding/json"
	initializer "go-cache-aside-service/init"
	"go-cache-aside-service/internal/helpers"
	"go-cache-aside-service/internal/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VacancyHandler struct {
}

func (handler *VacancyHandler) GetVacanciesCacheAside(ctx *gin.Context) {
	claim, exists := ctx.Get("claim")
	if !exists {
		ctx.JSON(http.StatusForbidden, gin.H{
			"status":  false,
			"error":   "key 'claim' doens't exist, probably authentication middleware",
			"message": "user has no credentials",
		})

		ctx.Abort()
		return
	}

	rdb, errRedis := initializer.GetRedisDB()
	if errRedis != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRedis.Error(),
			"message": "fail while gettting RDB instance",
		})

		ctx.Abort()
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

		ctx.Abort()
		return
	}

	errGetApplied := gormDB.Transaction(func(tx *gorm.DB) error {
		var candidateID string
		errGetCandidateID := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", claim).First(&candidateID).Error
		if errGetCandidateID != nil {
			applied = []string{}
		}

		getPipelines := tx.Model(&models.Pipeline{}).Select("vacancy_id").Where("candidate_id = ?", candidateID).Find(&applied)
		if getPipelines.RowsAffected == 0 {
			log.Println("candidate has zero apply")
		}
		return nil
	})

	if errGetApplied != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetApplied.Error(),
			"message": "error getting applied vacancies",
		})

		ctx.Abort()
		return
	}

	c := context.Background()
	cachedJSON, errCached := rdb.Get(c, claim.(string)).Result()
	if errCached != nil {
		log.Println("error cached \t:", errCached.Error())
		ctx.Set("CACHE_MISS", true) // set cache miss
		ctx.Set("CACHE_HIT", false) // set cache hit
	} else {
		log.Println("cached json \t:", cachedJSON)
		if errDecode := json.Unmarshal([]byte(cachedJSON), &vacancies); errDecode != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errDecode.Error(),
				"message": "error decoding cached JSON data",
			})

			ctx.Abort()
			return
		}
	}

	// sqlqueryCandidateProfile := `
	//   SELECT
	//     candidates.expertise,
	//     (
	//       SELECT
	//         educations.degree,
	//         educations.major
	//       FROM
	//         educations
	//       WHERE
	//         educations.candidate_id = candidates.id
	//       FOR JSON PATH
	//     ) AS educations,
	//     (
	//       SELECT
	//         skills.name
	//       FROM
	//         candidate_skills
	//         INNER JOIN skills ON skills.id = candidate_skills.skill_id
	//       WHERE
	//         candidate_skills.candidate_id = candidates.id
	//       FOR JSON PATH
	//     ) AS skills,
	//     (
	//       SELECT
	//         experiences.position,
	//         experiences.start_at,
	//         experiences.end_at
	//       FROM
	//         experiences
	//       WHERE
	//         experiences.candidate_id = candidates.id
	//       FOR JSON PATH
	//     ) AS experiences
	//   FROM
	//     candidates
	//   WHERE
	//     candidates.user_id = ?
	// `
	// soon should have line industry LIKE query
	sqlQueryVacancies := `
    SELECT TOP 5000
      vacancies.id,
      vacancies.position,
      vacancies.description,
      vacancies.qualification,
      vacancies.responsibility,
      vacancies.line_industry,
      vacancies.employee_type,
      vacancies.min_experience,
      vacancies.salary,
      vacancies.work_arrangement,
      vacancies.sla,
      vacancies.is_inactive,
      vacancies.created_at,
      employers.id,
      employers.name,
      employers.legal_name,
      employers.location,
      employers.profile_image_id
    FROM
      vacancies
      INNER JOIN employers ON employers.id = vacancies.employer_id 
    WHERE
      vacancies.is_inactive = ?
      AND
      vacancies.deleted_at IS NULL
    ORDER BY
      vacancies.created_at ASC
  `

	// var candidateProfiles map[string]interface{}

	getVacancies := gormDB.Raw(sqlQueryVacancies, false).Scan(&vacancies)

	if getVacancies.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   getVacancies.Error.Error(),
			"message": "there was an error with query builder",
		})

		ctx.Abort()
		return
	}

	// if educationsJSON, ok := candidateProfiles["educations"].(string); ok {
	// 	var educations []map[string]interface{}
	// 	if errDecode := json.Unmarshal([]byte(educationsJSON), &educations); errDecode != nil {
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{
	// 			"success": false,
	// 			"error":   errDecode.Error(),
	// 			"message": "fail unmarshall educations JSON",
	// 		})

	// 		ctx.Abort()
	// 		return
	// 	}

	// 	candidateProfiles["educations"] = educations
	// }
	// if skillsJSON, ok := candidateProfiles["skills"].(string); ok {
	// 	var skills []map[string]interface{}
	// 	if errDecode := json.Unmarshal([]byte(skillsJSON), &skills); errDecode != nil {
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{
	// 			"success": false,
	// 			"error":   errDecode.Error(),
	// 			"message": "fail unmarshall skills JSON",
	// 		})

	// 		ctx.Abort()
	// 		return
	// 	}

	// 	candidateProfiles["skills"] = skills
	// }
	// if experiencesJSON, ok := candidateProfiles["experiences"].(string); ok {
	// 	var experiences []map[string]interface{}
	// 	if errDecode := json.Unmarshal([]byte(experiencesJSON), &experiences); errDecode != nil {
	// 		ctx.JSON(http.StatusInternalServerError, gin.H{
	// 			"success": false,
	// 			"error":   errDecode.Error(),
	// 			"message": "fail unmarshall experiences JSON",
	// 		})

	// 		ctx.Abort()
	// 		return
	// 	}

	// 	candidateProfiles["experiences"] = experiences
	// }

	employerKeys := []string{
		"id",
		"name",
		"legal_name",
		"location",
		"profile_image_id",
	}
	for _, vacancy := range vacancies {
		employer := map[string]interface{}{}
		for _, key := range employerKeys {
			employer[key] = vacancy[key]
			delete(vacancy, key)
		}
		helpers.TransformsIdToPath([]string{"profile_image_id"}, employer)
		vacancy["employer"] = employer
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":   true,
		"applied":   applied,
		"vacancies": vacancies,
	})
}
