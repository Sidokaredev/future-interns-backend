package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type CandidatesHandler struct {
}

type CreateCandidateForm struct {
	Expertise   string `form:"expertise" binding:"required"`
	AboutMe     string `form:"about_me" binding:"required"`
	DateOfBirth string `form:"date_of_birth" binding:"required"`
}

type UpdateCandidateForm struct {
	Expertise   string `form:"expertise"`
	AboutMe     string `form:"about_me"`
	DateOfBirth string `form:"date_of_birth"`
}

type CreateEducationJSON struct {
	University  string     `json:"university" binding:"required"`
	Address     string     `json:"address" binding:"required"`
	Major       string     `json:"major" binding:"required"`
	IsGraduated bool       `json:"is_graduated"`
	StartAt     time.Time  `json:"start_at" binding:"required"`
	EndAt       *time.Time `json:"end_at"`
	GPA         float32    `json:"gpa" binding:"required"`
}

type UpdateEducationJSON struct {
	Id          uint       `json:"id" binding:"required"`
	University  *string    `json:"university"`
	Address     *string    `json:"address"`
	Major       *string    `json:"major"`
	IsGraduated *bool      `json:"is_graduated"`
	StartAt     *time.Time `json:"start_at"`
	EndAt       *time.Time `json:"end_at"`
	GPA         *float32   `json:"gpa"`
}

type CreateAddressJSON struct {
	Street       string `json:"street" binding:"required"`
	Neighborhood string `json:"neighborhood" binding:"required"`
	RuralArea    string `json:"rural_area" binding:"required"`
	SubDistrict  string `json:"sub_district" binding:"required"`
	City         string `json:"city" binding:"required"`
	Province     string `json:"province" binding:"required"`
	Country      string `json:"country" binding:"required"`
	PostalCode   int    `json:"postal_code" binding:"required"`
	Type         string `json:"type" binding:"required"`
}

type UpdateAddressJSON struct {
	Id           uint    `json:"id" binding:"required"`
	Street       *string `json:"street"`
	Neighborhood *string `json:"neighborhood"`
	RuralArea    *string `json:"rural_area"`
	SubDistrict  *string `json:"sub_district"`
	City         *string `json:"city"`
	Province     *string `json:"province"`
	Country      *string `json:"country"`
	PostalCode   *int    `json:"postal_code"`
	Type         *string `json:"type"`
}

type ChannelImage struct {
	Key     string
	Status  string
	ImageId uint
}

type ChannelDocument struct {
	Key        string
	Status     string
	DocumentId uint
}

/* helpers */
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

func SafelyNilPointer(v *uint) interface{} {
	if v != nil {
		return int(*v)
	}

	return nil
}

/* handlers */
func (c *CandidatesHandler) Create(context *gin.Context) {
	var candidateForm CreateCandidateForm
	if errBind := context.ShouldBind(&candidateForm); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your Form Data fields, kids",
		})
		context.Abort()
		return
	}
	bearerToken := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	tokenClaims := ParseJWT(bearerToken)
	/* channel */
	ch_storeImageStatus := make(chan ChannelImage, 2)
	ch_storeDocumentStatus := make(chan ChannelDocument)
	ch_uuid := make(chan string)
	imageChannelCounter := 2
	documentChannelCounter := 1

	backgroundProfileImg, errBackgroundProfileImg := context.FormFile("background_profile_img")
	profileImg, errProfileImg := context.FormFile("profile_img")
	cvDocument, errCvDocument := context.FormFile("cv_document")
	if errBackgroundProfileImg == nil {
		go StoreImage("background_profile_img", backgroundProfileImg, ch_storeImageStatus)
	} else {
		imageChannelCounter -= 1
	}
	if errProfileImg == nil {
		go StoreImage("profile_img", profileImg, ch_storeImageStatus)
	} else {
		imageChannelCounter -= 1
	}
	if errCvDocument == nil {
		go StoreDocument("cv_document", "curriculum vitae", cvDocument, ch_storeDocumentStatus)
	} else {
		documentChannelCounter -= 1
	}

	/* goroutine */
	go GenUuid(candidateForm.DateOfBirth, ch_uuid)
	TimeLocationIndonesian, _ := time.LoadLocation("Asia/Jakarta")
	parsedDateOfirth, _ := time.Parse(time.RFC3339, candidateForm.DateOfBirth)
	parsedToLocale := parsedDateOfirth.In(TimeLocationIndonesian)
	var (
		backgroundProfileImgStatus string = "no image attached"
		profileImgStatus           string = "no image attached"
		cvDocumentStatus           string = "no document attached"
	)
	m_candidate := &models.Candidate{
		UserId:      tokenClaims.Id,
		Id:          <-ch_uuid,
		Expertise:   candidateForm.Expertise,
		AboutMe:     candidateForm.AboutMe,
		DateOfBirth: parsedToLocale,
	}
	for i := 0; i < imageChannelCounter; i++ {
		data := <-ch_storeImageStatus
		switch data.Key {
		case "background_profile_img":
			if data.ImageId != 0 {
				m_candidate.BackgroundProfileImageId = &data.ImageId
				backgroundProfileImgStatus = data.Status
			} else {
				backgroundProfileImgStatus = data.Status
			}
		case "profile_img":
			if data.ImageId != 0 {
				m_candidate.ProfileImageId = &data.ImageId
				profileImgStatus = data.Status
			} else {
				profileImgStatus = data.Status
			}
		}
	}

	for i := 0; i < documentChannelCounter; i++ {
		data := <-ch_storeDocumentStatus
		if data.DocumentId != 0 {
			m_candidate.CVDocumentId = &data.DocumentId
			cvDocumentStatus = data.Status
		} else {
			cvDocumentStatus = data.Status
		}
	}

	gormDB, _ := initializer.GetGorm()
	if errStoreCandidate := gormDB.Create(&m_candidate).Error; errStoreCandidate != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreCandidate.Error(),
			"message": "there was a issue with database",
		})
		context.Abort()
		return
	}

	responseData := gin.H{
		"success": true,
		"data": gin.H{
			"background_profile_img_status": backgroundProfileImgStatus,
			"profile_img_status":            profileImgStatus,
			"cv_document_status":            cvDocumentStatus,
			"candidate_id":                  m_candidate.Id,
		},
	}

	context.JSON(http.StatusOK, responseData)
}

