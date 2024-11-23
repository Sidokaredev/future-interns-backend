package handlers

import (
	"encoding/json"
	"errors"
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
	"gorm.io/gorm"
)

type EmployerHandlers struct {
}

type CreateEmployerForm struct {
	Name            string `form:"name" binding:"required"`
	LegalName       string `form:"legal_name" binding:"required"`
	Location        string `form:"location" binding:"required"`
	Founded         uint   `form:"founded" binding:"required"`
	Founder         string `form:"founder" binding:"required"`
	TotalOfEmployee string `form:"total_of_employee" binding:"required"`
	Website         string `form:"website" binding:"required"`
	Description     string `form:"description" binding:"required"`
}
type UpdateEmployerForm struct {
	Name            *string `form:"name"`
	LegalName       *string `form:"legal_name"`
	Location        *string `form:"location"`
	Founded         *uint   `form:"founded"`
	Founder         *string `form:"founder"`
	TotalOfEmployee *string `form:"total_of_employee"`
	Website         *string `form:"website"`
	Description     *string `form:"description"`
}

type CreateHeadquarterForm struct {
	Name string `form:"name" binding:"required"`
	Type string `form:"type" binding:"required"`
	/* address */
	Street       string  `form:"street" binding:"required"`
	Neighborhood *string `form:"neighborhood"`
	RuralArea    *string `form:"rural_area"`
	SubDistrict  string  `form:"sub_district" binding:"required"`
	City         string  `form:"city" binding:"required"`
	Province     string  `form:"province" binding:"required"`
	Country      string  `form:"country" binding:"required"`
	PostalCode   int     `form:"postal_code" binding:"required"`
	/* address type should be headquater */
}
type UpdateHeadquarterForm struct {
	Name *string `form:"name"`
	Type *string `form:"type"`
	/* address */
	Street       *string `form:"street"`
	Neighborhood *string `form:"neighborhood"`
	RuralArea    *string `form:"rural_area"`
	SubDistrict  *string `form:"sub_district"`
	City         *string `form:"city"`
	Province     *string `form:"province"`
	Country      *string `form:"country"`
	PostalCode   *int    `form:"postal_code"`
	/* address type should be headquater */
}

type CreateEmployerSocialJSON struct {
	SocialId uint   `json:"social_id" binding:"required"`
	Url      string `json:"url" binding:"required"`
}

type CreateVacancyForm struct {
	Position        string `form:"position" binding:"required"`
	Description     string `form:"description" binding:"required"`
	Qualification   string `form:"qualification" binding:"required"`
	Responsibility  string `form:"responsibility" binding:"required"`
	LineIndustry    string `form:"line_industry" binding:"required"`
	EmployeeType    string `form:"employee_type" binding:"required"`
	MinExperience   string `form:"min_experience" binding:"required"`
	Salary          uint   `form:"salary" binding:"required"`
	WorkArrangement string `form:"work_arrangement" binding:"required"`
	SLA             int32  `form:"sla" binding:"required"` // should given a default value from Front-End when making request
	IsInactive      *bool  `form:"is_inactive" binding:"required"`
}

type UpdateVacancyForm struct {
	Position        *string `form:"position"`
	Description     *string `form:"description"`
	Qualification   *string `form:"qualification"`
	Responsibility  *string `form:"responsibility"`
	LineIndustry    *string `form:"line_industry"`
	EmployeeType    *string `form:"employee_type"`
	MinExperience   *string `form:"min_experience"`
	Salary          *uint   `form:"salary"`
	WorkArrangement *string `form:"work_arrangement"`
	SLA             *int32  `form:"sla"`
	IsInactive      *bool   `form:"is_inactive"`
}

type CreateAssessmentForm struct {
	Name           string    `form:"name" binding:"required"`
	Note           string    `form:"note" binding:"required"`
	AssessmentLink *string   `form:"assessment_link"`
	StartDate      time.Time `form:"start_date" binding:"required"`
	DueDate        time.Time `form:"due_date" binding:"required"`
	VacancyId      string    `form:"vacancy_id" binding:"required"`
}

type UpdateAssessmentForm struct {
	Name           *string    `form:"name"`
	Note           *string    `form:"note"`
	AssessmentLink *string    `form:"assessment_link"`
	StartDate      *time.Time `form:"start_date"`
	DueDate        *time.Time `form:"due_date"`
}

type CreateInterviewForm struct {
	PipelineId  string    `form:"pipeline_id" binding:"required"`
	VacancyId   string    `form:"vacancy_id" binding:"required"`
	Date        time.Time `form:"date" binding:"required"`
	Location    string    `form:"location" binding:"required"`
	LocationURL string    `form:"location_url" binding:"required"`
	// Status      string    `form:"status" binding:"required"`
	// Result      string    `form:"result" binding:"required"`
}

type UpdateInterviewForm struct {
	Date        *time.Time `form:"date" binding:"required"`
	Location    *string    `form:"location" binding:"required"`
	LocationURL *string    `form:"location_url" binding:"required"`
	Status      *string    `form:"status" binding:"required"`
	Result      *string    `form:"result" binding:"required"`
}

/* EMPLOYER */
func (e *EmployerHandlers) StoreEmployer(context *gin.Context) {
	/*
		check permissions here
	*/
	employerForm := CreateEmployerForm{}
	if errBind := context.ShouldBind(&employerForm); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your Form-Data fields, kids",
		})

		context.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(context.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	ch_uuid := make(chan string)
	go GenUuid(employerForm.Name, ch_uuid)

	var (
		background_profile_image_status string
		profile_image_status            string
	)
	backgroundProfileImage, errBackgroundProfile := context.FormFile("background_profile_image")
	profileImage, errProfileImage := context.FormFile("profile_image")

	gormDB, _ := initializer.GetGorm()

	errCreateEmployer := gormDB.Transaction(func(tx *gorm.DB) error {
		timeNow := time.Now()
		m_employer := models.Employer{
			Id:              <-ch_uuid,
			Name:            employerForm.Name,
			LegalName:       employerForm.LegalName,
			Location:        employerForm.Location,
			Founded:         employerForm.Founded,
			Founder:         employerForm.Founder,
			TotalOfEmployee: employerForm.TotalOfEmployee,
			Website:         employerForm.Website,
			Description:     employerForm.Description,
			UserId:          claims.Id,
			CreatedAt:       timeNow,
			UpdatedAt:       &timeNow,
		}

		if errBackgroundProfile == nil {
			imageData, errOpen := backgroundProfileImage.Open()
			if errOpen != nil {
				background_profile_image_status = errOpen.Error()
				goto StoreProfileImage
			}

			imageBinaryData := make([]byte, backgroundProfileImage.Size)
			_, errRead := imageData.Read(imageBinaryData)
			if errRead != nil {
				background_profile_image_status = errRead.Error()
				goto StoreProfileImage
			}

			m_image := models.Image{
				Name:      backgroundProfileImage.Filename,
				MimeType:  http.DetectContentType(imageBinaryData),
				Size:      backgroundProfileImage.Size,
				Blob:      imageBinaryData,
				CreatedAt: timeNow,
				UpdatedAt: &timeNow,
			}

			errStoreImage := tx.Create(&m_image).Error
			if errStoreImage != nil {
				background_profile_image_status = errStoreImage.Error()
				goto StoreProfileImage
			}

			background_profile_image_status = "stored successfully"
			m_employer.BackgroundProfileImageId = &m_image.ID

		}

	StoreProfileImage: // label

		if errProfileImage == nil {
			imageData, errOpen := profileImage.Open()
			if errOpen != nil {
				profile_image_status = errOpen.Error()
				goto StoreEmployer
			}

			imageBinaryData := make([]byte, profileImage.Size)
			_, errRead := imageData.Read(imageBinaryData)
			if errRead != nil {
				profile_image_status = errRead.Error()
				goto StoreEmployer
			}

			m_image := models.Image{
				Name:      profileImage.Filename,
				MimeType:  http.DetectContentType(imageBinaryData),
				Size:      profileImage.Size,
				Blob:      imageBinaryData,
				CreatedAt: timeNow,
				UpdatedAt: &timeNow,
			}

			errStoreImage := tx.Create(&m_image).Error
			if errStoreImage != nil {
				profile_image_status = errStoreImage.Error()
				goto StoreEmployer
			}

			profile_image_status = "stored successfully"
			m_employer.ProfileImageId = &m_image.ID
		}

	StoreEmployer: // label

		errStoreEmployer := tx.Create(&m_employer).Error
		if errStoreEmployer != nil {
			return errStoreEmployer
		}

		return nil
	})
	if errCreateEmployer != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCreateEmployer.Error(),
			"message": "failed creating employer",
		})

		context.Abort()
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message":                         "employer data stored successfully",
			"background_profile_image_status": background_profile_image_status,
			"profile_image_status":            profile_image_status,
		},
	})
}

