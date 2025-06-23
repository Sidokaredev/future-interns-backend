package cache

import (
	"context"
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func TransformsIdToPath(targets []string, record interface{}) {
	switch recordTyped := record.(type) {
	case []map[string]interface{}:
		for index, data := range recordTyped {
			for _, target := range targets {
				newKey := strings.Replace(target, "id", "path", 1)
				var pathType string
				if strings.Contains(target, "image") {
					pathType = "images"
				} else {
					pathType = "documents"
				}
				if value, exists := data[target]; exists {
					v := reflect.ValueOf(value)
					if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
						if !v.IsNil() {
							value = v.Elem().Interface()
						} else {
							value = nil
						}
					}

					if value != nil && value != 0 {
						recordTyped[index][newKey] = fmt.Sprintf("/%s/%v", pathType, value)
					} else {
						recordTyped[index][newKey] = nil
					}
					delete(recordTyped[index], target)
				}
			}
		}
	case map[string]interface{}:
		for _, target := range targets {
			newKey := strings.Replace(target, "id", "path", 1)
			var pathType string
			if strings.Contains(target, "image") {
				pathType = "images"
			} else {
				pathType = "documents"
			}
			if value, exists := recordTyped[target]; exists {
				v := reflect.ValueOf(value)
				if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
					if !v.IsNil() {
						value = v.Elem().Interface()
					} else {
						value = nil
					}
				}
				if value != nil && value != 0 {
					recordTyped[newKey] = fmt.Sprintf("/%s/%v", pathType, value)
				} else {
					recordTyped[newKey] = nil
				}
				delete(recordTyped, target)
			}
		}
	}
}

type TokenClaims struct {
	Id string
	jwt.RegisteredClaims
}

func ParseJWT(bearer string) *TokenClaims {
	secretKey := []byte(viper.GetString("authorization.jwt.secretKey"))
	token, _ := jwt.ParseWithClaims(bearer, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secretKey, nil
	})

	claims, _ := token.Claims.(*TokenClaims)
	return claims
}

type Cache struct {
	Limit        int64
	Page         int64
	Timestamp    string
	Keyword      string
	Location     string
	LineIndustry string
	EmployeeType string
}

