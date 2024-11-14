package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

/* TOKEN */
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

/* BLOB */
func StoreImage(imageFor string, image *multipart.FileHeader, ch_storeImageStatus chan ChannelImage) {
	imageData, errOpen := image.Open()
	if errOpen != nil {
		data := ChannelImage{
			Key:     imageFor,
			Status:  errOpen.Error(),
			ImageId: 0,
		}
		log.Println("err open \t:", data)
		ch_storeImageStatus <- data
		return
	}

	defer imageData.Close()

	imageInByte := make([]byte, image.Size)
	_, errRead := imageData.Read(imageInByte)
	if errRead != nil {
		log.Println("err read \t:", errRead.Error())
		data := ChannelImage{
			Key:     imageFor,
			Status:  errRead.Error(),
			ImageId: 0,
		}
		ch_storeImageStatus <- data
		return
	}

	gormDB, _ := initializer.GetGorm()
	m_image := &models.Image{
		Name:      image.Filename,
		MimeType:  http.DetectContentType(imageInByte),
		Size:      image.Size,
		Blob:      imageInByte,
		CreatedAt: time.Now(),
	}

	if errStoreImage := gormDB.Create(&m_image).Error; errStoreImage != nil {
		log.Println("err store image \t:", errStoreImage)
		data := ChannelImage{
			Key:     imageFor,
			Status:  errStoreImage.Error(),
			ImageId: 0,
		}
		ch_storeImageStatus <- data
		return
	}
	data := ChannelImage{
		Key:     imageFor,
		Status:  "stored successfully",
		ImageId: m_image.ID,
	}

	log.Println("stored successfully!")
	ch_storeImageStatus <- data
}
func UpdateImage(imageId uint, imageFor string, image *multipart.FileHeader, ch_updateImageStatus chan ChannelImage) {
	imageOpen, errOpen := image.Open()
	if errOpen != nil {
		data := ChannelImage{
			Key:     imageFor,
			Status:  errOpen.Error(),
			ImageId: 0,
		}

		ch_updateImageStatus <- data
		return
	}

	defer imageOpen.Close()

	imageBinaryData := make([]byte, image.Size)
	_, errRead := imageOpen.Read(imageBinaryData)
	if errRead != nil {
		data := ChannelImage{
			Key:     imageFor,
			Status:  errRead.Error(),
			ImageId: 0,
		}

		ch_updateImageStatus <- data
		return
	}

	gormDB, _ := initializer.GetGorm()
	m_image := &models.Image{
		ID: imageId,
	}
	updatedTime := time.Now()
	updateImageFields := gormDB.Model(&m_image).
		Updates(models.Image{
			Name:      image.Filename,
			Size:      image.Size,
			MimeType:  http.DetectContentType(imageBinaryData),
			Blob:      imageBinaryData,
			UpdatedAt: &updatedTime,
		})

	if updateImageFields.Error != nil {
		data := ChannelImage{
			Key:     imageFor,
			Status:  updateImageFields.Error.Error(),
			ImageId: 0,
		}
		ch_updateImageStatus <- data
		return
	}

	data := ChannelImage{
		Key:     imageFor,
		Status:  "updated successfully",
		ImageId: imageId,
	}

	ch_updateImageStatus <- data
}
func StoreDocument(documentFor string, purpose string, document *multipart.FileHeader, ch_storeDocumentStatus chan ChannelDocument) {
	docData, errOpen := document.Open()
	if errOpen != nil {
		data := ChannelDocument{
			Key:        documentFor,
			Status:     errOpen.Error(),
			DocumentId: 0,
		}

		ch_storeDocumentStatus <- data
		return
	}

	defer docData.Close()

	docBinaryData := make([]byte, document.Size)
	_, errRead := docData.Read(docBinaryData)
	if errRead != nil {
		data := ChannelDocument{
			Key:        documentFor,
			Status:     errRead.Error(),
			DocumentId: 0,
		}

		ch_storeDocumentStatus <- data
		return
	}

	gormDB, _ := initializer.GetGorm()
	m_document := &models.Document{
		Purpose:   purpose,
		Name:      document.Filename,
		MimeType:  http.DetectContentType(docBinaryData),
		Size:      document.Size,
		Blob:      docBinaryData,
		CreatedAt: time.Now(),
	}
	errStoreDocument := gormDB.Create(&m_document).Error

	if errStoreDocument != nil {
		data := ChannelDocument{
			Key:        documentFor,
			Status:     errStoreDocument.Error(),
			DocumentId: 0,
		}

		ch_storeDocumentStatus <- data
		return
	}

	data := ChannelDocument{
		Key:        documentFor,
		Status:     "document stored successfully",
		DocumentId: m_document.ID,
	}

	ch_storeDocumentStatus <- data
	close(ch_storeDocumentStatus)
}
func UpdateDocument(documentId uint, documentFor string, purpose string, document *multipart.FileHeader, ch_updateDocumentStatus chan ChannelDocument) {
	documentData, errOpen := document.Open()
	if errOpen != nil {
		data := ChannelDocument{
			Key:        documentFor,
			Status:     errOpen.Error(),
			DocumentId: 0,
		}

		ch_updateDocumentStatus <- data
		return
	}

	documentBinaryData := make([]byte, document.Size)
	_, errRead := documentData.Read(documentBinaryData)
	if errRead != nil {
		data := ChannelDocument{
			Key:        documentFor,
			Status:     errRead.Error(),
			DocumentId: 0,
		}

		ch_updateDocumentStatus <- data
		return
	}

	gormDB, _ := initializer.GetGorm()
	m_document := models.Document{
		ID: documentId,
	}
	updatedTime := time.Now()
	errUpdateDocument := gormDB.Model(&m_document).
		Updates(models.Document{
			Name:      document.Filename,
			MimeType:  http.DetectContentType(documentBinaryData),
			Size:      document.Size,
			Blob:      documentBinaryData,
			UpdatedAt: &updatedTime}).
		Error

	if errUpdateDocument != nil {
		data := ChannelDocument{
			Key:        documentFor,
			Status:     errUpdateDocument.Error(),
			DocumentId: 0,
		}

		ch_updateDocumentStatus <- data
		return
	}

	data := ChannelDocument{
		Key:        documentFor,
		Status:     "document updated successfully",
		DocumentId: documentId,
	}

	ch_updateDocumentStatus <- data
	close(ch_updateDocumentStatus)
}

