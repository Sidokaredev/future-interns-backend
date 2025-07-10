package handler_candidates

import (
	"context"
	"errors"
	"fmt"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/models"
	"go-write-behind-service/internal/services/caches"
	helper_services "go-write-behind-service/internal/services/helpers"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PipelinesWithWriteBehind(gctx *gin.Context) {
	var body struct {
		VacancyID string `json:"vacancy_id"`
	}
	if errBindJSON := gctx.ShouldBindJSON(&body); errBindJSON != nil {
		gctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBindJSON.Error(),
			"message": "Gagal melakukan binding, periksa kembali fields anda",
		})
		return
	}

	// middleware:AuthorizationWithBearer
	userID := gctx.GetString("user-id")

	DB, errDB := initializer.GetMssqlDB()
	if errDB != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDB.Error(),
			"message": "Gagal memanggil GORM Instance",
		})
		return
	}

	var candidateID string
	errPipeline := DB.Transaction(func(tx *gorm.DB) error {
		errCandidate := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", userID).First(&candidateID).Error
		if errCandidate != nil {
			return errCandidate
		}

		var pipelineExist int64
		errExist := tx.Model(&models.Pipeline{}).Where("vacancy_id = ? AND candidate_id = ?", body.VacancyID, candidateID).Count(&pipelineExist).Error
		if errExist != nil {
			return errExist
		}

		if pipelineExist != 0 {
			return errors.New("lamaran pekerjaan hanya dapat dilakukan sekali")
		}
		return nil
	})
	if errPipeline != nil {
		gctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   errPipeline.Error(),
			"message": "Gagal memproses data lamaran pekerjaan",
		})
		return
	}

	// -> Executor Func
	var executorFunc caches.ExecutorCall = func(data any) error {
		pipelineMap, ok := data.(map[string]any)
		if !ok {
			return errors.New("exec: data Pipeline bukan sebuah map")
		}

		rdb, err := initializer.GetRedisDB()
		if err != nil {
			return err
		}

		log.Printf("Executor Func: assigning [%v] into list of job:writer@pipelines ...", pipelineMap["id"])
		ctx := context.Background()
		_, errLPush := rdb.LPush(ctx, caches.JobPipelinesKey, pipelineMap["id"]).Result()
		if errLPush != nil {
			return errLPush
		}

		persistKey := fmt.Sprintf("pipelines:%v", candidateID)
		log.Printf("Executor Func: removing expiration time sortedset to key -> %v", persistKey)
		_, errPersist := rdb.Persist(ctx, persistKey).Result()
		if errPersist != nil {
			return errPersist
		}

		return nil
	}

	pipelineUUID := helper_services.GenerateUUID(body.VacancyID)
	newPipeline := map[string]any{
		"id":           pipelineUUID,
		"candidate_id": candidateID,
		"vacancy_id":   body.VacancyID,
		"stage":        "Screening",
		"status":       "Applied",
		"created_at":   time.Now(),
	}

	wt := caches.NewWriteBehind(executorFunc)
	errSetCache := wt.SetCache(newPipeline, caches.CacheArgs{
		Indexes: []string{
			fmt.Sprintf("pipelines:%s", candidateID),
		},
		CacheProps: caches.CacheProps{
			KeyPropName:    "id",
			ScorePropName:  "created_at",
			ScoreType:      "time.Time",
			MemberPropName: "id",
		},
	})
	if errSetCache != nil {
		gctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errSetCache.Error(),
			"message": "Gagal menerapkan cache Pipelines dengan write-through",
		})
		return
	}

	gctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Berhasil mengirimkan lamaran pekerjaan",
		},
	})
}