func (c *CandidatesHandler) Update(context *gin.Context) {
	// get bearer token
	bearerToken := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	tokenClaims := ParseJWT(bearerToken)
	// bind form-data requerst
	var candidateForm UpdateCandidateForm
	context.ShouldBind(&candidateForm)
	// create map from form-data request
	mapCandidateFields := make(map[string]interface{})
	formValue := reflect.ValueOf(candidateForm)
	formField := reflect.TypeOf(candidateForm)

	for i := 0; i < formValue.NumField(); i++ {
		fieldName := formField.Field(i).Tag.Get("form")
		value := formValue.Field(i).Interface()

		if value != "" {
			if fieldName == "date_of_birth" {
				TimeLocationIndonesian, _ := time.LoadLocation("Asia/Jakarta")
				parsedDateOfirth, _ := time.Parse(time.RFC3339, value.(string))
				parsedToLocale := parsedDateOfirth.In(TimeLocationIndonesian)
				mapCandidateFields[fieldName] = parsedToLocale
			} else {
				mapCandidateFields[fieldName] = value
			}
		}
	}
	// get candidate few candidate field from database
	gormDB, _ := initializer.GetGorm()
	m_candidate := models.Candidate{}
	errGetCandidate := gormDB.
		Select("id", "background_profile_image_id", "profile_image_id", "cv_document_id").
		First(&m_candidate, "user_id = ?", tokenClaims.Id).
		Error

	if errGetCandidate != nil {
		message := fmt.Sprintf("candidate with user_id (%s) not found", tokenClaims.Id)
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errGetCandidate.Error(),
			"message": message,
		})

		context.Abort()
		return
	}
	// creating channel for getting updated image status
	ch_updateImageStatus := make(chan ChannelImage, 2)
	ch_updateDocumentStatus := make(chan ChannelDocument, 1)
	updateImageCounter := 2
	var (
		background_profile_img_status string
		profile_image_status          string
		cv_document_status            string
	)
	// getting file attached on request (multipart)
	backgroundProfileImg, errBackground := context.FormFile("background_profile_img")
	profileImg, errProfile := context.FormFile("profile_img")
	cvDocument, errCVDocument := context.FormFile("cv_document")

	if errBackground == nil {
		go UpdateImage(*m_candidate.BackgroundProfileImageId, "background_profile_img", backgroundProfileImg, ch_updateImageStatus)
	} else {
		background_profile_img_status = errBackground.Error()
		updateImageCounter -= 1
	}

	if errProfile == nil {
		go UpdateImage(*m_candidate.ProfileImageId, "profile_img", profileImg, ch_updateImageStatus)
	} else {
		profile_image_status = errProfile.Error()
		updateImageCounter -= 1
	}

	if errCVDocument == nil {
		go UpdateDocument(*m_candidate.CVDocumentId, "cv_document", "curriculum_vitae", cvDocument, ch_updateDocumentStatus)
	}

	for i := 0; i < updateImageCounter; i++ {
		data := <-ch_updateImageStatus

		switch data.Key {
		case "background_profile_img":
			background_profile_img_status = data.Status

		case "profile_img":
			profile_image_status = data.Status
		}
	}

	if ch_data, ok := <-ch_updateDocumentStatus; ok {
		cv_document_status = ch_data.Status
	}
	//  storing candidate field prop
	var updated_status string
	if len(mapCandidateFields) != 0 {
		log.Println("updated query still running even there are no data available...", mapCandidateFields)
		updateCandidate := gormDB.Model(&models.Candidate{}).Where("id = ?", m_candidate.Id).Updates(mapCandidateFields)
		if updateCandidate.Error != nil {
			context.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   updateCandidate.Error.Error(),
				"message": "failed to update candidate data",
			})
			context.Abort()
			return
		}
		updated_status = fmt.Sprintf("successfully update candidate with ID (%s)", m_candidate.Id)
	} else {
		updated_status = fmt.Sprintf("no data available for candidate with ID (%s)", m_candidate.Id)
	}
	response := gin.H{
		"success": true,
		"data": gin.H{
			"background_profile_img_update_status": background_profile_img_status,
			"profile_img_update_status":            profile_image_status,
			"cv_document_update_status":            cv_document_status,
			"updated_status":                       updated_status,
		},
	}
	context.JSON(http.StatusOK, response)
}

