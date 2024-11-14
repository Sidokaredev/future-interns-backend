package handlers

import (
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"net/http"

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
