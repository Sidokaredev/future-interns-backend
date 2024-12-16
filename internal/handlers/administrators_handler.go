package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AdministratorHandlers struct {
}

type CreateEmployerUserJSON struct {
	Fullname string `json:"fullname" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (a *AdministratorHandlers) CreateEmployerUser(ctx *gin.Context) {
	/*
			  get key "permissions"
			  check the permission name before executing a job
		    if doesn't have required permissions, just throw an fail response
	*/
	permissions, _ := ctx.Get("permissions")
	if !permissions.((map[string]bool))["users.employer.create"] {
		ctx.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Lacks the [users.employer.create] permission",
			"message": "you are not allowed to access this resource",
		})

		ctx.Abort()
		return
	}

	user_data := CreateEmployerUserJSON{}
	if errBind := ctx.ShouldBindJSON(&user_data); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kid",
		})

		ctx.Abort()
		return
	}

	ch_hash_password := make(chan string, 1)
	ch_uuid := make(chan string, 1)
	go GenUuid(user_data.Fullname, ch_uuid)
	go HashPassword(user_data.Password, ch_hash_password)

	const EmployerRoleId = 2

	gormDB, _ := initializer.GetGorm()
	m_user := models.User{
		Id:       <-ch_uuid,
		Fullname: user_data.Fullname,
		Email:    user_data.Email,
		Password: <-ch_hash_password,
	}

	errCreateEmployerUser := gormDB.Transaction(func(tx *gorm.DB) error {
		errCreateUser := tx.Create(&m_user).Error
		if errCreateUser != nil {
			return errCreateUser
		}

		errAssignIdentity := tx.Create(&models.IdentityAccess{
			UserId: m_user.Id,
			RoleId: EmployerRoleId,
			Type:   "employer",
		}).Error
		if errAssignIdentity != nil {
			return errAssignIdentity
		}

		return nil
	})

	if errCreateEmployerUser != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCreateEmployerUser.Error(),
			"message": "failed creating new employer user",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    "Employer user created successfully",
	})
}
func (a *AdministratorHandlers) UpdateEmployerUser(ctx *gin.Context) {

}

func (a *AdministratorHandlers) GetEmployerUserById(ctx *gin.Context) {

}

func (a *AdministratorHandlers) DeleteEmployerUserById(ctx *gin.Context) {

}

func (a *AdministratorHandlers) ListEmployerUsers(ctx *gin.Context) {

}

/* Skill */
func (a *AdministratorHandlers) CreateSkills(ctx *gin.Context) {
	var skillForm struct {
		Name string `form:"name" binding:"required"`
	}
	if errBind := ctx.ShouldBind(&skillForm); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "name for skill is required",
		})

		ctx.Abort()
		return
	}

	skillIcon, errSkillIcon := ctx.FormFile("skill_icon_image")
	if errSkillIcon != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errSkillIcon.Error(),
			"message": "skill_icon_image is required",
		})

		ctx.Abort()
		return
	}

	m_image, errImage := ImageData(skillIcon)
	if errImage != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errImage.Error(),
			"message": "there was an error with your file image",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errStoreSkill := gormDB.Transaction(func(tx *gorm.DB) error {
		errCreateImage := tx.Create(&m_image).Error
		if errCreateImage != nil {
			return errCreateImage
		}

		m_skill := models.Skill{
			Name:             skillForm.Name,
			SkillIconImageId: m_image.ID,
			CreatedAt:        time.Now(),
			UpdatedAt:        nil,
		}
		errCreateSkill := tx.Create(&m_skill).Error
		if errCreateSkill != nil {
			return errCreateSkill
		}

		return nil
	})

	if errStoreSkill != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errStoreSkill.Error(),
			"message": "error database operation",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    "skill stored successfully",
	})
}

func (a *AdministratorHandlers) DeleteSkills(ctx *gin.Context) {
	skillId := ctx.Param("id")
	if _, errParse := strconv.Atoi(skillId); errParse != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errParse.Error(),
			"message": "skill id must be a valid number",
		})

		ctx.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	errDeleteSkill := gormDB.Transaction(func(tx *gorm.DB) error {
		var imageID int
		errGetImageID := tx.Model(&models.Skill{}).
			Select("skill_icon_image_id").
			Where("id = ?", skillId).
			First(&imageID).Error
		if errGetImageID != nil {
			return errGetImageID
		}

		log.Println("image id \t:", imageID)

		deleteSkillRow := tx.Delete(&models.Skill{}, skillId).RowsAffected
		if deleteSkillRow == 0 {
			return fmt.Errorf("failed deleting skill with id %v", skillId)
		}

		deleteImageRow := tx.Delete(&models.Image{}, imageID).RowsAffected
		if deleteImageRow == 0 {
			return fmt.Errorf("failed deleting image with id %v", imageID)
		}

		return nil
	})

	if errDeleteSkill != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errDeleteSkill.Error(),
			"message": "error database operation",
		})

		ctx.Abort()
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fmt.Sprintf("skill with id %v deleted successfully", skillId),
	})
}