func (c *CandidatesHandler) Get(context *gin.Context) {
	queries, _ := context.GetQuery("includes")
	gormDB, _ := initializer.GetGorm()
	final_candidates := []map[string]any{}
	candidates := []models.Candidate{}

	errGetCandidateRows := gormDB.Transaction(func(tx *gorm.DB) error {
		candidateRows := tx.Model(&models.Candidate{})
		if strings.Contains(queries, "user") {
			candidateRows = candidateRows.Preload("User")
		}
		if strings.Contains(queries, "address") {
			candidateRows = candidateRows.Preload("Addresses", func(db *gorm.DB) *gorm.DB {
				return db.Where("type = ?", "home").Limit(1).Order("created_at DESC")
			})
		}
		if strings.Contains(queries, "skills") {
			candidateRows = candidateRows.Preload("Skills")
		}
		if strings.Contains(queries, "educations") {
			candidateRows = candidateRows.Preload("Educations")
		}
		if strings.Contains(queries, "experiences") {
			candidateRows = candidateRows.Preload("Experiences")
		}
		if strings.Contains(queries, "socials") {
			candidateRows = candidateRows.Preload("CandidateSocials.Social")
		}

		candidateRows = candidateRows.Find(&candidates)

		if candidateRows.Error != nil {
			return candidateRows.Error
		} else if candidateRows.RowsAffected == 0 {
			return fmt.Errorf("no candidate record, %d rows affected", candidateRows.RowsAffected)
		}

		return nil
	})

	if errGetCandidateRows != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errGetCandidateRows.Error(),
			"message": "failed getting all candidate records",
		})
		context.Abort()
		return
	}

	for _, candidate := range candidates {
		candidateMap := map[string]interface{}{
			"id":                          candidate.Id,
			"expertise":                   candidate.Expertise,
			"about_me":                    candidate.AboutMe,
			"date_of_birth":               candidate.DateOfBirth,
			"background_profile_image_id": SafelyNilPointer(candidate.BackgroundProfileImageId),
			"profile_image_id":            SafelyNilPointer(candidate.ProfileImageId),
			"cv_document_id":              SafelyNilPointer(candidate.CVDocumentId),
		}

		TransformsIdToPath([]string{
			"background_profile_image_id",
			"profile_image_id",
			"cv_document_id",
		}, candidateMap)

		if strings.Contains(queries, "user") {
			userMap := map[string]interface{}{
				"id":       candidate.User.Id,
				"fullname": candidate.User.Fullname,
				"email":    candidate.User.Email,
			}

			candidateMap["user"] = userMap
		}
		if strings.Contains(queries, "address") {
			if len(candidate.Addresses) != 0 {
				addressMap := map[string]interface{}{
					"id":           candidate.Addresses[0].ID,
					"street":       candidate.Addresses[0].Street,
					"neighborhood": candidate.Addresses[0].Neighborhood,
					"rural_area":   candidate.Addresses[0].RuralArea,
					"sub_district": candidate.Addresses[0].SubDistrict,
					"city":         candidate.Addresses[0].City,
					"province":     candidate.Addresses[0].Province,
					"country":      candidate.Addresses[0].Country,
				}

				candidateMap["address"] = addressMap
			} else {
				candidateMap["address"] = nil
			}
		}

		if strings.Contains(queries, "skills") {
			if len(candidate.Skills) != 0 {
				skillsMap := []map[string]interface{}{}
				for _, skill := range candidate.Skills {

					skillsMap = append(skillsMap, map[string]interface{}{
						"id":                    skill.ID,
						"name":                  skill.Name,
						"skill_icon_image_path": fmt.Sprintf("/api/v1/images/%d", skill.SkillIconImageId),
					})
				}

				candidateMap["skills"] = skillsMap
			} else {
				candidateMap["skills"] = nil
			}
		}

		if strings.Contains(queries, "educations") {
			if len(candidate.Educations) != 0 {
				educationsMap := []map[string]interface{}{}
				for _, education := range candidate.Educations {
					educationsMap = append(educationsMap, map[string]interface{}{
						"id":           education.ID,
						"university":   education.University,
						"major":        education.Major,
						"address":      education.Address,
						"is_graduated": education.IsGraduated,
						"start_at":     education.StartAt,
						"end_at":       education.EndAt,
						"gpa":          education.GPA,
					})
				}

				candidateMap["educations"] = educationsMap
			} else {
				candidateMap["educations"] = nil
			}
		}

		if strings.Contains(queries, "experiences") {
			if len(candidate.Experiences) != 0 {
				experiencesMap := []map[string]interface{}{}
				for _, experience := range candidate.Experiences {
					experiencesMap = append(experiencesMap, map[string]interface{}{
						"id":                     experience.ID,
						"company_name":           experience.CompanyName,
						"position":               experience.Position,
						"location_address":       experience.LocationAddress,
						"type":                   experience.Type,
						"is_current":             experience.IsCurrent,
						"start_at":               experience.StartAt,
						"end_at":                 experience.EndAt,
						"description":            experience.Description,
						"attachment_document_id": SafelyNilPointer(&experience.AttachmentDocumentId),
					})

				}

				TransformsIdToPath([]string{
					"attachment_document_id",
				}, experiencesMap)

				candidateMap["experiences"] = experiencesMap
			} else {
				candidateMap["experiences"] = nil
			}
		}

		if strings.Contains(queries, "socials") {
			if len(candidate.CandidateSocials) != 0 {
				socialsMap := []map[string]interface{}{}
				for _, candidateSocial := range candidate.CandidateSocials {
					icon_image_id := uint(candidateSocial.Social.IconImageId)
					socialsMap = append(socialsMap, map[string]interface{}{
						"id":              candidateSocial.Social.ID,
						"name":            candidateSocial.Social.Name,
						"url":             candidateSocial.Url,
						"icon_image_path": fmt.Sprintf("/api/v1/images/%v", SafelyNilPointer(&icon_image_id)),
					})
				}

				candidateMap["socials"] = socialsMap
			} else {
				candidateMap["socials"] = nil
			}
		}

		final_candidates = append(final_candidates, candidateMap)
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    final_candidates,
	})
}