func (e *EmployerHandlers) UpdateEmployer(ctx *gin.Context) {
	/*
		check permissions here
	*/
	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	employerForm := UpdateEmployerForm{}
	ctx.ShouldBind(&employerForm)

	employerFormMap := map[string]interface{}{}
	v := reflect.ValueOf(employerForm)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Tag.Get("form")
		value := v.Field(i)

		if value.Kind() == reflect.Pointer && !value.IsNil() {
			employerFormMap[fieldName] = value.Interface()
		}
	}

	ch_update_image := make(chan ChannelImage, 2)
	imageCounter := 2
	var (
		background_profile_image_status string
		profile_image_status            string
	)
	backgroundProfileImage, errBackgroudProfileImage := ctx.FormFile("background_profile_image")
	profileImage, errProfileImage := ctx.FormFile("profile_image")

	gormDB, _ := initializer.GetGorm()
	errUpdateEmployer := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).Select([]string{
			"id",
			"background_profile_image_id",
			"profile_image_id",
		}).Where("user_id = ?", claims.Id).First(&employer).Error

		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		if errBackgroudProfileImage == nil {
			if employer["background_profile_image_id"] != nil {
				imageId, _ := employer["background_profile_image_id"].(*uint)
				go UpdateImage(*imageId, "background_profile_image", backgroundProfileImage, ch_update_image)
			} else {
				imageCounter -= 1 // decrement channel counter

				imageData, errOpen := backgroundProfileImage.Open()
				if errOpen != nil {
					background_profile_image_status = errOpen.Error()
					goto StoreProfileImage
				}

				imageBinaryData := make([]byte, backgroundProfileImage.Size)
				_, errRead := imageData.Read(imageBinaryData)
				if errRead != nil {
					background_profile_image_status = errRead.Error()
					goto StoreProfileImage
				}

				m_image := models.Image{
					Name:     backgroundProfileImage.Filename,
					MimeType: http.DetectContentType(imageBinaryData),
					Size:     backgroundProfileImage.Size,
					Blob:     imageBinaryData,
				}

				errStoreImage := tx.Create(&m_image).Error
				if errStoreImage != nil {
					background_profile_image_status = errStoreImage.Error()
					goto StoreProfileImage
				}

				background_profile_image_status = "new background profile image added successfully"
				employerFormMap["background_profile_image_id"] = m_image.ID
			}
		} else {
			background_profile_image_status = errBackgroudProfileImage.Error()
			imageCounter -= 1
		}

	StoreProfileImage: // label

		if errProfileImage == nil {
			if employer["profile_image_id"] != nil {
				imageId, _ := employer["profile_image_id"].(*uint)
				go UpdateImage(*imageId, "profile_image", profileImage, ch_update_image)
			} else {
				imageCounter -= 1

				imageData, errOpen := profileImage.Open()
				if errOpen != nil {
					profile_image_status = errOpen.Error()
					goto UpdateEmployer
				}

				imageBinaryData := make([]byte, profileImage.Size)
				_, errRead := imageData.Read(imageBinaryData)
				if errRead != nil {
					profile_image_status = errRead.Error()
					goto UpdateEmployer
				}

				m_image := models.Image{
					Name:     profileImage.Filename,
					MimeType: http.DetectContentType(imageBinaryData),
					Size:     profileImage.Size,
					Blob:     imageBinaryData,
				}

				errStoreImage := tx.Create(&m_image).Error
				if errStoreImage != nil {
					profile_image_status = errStoreImage.Error()
					goto UpdateEmployer
				}

				profile_image_status = "new profile image added successfully"
				employerFormMap["profile_image_id"] = m_image.ID
			}
		} else {
			profile_image_status = errProfileImage.Error()
			imageCounter -= 1 // decrement channel counter
		}

	UpdateEmployer:

		update := tx.Model(&models.Employer{}).Where("id = ?", employer["id"]).Updates(employerFormMap)
		if update.RowsAffected == 0 {
			log.Printf("%v employer row affected", update.RowsAffected)
			return errors.New("unable updating employer. data might not available in the database")
		}

		return nil
	})

	for i := 0; i < imageCounter; i++ {
		data := <-ch_update_image
		switch data.Key {
		case "background_profile_image":
			background_profile_image_status = data.Status
		case "profile_image":
			profile_image_status = data.Status
		}
	}

	if errUpdateEmployer != nil {
		log.Println(background_profile_image_status) // update status background profile image
		log.Println(profile_image_status)            // update status profile image status
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errUpdateEmployer.Error(),
			"message": "failed updating employer data",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message":                   "employer data updated successfully",
			"background_profile_status": background_profile_image_status,
			"profile_image_status":      profile_image_status,
		},
	})
}

func (e *EmployerHandlers) GetEmployer(ctx *gin.Context) {
	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	includes := ctx.Query("includes")

	gormDB, _ := initializer.GetGorm()
	m_employer := models.Employer{}
	errGetEmployer := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select("id").
			Where("user_id = ?", claims.Id).
			First(&employer).Error

		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		db := tx.Model(&models.Employer{})

		if strings.Contains(includes, "user") {
			db = db.Preload("User")
		}

		if strings.Contains(includes, "headquarters") {
			db = db.Preload("Headquarters.Address")
		}

		if strings.Contains(includes, "socials") {
			db = db.Preload("EmployerSocials.Social")
		}

		errSearchEmployer := db.First(&m_employer).Error
		if errSearchEmployer != nil {
			return errSearchEmployer
		}

		return nil
	})

	if errGetEmployer != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errGetEmployer.Error(),
			"message": "failed getting employer data",
		})

		ctx.Abort()
		return
	}

	/*
		transform data
	*/
	employerMap := map[string]interface{}{
		"name":                        m_employer.Name,
		"legal_name":                  m_employer.LegalName,
		"location":                    m_employer.Location,
		"founded":                     m_employer.Founded,
		"founder":                     m_employer.Founder,
		"total_of_employee":           m_employer.TotalOfEmployee,
		"website":                     m_employer.Website,
		"description":                 m_employer.Description,
		"background_profile_image_id": SafelyNilPointer(m_employer.BackgroundProfileImageId),
		"profile_image_id":            SafelyNilPointer(m_employer.ProfileImageId),
		"headquarters":                []gin.H{},
		"socials":                     []map[string]interface{}{},
	}

	TransformsIdToPath([]string{
		"background_profile_image_id",
		"profile_image_id",
	}, employerMap)

	if m_employer.User != nil {
		employerMap["user"] = gin.H{
			"id":       m_employer.User.Id,
			"fullname": m_employer.User.Fullname,
			"email":    m_employer.User.Email,
		}
	}

	for _, headquarter := range m_employer.Headquarters {
		employerMap["headquarters"] = append(employerMap["headquarters"].([]gin.H), gin.H{
			"address_id": headquarter.AddressId,
			"name":       headquarter.Name,
			"type":       headquarter.Type,
			"address": gin.H{
				"street":       headquarter.Address.Street,
				"neighborhood": headquarter.Address.Neighborhood,
				"rural_area":   headquarter.Address.RuralArea,
				"sub_district": headquarter.Address.SubDistrict,
				"city":         headquarter.Address.City,
				"province":     headquarter.Address.Province,
				"country":      headquarter.Address.Country,
				"postal_code":  headquarter.Address.PostalCode,
				"type":         headquarter.Address.Type,
			},
		})
	}

	for _, social := range m_employer.EmployerSocials {
		employerMap["socials"] = append(employerMap["socials"].([]map[string]interface{}), map[string]interface{}{
			"name":          social.Social.Name,
			"url":           social.Url,
			"icon_image_id": social.Social.IconImageId,
		})
	}

	TransformsIdToPath([]string{"icon_image_id"}, employerMap["socials"])
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    employerMap,
	})
}

func (e *EmployerHandlers) DeleteEmployer(ctx *gin.Context) {
	/*
		check permissions here
	*/

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	errDeleteEmployer := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select([]string{
				"id",
			}).
			Where("user_id = ?", claims.Id).
			First(&employer).Error

		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		deleteEmployer := tx.Where("id = ?", employer["id"]).Delete(&models.Employer{})
		if deleteEmployer.RowsAffected == 0 {
			return errors.New("unable deleting employer. data might be not available in the database")
		}

		return nil
	})

	if errDeleteEmployer != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteEmployer.Error(),
			"message": "failed deleting employer",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "employer data deleted successfully",
		},
	})
}