func (c *Cache) CacheAside(ctx *gin.Context) {
	authenticated, _ := ctx.Get("authenticated")
	identity, _ := ctx.Get("identity-access")
	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")

	rdb, errRdb := initializer.GetRedisDB()
	if errRdb != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errRdb.Error(),
			"message": "fail getting Redis instance connection",
		})
		return
	}

	gormDB, errGorm := initializer.GetGorm()
	if errGorm != nil {
		ctx.Set("CACHE_TYPE", "invalid-cache-aside")

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGorm.Error(),
			"message": "fail getting GORM instance connection",
		})
		return
	}

	intersectionKey := ""
	zinterkeys := []string{}
	v := reflect.ValueOf(*c)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Name
		if fieldName == "Timestamp" {
			continue // skip field Timestamp
		}

		if v.Field(i).Kind() == reflect.String {
			if v.Field(i).String() == "" {
				continue
			}

			intersectionKey += fmt.Sprintf("%s,", v.Field(i).String())
			zinterkeys = append(zinterkeys, fmt.Sprintf("index:%s", v.Field(i).String()))
		}
	}
	intersectionKey = strings.TrimSuffix(intersectionKey, ",")
	t, err := time.Parse(time.RFC3339, c.Timestamp)
	if err != nil {
		fmt.Println("Error parsing time:", err)
		return
	}

	rdbCtx := context.Background()
	ZInter, errZInter := rdb.ZRevRangeByScore(rdbCtx, intersectionKey, &redis.ZRangeBy{
		Min:    "-inf",
		Max:    strconv.FormatInt(t.UnixNano(), 10),
		Offset: ((c.Limit * c.Page) - c.Limit),
		Count:  c.Limit,
	}).Result()
	if errZInter != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errZInter.Error(),
			"message": fmt.Sprintf("kegagalan intersection: %s", intersectionKey),
		})
		return
	}

	if len(ZInter) == 0 {
		var vacancies []map[string]interface{}
		var applied []string
		var vacanciesCount int64

		queryParams := []interface{}{
			false,
			fmt.Sprintf("%%%s%%", c.Location),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.LineIndustry),
			fmt.Sprintf("%%%s%%", c.EmployeeType),
		}

		errGetVacancies := gormDB.Transaction(func(tx *gorm.DB) error {
			errCountVacancies := tx.Model(&models.Vacancy{}).
				Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
				Order("vacancies.created_at DESC").
				Where(`vacancies.is_inactive = ? AND
        employers.location LIKE ? AND
        (vacancies.position LIKE ? OR
        vacancies.description LIKE ? OR
        vacancies.qualification LIKE ? OR
        vacancies.responsibility LIKE ?) AND
        vacancies.line_industry LIKE ? AND
        vacancies.employee_type LIKE ?`, queryParams...).
				Count(&vacanciesCount).Error
			if errCountVacancies != nil {
				return errCountVacancies
			}
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
				Where(`vacancies.is_inactive = ? AND
               employers.location LIKE ? AND
               (vacancies.position LIKE ? OR
               vacancies.description LIKE ? OR
               vacancies.qualification LIKE ? OR
               vacancies.responsibility LIKE ?) AND
               vacancies.line_industry LIKE ? AND
               vacancies.employee_type LIKE ?`, queryParams...).
				Limit(int(c.Limit)).
				Offset(int((c.Limit * c.Page) - c.Limit)).
				Find(&vacancies)

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

		pipe := rdb.Pipeline()
		members := []redis.Z{}
		for _, vacancy := range vacancies {
			hfields := []string{}

			for key := range vacancy {
				hfields = append(hfields, key)
			}
			key := fmt.Sprintf("CA:%s", vacancy["id"])
			// t, err := time.Parse(time.RFC3339, vacancy["created_at"].(string))
			// if err != nil {
			// 	fmt.Println("Error parsing time:", err)
			// 	return
			// }
			members = append(members, redis.Z{
				Score:  float64(vacancy["created_at"].(time.Time).UnixNano()),
				Member: key,
			})
			pipe.HSet(rdbCtx, key, vacancy)
			pipe.HExpire(rdbCtx, key, 30*time.Minute, hfields...)

			employer := map[string]interface{}{}
			for _, key := range employerKeys {
				employer[key] = vacancy[key]
			}
			TransformsIdToPath([]string{"profile_image_id"}, employer)
			vacancy["employer"] = employer
		}
		for _, ZKey := range zinterkeys {
			pipe.ZAddArgs(rdbCtx, ZKey, redis.ZAddArgs{
				GT:      true,
				Members: members,
			})
		}
		pipe.ZInterStore(rdbCtx, intersectionKey, &redis.ZStore{
			Keys:      zinterkeys[:],
			Aggregate: "MAX",
		})
		pipe.Expire(rdbCtx, intersectionKey, 30*time.Minute)
		_, errExec := pipe.Exec(rdbCtx)
		if errExec != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": true,
				"error":   errExec.Error(),
				"message": "kegagalan saat melakukan caching:cache-aside",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"status":    "cache missed",
				"vacancies": vacancies,
				"applied":   applied,
				"count":     vacanciesCount,
			},
		})
		return
	}

	// GET HASH BY ZINTER MEMBER
	// QUERY AND SET MISSED CACHED
	var cachedVacancies []map[string]interface{}
	var applied []string
	var vacanciesCount int64
	unCachedVacancyKeys := []string{}
	for _, key := range ZInter {
		cmd := rdb.HGetAll(rdbCtx, key)
		if len(cmd.Val()) == 0 {
			unCachedVacancyKeys = append(unCachedVacancyKeys, strings.TrimPrefix(key, "CA:"))
		} else {
			vacancyCached, errCached := cmd.Result()
			if errCached != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   errCached.Error(),
					"message": fmt.Sprintf("gagal mendapatkan hasil Hash, key: %s", key),
				})
				return
			}
			vacancyCachedInterface := make(map[string]interface{})
			for key, value := range vacancyCached {
				vacancyCachedInterface[key] = value
			}

			employerKeys := []string{
				"name",
				"legal_name",
				"location",
				"profile_image_id",
			}
			employer := map[string]interface{}{}
			for _, key := range employerKeys {
				employer[key] = vacancyCachedInterface[key]
			}
			TransformsIdToPath([]string{"profile_image_id"}, employer)
			vacancyCachedInterface["employer"] = employer

			cachedVacancies = append(cachedVacancies, vacancyCachedInterface) // collect data
		}
	}

	log.Printf("UNCACHED IN TOTAL: %v of %v", len(unCachedVacancyKeys), len(ZInter))

	if len(unCachedVacancyKeys) > 0 {
		var uncachedVacancies []map[string]interface{}
		queryParams := []interface{}{
			false,
			fmt.Sprintf("%%%s%%", c.Location),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.LineIndustry),
			fmt.Sprintf("%%%s%%", c.EmployeeType),
		}

		errGetVacancies := gormDB.Transaction(func(tx *gorm.DB) error {
			errCountVacancies := tx.Model(&models.Vacancy{}).
				Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
				Order("vacancies.created_at DESC").
				Where(`vacancies.is_inactive = ? AND
        employers.location LIKE ? AND
        (vacancies.position LIKE ? OR
        vacancies.description LIKE ? OR
        vacancies.qualification LIKE ? OR
        vacancies.responsibility LIKE ?) AND
        vacancies.line_industry LIKE ? AND
        vacancies.employee_type LIKE ?`, queryParams...).
				Count(&vacanciesCount).Error
			if errCountVacancies != nil {
				return errCountVacancies
			}
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
				Where(`vacancies.id IN ?`, unCachedVacancyKeys).
				Find(&uncachedVacancies)

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

		pipe := rdb.Pipeline()
		members := []redis.Z{}
		for _, vacancy := range uncachedVacancies {
			hfields := []string{}

			for key := range vacancy {
				hfields = append(hfields, key)
			}
			key := fmt.Sprintf("CA:%s", vacancy["id"])
			// t, err := time.Parse(time.RFC3339, vacancy["created_at"].(string))
			// if err != nil {
			// 	fmt.Println("Error parsing time:", err)
			// 	return
			// }
			members = append(members, redis.Z{
				Score:  float64(vacancy["created_at"].(time.Time).UnixNano()),
				Member: key,
			})
			pipe.HSet(rdbCtx, key, vacancy)
			pipe.HExpire(rdbCtx, key, 30*time.Minute, hfields...)

			employer := map[string]interface{}{}
			for _, key := range employerKeys {
				employer[key] = vacancy[key]
			}
			TransformsIdToPath([]string{"profile_image_id"}, employer)
			vacancy["employer"] = employer

			cachedVacancies = append(cachedVacancies, vacancy) // collect data
		}
		for _, ZKey := range zinterkeys {
			pipe.ZAddArgs(rdbCtx, ZKey, redis.ZAddArgs{
				GT:      true,
				Members: members,
			})
		}
		pipe.ZInterStore(rdbCtx, intersectionKey, &redis.ZStore{
			Keys:      zinterkeys[:],
			Aggregate: "MAX",
		})
		pipe.Expire(rdbCtx, intersectionKey, 30*time.Minute)
		_, errExec := pipe.Exec(rdbCtx)
		if errExec != nil {

			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": true,
				"error":   errExec.Error(),
				"message": "kegagalan saat melakukan caching:cache-aside",
			})
			return
		}
	} else {
		queryParams := []interface{}{
			false,
			fmt.Sprintf("%%%s%%", c.Location),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.Keyword),
			fmt.Sprintf("%%%s%%", c.LineIndustry),
			fmt.Sprintf("%%%s%%", c.EmployeeType),
		}

		errGetVacancies := gormDB.Transaction(func(tx *gorm.DB) error {
			errCountVacancies := tx.Model(&models.Vacancy{}).
				Joins("INNER JOIN employers ON employers.id = vacancies.employer_id").
				Order("vacancies.created_at DESC").
				Where(`vacancies.is_inactive = ? AND
        employers.location LIKE ? AND
        (vacancies.position LIKE ? OR
        vacancies.description LIKE ? OR
        vacancies.qualification LIKE ? OR
        vacancies.responsibility LIKE ?) AND
        vacancies.line_industry LIKE ? AND
        vacancies.employee_type LIKE ?`, queryParams...).
				Count(&vacanciesCount).Error
			if errCountVacancies != nil {
				return errCountVacancies
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
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":    "cache hitted",
			"vacancies": cachedVacancies,
			"applied":   applied,
			"count":     vacanciesCount,
		},
	})
	ctx.Abort()
}