func (c *CandidatesHandler) GetById(context *gin.Context) {
	queries, _ := context.GetQuery("includes")

	param_candidateId := context.Param("id")

	var candidate map[string]interface{}
	m_candidate := models.Candidate{}

	gormDB, _ := initializer.GetGorm()
	errCandidate := gormDB.Transaction(func(tx *gorm.DB) error {
		getCandidate := tx.Model(&models.Candidate{})
		if strings.Contains(queries, "user") {
			getCandidate = getCandidate.Preload("User")
		}
		if strings.Contains(queries, "address") {
			getCandidate = getCandidate.Preload("Addresses", func(db *gorm.DB) *gorm.DB {
				return db.Where("type = ?", "home").Limit(1).Order("created_at DESC")
			})
		}
		if strings.Contains(queries, "skills") {
			getCandidate = getCandidate.Preload("Skills")
		}
		if strings.Contains(queries, "educations") {
			getCandidate = getCandidate.Preload("Educations")
		}
		if strings.Contains(queries, "experiences") {
			getCandidate = getCandidate.Preload("Experiences")
		}
		if strings.Contains(queries, "socials") {
			getCandidate = getCandidate.Preload("CandidateSocials.Social")
		}

		getCandidate = getCandidate.Model(&models.Candidate{}).Where("id = ?", param_candidateId).First(&m_candidate)

		if getCandidate.Error != nil {
			return getCandidate.Error
		} else if getCandidate.RowsAffected == 0 {
			return fmt.Errorf("no candidate record, %d rows affected", getCandidate.RowsAffected)
		}

		return nil
	})

	if errCandidate != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errCandidate.Error(),
			"message": fmt.Sprintf("failed getting candidate with id (%v)", param_candidateId),
		})

		context.Abort()
		return
	}
	// WAIT
	candidate = map[string]interface{}{
		"expertise":                   m_candidate.Expertise,
		"about_me":                    m_candidate.AboutMe,
		"date_of_birth":               m_candidate.DateOfBirth,
		"background_profile_image_id": SafelyNilPointer(m_candidate.BackgroundProfileImageId),
		"profile_image_id":            SafelyNilPointer(m_candidate.ProfileImageId),
		"cv_document_id":              SafelyNilPointer(m_candidate.CVDocumentId),
	}

	TransformsIdToPath([]string{
		"background_profile_image_id",
		"profile_image_id",
		"cv_document_id",
	}, candidate)

	if strings.Contains(queries, "user") {
		userMap := map[string]interface{}{
			"id":       m_candidate.User.Id,
			"fullname": m_candidate.User.Fullname,
			"email":    m_candidate.User.Email,
		}

		candidate["user"] = userMap
	}
	if strings.Contains(queries, "address") {
		if len(m_candidate.Addresses) != 0 {
			addressMap := map[string]interface{}{
				"street":       m_candidate.Addresses[0].Street,
				"neighborhood": m_candidate.Addresses[0].Neighborhood,
				"rural_area":   m_candidate.Addresses[0].RuralArea,
				"sub_district": m_candidate.Addresses[0].SubDistrict,
				"city":         m_candidate.Addresses[0].City,
				"province":     m_candidate.Addresses[0].Province,
				"country":      m_candidate.Addresses[0].Country,
			}

			candidate["address"] = addressMap
		} else {
			candidate["address"] = nil
		}
	}

	if strings.Contains(queries, "skills") {
		if len(m_candidate.Skills) != 0 {
			skillsMap := []map[string]interface{}{}
			for _, skill := range m_candidate.Skills {

				skillsMap = append(skillsMap, map[string]interface{}{
					"name":                  skill.Name,
					"skill_icon_image_path": fmt.Sprintf("/api/v1/images/%d", skill.SkillIconImageId),
				})
			}

			candidate["skills"] = skillsMap
		} else {
			candidate["skills"] = nil
		}
	}

	if strings.Contains(queries, "educations") {
		if len(m_candidate.Educations) != 0 {
			educationsMap := []map[string]interface{}{}
			for _, education := range m_candidate.Educations {
				educationsMap = append(educationsMap, map[string]interface{}{
					"university":   education.University,
					"major":        education.Major,
					"address":      education.Address,
					"is_graduated": education.IsGraduated,
					"start_at":     education.StartAt,
					"end_at":       education.EndAt,
					"gpa":          education.GPA,
				})
			}

			candidate["educations"] = educationsMap
		} else {
			candidate["educations"] = nil
		}
	}

	if strings.Contains(queries, "experiences") {
		if len(m_candidate.Experiences) != 0 {
			experiencesMap := []map[string]interface{}{}
			for _, experience := range m_candidate.Experiences {
				experiencesMap = append(experiencesMap, map[string]interface{}{
					"company_name":           experience.CompanyName,
					"position":               experience.Position,
					"location_address":       experience.LocationAddress,
					"type":                   experience.Type,
					"is_current":             experience.IsCurrent,
					"start_at":               experience.StartAt,
					"end_at":                 experience.EndAt,
					"description":            experience.Description,
					"attachment_document_id": SafelyNilPointer(&experience.AttachmentDocumentId),
				})

			}

			TransformsIdToPath([]string{
				"attachment_document_id",
			}, experiencesMap)

			candidate["experiences"] = experiencesMap
		} else {
			candidate["experiences"] = nil
		}
	}

	if strings.Contains(queries, "socials") {
		if len(m_candidate.CandidateSocials) != 0 {
			socialsMap := []map[string]interface{}{}
			for _, candidateSocial := range m_candidate.CandidateSocials {
				icon_image_id := uint(candidateSocial.Social.IconImageId)
				socialsMap = append(socialsMap, map[string]interface{}{
					"name":            candidateSocial.Social.Name,
					"url":             candidateSocial.Url,
					"icon_image_path": fmt.Sprintf("/api/v1/images/%v", SafelyNilPointer(&icon_image_id)),
				})
			}

			candidate["socials"] = socialsMap
		} else {
			candidate["socials"] = nil
		}
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    candidate,
	})
}