/* HEADQUARTER */
func (e *EmployerHandlers) StoreHeadquarter(ctx *gin.Context) {
	/*
		check permissions here
	*/
	headquarterForm := CreateHeadquarterForm{}
	if errBind := ctx.ShouldBind(&headquarterForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your Form-Data fields, kids",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	errCreateHeadquarter := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select([]string{
				"id",
			}).
			Where("user_id = ?", claims.Id).
			First(&employer).Error

		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		m_address := models.Address{
			Street:       headquarterForm.Street,
			Neighborhood: headquarterForm.Neighborhood,
			RuralArea:    headquarterForm.RuralArea,
			SubDistrict:  headquarterForm.SubDistrict,
			City:         headquarterForm.City,
			Province:     headquarterForm.Province,
			Country:      headquarterForm.Country,
			PostalCode:   headquarterForm.PostalCode,
			Type:         "headquarter",
			CreatedAt:    time.Now(),
		}
		errCreateAddress := tx.Create(&m_address).Error
		if errCreateAddress != nil {
			return errCreateAddress
		}

		m_headquarter := models.Headquarter{
			EmployerId: employer["id"].(string),
			AddressId:  m_address.ID,
			Name:       headquarterForm.Name,
			Type:       "branch",
			CreatedAt:  time.Now(),
		}
		errCreateHeadquarter := tx.Create(&m_headquarter).Error
		if errCreateHeadquarter != nil {
			return errCreateHeadquarter
		}

		return nil
	})

	if errCreateHeadquarter != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCreateHeadquarter.Error(),
			"message": "failed creating headquarter",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": "headquarter created successfully",
		},
	})
}

func (e *EmployerHandlers) UpdateHeadquarter(ctx *gin.Context) {
	/*
		check permissions here
	*/
	addressId := ctx.Param("id")
	if addressId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing address id as url parameter",
			"message": "address id is required as url parameter",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	headquarterForm := UpdateHeadquarterForm{}
	ctx.ShouldBind(&headquarterForm)

	v := reflect.ValueOf(headquarterForm)

	addressMap := map[string]interface{}{}
	headquarterMap := map[string]interface{}{}
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Tag.Get("form")
		value := v.Field(i)

		if value.Kind() == reflect.Pointer && !value.IsNil() {
			if fieldName != "name" && fieldName != "type" {
				addressMap[fieldName] = value.Interface()
			} else {
				headquarterMap[fieldName] = value.Interface()
			}
		}
	}

	gormDB, _ := initializer.GetGorm()
	errUpdateHeadquarter := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).Select("id").Where("user_id = ?", claims.Id).First(&employer).Error
		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		updateAddress := tx.Model(&models.Address{}).Where("id = ?", addressId).Updates(addressMap)
		if updateAddress.RowsAffected == 0 {
			return fmt.Errorf("unable updating address. address with id (%v) might not be available in the database", addressId)
		}

		updateHeadquarter := tx.Model(&models.Headquarter{}).Where("employer_id = ? AND address_id = ?", employer["id"], addressId).Updates(headquarterMap)
		if updateHeadquarter.RowsAffected == 0 {
			return fmt.Errorf("unable updating headquarter. data might not be available in the database")
		}

		return nil
	})

	if errUpdateHeadquarter != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errUpdateHeadquarter.Error(),
			"message": fmt.Sprintf("failed updating headquarter with address id (%v)", addressId),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "headquarter updated successfully",
		},
	})
}

func (e *EmployerHandlers) GetHeadquarter(ctx *gin.Context) {
	/*
		confused -> do I need to retrieve a single information about headquarter?
	*/
}

func (e *EmployerHandlers) DeleteHeadquarter(ctx *gin.Context) {
	addressId := ctx.Param("id")
	if addressId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing address id as url parameter",
			"message": "address id is required, please specify as url parameter",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	errDeleteHeadquarter := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select("id").
			Where("user_id = ?", claims.Id).
			First(&employer).Error

		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		deleteHeadquarter := tx.Where("employer_id = ? AND address_id = ?", employer["id"], addressId).Delete(&models.Headquarter{})
		if deleteHeadquarter.RowsAffected == 0 {
			return errors.New("unable deleting headquarter. data might not available in the database")
		}

		deleteAddress := tx.Where("id = ?", addressId).Delete(&models.Address{})
		if deleteAddress.RowsAffected == 0 {
			return errors.New("unable deleting address. data might not available in the database")
		}

		return nil
	})

	if errDeleteHeadquarter != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteHeadquarter.Error(),
			"message": fmt.Sprintf("failed deleting headquarter with address id %v", addressId),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "headquarter deleted successfully",
		},
	})
}

func (e *EmployerHandlers) ListHeadquarter(ctx *gin.Context) {
	/*
		check permissions here
	*/
	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	headquarters := []models.Headquarter{}

	gormDB, _ := initializer.GetGorm()
	errListHeadquarter := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select("id").
			Where("user_id = ?", claims.Id).
			First(&employer).Error

		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		listHeadquarter := tx.Preload("Address").Where("employer_id = ?", employer["id"]).Find(&headquarters)
		if listHeadquarter.RowsAffected == 0 {
			return fmt.Errorf("%v records found. no headquarters data availbale at the moment", listHeadquarter.RowsAffected)
		}

		return nil
	})

	if errListHeadquarter != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errListHeadquarter.Error(),
			"message": "failed getting list of headquater",
		})

		ctx.Abort()
		return
	}

	headquartersMap := []map[string]interface{}{}
	for _, headquarter := range headquarters {
		headquartersMap = append(headquartersMap, map[string]interface{}{
			"name": headquarter.Name,
			"type": headquarter.Type,
			"address": map[string]interface{}{
				"id":           headquarter.Address.ID,
				"street":       headquarter.Address.Street,
				"neighborhood": headquarter.Address.Neighborhood,
				"rural_area":   headquarter.Address.RuralArea,
				"sub_district": headquarter.Address.SubDistrict,
				"city":         headquarter.Address.City,
				"province":     headquarter.Address.Province,
				"country":      headquarter.Address.Country,
				"postal_code":  headquarter.Address.PostalCode,
			},
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": false,
		"data":    headquartersMap,
	})
}

/* OFFICE IMAGE */
func (e *EmployerHandlers) StoreOfficeImage(ctx *gin.Context) {
	/*
		check permisions here
	*/
	officeImages, errOfficeImages := ctx.MultipartForm()
	if errOfficeImages != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errOfficeImages.Error(),
			"message": "there was an issue with your image files.",
		})

		ctx.Abort()
		return
	}
	imageFiles := officeImages.File["office_images"]

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	m_images, status := MultipleImageData(imageFiles)
	if len(m_images) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "empty images or invalid image files",
			"message": "no images attached. it might might be invalid image files. make sure you select a valid image file",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errStoreOfficeImages := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		getEmployerId := tx.Model(&models.Employer{}).
			Select("id").
			Where("user_id = ?", claims.Id).
			First(&employer)
		if getEmployerId.Error != nil {
			return getEmployerId.Error
		}

		storeImages := tx.Create(&m_images)
		if storeImages.Error != nil {
			return storeImages.Error
		}

		m_officeImages := []models.OfficeImage{}
		for _, image := range m_images {
			m_officeImages = append(m_officeImages, models.OfficeImage{
				EmployerId: employer["id"].(string),
				ImageId:    image.ID,
				CreatedAt:  time.Now(),
			})
		}

		storeOfficeImages := tx.Create(m_officeImages)
		if storeOfficeImages.Error != nil {
			return storeOfficeImages.Error
		}

		return nil
	})

	if errStoreOfficeImages != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreOfficeImages.Error(),
			"message": "failed storing office images",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("%v images stored successfully", len(m_images)),
			"status":  status,
		},
	})
}

