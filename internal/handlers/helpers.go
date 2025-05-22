package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"gorm.io/gorm"
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

func MultipleDocumentData(listOfDocuments []*multipart.FileHeader, purposeOfDocuments string) ([]models.Document, map[string]string) {
	documents := []models.Document{}
	document_status := map[string]string{}
	for _, document := range listOfDocuments {
		documentData, errOpen := document.Open()
		if errOpen != nil {
			document_status[document.Filename] = errOpen.Error()
			continue
		}

		documentBinaryData := make([]byte, document.Size)
		_, errRead := documentData.Read(documentBinaryData)
		if errRead != nil {
			document_status[document.Filename] = errRead.Error()
			continue
		}

		mimeType := http.DetectContentType(documentBinaryData)
		if !strings.HasPrefix(mimeType, "application/") {
			document_status[document.Filename] = "not a valid document. file type should be application/*"
			continue
		}

		documents = append(documents, models.Document{
			Purpose:  purposeOfDocuments,
			Name:     document.Filename,
			Size:     document.Size,
			MimeType: mimeType,
			Blob:     documentBinaryData,
		})

		document_status[document.Filename] = "valid document file"
	}

	return documents, document_status
}

func DocumentData(document *multipart.FileHeader, purposeOfDocument string) (*models.Document, error) {
	documentData, errOpen := document.Open()
	if errOpen != nil {
		return nil, errOpen
	}

	documentInBinary := make([]byte, document.Size)
	_, errRead := documentData.Read(documentInBinary)
	if errRead != nil {
		return nil, errRead
	}

	m_document := models.Document{
		Purpose:  purposeOfDocument,
		Name:     document.Filename,
		Size:     document.Size,
		MimeType: http.DetectContentType(documentInBinary),
		Blob:     documentInBinary,
	}

	return &m_document, nil
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

/* CHECKER */
func SafelyNilPointer(v *uint) interface{} {
	if v != nil {
		return int(*v)
	}

	return nil
}

func SafelyNilPointerV2(v *interface{}) interface{} {
	if v != nil {
		return v
	}

	return nil
}

/* DATABASE GUARDS */
func SLAGuard(vacancyId string, tx *gorm.DB) error {
	expirationGuard := tx.Exec(`
		UPDATE vacancies
			SET sla = CASE
								WHEN is_inactive = 1 THEN 0
								WHEN sla > 0 THEN sla - DATEDIFF(HOUR, updated_at, GETDATE())
								ELSE 0
							END,
				is_inactive = CASE
												WHEN sla = 0 THEN 1
												ELse is_inactive
											END,
				updated_at = GETDATE()
		WHERE id = ? 
	`, vacancyId)

	if expirationGuard.Error != nil {
		return expirationGuard.Error
	}

	RowsAffected := expirationGuard.RowsAffected
	if RowsAffected == 0 {
		log.Printf("no rows were updated, this might record with id %v doesn't exist in database", vacancyId)
	} else {
		log.Printf("SLA guard do their work!")
	}

	return nil
}