func (c *CandidatesHandler) DeleteById(context *gin.Context) {
	candidateId := context.Param("id")
	if candidateId == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "param candidate :id required",
			"message": "specify your candidate id as param",
		})

		context.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	tokenClaims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()

	errDeleting := gormDB.Transaction(func(tx *gorm.DB) error {
		deleteCandidateRowAffected := tx.Delete(&models.Candidate{}, "id = ?", candidateId).RowsAffected
		if deleteCandidateRowAffected == 0 {
			return fmt.Errorf("failed deleting candidate with id (%s)", candidateId)
		}

		deleteUserRowAffected := tx.Delete(&models.User{}, "id = ?", tokenClaims.Id).RowsAffected
		if deleteUserRowAffected == 0 {
			return fmt.Errorf("failed deleting user with id (%s)", tokenClaims.Id)
		}
		return nil
	})

	if errDeleting != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errDeleting.Error(),
			"message": "be carefull, this request directed to /api/v1/candidates/:id. check your url path and specify the :id param",
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("candidate with id (%s) deleted successfully", candidateId),
	})
}

func (c *CandidatesHandler) Unscoped(context *gin.Context) {
	queries, _ := context.GetQuery("includes")

	unscoped_candidates := []map[string]interface{}{}
	m_unscoped_candidates := []models.Candidate{}

	gormDB, _ := initializer.GetGorm()
	errUnscopedCandidates := gormDB.Transaction(func(tx *gorm.DB) error {
		unscopedCandidates := tx.Model(&models.Candidate{})
		if strings.Contains(queries, "user") {
			unscopedCandidates = unscopedCandidates.Preload("User")
		}
		if strings.Contains(queries, "address") {
			unscopedCandidates = unscopedCandidates.Preload("Addresses", func(db *gorm.DB) *gorm.DB {
				return db.Where("type = ?", "home").Limit(1).Order("created_at DESC")
			})
		}
		if strings.Contains(queries, "skills") {
			unscopedCandidates = unscopedCandidates.Preload("Skills")
		}
		if strings.Contains(queries, "educations") {
			unscopedCandidates = unscopedCandidates.Preload("Educations")
		}
		if strings.Contains(queries, "experiences") {
			unscopedCandidates = unscopedCandidates.Preload("Experiences")
		}
		if strings.Contains(queries, "socials") {
			unscopedCandidates = unscopedCandidates.Preload("CandidateSocials.Social")
		}

		unscopedCandidates = unscopedCandidates.Unscoped().Not("delete_at", nil).Find(&m_unscoped_candidates)

		if unscopedCandidates.Error != nil {
			return unscopedCandidates.Error
		} else if unscopedCandidates.RowsAffected == 0 {
			return fmt.Errorf("no candidate record, %d rows affected", unscopedCandidates.RowsAffected)
		}

		return nil
	})

	if errUnscopedCandidates != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errUnscopedCandidates.Error(),
			"message": "failed getting unscoped candidate records",
		})

		context.Abort()
		return
	}

	for _, candidate := range m_unscoped_candidates {
		candidateMap := map[string]interface{}{
			"expertise":                   candidate.Expertise,
			"about_me":                    candidate.AboutMe,
			"date_of_birth":               candidate.DateOfBirth,
			"background_profile_image_id": SafelyNilPointer(candidate.BackgroundProfileImageId),
			"profile_image_id":            SafelyNilPointer(candidate.ProfileImageId),
			"cv_document_id":              SafelyNilPointer(candidate.CVDocumentId),
		}

		TransformsIdToPath([]string{
			"background_profile_image_id",
			"profile_image_id",
			"cv_document_id",
		}, candidateMap)

		if strings.Contains(queries, "user") {
			userMap := map[string]interface{}{
				"id":       candidate.User.Id,
				"fullname": candidate.User.Fullname,
				"email":    candidate.User.Email,
			}

			candidateMap["user"] = userMap
		}
		if strings.Contains(queries, "address") {
			if len(candidate.Addresses) != 0 {
				addressMap := map[string]interface{}{
					"street":       candidate.Addresses[0].Street,
					"neighborhood": candidate.Addresses[0].Neighborhood,
					"rural_area":   candidate.Addresses[0].RuralArea,
					"sub_district": candidate.Addresses[0].SubDistrict,
					"city":         candidate.Addresses[0].City,
					"province":     candidate.Addresses[0].Province,
					"country":      candidate.Addresses[0].Country,
				}

				candidateMap["address"] = addressMap
			} else {
				candidateMap["address"] = nil
			}
		}

		if strings.Contains(queries, "skills") {
			if len(candidate.Skills) != 0 {
				skillsMap := []map[string]interface{}{}
				for _, skill := range candidate.Skills {

					skillsMap = append(skillsMap, map[string]interface{}{
						"name":                  skill.Name,
						"skill_icon_image_path": fmt.Sprintf("/api/v1/images/%d", skill.SkillIconImageId),
					})
				}

				candidateMap["skills"] = skillsMap
			} else {
				candidateMap["skills"] = nil
			}
		}

		if strings.Contains(queries, "educations") {
			if len(candidate.Educations) != 0 {
				educationsMap := []map[string]interface{}{}
				for _, education := range candidate.Educations {
					educationsMap = append(educationsMap, map[string]interface{}{
						"university":   education.University,
						"major":        education.Major,
						"address":      education.Address,
						"is_graduated": education.IsGraduated,
						"start_at":     education.StartAt,
						"end_at":       education.EndAt,
						"gpa":          education.GPA,
					})
				}

				candidateMap["educations"] = educationsMap
			} else {
				candidateMap["educations"] = nil
			}
		}

		if strings.Contains(queries, "experiences") {
			if len(candidate.Experiences) != 0 {
				experiencesMap := []map[string]interface{}{}
				for _, experience := range candidate.Experiences {
					experiencesMap = append(experiencesMap, map[string]interface{}{
						"company_name":           experience.CompanyName,
						"position":               experience.Position,
						"location_address":       experience.LocationAddress,
						"type":                   experience.Type,
						"is_current":             experience.IsCurrent,
						"start_at":               experience.StartAt,
						"end_at":                 experience.EndAt,
						"description":            experience.Description,
						"attachment_document_id": SafelyNilPointer(&experience.AttachmentDocumentId),
					})

				}

				TransformsIdToPath([]string{
					"attachment_document_id",
				}, experiencesMap)

				candidateMap["experiences"] = experiencesMap
			} else {
				candidateMap["experiences"] = nil
			}
		}

		if strings.Contains(queries, "socials") {
			if len(candidate.CandidateSocials) != 0 {
				socialsMap := []map[string]interface{}{}
				for _, candidateSocial := range candidate.CandidateSocials {
					icon_image_id := uint(candidateSocial.Social.IconImageId)
					socialsMap = append(socialsMap, map[string]interface{}{
						"name":            candidateSocial.Social.Name,
						"url":             candidateSocial.Url,
						"icon_image_path": fmt.Sprintf("/api/v1/images/%v", SafelyNilPointer(&icon_image_id)),
					})
				}

				candidateMap["socials"] = socialsMap
			} else {
				candidateMap["socials"] = nil
			}
		}

		unscoped_candidates = append(unscoped_candidates, candidateMap)
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    unscoped_candidates,
	})
}