func (e *EmployerHandlers) UpdateOfficeImage(ctx *gin.Context) {
	/*
		check permission here
	*/
	imageId, errParse := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "url parameter of image id should be a valid number",
		})

		ctx.Abort()
		return
	}

	officeImage, errOfficeImage := ctx.FormFile("office_image")
	if errOfficeImage != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errOfficeImage.Error(),
			"message": "no image file attached, select the image first",
		})

		ctx.Abort()
		return
	}

	m_image, errImage := ImageData(officeImage)
	if errImage != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errImage.Error(),
			"message": "failed while processing an image",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	m_image.ID = uint(imageId) // assign image id from url parameter

	errUpdateOfficeImage := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select("id").
			Where("user_id = ?", claims.Id).
			First(&employer).Error
		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		updateImage := tx.Updates(m_image) // no need ampersand (&) because the value already pointer
		if updateImage.RowsAffected == 0 {
			return fmt.Errorf("image with id %v might not available in database", imageId)
		}

		updateOfficeImage := tx.Model(&models.OfficeImage{}).
			Where("employer_id = ? AND image_id = ?", employer["id"], imageId).
			Update("updated_at", time.Now())

		if updateOfficeImage.RowsAffected == 0 {
			return fmt.Errorf("office image with employer id %v and image id %v might not available in database", employer["id"], imageId)
		}

		return nil
	})
	if errUpdateOfficeImage != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errUpdateOfficeImage.Error(),
			"message": fmt.Sprintf("unable updating office image with image id %v", imageId),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("office image with image id %v updated successfully", imageId),
		},
	})
}

func (e *EmployerHandlers) DeleteOfficeImage(ctx *gin.Context) {
	/*
	   check permissions here
	*/
	imageId, errParse := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "url parameter of image id should be a valid number",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	errDeleteOfficeImage := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).Select("id").Where("user_id = ?", claims.Id).First(&employer).Error
		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		deleteOfficeImage := tx.Delete(&models.OfficeImage{}, "employer_id = ? AND image_id = ?", employer["id"], imageId)
		if deleteOfficeImage.RowsAffected == 0 {
			return fmt.Errorf("office image with employer id %v and image id %v might not available in database", employer["id"], imageId)
		}

		deleteImage := tx.Delete(&models.Image{}, "id = ?", imageId)
		if deleteImage.RowsAffected == 0 {
			return fmt.Errorf("image with id %v might not available in database", imageId)
		}

		return nil
	})

	if errDeleteOfficeImage != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errDeleteOfficeImage.Error(),
			"message": "failed deleting office image",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("office image with image id %v deleted successfully", imageId),
		},
	})
}

/* SOCIAL */
func (e *EmployerHandlers) StoreEmployerSocials(ctx *gin.Context) {
	employerSocials := []CreateEmployerSocialJSON{}
	if errBind := ctx.ShouldBindJSON(&employerSocials); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	m_employerSocials := []models.EmployerSocial{}
	for _, esocial := range employerSocials {
		m_employerSocials = append(m_employerSocials, models.EmployerSocial{
			SocialId:  esocial.SocialId,
			Url:       esocial.Url,
			CreatedAt: time.Now(),
		})
	}

	gormDB, _ := initializer.GetGorm()
	errStoreEmployerSocials := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).Select("id").Where("user_id = ?", claims.Id).First(&employer).Error
		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		for index := 0; index < len(m_employerSocials); index++ {
			m_employerSocials[index].EmployerId = employer["id"].(string)
		}

		storeEmployerSocials := tx.Create(&m_employerSocials)
		if storeEmployerSocials.Error != nil {
			return storeEmployerSocials.Error
		}

		return nil
	})

	if errStoreEmployerSocials != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreEmployerSocials.Error(),
			"message": "failed creating employer socials",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("%v employer socials created successfully", len(m_employerSocials)),
		},
	})
}
func (e *EmployerHandlers) UpdateEmployerSocial(ctx *gin.Context) {
	socialId, errParse := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "social id as url parameter invalid, value must be a valid number",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	employerSocial := struct {
		SocialId *uint   `json:"social_id"`
		Url      *string `json:"url"`
	}{}
	ctx.ShouldBindJSON(&employerSocial)

	gormDB, _ := initializer.GetGorm()
	errUpdateEmployerSocial := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select("id").
			Where("user_id = ?", claims.Id).
			First(&employer).Error
		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		employerSocialMap := map[string]interface{}{}

		v := reflect.ValueOf(employerSocial)
		for i := 0; i < v.NumField(); i++ {
			fieldName := v.Type().Field(i).Tag.Get("json")
			value := v.Field(i)

			if value.Kind() == reflect.Pointer && !value.IsNil() {
				employerSocialMap[fieldName] = value.Interface()
			}
		}

		updateEmployerSocial := tx.Model(&models.EmployerSocial{}).
			Where("employer_id = ? AND social_id = ?", employer["id"], socialId).
			Updates(employerSocialMap)

		if updateEmployerSocial.RowsAffected == 0 {
			return fmt.Errorf("employer social with social id %v might not available in database", socialId)
		}

		return nil
	})

	if errUpdateEmployerSocial != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errUpdateEmployerSocial.Error(),
			"message": fmt.Sprintf("failed updating employer social with social id %v", socialId),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "employer social updated successfully",
		},
	})
}
func (e *EmployerHandlers) DeleteEmployerSocial(ctx *gin.Context) {
	socialId, errParse := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "social id as url parameter invalid, value must be a valid number",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	gormDB, _ := initializer.GetGorm()
	errDeleteEmployerSocial := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).
			Select("id").
			Where("user_id = ?", claims.Id).
			First(&employer).Error

		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		deleteEmployerSocial := tx.
			Where("employer_id = ? AND social_id = ?", employer["id"], socialId).
			Delete(&models.EmployerSocial{})

		if deleteEmployerSocial.RowsAffected == 0 {
			return fmt.Errorf("employer social with social id %v might not available in database", socialId)
		}

		return nil
	})

	if errDeleteEmployerSocial != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteEmployerSocial.Error(),
			"message": fmt.Sprintf("failed deleting employer social with social id %v", socialId),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("employer social with social id %v deleted successfully", socialId),
		},
	})
}

/* VACANCIES */
func (e *EmployerHandlers) StoreVacancy(ctx *gin.Context) {
	/*
		check permissions here
	*/
	vacancyForm := CreateVacancyForm{}
	if errBind := ctx.ShouldBind(&vacancyForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your Form-Data fields, kids",
		})

		ctx.Abort()
		return
	}

	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	ch_uuid := make(chan string, 1)
	go GenUuid(vacancyForm.Position, ch_uuid)

	m_vacancy := models.Vacancy{
		Position:        vacancyForm.Position,
		Description:     vacancyForm.Description,
		Qualification:   vacancyForm.Qualification,
		Responsibility:  vacancyForm.Responsibility,
		LineIndustry:    vacancyForm.LineIndustry,
		EmployeeType:    vacancyForm.EmployeeType,
		MinExperience:   vacancyForm.MinExperience,
		Salary:          vacancyForm.Salary,
		WorkArrangement: vacancyForm.WorkArrangement,
		SLA:             vacancyForm.SLA, // in a hour
		IsInactive:      *vacancyForm.IsInactive,
		CreatedAt:       time.Now(),
		UpdatedAt:       nil,
	}

	gormDB, _ := initializer.GetGorm()
	errStoreVacancy := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).Select("id").Where("user_id = ?", claims.Id).First(&employer).Error
		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		m_vacancy.EmployerId = employer["id"].(string)
		m_vacancy.Id = <-ch_uuid
		storeVacancy := tx.Create(&m_vacancy)
		if storeVacancy.Error != nil {
			return storeVacancy.Error
		}

		return nil
	})

	if errStoreVacancy != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreVacancy.Error(),
			"message": "failed creating vacancy",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": "vacancy created successfully",
		},
	})
}

func (e *EmployerHandlers) UpdateVacancy(ctx *gin.Context) {
	/*
	 perform check for SLA expiration
	 no one can change the sla count whether increement or decreement
	*/
	vacancyId := ctx.Param("id")
	if vacancyId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "vacancy id is required",
			"message": "missing vacancy id as url parameter",
		})

		ctx.Abort()
		return
	}

	vacancyForm := UpdateVacancyForm{}
	ctx.ShouldBind(&vacancyForm)

	vacancyFormMap := map[string]interface{}{}
	v := reflect.ValueOf(vacancyForm)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Tag.Get("form")
		value := v.Field(i)

		if value.Kind() == reflect.Pointer && !value.IsNil() {
			vacancyFormMap[fieldName] = value.Interface()
		}
	}

	gormDB, _ := initializer.GetGorm()
	errUpdateVacancy := gormDB.Transaction(func(tx *gorm.DB) error {
		// chek expiration here
		/*
			- set inactive or updating sla
			- the RowsAffected always counted as 1, indicate the operation success whether updating sla or updating is_inactive
			- if the operation go for updating is_inactive, create or assign variable additional attribute with string value indicate that the vacancy is inactive
		*/
		slaError := SLAGuard(vacancyId, tx)
		if slaError != nil {
			return slaError
		}

		/*
			check permissions for allowing employer wants to modify/updating the SLA count
		*/
		updateVacancy := tx.Model(&models.Vacancy{}).
			Where("id = ?", vacancyId).
			Updates(vacancyFormMap)

		if updateVacancy.RowsAffected == 0 {
			return fmt.Errorf("unable updating vacancy, vacancy with id %v might not exists in database", vacancyId)
		}

		return nil
	})

	if errUpdateVacancy != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errUpdateVacancy.Error(),
			"message": fmt.Sprintf("failed updating vacancy with id %v", vacancyId),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("vacancy with id %v updated successfully", vacancyId),
		},
	})
}

