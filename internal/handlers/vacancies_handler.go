package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"regexp"
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

func CountAndApplied(candidateID string, queries []any, count *int64, applied *[]string) error {
	DB, err := initializer.GetGorm()
	if err != nil {
		return err
	}

	errTransacQuery := DB.Transaction(func(tx *gorm.DB) error {
		if candidateID == "" {
			applied = &[]string{}
		} else {
			fetchApplied := tx.Model(&models.Pipeline{}).Select("vacancy_id").Where("candidate_id = ?", candidateID).Find(applied)
			if fetchApplied.Error != nil {
				return fetchApplied.Error
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
	candidateID := ctx.GetString("candidate-id")

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
			"error":   "invalid access token",
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

	queryValues := []string{}
	re := regexp.MustCompile("%(.+?)%") // mengambil search query parameter yang hanya memiliki nilai [%SEARCH%]
	for _, val := range searchQueriesWilcards {
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

	cacheQuery := ctx.Query("cache")

	var scoreMax string
	tmScore, errParse := time.Parse(time.RFC3339, ctx.Query("time")) // mengambil nilai [time] sebagai UnixNano untuk Redis dan RFC3339 untuk MS SQL Server
	if errParse != nil {
		now := time.Now().UnixNano()
		scoreMax = strconv.Itoa(int(now))
	} else {
		scoreMax = strconv.Itoa(int(tmScore.UnixNano()))
	}

	// how if query("time") empty? tmScore?

	var count int64
	var applied []string
	errCountApplied := CountAndApplied(candidateID, []any{ // mengambil jumlah data berdasarkan search query parameter dan ID lowongan pekerjaan yang telah di apply (hanya kandidat)
		tmScore.Format(time.RFC3339),
		limit, offset,
		searchQueriesWilcards[0],
		searchQueriesWilcards[1],
		searchQueriesWilcards[2],
		searchQueriesWilcards[3],
	}, &count, &applied)
	if errCountApplied != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCountApplied.Error(),
			"message": "Gagal melakukan query [count, applied]",
		})
		return
	}

	cache := caching.NewCacheStrategy(cacheQuery, FallbackVacancies, []any{ // membuat instance cache berdasarkan pola [cache-aside, read-through] :!READ ONLY
		tmScore.Format(time.RFC3339),
		limit, offset,
		searchQueriesWilcards[0],
		searchQueriesWilcards[1],
		searchQueriesWilcards[2],
		searchQueriesWilcards[3],
	})

	var vacancies []map[string]any
	errCache := cache.GetCache(caching.GetCacheArgs{ // mengambil nilai data yang ada pada cache, dan menjalankan FallbackCall jika cache kosong
		Intersection: intersectionKey,
		Min:          "-inf",
		Max:          scoreMax,
		Count:        int64(limit),
		Offset:       int64(offset),
		Indexes:      queryValues,
		CacheArgs: caching.CacheProps{
			KeyPropName:    "id",
			ScorePropName:  "created_at",
			ScoreType:      "time.Time",
			MemberPropName: "id",
		},
	}, &vacancies)

	if errCache != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCache.Error(),
			"message": fmt.Sprintf("Gagal melakukan cache dengan pola [:%v]", cacheQuery),
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"count":     count,
			"applied":   applied,
			"vacancies": vacancies,
			"last_time": vacancies[len(vacancies)-1]["created_at"], // validate if len == 0
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