func MultipleImageData(listOfImages []*multipart.FileHeader) ([]models.Image, map[string]string) {
	images := []models.Image{}
	image_status := map[string]string{}
	for _, image := range listOfImages {
		imageData, errOpen := image.Open()
		if errOpen != nil {
			image_status[image.Filename] = errOpen.Error()
			continue
		}

		imageBinaryData := make([]byte, image.Size)
		_, errRead := imageData.Read(imageBinaryData)
		if errRead != nil {
			image_status[image.Filename] = errRead.Error()
			continue
		}

		mimeType := http.DetectContentType(imageBinaryData)
		if !strings.HasPrefix(mimeType, "image/") {
			image_status[image.Filename] = "not a vaid image. file type should be image/*"
			continue
		}

		images = append(images, models.Image{
			Name:     image.Filename,
			Size:     image.Size,
			MimeType: mimeType,
			Blob:     imageBinaryData,
		})

		image_status[image.Filename] = "valid image file"
	}

	return images, image_status
}

func ImageData(image *multipart.FileHeader) (*models.Image, error) {
	imageData, errOpen := image.Open()
	if errOpen != nil {
		return nil, errOpen
	}

	imageBinaryData := make([]byte, image.Size)
	_, errRead := imageData.Read(imageBinaryData)
	if errRead != nil {
		return nil, errRead
	}

	timeNow := time.Now()

	m_image := models.Image{
		Name:      image.Filename,
		Size:      image.Size,
		MimeType:  http.DetectContentType(imageBinaryData),
		Blob:      imageBinaryData,
		UpdatedAt: &timeNow,
	}

	return &m_image, nil
}

/* DATA TRANSFORM */
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
					if value != nil && value != 0 {
						recordTyped[index][newKey] = fmt.Sprintf("/api/v1/%s/%v", pathType, value)
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
				if value != nil && value != 0 {
					recordTyped[newKey] = fmt.Sprintf("/api/v1/%s/%v", pathType, value)
				} else {
					recordTyped[newKey] = nil
				}
				delete(recordTyped, target)
			}
		}
	}
}

/* CHECKER */
func SafelyNilPointer(v *uint) interface{} {
	if v != nil {
		return int(*v)
	}

	return nil
}