func (e *EmployerHandlers) GetVacancy(ctx *gin.Context) {
	/*
		if just getting the details of spesific vacancy, just utilize the /api/v1/vacancies/:id
	*/
}

func (e *EmployerHandlers) DeleteVacancy(ctx *gin.Context) {
	vacancyId := ctx.Param("id")
	if vacancyId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing value for url parameter vacancy id",
			"message": "provide a valid value for vacancy id as url parameter",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	deleteVacancy := gormDB.Where("id = ?", vacancyId).Delete(&models.Vacancy{})
	if deleteVacancy.RowsAffected == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("unable delete vacancy with id %v, this might the record doesn't exist in database", vacancyId),
			"message": "no vacancy data deleted",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("vacancy with id %v deleted successfully", vacancyId),
		},
	})
}

func (e *EmployerHandlers) ListVacancy(ctx *gin.Context) {
	bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
	claims := ParseJWT(bearerToken)

	listVacancies := []map[string]interface{}{}
	gormDB, _ := initializer.GetGorm()
	errListVacanciesByEmployerId := gormDB.Transaction(func(tx *gorm.DB) error {
		employer := map[string]interface{}{}
		errGetEmployerId := tx.Model(&models.Employer{}).Select("id").Where("user_id = ?", claims.Id).First(&employer).Error
		if errGetEmployerId != nil {
			return errGetEmployerId
		}

		getVacancies := tx.Model(&models.Vacancy{}).Select([]string{
			"id",
			"position",
			"description",
			"qualification",
			"responsibility",
			"line_industry",
			"employee_type",
			"min_experience",
			"salary",
			"work_arrangement",
			"sla",
			"is_inactive",
		}).Where("employer_id = ?", employer["id"]).Find(&listVacancies)
		if getVacancies.RowsAffected == 0 {
			return fmt.Errorf("no vacancy data found for employer id: %v", employer["id"])
		}

		return nil
	})

	if errListVacanciesByEmployerId != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   errListVacanciesByEmployerId.Error(),
			"message": "0 vacancies data found",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listVacancies,
	})
}

/* ASSESSMENT */
func (e *EmployerHandlers) StoreAssessment(ctx *gin.Context) {
	assessmentForm := CreateAssessmentForm{}
	if errBind := ctx.ShouldBind(&assessmentForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your Form-Data fields, kids",
		})

		ctx.Abort()
		return
	}
	m_assessment := models.Assessment{
		Name:           assessmentForm.Name,
		Note:           assessmentForm.Note,
		AssessmentLink: assessmentForm.AssessmentLink,
		StartAt:        assessmentForm.StartDate,
		DueDate:        assessmentForm.DueDate,
		VacancyId:      assessmentForm.VacancyId,
		CreatedAt:      time.Now(),
		UpdatedAt:      nil,
	}

	form, _ := ctx.MultipartForm()
	assessmentDocuments := form.File["assessment_documents"]
	/* extract document data */
	m_documents, document_status := MultipleDocumentData(assessmentDocuments, "assessment_document")

	gormDB, _ := initializer.GetGorm()
	errStoreAssessment := gormDB.Transaction(func(tx *gorm.DB) error {
		errCreateAssessment := tx.Create(&m_assessment).Error
		if errCreateAssessment != nil {
			return errCreateAssessment
		}

		if len(m_documents) > 0 {
			errStoreDocuments := tx.Create(&m_documents).Error
			if errStoreDocuments != nil {
				return errStoreDocuments
			}

			m_assessmentDocuments := []models.AssessmentDocument{}
			for _, document := range m_documents {
				m_assessmentDocuments = append(m_assessmentDocuments, models.AssessmentDocument{
					AssessmentId: m_assessment.ID,
					DocumentId:   document.ID,
					CreatedAt:    time.Now(),
					UpdatedAt:    nil,
				})
			}

			errStoreAssessmentDocuments := tx.Create(&m_assessmentDocuments).Error
			if errStoreAssessmentDocuments != nil {
				return errStoreAssessmentDocuments
			}
		}
		return nil
	})

	if errStoreAssessment != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreAssessment.Error(),
			"message": "failed creating assessment",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message":          fmt.Sprintf("assessment created successfully with %v documents attached", len(m_documents)),
			"documents_status": document_status,
		},
	})
}

func (e *EmployerHandlers) UpdateAssessment(ctx *gin.Context) {
	assessmentId, errParse := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "assessment id should be a valid number",
		})

		ctx.Abort()
		return
	}

	/*
		updating assessment document should be handled by /api/v1/documents/:id to updating spesific documents by ID
	*/

	assessmentForm := UpdateAssessmentForm{}
	ctx.ShouldBind(&assessmentForm)

	assessmentFormMap := map[string]interface{}{}
	v := reflect.ValueOf(assessmentForm)
	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Tag.Get("form")
		value := v.Field(i)

		if value.Kind() == reflect.Pointer && !value.IsNil() {

			if value.Elem().Kind() == reflect.String && value.Elem().String() == "" {
				continue
			}

			assessmentFormMap[fieldName] = value.Interface()
		}
	}

	gormDB, _ := initializer.GetGorm()
	if len(assessmentFormMap) != 0 {
		updateAssessment := gormDB.Model(&models.Assessment{}).Where("id = ?", assessmentId).Updates(&assessmentFormMap)
		if updateAssessment.RowsAffected == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   fmt.Sprintf("no rows updated for assessment with id %d", assessmentId),
				"message": fmt.Sprintf("unable updating assessment with id %d, this might the record doesn't exist in database", assessmentId),
			})

			ctx.Abort()
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("successfully updated assessment with id %d", assessmentId),
		},
	})
}

func (e *EmployerHandlers) GetAssessment(ctx *gin.Context) {
	/*
		no need this handler at the moment
	*/
}

