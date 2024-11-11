package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ImageHandlers struct {
}

func (i *ImageHandlers) Create(context *gin.Context) {
	imageFile, errImage := context.FormFile("image")
	if errImage != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errImage.Error(),
			"message": "no image file atached",
		})

		context.Abort()
		return
	}

	imageData, errOpen := imageFile.Open()
	if errOpen != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errOpen.Error(),
			"message": "failed opening image file, double check your image file and make sure it is a valid image",
		})

		context.Abort()
		return
	}

	imageBinaryData := make([]byte, imageFile.Size)
	_, errRead := imageData.Read(imageBinaryData)
	if errRead != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errRead.Error(),
			"message": "failed reading image file, double check your image file and make sure it is a valid image",
		})

		return
	}

	gormDB, _ := initializer.GetGorm()
	timeNow := time.Now()
	m_image := &models.Image{
		Name:      imageFile.Filename,
		MimeType:  http.DetectContentType(imageBinaryData),
		Size:      imageFile.Size,
		Blob:      imageBinaryData,
		CreatedAt: timeNow,
		UpdatedAt: &timeNow,
	}
	errStoreImage := gormDB.Create(&m_image).Error
	if errStoreImage != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreImage.Error(),
			"message": fmt.Sprintf("failed storing image (%s)", imageFile.Filename),
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("%s stored successfully", imageFile.Filename),
	})
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
