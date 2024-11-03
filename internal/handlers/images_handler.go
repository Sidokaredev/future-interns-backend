package handlers

import (
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ImageHandlers struct {
}

func (i *ImageHandlers) GetById(context *gin.Context) {
	param_imageId := context.Param("id")

	gormDB, _ := initializer.GetGorm()
	m_image := models.Image{}
	errImage := gormDB.Model(&models.Image{}).
		Select([]string{
			"mime_type",
			"blob",
		}).
		Where("id = ?", param_imageId).
		First(&m_image).Error

	if errImage != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errImage.Error(),
			"message": "failed getting binary data from image",
		})

		context.Abort()
		return
	}

	context.Data(http.StatusOK, m_image.MimeType, m_image.Blob)
}