func (e *EmployerHandlers) DeleteAssessment(ctx *gin.Context) {
	/*
		deleting assessment inlcuding related table
		1. delete assessment_assignees within assessment_assignee_submission and documents
		2. delete assessment_documents within documents
		3. delete assessment itself
	*/
	assessmentId, errParse := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "assessment id must be a valid number as url parameter",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errDeleteAssessment := gormDB.Transaction(func(tx *gorm.DB) error {
		// existence check for assessment
		var exists bool
		tx.Model(&models.Assessment{}).Select("1").Where("id = ?", assessmentId).Limit(1).Scan(&exists)
		if !exists {
			return fmt.Errorf("no assessment record found with id %v", assessmentId)
		}

		listDeletedDocuments := []uint{}

		assessmentDocumentsId := []map[string]interface{}{}
		// getting assessment_documents if there any assessments with id exists.
		// it mean that if there are no assessment with id, the assessment_documents will not return any values.
		getAssessmentDocumentsId := tx.Model(&models.AssessmentDocument{}).
			Select("document_id").
			Where("assessment_id = ?", assessmentId).
			Find(&assessmentDocumentsId)

		// check if above query return values
		if getAssessmentDocumentsId.RowsAffected > 0 {
			// append document_id into listed document that will be deleted soon
			for _, document := range assessmentDocumentsId {
				listDeletedDocuments = append(listDeletedDocuments, document["document_id"].(uint))
			}
			// delete assessment_documents when record found
			tx.Where("assessment_id = ?", assessmentId).Delete(&models.AssessmentDocument{})
		} else {
			// when no values returned, it means there are no assessment_documents record match that id
			log.Println("no assessment documents records were found")
		}

		// after deleting the assessment_documents record.
		// next just deleting the assessment_assignees record
		deleteAssessmentAssignees := tx.Where("assessment_id = ?", assessmentId).Delete(&models.AssessmentAssignee{})
		// check if there are no rows deleted, it means there are no assignee attached to that assessment
		if deleteAssessmentAssignees.RowsAffected == 0 {
			// if there are no asignees, it means there are no submissions too
			// then, delete the assessment by its id
			tx.Where("id = ?", assessmentId).Delete(&models.Assessment{})
			// then, delete the documents that exists in list deleted documents
			tx.Where("id IN ?", listDeletedDocuments).Delete(&models.Document{})

			log.Println("no assignees rows affected. it might because the record doesn't exist in database")
			return nil
		}

		// the assignees record are successfully deleted
		// then, get the list of submission document id
		submissionDocumentsId := []map[string]interface{}{}
		getSubmissionDocumentsId := tx.Model(&models.AssessmentAssigneeSubmission{}).
			Select("submission_document_id").
			Where("assessment_id = ?", assessmentId).
			Find(&submissionDocumentsId)

		// check if the above query does not return values
		if getSubmissionDocumentsId.RowsAffected == 0 {
			// if there are no submissions, it means those assignees are not submitted documents yet
			// then, delete the assessment by its id
			tx.Where("id = ?", assessmentId).Delete(&models.Assessment{})
			// then, delete the documents that exists in list deleted documents
			tx.Where("id IN ?", listDeletedDocuments).Delete(&models.Document{})

			log.Println("no submissions rows affected. it might because the record doesn't exist in database")
			return nil
		}

		// the submissions record are exists
		// then, append to the list of submission document id
		for _, document := range submissionDocumentsId {
			listDeletedDocuments = append(listDeletedDocuments, document["submission_document_id"].(uint))
		}
		log.Println("list deleted documents \t", listDeletedDocuments)

		// then, delete assessment_assignee_documents
		tx.Where("assessment_id = ?", assessmentId).Delete(&models.AssessmentAssigneeSubmission{})
		// then, delete the assessment by its id
		tx.Where("id = ?", assessmentId).Delete(&models.Assessment{})
		// then, delete the listed documents id. it combining all assessment_documents and assessment_assignee_documents
		tx.Where("id IN ?", listDeletedDocuments).Delete(&models.Document{})
		return nil
	})

	if errDeleteAssessment != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteAssessment.Error(),
			"message": "failed deleting assessment",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "assessment successfully deleted",
		},
	})
}
func (e *EmployerHandlers) ListAssessment(ctx *gin.Context) {
	// url parameter
	vacancyId := ctx.Param("id")

	gormDB, _ := initializer.GetGorm()

	// result query as single JSON string
	var assessmentsJSON string
	assessmentQueryBuilder := gormDB.Raw(`
		SELECT
			assessments.id,
			assessments.name,
			(
				SELECT
					assessment_documents.document_id,
					(
						SELECT
							documents.name
						FROM
							documents
						WHERE
							documents.id = assessment_documents.document_id
						FOR JSON PATH
					) as name
				FROM
					assessment_documents
				WHERE
					assessment_documents.assessment_id = assessments.id
				FOR JSON PATH
			) as assessment_documents,
			(
				SELECT
					assessment_assignees.pipeline_id,
					assessment_assignees.submission_status,
					assessment_assignees.submission_result,
					(
						SELECT
							pipelines.candidate_id,
							(
								SELECT
									candidates.id,
									candidates.expertise,
									candidates.profile_image_id,
									(
										SELECT
											users.fullname,
											users.email
										FROM
											users
										WHERE
											users.id = candidates.user_id
										FOR JSON PATH
									) AS [user]
								FROM
									candidates
								WHERE
									candidates.id = pipelines.candidate_id
								FOR JSON PATH
							) as candidate
						FROM
							pipelines
						WHERE
							pipelines.id = assessment_assignees.pipeline_id
						FOR JSON PATH
					) as pipeline,
					(
						SELECT
							assessment_assignee_submissions.submission_document_id,
							(
								SELECT
									documents.name
								FROM
									documents
								WHERE
									documents.id = assessment_assignee_submissions.submission_document_id
								FOR JSON PATH
							) as documents
						FROM
							assessment_assignee_submissions
						WHERE
							assessment_assignee_submissions.assessment_id = assessment_assignees.assessment_id
							AND
							assessment_assignee_submissions.pipeline_id = assessment_assignees.pipeline_id
						FOR JSON PATH
					) as submissions
				FROM
					assessment_assignees
				WHERE
					assessment_assignees.assessment_id = assessments.id
				FOR JSON PATH
			) as assessment_assignees
		FROM
			assessments
		WHERE
			vacancy_id = ?
		FOR JSON PATH;
	`, vacancyId).Scan(&assessmentsJSON)

	if assessmentQueryBuilder.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   assessmentQueryBuilder.Error.Error(),
			"message": "there was an issue with assessment query builder",
		})

		ctx.Abort()
		return
	}

	tempAssessments := []map[string]interface{}{}
	errDecode := json.Unmarshal([]byte(assessmentsJSON), &tempAssessments)
	if errDecode != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": errDecode.Error(),
		})
	}

	var listAssessment []map[string]interface{}

	for _, assessment := range tempAssessments {
		assesmentMap := map[string]interface{}{
			"id":   assessment["id"],
			"name": assessment["name"],
		}

		documents := []map[string]interface{}{}
		for _, document := range assessment["assessment_documents"].([]interface{}) {
			documents = append(documents, map[string]interface{}{
				"id":   document.(map[string]interface{})["document_id"],
				"name": document.(map[string]interface{})["name"].([]interface{})[0].(map[string]interface{})["name"],
			})
		}

		assessmentAssignees := []map[string]interface{}{}
		if assignees, ok := assessment["assessment_assignees"].([]interface{}); ok {
			for _, assignee := range assignees {
				if assignee == nil {
					continue
				}

				candidate := assignee.(map[string]interface{})["pipeline"].([]interface{})[0].(map[string]interface{})["candidate"].([]interface{})[0].(map[string]interface{})
				user := assignee.(map[string]interface{})["pipeline"].([]interface{})[0].(map[string]interface{})["candidate"].([]interface{})[0].(map[string]interface{})["user"].([]interface{})[0].(map[string]interface{})

				assessmentSubmissions := []map[string]interface{}{}
				if submissions, ok := assignee.(map[string]interface{})["submissions"].([]interface{}); ok {
					for _, submission := range submissions {
						assessmentSubmissions = append(assessmentSubmissions, map[string]interface{}{
							"id":   submission.(map[string]interface{})["submission_document_id"],
							"name": submission.(map[string]interface{})["documents"].([]interface{})[0].(map[string]interface{})["name"],
						})
					}
				}

				assessmentAssignee := map[string]interface{}{
					"pipeline_id":       assignee.(map[string]interface{})["pipeline_id"],
					"submission_status": assignee.(map[string]interface{})["submission_status"],
					"submission_result": assignee.(map[string]interface{})["submission_result"],
					"candidate": map[string]interface{}{
						"fullname":         user["fullname"],
						"email":            user["email"],
						"expertise":        candidate["expertise"],
						"profile_image_id": candidate["profile_image_id"],
					},
					"submission_documents": assessmentSubmissions,
				}
				assessmentAssignees = append(assessmentAssignees, assessmentAssignee)
			}
		}

		assesmentMap["assessment_documents"] = documents
		assesmentMap["assessment_assignees"] = assessmentAssignees

		listAssessment = append(listAssessment, assesmentMap)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listAssessment,
	})
}

/* ASSESSMENT ASSIGNEE */
func (e *EmployerHandlers) StoreAssignees(ctx *gin.Context) {
	assignees := []struct {
		AssessmentId uint   `json:"assessment_id" binding:"required"`
		PipelineId   string `json:"pipeline_id" binding:"required"`
	}{}
	if errBind := ctx.ShouldBindJSON(&assignees); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})

		ctx.Abort()
		return
	}

	m_assessment_assignees := []models.AssessmentAssignee{}
	pipelineList := []string{}

	for _, assignee := range assignees {
		m_assessment_assignees = append(m_assessment_assignees, models.AssessmentAssignee{
			AssessmentId:     assignee.AssessmentId,
			PipelineId:       assignee.PipelineId,
			SubmissionStatus: "assigned",
			SubmissionResult: nil,
			CreatedAt:        time.Now(),
			UpdatedAt:        nil,
		})

		pipelineList = append(pipelineList, assignee.PipelineId)
	}

	gormDB, _ := initializer.GetGorm()
	errStoreAssignees := gormDB.Transaction(func(tx *gorm.DB) error {
		errCreateAssignees := gormDB.Create(&m_assessment_assignees).Error
		if errCreateAssignees != nil {
			return errCreateAssignees
		}

		updatePipelines := gormDB.Model(&models.Pipeline{}).Where("id IN (?)", pipelineList).Updates(map[string]interface{}{
			"stage":  "Assessment",
			"status": "On Process",
		})
		if updatePipelines.RowsAffected == 0 {
			return fmt.Errorf("%v rows affected. failed updating pipeline (stage and status)", updatePipelines.RowsAffected)
		}

		return nil
	})

	if errStoreAssignees != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreAssignees.Error(),
			"message": fmt.Sprintf("unable storing %v assignees", len(m_assessment_assignees)),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("%v assignees stored successfully", len(m_assessment_assignees)),
		},
	})
}