func (c *CandidatesHandler) UnscopedById(context *gin.Context) {
	queries, _ := context.GetQuery("includes")

	param_candidateId := context.Param("id")

	var unscoped_candidate map[string]interface{}
	m_unscoped_candidate := models.Candidate{}

	gormDB, _ := initializer.GetGorm()
	errUnscopedCandidate := gormDB.Transaction(func(tx *gorm.DB) error {
		unscopedCandidate := tx.Model(&models.Candidate{})
		if strings.Contains(queries, "user") {
			unscopedCandidate = unscopedCandidate.Preload("User")
		}
		if strings.Contains(queries, "address") {
			unscopedCandidate = unscopedCandidate.Preload("Addresses", func(db *gorm.DB) *gorm.DB {
				return db.Where("type = ?", "home").Limit(1).Order("created_at DESC")
			})
		}
		if strings.Contains(queries, "skills") {
			unscopedCandidate = unscopedCandidate.Preload("Skills")
		}
		if strings.Contains(queries, "educations") {
			unscopedCandidate = unscopedCandidate.Preload("Educations")
		}
		if strings.Contains(queries, "experiences") {
			unscopedCandidate = unscopedCandidate.Preload("Experiences")
		}
		if strings.Contains(queries, "socials") {
			unscopedCandidate = unscopedCandidate.Preload("CandidateSocials.Social")
		}

		unscopedCandidate = unscopedCandidate.Unscoped().Not("delete_at", nil).Where("id = ?", param_candidateId).First(&m_unscoped_candidate)

		if unscopedCandidate.Error != nil {
			return unscopedCandidate.Error
		} else if unscopedCandidate.RowsAffected == 0 {
			return fmt.Errorf("no candidate record, %d rows affected", unscopedCandidate.RowsAffected)
		}

		return nil
	})

	if errUnscopedCandidate != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errUnscopedCandidate.Error(),
			"message": "failed getting unscoped candidate records",
		})

		context.Abort()
		return
	}
	// WAIT
	unscoped_candidate = map[string]interface{}{
		"expertise":                   m_unscoped_candidate.Expertise,
		"about_me":                    m_unscoped_candidate.AboutMe,
		"date_of_birth":               m_unscoped_candidate.DateOfBirth,
		"background_profile_image_id": SafelyNilPointer(m_unscoped_candidate.BackgroundProfileImageId),
		"profile_image_id":            SafelyNilPointer(m_unscoped_candidate.ProfileImageId),
		"cv_document_id":              SafelyNilPointer(m_unscoped_candidate.CVDocumentId),
	}

	TransformsIdToPath([]string{
		"background_profile_image_id",
		"profile_image_id",
		"cv_document_id",
	}, unscoped_candidate)

	if strings.Contains(queries, "user") {
		userMap := map[string]interface{}{
			"id":       m_unscoped_candidate.User.Id,
			"fullname": m_unscoped_candidate.User.Fullname,
			"email":    m_unscoped_candidate.User.Email,
		}

		unscoped_candidate["user"] = userMap
	}
	if strings.Contains(queries, "address") {
		if len(m_unscoped_candidate.Addresses) != 0 {
			addressMap := map[string]interface{}{
				"street":       m_unscoped_candidate.Addresses[0].Street,
				"neighborhood": m_unscoped_candidate.Addresses[0].Neighborhood,
				"rural_area":   m_unscoped_candidate.Addresses[0].RuralArea,
				"sub_district": m_unscoped_candidate.Addresses[0].SubDistrict,
				"city":         m_unscoped_candidate.Addresses[0].City,
				"province":     m_unscoped_candidate.Addresses[0].Province,
				"country":      m_unscoped_candidate.Addresses[0].Country,
			}

			unscoped_candidate["address"] = addressMap
		} else {
			unscoped_candidate["address"] = nil
		}
	}

	if strings.Contains(queries, "skills") {
		if len(m_unscoped_candidate.Skills) != 0 {
			skillsMap := []map[string]interface{}{}
			for _, skill := range m_unscoped_candidate.Skills {

				skillsMap = append(skillsMap, map[string]interface{}{
					"name":                  skill.Name,
					"skill_icon_image_path": fmt.Sprintf("/api/v1/images/%d", skill.SkillIconImageId),
				})
			}

			unscoped_candidate["skills"] = skillsMap
		} else {
			unscoped_candidate["skills"] = nil
		}
	}

	if strings.Contains(queries, "educations") {
		if len(m_unscoped_candidate.Educations) != 0 {
			educationsMap := []map[string]interface{}{}
			for _, education := range m_unscoped_candidate.Educations {
				educationsMap = append(educationsMap, map[string]interface{}{
					"university":   education.University,
					"major":        education.Major,
					"address":      education.Address,
					"is_graduated": education.IsGraduated,
					"start_at":     education.StartAt,
					"end_at":       education.EndAt,
					"gpa":          education.GPA,
				})
			}

			unscoped_candidate["educations"] = educationsMap
		} else {
			unscoped_candidate["educations"] = nil
		}
	}

	if strings.Contains(queries, "experiences") {
		if len(m_unscoped_candidate.Experiences) != 0 {
			experiencesMap := []map[string]interface{}{}
			for _, experience := range m_unscoped_candidate.Experiences {
				experiencesMap = append(experiencesMap, map[string]interface{}{
					"company_name":           experience.CompanyName,
					"position":               experience.Position,
					"location_address":       experience.LocationAddress,
					"type":                   experience.Type,
					"is_current":             experience.IsCurrent,
					"start_at":               experience.StartAt,
					"end_at":                 experience.EndAt,
					"description":            experience.Description,
					"attachment_document_id": SafelyNilPointer(&experience.AttachmentDocumentId),
				})

			}

			TransformsIdToPath([]string{
				"attachment_document_id",
			}, experiencesMap)

			unscoped_candidate["experiences"] = experiencesMap
		} else {
			unscoped_candidate["experiences"] = nil
		}
	}

	if strings.Contains(queries, "socials") {
		if len(m_unscoped_candidate.CandidateSocials) != 0 {
			socialsMap := []map[string]interface{}{}
			for _, candidateSocial := range m_unscoped_candidate.CandidateSocials {
				icon_image_id := uint(candidateSocial.Social.IconImageId)
				socialsMap = append(socialsMap, map[string]interface{}{
					"name":            candidateSocial.Social.Name,
					"url":             candidateSocial.Url,
					"icon_image_path": fmt.Sprintf("/api/v1/images/%v", SafelyNilPointer(&icon_image_id)),
				})
			}

			unscoped_candidate["socials"] = socialsMap
		} else {
			unscoped_candidate["socials"] = nil
		}
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    unscoped_candidate,
	})
}

