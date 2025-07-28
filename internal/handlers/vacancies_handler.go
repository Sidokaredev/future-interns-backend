package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"time"

	"future-interns-backend/internal/models"
	"future-interns-backend/internal/services/caching"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VacancyHandlers struct {
}

// variadic -> [time,  limit, offset, location, keyword, line industry, employee type]
func FallbackVacancies(queries ...any) (*caching.FallbackReturnValues, error) {
	var vacancies []map[string]any
	DB, _ := initializer.GetGorm()
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
		vacancies.employee_type LIKE ?`, false, queries[0], queries[3], queries[4], queries[4], queries[4], queries[4], queries[5], queries[6]).
		Order("vacancies.created_at DESC").
		Limit(queries[1].(int)).
		Offset(queries[2].(int)).
		Find(&vacancies)

	if getVacancies.Error != nil {
		return nil, getVacancies.Error
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

	return &caching.FallbackReturnValues{
		Data: vacancies,
		Indexes: []string{
			queries[3].(string),
			queries[4].(string),
			queries[5].(string),
			queries[6].(string),
		},
	}, nil
}

func CountAndApplied(userID string, queries []any, count *int64, applied *[]string) error {
	DB, err := initializer.GetGorm()
	if err != nil {
		return err
	}

	errTransacQuery := DB.Transaction(func(tx *gorm.DB) error {
		if userID == "" {
			applied = &[]string{}
		} else {
			var candidateID string
			fetchCandidate := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", userID).First(&candidateID)
			if fetchCandidate.Error != nil {
				log.Printf("candidate id: %v", fetchCandidate.Error.Error())
			} else {
				fetchApplied := tx.Model(&models.Pipeline{}).Select("vacancy_id").Where("candidate_id = ?", candidateID).Find(applied)
				if fetchApplied.Error != nil {
					return fetchApplied.Error
				}
			}
		}

		fetchCount := tx.Model(&models.Vacancy{}).
			Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
			Where(`vacancies.is_inactive = ? AND
		vacancies.created_at <= ? AND
		employers.location LIKE ? AND
		(vacancies.position LIKE ? OR
		vacancies.description LIKE ? OR
		vacancies.qualification LIKE ? OR
		vacancies.responsibility LIKE ?) AND
		vacancies.line_industry LIKE ? AND
		vacancies.employee_type LIKE ?`, false, queries[0], queries[3], queries[4], queries[4], queries[4], queries[4], queries[5], queries[6]).
			Order("vacancies.created_at DESC").Count(count)

		if fetchCount.Error != nil {
			return fetchCount.Error
		}

		return nil
	})

	if errTransacQuery != nil {
		return errTransacQuery
	}

	return nil
}

func (h *VacancyHandlers) GetVacancies(ctx *gin.Context) {
	// middleware:public_identity_check
	authenticated := ctx.GetBool("authenticated")
	identity := ctx.GetString("identity-access")
	userID := ctx.GetString("user-id")

	page, errConvPage := strconv.Atoi(ctx.Query("page"))
	if errConvPage != nil {
		page = 1
	}
	limit, errConvLimit := strconv.Atoi(ctx.Query("limit"))
	if errConvLimit != nil {
		limit = 10
	}

	if !authenticated && (page > 1 || limit > 10) { // pengguna yang belum terauthentikasi hanya tidak bisa melihat lebih dari 10 data
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "pengguna tidak ter-autentikasi",
			"message": "Masuk terlebih dahulu untuk melihat lebih banyak lowongan pekerjaan",
		})
		return
	}

	offset := (limit * page) - limit

	searchQueriesWilcards := []string{
		fmt.Sprintf("%%%s%%", strings.TrimSpace(ctx.Query("keyword"))),
		fmt.Sprintf("%%%s%%", strings.TrimSpace(ctx.Query("location"))),
		fmt.Sprintf("%%%s%%", strings.TrimSpace(ctx.Query("lineIndustry"))),
		fmt.Sprintf("%%%s%%", strings.TrimSpace(ctx.Query("employeeType"))),
	}

	tmScore, errParse := time.Parse(time.RFC3339, ctx.Query("time")) // mengambil nilai [time] sebagai UnixNano untuk Redis dan RFC3339 untuk MS SQL Server
	if errParse != nil {
		tmScore = time.Now()
	}

	DB, errDB := initializer.GetGorm()
	if errDB != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDB.Error(),
			"message": "Gagal memanggil GORM Instance connection",
		})
		return
	}

	var count int64
	var applied []string
	var vacancies []map[string]any
	errQuery := DB.Transaction(func(tx *gorm.DB) error {
		if authenticated && identity == "candidate" {
			var candidateID string
			errCandidate := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", userID).First(&candidateID).Error
			if errCandidate != nil {
				return errCandidate
			}

			errPipelines := tx.Model(&models.Pipeline{}).Select("vacancy_id").Where("candidate_id = ?", candidateID).Find(&applied).Error
			if errPipelines != nil {
				return errPipelines
			}
		} else {
			applied = []string{}
		}

		stateQuery := tx.Model(&models.Vacancy{}).Select([]string{
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
				tmScore.Format(time.RFC3339),
				searchQueriesWilcards[1],
				searchQueriesWilcards[0],
				searchQueriesWilcards[0],
				searchQueriesWilcards[0],
				searchQueriesWilcards[0],
				searchQueriesWilcards[2],
				searchQueriesWilcards[3],
			)

		errCount := stateQuery.Count(&count).Error
		if errCount != nil {
			return errCount
		}

		errVacancies := stateQuery.Order("vacancies.created_at DESC").Limit(limit).Offset(offset).Find(&vacancies).Error
		if errVacancies != nil {
			return errVacancies
		}

		return nil
	})

	if errQuery != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errQuery.Error(),
			"message": "Gagal mendapatkan data lowongan pekerjaan berdasarkan filter",
		})
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

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count":     count,
			"applied":   applied,
			"vacancies": vacancies,
			"last_time": last_time, // validate if len == 0
		},
	})
}

func (h *VacancyHandlers) GetVacancyDetail(ctx *gin.Context) {
	authenticated := ctx.GetBool("authenticated")
	identity := ctx.GetString("identity-access")
	userID := ctx.GetString("user-id")

	vacancyID := ctx.Param("id")
	if vacancyID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "parameter :id untuk vacancy_id tidak tersedia",
			"message": "Periksa kembali dan pastikan nilai parameter :id untuk vacancy_id",
		})
		return
	}

	DB, errDB := initializer.GetGorm()
	if errDB != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDB.Error(),
			"message": "Gagal memanggil GORM Instance connection",
		})
		return
	}
	vacancy := map[string]any{}
	var applied []string
	errQuery := DB.Transaction(func(tx *gorm.DB) error {
		errVacancy := tx.Model(&models.Vacancy{}).Select([]string{
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
		if errVacancy != nil {
			return errVacancy
		}

		if authenticated && identity == "candidate" {
			var candidateID string
			errGetCandidateID := tx.Model(&models.Candidate{}).
				Select("id").
				Where("user_id = ?", userID).
				First(&candidateID).Error
			if errGetCandidateID != nil {
				return errGetCandidateID
			}

			errAppliedCheck := tx.Model(&models.Pipeline{}).
				Select("vacancy_id").
				Where("candidate_id = ?", candidateID).
				Find(&applied).Error
			if errAppliedCheck != nil {
				log.Println(errAppliedCheck.Error())
			}
		} else {
			applied = []string{}
		}

		return nil
	})

	if errQuery != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errQuery.Error(),
			"message": "Gagal melakukan query untuk data pipelines dan vacancy",
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
		employer[key] = vacancy[key]
		delete(vacancy, key)
	}
	TransformsIdToPath([]string{"profile_image_id"}, employer)
	vacancy["employer"] = employer

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"applied": applied,
			"vacancy": vacancy,
		},
	})
}