func (e *EmployerHandlers) UpdateAssignee(ctx *gin.Context) {
	assignee := struct {
		AssessmentId     uint    `json:"assessment_id" binding:"required"`
		PipelineId       string  `json:"pipeline_id" binding:"required"`
		SubmissionStatus *string `json:"submission_status"`
		SubmissionResult *string `json:"submission_result"`
	}{}

	if errBind := ctx.ShouldBindJSON(&assignee); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})

		ctx.Abort()
		return
	}

	assigneeMap := map[string]interface{}{}
	v := reflect.ValueOf(assignee)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Tag.Get("json")
		value := v.Field(i)

		if value.Kind() == reflect.Pointer && !value.IsNil() {
			assigneeMap[fieldName] = value.Interface()
		}
	}

	gormDB, _ := initializer.GetGorm()
	updateAssignee := gormDB.Model(&models.AssessmentAssignee{}).
		Where("assessment_id = ? AND pipeline_id = ?", assignee.AssessmentId, assignee.PipelineId).
		Updates(assigneeMap)

	if updateAssignee.RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   fmt.Sprintf("assignee with assessment_id %v and pipeline_id %v might not exist in database", assignee.AssessmentId, assignee.PipelineId),
			"message": "failed updating assessment assignee",
		})

		ctx.Abort()
		return
	}

	ctx.JSONP(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "assignee updated successfully",
		},
	})
}

func (e *EmployerHandlers) DeleteAssignee(ctx *gin.Context) {
	assessmentId := ctx.Param("assessmentId")
	pipelineId := ctx.Param("pipelineId")

	if assessmentId == "" || pipelineId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "missing assessmentId or pipelineId as url parameter",
			"message": "double check your URL Parameter, kids",
		})

		ctx.Abort()
		return
	}
	/*
		this deletion should delete the assessment_assignees, assessment_assignee_submissions, documents
	*/
	gormDB, _ := initializer.GetGorm()
	errDeleteAssignee := gormDB.Transaction(func(tx *gorm.DB) error {
		deleteAssignees := tx.Where("assessment_id = ? AND pipeline_id = ?", assessmentId, pipelineId).
			Delete(&models.AssessmentAssignee{})
		if deleteAssignees.RowsAffected == 0 {
			return fmt.Errorf("no rows deleted for assignee with assessment id %v and pipeline id %v", assessmentId, pipelineId)
		}

		submissionDocumentId := []map[string]interface{}{}
		getDocumentsId := tx.Model(&models.AssessmentAssigneeSubmission{}).
			Select("submission_document_id").
			Where("assessment_id = ? AND pipeline_id = ?", assessmentId, pipelineId).
			Find(&submissionDocumentId)
		if getDocumentsId.RowsAffected == 0 {
			// Log Test
			log.Println("candidates didn't submit any documents yet, just being assigned to the assessment")
			return nil
		}

		listDeletedDocumentId := []uint{}
		for _, document := range submissionDocumentId {
			listDeletedDocumentId = append(listDeletedDocumentId, document["submission_document_id"].(uint))
		}

		deleteSubmissions := tx.Where("assessment_id = ? AND pipeline_id = ?", assessmentId, pipelineId).
			Delete(&models.AssessmentAssigneeSubmission{})
		if deleteSubmissions.RowsAffected == 0 {
			return fmt.Errorf("no rows deleted for assignee submissions with assessment id %v and pipeline id %v", assessmentId, pipelineId)
		}
		deleteDocument := tx.Where("id IN ?", listDeletedDocumentId).
			Delete(&models.Document{})
		if deleteDocument.RowsAffected == 0 {
			return fmt.Errorf("no rows deleted for documents with id IN %v", listDeletedDocumentId)
		}
		return nil
	})

	if errDeleteAssignee != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteAssignee.Error(),
			"message": "failed deleting assignee",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "assignee deleted successfully",
		},
	})
}

/* INTERVIEW */
func (e *EmployerHandlers) StoreInterview(ctx *gin.Context) {
	interviewForm := CreateInterviewForm{}
	if errBind := ctx.ShouldBind(&interviewForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your Form-Data fields, kids",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errStoreInterview := gormDB.Transaction(func(tx *gorm.DB) error {
		/*
			check if the candidate doesn't have any interview schedules, then should update the pipeline stage with "Interviews"
		*/
		var hasNoInterview bool
		checkInterviews := tx.Raw(`
			SELECT
				CASE
					WHEN NOT EXISTS (
						SELECT 1
						FROM interviews
						WHERE pipeline_id = ?
						GROUP BY pipeline_id
						HAVING count(*) > 0
					)
						THEN 1
					ELSE 0
				END AS hasNoInterview
		`, interviewForm.PipelineId).Scan(&hasNoInterview)

		if checkInterviews.Error != nil {
			return checkInterviews.Error
		}

		if hasNoInterview {
			log.Println("candidate has no interviews")
			updateStage := tx.Model(&models.Pipeline{}).Where("id = ?", interviewForm.PipelineId).Update("stage", "Interview")
			if updateStage.RowsAffected == 0 {
				log.Println("no pipeline updated. failed")
			}
		}

		m_interview := models.Interview{
			Date:        interviewForm.Date,
			Location:    interviewForm.Location,
			LocationURL: interviewForm.LocationURL,
			Status:      "Scheduled",
			Result:      nil,
			PipelineId:  interviewForm.PipelineId,
			VacancyId:   interviewForm.VacancyId,
			CreatedAt:   time.Now(),
			UpdatedAt:   nil,
		}
		errCreateInterview := tx.Create(&m_interview).Error
		if errCreateInterview != nil {
			return errCreateInterview
		}

		return nil
	})

	if errStoreInterview != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreInterview.Error(),
			"message": fmt.Sprintf("failed storing Interview for pipeline_id %v", interviewForm.PipelineId),
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": fmt.Sprintf("interview schedule created successfully for pipeline_id %v", interviewForm.PipelineId),
		},
	})
}

func (e *EmployerHandlers) UpdateInterview(ctx *gin.Context) {
	interviewId := ctx.Param("id")
	if _, errParse := strconv.ParseUint(interviewId, 10, 32); errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "interview ID must be a valid number as url parameter",
		})

		ctx.Abort()
		return
	}

	interviewForm := UpdateInterviewForm{}
	ctx.ShouldBind(&interviewForm)

	log.Println("interview form \t:", interviewForm)
	interviewFormMap := map[string]interface{}{}
	v := reflect.ValueOf(interviewForm)

	for i := 0; i < v.NumField(); i++ {
		fieldName := v.Type().Field(i).Tag.Get("form")
		value := v.Field(i)

		if value.Elem().String() == "" {
			continue
		}

		if value.Kind() == reflect.Pointer && !value.IsNil() {
			interviewFormMap[fieldName] = value.Interface()
		}
	}
	log.Println("interview form map \t:", interviewFormMap)

	gormDB, _ := initializer.GetGorm()
	updateInterview := gormDB.Model(&models.Interview{}).Where("id = ?", interviewId).Updates(interviewFormMap)
	if updateInterview.RowsAffected == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("%v rows affected. it might interview with id %v is not available in database", updateInterview.RowsAffected, interviewId),
			"message": "failed updating interview",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "interviews updated successfully",
		},
	})
}

func (e *EmployerHandlers) GetInterview(ctx *gin.Context) {
	/*
		doesn't need this handler at the moment
	*/
}

func (e *EmployerHandlers) DeleteInterview(ctx *gin.Context) {
	interviewId := ctx.Param("id")

	if _, errParse := strconv.ParseUint(interviewId, 10, 32); errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "interview id must be a valid number as url parameter",
			"message": "failed deleting interview",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	deleteInterview := gormDB.Where("id = ?", interviewId).Delete(&models.Interview{})
	if deleteInterview.RowsAffected == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("interview with id %v not found", interviewId),
			"message": "failed deleting interview",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "interview deleted successfully",
		},
	})
}

