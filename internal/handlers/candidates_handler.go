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

type CandidateById struct {
	Id                       string
	Expertise                string
	AboutMe                  string
	DateOfBirth              time.Time
	BackgroundProfileImageId uint
	ProfileImageId           uint
	CVDocumentId             uint
	UserId                   string
	User                     *models.User `gorm:"foreignKey:UserId"`
	Educations               []*models.Education
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
		Name:     image.Filename,
		MimeType: http.DetectContentType(imageInByte),
		Size:     image.Size,
		Blob:     imageInByte,
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
	m_image := &models.Image{}
	updateImageFields := gormDB.Model(&m_image).Where("id = ?", imageId).Updates(models.Image{Name: image.Filename, Size: image.Size, Blob: imageBinaryData})

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
		Purpose:  purpose,
		Name:     document.Filename,
		MimeType: http.DetectContentType(docBinaryData),
		Size:     document.Size,
		Blob:     docBinaryData,
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
		Model: gorm.Model{ID: documentId},
	}

	errUpdateDocument := gormDB.Model(&m_document).
		Updates(models.Document{
			Name:     document.Filename,
			MimeType: http.DetectContentType(documentBinaryData),
			Size:     document.Size,
			Blob:     documentBinaryData}).
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

		final_candidates = append(final_candidates, candidateMap)
	}

	context.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    final_candidates,
	})
}

func (c *CandidatesHandler) GetById(context *gin.Context) {
	candidateId := context.Param("id")
	querys, _ := context.GetQuery("includes")

	if candidateId == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing parameter :id",
			"message": "parameter :id not specified (e.g /candidates/:your_candidate_id)",
		})

		context.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	candidate := make(map[string]interface{})
	var (
		user        map[string]interface{}
		address     map[string]interface{}
		skills      []map[string]interface{}
		educations  []map[string]interface{}
		experiences []map[string]interface{}
		socials     []map[string]interface{}
	)
	errGetCandidate := gormDB.Transaction(func(tx *gorm.DB) error {
		getCandidate := tx.Model(&models.Candidate{}).Where("id = ?", candidateId).Scan(&candidate)
		if getCandidate.Error != nil {
			return getCandidate.Error
		} else if getCandidate.RowsAffected == 0 {
			return fmt.Errorf(fmt.Sprintf("not found candidate with id (%s)", candidateId))
		}

		TransformsIdToPath([]string{
			"background_profile_image_id",
			"profile_image_id",
			"cv_document_id",
		}, candidate)

		if strings.Contains(querys, "user") {
			errUser := tx.Model(&models.User{}).Where("id = ?", candidate["user_id"]).First(&user).Error
			if errUser != nil {
				return errUser
			}
			candidate["user"] = &user
		}

		if strings.Contains(querys, "address") {
			errAddress := tx.Model(&models.CandidateAddress{}).
				Select([]string{
					"addresses.street",
					"addresses.neighborhood",
					"addresses.rural_area",
					"addresses.sub_district",
					"addresses.city",
					"addresses.province",
					"addresses.country",
				}).
				Joins("inner join addresses on candidate_addresses.address_id = addresses.id").
				Where("candidate_addresses.candidate_id = ? AND addresses.type = ?", candidateId, "home").
				Limit(1).
				Find(&address).Error

			if errAddress != nil {
				return errAddress
			}

			candidate["address"] = &address
		}

		if strings.Contains(querys, "skills") {
			errSkills := tx.Model(&models.CandidateSkill{}).
				Select([]string{
					"skills.name",
					"skills.skill_icon_image_id",
				}).
				Joins("INNER JOIN skills ON candidate_skills.skill_id = skills.id").
				Where("candidate_skills.candidate_id = ?", candidateId).
				Find(&skills).Error

			if errSkills != nil {
				return errSkills
			}

			TransformsIdToPath([]string{
				"skill_icon_image_id",
			}, skills)

			candidate["skills"] = &skills
		}

		if strings.Contains(querys, "educations") {
			errEducations := tx.Model(&models.Education{}).Select("id", "university", "major", "gpa", "address").Where("candidate_id = ?", candidateId).Scan(&educations).Error
			if errEducations != nil {
				return errEducations
			}

			for _, education := range educations {
				if value, exists := education["gpa"].([]byte); exists {
					gpa, errParseFloat := strconv.ParseFloat(string(value), 64)
					if errParseFloat != nil {
						return errParseFloat
					}
					education["gpa"] = gpa
				}
			}

			candidate["educations"] = &educations
		}

		if strings.Contains(querys, "experiences") {
			errExperiences := tx.Model(&models.Experience{}).
				Select([]string{
					"company_name",
					"position",
					"type",
					"location_address",
					"start_at",
					"end_at",
					"description",
					"attachment_document_id",
				}).
				Where("candidate_id = ?", candidateId).
				Find(&experiences).Error

			if errExperiences != nil {
				return errExperiences
			}

			TransformsIdToPath([]string{
				"attachment_document_id",
			}, experiences)

			candidate["experiences"] = &experiences
		}

		if strings.Contains(querys, "socials") {
			errSocials := tx.Model(&models.CandidateSocial{}).
				Select([]string{
					"candidate_socials.url",
					"socials.name",
					"socials.icon_image_id",
				}).
				Joins("INNER JOIN socials ON candidate_socials.social_id = socials.id").
				Where("candidate_id = ?", candidateId).
				Find(&socials).Error

			if errSocials != nil {
				return errSocials
			}

			TransformsIdToPath([]string{
				"icon_image_id",
			}, socials)

			candidate["socials"] = &socials
		}

		return nil
	})

	if errGetCandidate != nil {
		message := fmt.Sprintf("database error, failed getting candidate by id (%s)", candidateId)
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetCandidate.Error(),
			"message": message,
		})

		context.Abort()
		return
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
		errDeleteCandidate := tx.Delete(&models.Candidate{}, "id = ?", candidateId).Error
		if errDeleteCandidate != nil {
			return errDeleteCandidate
		}

		errDeleteUser := tx.Delete(&models.User{}, "id = ?", tokenClaims.Id).Error
		if errDeleteUser != nil {
			return errDeleteUser
		}
		return nil
	})

	if errDeleting != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errDeleting.Error(),
			"message": fmt.Sprintf("failed deleting candidate with id (%s)", candidateId),
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
	// if param_candidateId == "" {
	// 	context.JSON(http.StatusBadRequest, gin.H{
	// 		"success": false,
	// 		"error":   "candidate :id param required",
	// 		"message": "specify you candidate :id as url param, kids. (e.g /api/v1/candidates/unscoped/:your_candidate_id)",
	// 	})

	// 	context.Abort()
	// 	return
	// }

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
	// like Get y Id Hanlder, but with .Unscoped() method
}