func (c *CandidatesHandler) StoreAddresses(context *gin.Context) {
	addressesJSON := []CreateAddressJSON{}
	if errBind := context.ShouldBindJSON(&addressesJSON); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})

		context.Abort()
		return
	}

	addressesData := []models.Address{}
	for _, address := range addressesJSON {
		addressesData = append(addressesData, models.Address{
			Street:       address.Street,
			Neighborhood: address.Neighborhood,
			RuralArea:    address.RuralArea,
			SubDistrict:  address.SubDistrict,
			City:         address.City,
			Province:     address.Province,
			Country:      address.Country,
			PostalCode:   address.PostalCode,
			Type:         address.Type,
		})
	}

	bearerToken := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	tokenClaims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	errStoreAddresses := gormDB.Transaction(func(tx *gorm.DB) error {
		m_candidate := models.Candidate{}
		errGetCandidate := tx.Model(&models.Candidate{}).Select("id").Where("user_id = ?", tokenClaims.Id).First(&m_candidate).Error
		if errGetCandidate != nil {
			return errGetCandidate
		}

		errStoreAddress := tx.Create(addressesData).Error
		if errStoreAddress != nil {
			return errStoreAddress
		}

		m_candidate_address := []models.CandidateAddress{}
		for _, address := range addressesData {
			m_candidate_address = append(m_candidate_address, models.CandidateAddress{
				CandidateId: m_candidate.Id,
				AddressId:   address.ID,
				CreatedAt:   time.Now(),
			})
		}

		storeCandidateAddress := tx.Create(&m_candidate_address)
		if storeCandidateAddress.Error != nil {
			return storeCandidateAddress.Error
		}

		return nil
	})

	if errStoreAddresses != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreAddresses.Error(),
			"message": "failed storing candidate addresses",
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    fmt.Sprintf("%v candidate address stored successfully", len(addressesData)),
	})
}

func (c *CandidatesHandler) UpdateAddress(context *gin.Context) {
	updateAddressJSON := UpdateAddressJSON{}
	if errBind := context.ShouldBindJSON(&updateAddressJSON); errBind != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})

		context.Abort()
		return
	}

	addressFields := map[string]interface{}{}
	v := reflect.ValueOf(updateAddressJSON)

	for i := 0; i < v.NumField(); i++ {
		keyName := v.Type().Field(i).Tag.Get("json")
		value := v.Field(i)

		if value.Kind() == reflect.Pointer && !value.IsNil() {
			addressFields[keyName] = value.Elem().Interface()
		}
	}

	addressFields["updated_at"] = time.Now()

	gormDB, _ := initializer.GetGorm()
	updateAddress := gormDB.Model(&models.Address{ID: updateAddressJSON.Id}).Updates(addressFields)
	if updateAddress.RowsAffected == 0 {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("couldn't update address with id (%v) or record not found", updateAddressJSON.Id),
			"message": fmt.Sprintf("failed updating address with id (%v)", updateAddressJSON.Id),
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("%v address updated successfully", updateAddress.RowsAffected),
	})
}