func (e *EmployerHandlers) ListInterviewHistory(ctx *gin.Context) {
	pipelineId := ctx.Param("pipelineId")
	vacacncyId := ctx.Param("vacancyId")

	listInterview := []struct {
		Id          uint      `json:"id"`
		Date        time.Time `json:"date"`
		Location    string    `json:"location"`
		LocationURL string    `json:"location_url"`
		Status      string    `json:"status"`
		Result      *string   `json:"result"`
	}{}

	gormDB, _ := initializer.GetGorm()
	findInterviews := gormDB.Model(&models.Interview{}).
		Select([]string{
			"id",
			"date",
			"location",
			"location_url",
			"status",
			"result",
		}).
		Where("pipeline_id = ? AND vacancy_id = ?", pipelineId, vacacncyId).
		Find(&listInterview)

	if findInterviews.RowsAffected == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   fmt.Sprintf("interview history for pipeline id %v not found", pipelineId),
			"message": "has no interview history",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listInterview,
	})
}

/* OFFERING */
func (e *EmployerHandlers) StoreOffering(ctx *gin.Context) {
	offeringForm := struct {
		VacancyId  string    `form:"vacancy_id" binding:"required"`
		PipelineId string    `form:"pipeline_id" binding:"required"`
		EndOn      time.Time `form:"end_on" binding:"required"`
	}{}

	if errBind := ctx.ShouldBind(&offeringForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your Form-Data fields, kids",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errStoreOffering := gormDB.Transaction(func(tx *gorm.DB) error {
		m_offering := models.Offering{
			EndOn:         offeringForm.EndOn,
			Status:        "Offered",
			PipelineId:    offeringForm.PipelineId,
			VacancyId:     offeringForm.VacancyId,
			LoaDocumentId: nil,
			CreatedAt:     time.Now(),
			UpdatedAt:     nil,
		}
		errCreateOffering := tx.Create(&m_offering).Error
		if errCreateOffering != nil {
			return errCreateOffering
		}

		updatePipeline := tx.Model(&models.Pipeline{}).Where("id = ?", offeringForm.PipelineId).Update("stage", "Offering")
		if updatePipeline.RowsAffected == 0 {
			return fmt.Errorf("%v rows affected, it might record doesn't exists in database", updatePipeline.RowsAffected)
		}

		return nil
	})
	if errStoreOffering != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreOffering.Error(),
			"message": "failed creating offering",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": "offering created successfully",
		},
	})
}

func (e *EmployerHandlers) UpdateOffering(ctx *gin.Context) {
	offeringId := ctx.Param("id")
	if _, errParse := strconv.ParseUint(offeringId, 10, 32); errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "offering id must be a valid number as url parameter",
		})

		ctx.Abort()
		return
	}

	offeringFormMap := map[string]interface{}{}
	offeringForm := struct {
		EndOn *time.Time `form:"end_on"`
	}{}
	ctx.ShouldBind(&offeringForm)
	if offeringForm.EndOn != nil {
		offeringFormMap["end_on"] = offeringForm.EndOn
	}

	var document_status string
	loaDocument, errLoaDocument := ctx.FormFile("loa_document")

	gormDB, _ := initializer.GetGorm()
	errUpdateOffering := gormDB.Transaction(func(tx *gorm.DB) error {
		timeNow := time.Now()

		offering := map[string]interface{}{}
		errGetLoaDocumentId := tx.Model(&models.Offering{}).Select("loa_document_id").Where("id = ?", offeringId).First(&offering).Error

		if errGetLoaDocumentId != nil {
			return errGetLoaDocumentId
		}

		if errLoaDocument == nil {
			m_document, errExtract := DocumentData(loaDocument, "loa_document")
			if errExtract != nil {
				document_status = errExtract.Error()
				goto UpdateAttributesOnly
			}

			// when loa_document is nil, then create new document and update loa_document_id
			if offering["loa_document_id"] == nil {
				m_document.CreatedAt = time.Now()
				m_document.UpdatedAt = nil
				errStoreLoaDocument := tx.Create(&m_document).Error
				if errStoreLoaDocument != nil {
					document_status = errStoreLoaDocument.Error()
					goto UpdateAttributesOnly
				}
				updateOfferingLoaDocumentId := tx.Model(&models.Offering{}).
					Where("id = ?", offeringId).
					Updates(map[string]interface{}{
						"loa_document_id": m_document.ID,
						"updated_at":      &timeNow,
					})
				if updateOfferingLoaDocumentId.RowsAffected == 0 {
					document_status = "failed updating loa_document_id while storing new document"
				}

				document_status = "document added successfully"
			} else { // when loa_document_id exists, then just update document within that id
				m_document.UpdatedAt = &timeNow
				updateDocument := tx.Model(&models.Document{}).Where("id = ?", offering["loa_document_id"].(uint)).Updates(m_document)
				if updateDocument.RowsAffected == 0 {
					document_status = "no rows of document updated"
				}

				document_status = "document updated successfully"
			}
		} else {
			document_status = "no document updated or added"
		}

	UpdateAttributesOnly: // label

		offeringFormMap["updated_at"] = &timeNow
		updateOffering := tx.Model(&models.Offering{}).Where("id = ?", offeringId).Updates(offeringFormMap)

		if updateOffering.RowsAffected == 0 {
			return fmt.Errorf("offeringn with id %v might not exists in database", offeringId)
		}

		return nil
	})

	if errUpdateOffering != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errUpdateOffering.Error(),
			"message": "failed updating offering",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message":         "offering updated successfully",
			"document_status": document_status,
		},
	})
}

func (e *EmployerHandlers) GetOffering(ctx *gin.Context) {
	/*
		no need this handler at the moment
	*/
}

func (e *EmployerHandlers) DeleteOffering(ctx *gin.Context) {
	// delete offerings
	// delete document
	// update pipeline stage to previous stage
	offeringId := ctx.Param("id")
	if _, errParse := strconv.ParseUint(offeringId, 10, 32); errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "offering id must be a valid number as url parameter",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errDeleteOffering := gormDB.Transaction(func(tx *gorm.DB) error {
		offering := map[string]interface{}{}
		errGetDocumentId := tx.Model(&models.Offering{}).
			Select([]string{
				"pipeline_id",
				"loa_document_id",
			}).
			Where("id = ?", offeringId).
			First(&offering).Error
		if errGetDocumentId != nil {
			return errGetDocumentId
		}

		deleteOffering := tx.Where("id = ?", offeringId).Delete(&models.Offering{})
		if deleteOffering.RowsAffected == 0 {
			return fmt.Errorf("offering with id %v does not exists in database", offeringId)
		}

		deleteLoaDocument := tx.Where("id = ?", offering["loa_document_id"]).Delete(&models.Document{})
		if deleteLoaDocument.RowsAffected == 0 {
			return fmt.Errorf("document with id %v does not exists in database", offering["loa_document_id"])
		}

		updatePipelineStage := tx.Model(&models.Pipeline{}).Where("id = ?", offering["pipeline_id"]).Update("stage", "Assessment")
		if updatePipelineStage.RowsAffected == 0 {
			return fmt.Errorf("pipeline with id %v does not exists in database", offering["pipeline_id"])
		}
		return nil
	})

	if errDeleteOffering != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteOffering.Error(),
			"message": "failed deleting offering",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "offering deleted successfully",
		},
	})
}

func (e *EmployerHandlers) ListOffering(ctx *gin.Context) {
	vacancyId := ctx.Param("id")

	listOffering := []map[string]interface{}{}
	gormDB, _ := initializer.GetGorm()
	getListOffering := gormDB.Model(&models.Offering{}).
		Select([]string{
			"offerings.id",
			"offerings.end_on",
			"offerings.status",
			"offerings.loa_document_id",
			"documents.name AS loa_document_name",
			"candidates.expertise",
			"users.fullname",
		}).
		Joins("INNER JOIN pipelines ON pipelines.id = offerings.pipeline_id").
		Joins("LEFT JOIN documents ON documents.id = offerings.loa_document_id").
		Joins("INNER JOIN candidates ON candidates.id = pipelines.candidate_id").
		Joins("INNER JOIN users ON users.id = candidates.user_id").
		Where("offerings.vacancy_id = ?", vacancyId).Find(&listOffering)

	if getListOffering.RowsAffected == 0 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "no data found",
			"message": "no data",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listOffering,
	})
}
