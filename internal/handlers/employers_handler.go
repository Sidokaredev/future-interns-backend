package handlers

import (
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
