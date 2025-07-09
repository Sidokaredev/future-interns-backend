package handler_candidates

import (
	"errors"
	"fmt"
	initializer "go-write-through-service/init"
	"go-write-through-service/internal/models"
	"go-write-through-service/internal/services/caches"
	helper_services "go-write-through-service/internal/services/helpers"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func PipelinesWithWriteThrough(gctx *gin.Context) {
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

	var executorFunc caches.ExecutorCall = func(data any, hash caches.HashCollection, sortedset *caches.SortedSetCollection) error {
		pipelineMap, ok := data.(map[string]any)
		if !ok {
			return errors.New("executor args bukan map[string]any")
		}

		log.Println("Executor Func: storing pipeline to database ...")
		errPipeline := DB.Model(&models.Pipeline{}).Create(pipelineMap).Error
		if errPipeline != nil {
			return errPipeline
		}

		log.Println("Executor Func: setting expire sortedset and hash ...")
		errExpireSortedSet := sortedset.SetExpireByKeys(1 * time.Hour)
		if errExpireSortedSet != nil {
			return errExpireSortedSet
		}

		errExpireHash := hash.SetExpireCollection(1 * time.Hour)
		if errExpireHash != nil {
			return errExpireHash
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

	wt := caches.NewWriteThrough(executorFunc)
	errSetCache := wt.SetCache(newPipeline, caches.CacheArgs{
		Indexes: []string{
			fmt.Sprintf("pipe:%s", candidateID),
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