func (c *CandidatesHandler) AddressGetById(context *gin.Context) {
	param_addressId, _ := strconv.ParseUint(context.Param("id"), 10, 64)

	gormDB, _ := initializer.GetGorm()
	address := map[string]interface{}{}
	errGetAddress := gormDB.Model(&models.Address{}).Select([]string{
		"id",
		"street",
		"neighborhood",
		"rural_area",
		"sub_district",
		"city",
		"province",
		"country",
		"postal_code",
		"type",
	}).
		Where("id = ?", uint(param_addressId)).
		First(&address).Error

	if errGetAddress != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetAddress.Error(),
			"message": fmt.Sprintf("failed getting address with id (%v)", param_addressId),
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    address,
	})
}

func (c *CandidatesHandler) AddressDeleteById(context *gin.Context) {
	param_addressId, _ := strconv.ParseUint(context.Param("id"), 10, 32)

	gormDB, _ := initializer.GetGorm()

	deleteAddress := gormDB.Transaction(func(tx *gorm.DB) error {
		m_candidateAddress := models.CandidateAddress{}
		deleteCandidateAddress := tx.Model(&models.CandidateAddress{}).Where("address_id = ?", param_addressId).Delete(&m_candidateAddress)
		if deleteCandidateAddress.RowsAffected == 0 {
			return fmt.Errorf("couldn't delete candidate_address with id (%v)", param_addressId)
		}

		m_address := models.Address{ID: uint(param_addressId)}
		deleteAddress := tx.Delete(&m_address)
		if deleteAddress.RowsAffected == 0 {
			return fmt.Errorf("couldn't delete address with id (%v)", param_addressId)
		}

		return nil
	})

	if deleteAddress != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "couldn't delete address or record not found",
			"message": fmt.Sprintf("failed deleting address with id (%v)", param_addressId),
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("address with id (%v) deleted successfully", param_addressId),
	})
}

func (c *CandidatesHandler) StoreEducations(context *gin.Context) {
	bearerToken := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	tokenClaims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	m_candidate := map[string]interface{}{}
	getCandidateId := gormDB.Model(&models.Candidate{}).Select("id").Where("user_id = ?", tokenClaims.Id).First(&m_candidate)

	if getCandidateId.Error != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   getCandidateId.Error.Error(),
			"message": "failed getting candidate id by user_id",
		})

		context.Abort()
		return
	}

	education_data := []models.Education{}
	educationsJSON := []CreateEducationJSON{}
	if errBind := context.ShouldBindJSON(&educationsJSON); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})

		context.Abort()
		return
	}

	for _, education := range educationsJSON {
		education_data = append(education_data, models.Education{
			University:  education.University,
			Address:     education.Address,
			Major:       education.Major,
			IsGraduated: education.IsGraduated,
			StartAt:     education.StartAt,
			EndAt:       education.EndAt,
			GPA:         education.GPA,
			CandidateId: m_candidate["id"].(string),
			CreatedAt:   time.Now(),
		})
	}

	storeEducations := gormDB.Model(&models.Education{}).Create(education_data)
	if storeEducations.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   storeEducations.Error.Error(),
			"message": "failed storing educations data",
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("%v educations stored successfully", storeEducations.RowsAffected),
	})
}

func (c *CandidatesHandler) UpdateEducation(context *gin.Context) {
	educationJSON := UpdateEducationJSON{}
	if errBind := context.ShouldBindJSON(&educationJSON); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})

		context.Abort()
		return
	}

	educationFields := map[string]interface{}{}
	v := reflect.ValueOf(educationJSON)

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		jsonTag := v.Type().Field(i).Tag.Get("json")

		if jsonTag == "" {
			continue
		}

		if value.Kind() == reflect.Ptr {
			if !value.IsNil() {
				educationFields[jsonTag] = value.Elem().Interface()
			}
		}
	}

	gormDB, _ := initializer.GetGorm()
	updateEducation := gormDB.Model(&models.Education{ID: educationJSON.Id}).Updates(educationFields)
	if updateEducation.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   updateEducation.Error.Error(),
			"message": fmt.Sprintf("failed updating education with id (%v)", educationJSON.Id),
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success":        true,
		"data":           fmt.Sprintf("%v education updated successfully", updateEducation.RowsAffected),
		"updated_fields": educationFields,
	})
}

func (c *CandidatesHandler) EducationGetById(context *gin.Context) {
	param_educationId := context.Param("id")

	m_education := map[string]interface{}{}
	gormDB, _ := initializer.GetGorm()
	errGetEducation := gormDB.Model(&models.Education{}).
		Select([]string{
			"university",
			"address",
			"major",
			"is_graduated",
			"start_at",
			"end_at",
			"gpa",
		}).Where("id = ?", param_educationId).
		First(&m_education).Error

	if errGetEducation != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetEducation.Error(),
			"message": fmt.Sprintf("failed getting education with id (%v)", param_educationId),
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    m_education,
	})
}

func (c *CandidatesHandler) EducationDeleteById(context *gin.Context) {
	gormDB, _ := initializer.GetGorm()
	educationId, _ := strconv.ParseUint(context.Param("id"), 10, 64)
	m_education := models.Education{
		ID: uint(educationId),
	}
	deleteEducation := gormDB.Delete(&m_education)
	if deleteEducation.RowsAffected == 0 {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("education with ID %v maybe not found", educationId),
			"message": fmt.Sprintf("failed deleting education with id (%v)", educationId),
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("%v education deleted successfully", deleteEducation.RowsAffected),
	})
}
